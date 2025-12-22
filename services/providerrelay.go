package services

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/daodao97/xgo/xdb"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	_ "modernc.org/sqlite"
)

type ProviderRelayService struct {
	providerService *ProviderService
	logService      *LogService
	server          *http.Server
	addr            string
	
	// 记录每个供应商上次被禁用/启用时的请求数
	// key: "platform:providerName", value: 上次检查时的请求数
	lastCheckRequests map[string]int64
	lastCheckMu       sync.Mutex
}

// 自动禁用阈值
const (
	AutoDisableSuccessRateThreshold = 0.50 // 成功率低于50%时自动禁用
	AutoDisableMinNewRequests       = 20   // 手动启用后至少20个新请求才重新检查
)

func NewProviderRelayService(providerService *ProviderService, logService *LogService, addr string) *ProviderRelayService {
	if addr == "" {
		addr = ":18100"
	}

	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".code-relay")
	
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Printf("创建数据目录失败: %v", err)
	}
	
	const sqliteOptions = "?cache=shared&mode=rwc&_busy_timeout=5000&_journal_mode=WAL"

	dbPath := filepath.Join(dataDir, "app.db"+sqliteOptions)
	log.Printf("[DB] 初始化数据库: %s", dbPath)
	
	if err := xdb.Inits([]xdb.Config{
		{
			Name:        "default",
			Driver:      "sqlite",
			DSN:         dbPath,
			MaxOpenConn: 5,  // 增加连接数
			MaxIdleConn: 2,
		},
	}); err != nil {
		log.Printf("[DB] 初始化数据库失败: %v", err)
	} else if err := ensureRequestLogTable(); err != nil {
		log.Printf("[DB] 初始化 request_log 表失败: %v", err)
	} else {
		log.Printf("[DB] 数据库初始化成功")
	}

	return &ProviderRelayService{
		providerService:   providerService,
		logService:        logService,
		addr:              addr,
		lastCheckRequests: make(map[string]int64),
	}
}

func (prs *ProviderRelayService) Start() error {
	// 启动前验证配置
	warnings := prs.validateConfig()
	for _, w := range warnings {
		log.Printf("[Relay] 配置警告: %s", w)
	}

	// 生产模式：禁用 GIN 控制台输出
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	
	// 添加请求日志中间件
	router.Use(func(c *gin.Context) {
		log.Printf("[Relay] 中间件: 收到 %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})
	
	prs.registerRoutes(router)

	prs.server = &http.Server{
		Addr:    prs.addr,
		Handler: router,
	}

	log.Printf("[Relay] ========================================")
	log.Printf("[Relay] 代理服务启动中，监听地址: %s", prs.addr)
	log.Printf("[Relay] Claude API: http://127.0.0.1%s/v1/messages", prs.addr)
	log.Printf("[Relay] Codex API: http://127.0.0.1%s/responses", prs.addr)
	log.Printf("[Relay] ========================================")

	go func() {
		if err := prs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[Relay] 服务器错误: %v", err)
		}
	}()
	return nil
}

// validateConfig 验证所有 provider 的配置
// 返回警告列表（非阻塞性错误）
func (prs *ProviderRelayService) validateConfig() []string {
	warnings := make([]string, 0)

	for _, kind := range []string{"claude", "codex", "gemini"} {
		providers, err := prs.providerService.LoadProviders(kind)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("[%s] 加载配置失败: %v", kind, err))
			continue
		}

		enabledCount := 0
		for _, p := range providers {
			if !p.Enabled {
				continue
			}
			enabledCount++

			// 验证每个启用的 provider
			if errs := p.ValidateConfiguration(); len(errs) > 0 {
				for _, errMsg := range errs {
					warnings = append(warnings, fmt.Sprintf("[%s/%s] %s", kind, p.Name, errMsg))
				}
			}

			// 检查是否配置了模型白名单或映射
			if (p.SupportedModels == nil || len(p.SupportedModels) == 0) &&
				(p.ModelMapping == nil || len(p.ModelMapping) == 0) {
				warnings = append(warnings, fmt.Sprintf(
					"[%s/%s] 未配置 supportedModels 或 modelMapping，将假设支持所有模型（可能导致降级失败）",
					kind, p.Name))
			}
		}

		if enabledCount == 0 {
			warnings = append(warnings, fmt.Sprintf("[%s] 没有启用的 provider", kind))
		}
	}

	return warnings
}

func (prs *ProviderRelayService) Stop() error {
	if prs.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return prs.server.Shutdown(ctx)
}

func (prs *ProviderRelayService) Addr() string {
	return prs.addr
}

func (prs *ProviderRelayService) registerRoutes(router gin.IRouter) {
	router.POST("/v1/messages", prs.proxyHandler("claude", "/v1/messages"))
	router.POST("/responses", prs.proxyHandler("codex", "/responses"))
	// Gemini OpenAI-compatible routes
	router.POST("/v1/chat/completions", prs.proxyHandler("gemini", "/v1/chat/completions"))
	router.POST("/v1/embeddings", prs.proxyHandler("gemini", "/v1/embeddings"))
}

func (prs *ProviderRelayService) proxyHandler(kind string, endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[Relay] ========== 收到请求 ==========")
		log.Printf("[Relay] 请求路径: %s %s", c.Request.Method, c.Request.URL.Path)
		log.Printf("[Relay] 客户端: %s", c.ClientIP())
		
		var bodyBytes []byte
		if c.Request.Body != nil {
			data, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Printf("[Relay] 读取请求体失败: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
				return
			}
			bodyBytes = data
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		isStream := gjson.GetBytes(bodyBytes, "stream").Bool()
		requestedModel := gjson.GetBytes(bodyBytes, "model").String()
		
		log.Printf("[Relay] 请求模型: %s, 流式: %v, 请求体大小: %d bytes", requestedModel, isStream, len(bodyBytes))

		providers, err := prs.providerService.LoadProviders(kind)
		if err != nil {
			log.Printf("[Relay] 加载 providers 失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load providers"})
			return
		}
		
		log.Printf("[Relay] 加载到 %d 个 providers", len(providers))

		active := make([]Provider, 0, len(providers))
		for _, provider := range providers {
			// 基础过滤：enabled、URL、APIKey
			if !provider.Enabled {
				log.Printf("[Relay] 跳过 provider %s: 未启用", provider.Name)
				continue
			}
			if provider.APIURL == "" {
				log.Printf("[Relay] 跳过 provider %s: 无 API URL", provider.Name)
				continue
			}
			if provider.APIKey == "" {
				log.Printf("[Relay] 跳过 provider %s: 无 API Key", provider.Name)
				continue
			}

			// 配置验证：失败则自动跳过
			if errs := provider.ValidateConfiguration(); len(errs) > 0 {
				log.Printf("[Relay] 跳过 provider %s: 配置验证失败 %v", provider.Name, errs)
				continue
			}

			// 核心过滤：只保留支持请求模型的 provider
			if requestedModel != "" && !provider.IsModelSupported(requestedModel) {
				log.Printf("[Relay] 跳过 provider %s: 不支持模型 %s", provider.Name, requestedModel)
				continue
			}

			log.Printf("[Relay] 添加 provider: %s", provider.Name)
			active = append(active, provider)
		}
		
		log.Printf("[Relay] 可用 providers: %d 个", len(active))

		if len(active) == 0 {
			// 记录 404 错误日志
			go func() {
				if _, err := xdb.New("request_log").Insert(xdb.Record{
					"platform":   kind,
					"model":      requestedModel,
					"provider":   "",
					"http_code":  http.StatusNotFound,
					"created_at": time.Now().Format("2006-01-02 15:04:05"),
				}); err != nil {
					// 静默处理
				}
			}()
			
			// 返回符合 Anthropic API 格式的错误响应
			message := "no providers available"
			if requestedModel != "" {
				message = "没有可用的 provider 支持模型 '" + requestedModel + "'"
			}
			log.Printf("[Relay] 没有可用 provider: %s", message)
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{
				"type": "error",
				"error": gin.H{
					"type":    "not_found_error",
					"message": message,
				},
			})
			c.Writer.Flush()
			return
		}

		query := flattenQuery(c.Request.URL.Query())
		// 性能优化：只克隆白名单内的请求头，避免无效遍历
		clientHeaders := filterHeaders(c.Request.Header)

		totalProviders := len(active)
		var lastErr error
		var lastStatus int
		var lastBody []byte
		var lastHeaders http.Header
		bodyCache := make(map[string][]byte)
		
		for i, provider := range active {
			isLastProvider := (i == totalProviders - 1)
			
			effectiveModel := provider.GetEffectiveModel(requestedModel)

			currentBodyBytes := bodyBytes
			if effectiveModel != requestedModel && requestedModel != "" {
				// 性能优化：缓存已替换的模型体，避免在 provider 轮询中重复执行 sjson 操作
				if cachedBody, exists := bodyCache[effectiveModel]; exists {
					currentBodyBytes = cachedBody
				} else {
					modifiedBody, err := ReplaceModelInRequestBody(bodyBytes, effectiveModel)
					if err != nil {
						lastErr = err
						log.Printf("[Relay] 替换模型失败: %v", err)
						continue
					}
					currentBodyBytes = modifiedBody
					bodyCache[effectiveModel] = modifiedBody
				}
			}

			status, headers, body, err := prs.forwardRequest(c, kind, provider, endpoint, query, clientHeaders, currentBodyBytes, isStream, effectiveModel)

			if err != nil {
				log.Printf("[Relay] Provider %s 请求失败: %v", provider.Name, err)
				lastErr = err
				continue
			}

			// status = -1 表示流式响应已经直接写入客户端，直接返回
			if status == -1 {
				log.Printf("[Relay] Provider %s 流式转发完成", provider.Name)
				return
			}

			// 保存最后一次响应
			lastStatus = status
			lastHeaders = headers
			lastBody = body

			// 如果成功 (2xx)，立即返回
			if status >= 200 && status < 300 {
				log.Printf("[Relay] Provider %s 成功, status=%d", provider.Name, status)
				prs.writeResponse(c, status, headers, body)
				return
			}

			// 如果失败但是最后一个 provider，返回错误响应
			if isLastProvider {
				log.Printf("[Relay] 最后一个 provider %s 失败, status=%d, 返回错误给客户端", provider.Name, status)
				prs.writeResponse(c, status, headers, body)
				return
			}

			// 如果失败且还有其他 provider，继续尝试
			log.Printf("[Relay] Provider %s 失败, status=%d, 尝试下一个 provider", provider.Name, status)
		}

		// 如果所有 provider 都失败了（可能是网络错误等）
		if lastBody != nil {
			// 返回最后一个 provider 的响应
			prs.writeResponse(c, lastStatus, lastHeaders, lastBody)
			return
		}

		// 如果连响应都没有（所有请求都失败了）
		message := "所有 provider 均失败"
		if lastErr != nil {
			message = message + ": " + lastErr.Error()
		}
		log.Printf("[Relay] 所有 provider 失败: %s", message)
		
		// 返回符合 Anthropic API 格式的错误响应
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusBadGateway, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "api_error",
				"message": message,
			},
		})
		c.Writer.Flush()
	}
}

// writeResponse 将响应写入客户端
func (prs *ProviderRelayService) writeResponse(c *gin.Context, status int, headers http.Header, body []byte) {
	for k, vv := range headers {
		for _, v := range vv {
			c.Writer.Header().Add(k, v)
		}
	}
	c.Writer.WriteHeader(status)
	if _, err := c.Writer.Write(body); err != nil {
		log.Printf("[Relay] 写入客户端失败: %v", err)
	}
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// httpClient 用于转发请求，支持长连接和流式响应
var httpClient = &http.Client{
	Timeout: 0,
	Transport: &http.Transport{
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 100, // 显著提升并发处理能力
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false,
	},
}

// forwardRequest 转发请求到上游 provider
// 返回值: (状态码, 响应头, 响应体, 错误)
// - 返回响应数据，由调用者决定是否写入客户端
// - 如果发生网络错误，返回 error，调用者可以尝试下一个 provider
func (prs *ProviderRelayService) forwardRequest(
	c *gin.Context,
	kind string,
	provider Provider,
	endpoint string,
	query map[string]string,
	clientHeaders map[string]string,
	bodyBytes []byte,
	isStream bool,
	model string,
) (int, http.Header, []byte, error) {
	targetURL := joinURL(provider.APIURL, endpoint)
	
	// 构建查询参数
	if len(query) > 0 {
		params := make([]string, 0, len(query))
		for k, v := range query {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
		targetURL = targetURL + "?" + strings.Join(params, "&")
	}

	requestLog := &RequestLog{
		Platform: kind,
		Provider: provider.Name,
		Model:    model,
		IsStream: isStream,
	}
	start := time.Now()
	
	// 写入日志的函数
	writeLog := func() {
		requestLog.DurationSec = time.Since(start).Seconds()
		go func(rl *RequestLog) {
			if rl.Platform == "" {
				return
			}
			if _, err := xdb.New("request_log").Insert(xdb.Record{
				"platform":            rl.Platform,
				"model":               rl.Model,
				"provider":            rl.Provider,
				"http_code":           rl.HttpCode,
				"input_tokens":        rl.InputTokens,
				"output_tokens":       rl.OutputTokens,
				"cache_create_tokens": rl.CacheCreateTokens,
				"cache_read_tokens":   rl.CacheReadTokens,
				"reasoning_tokens":    rl.ReasoningTokens,
				"is_stream":           boolToInt(rl.IsStream),
				"duration_sec":        rl.DurationSec,
				"created_at":          time.Now().Format("2006-01-02 15:04:05"),
			}); err != nil {
				// 生产环境不频繁记录日志失败
			}
		}(requestLog)
	}

	// 创建请求并绑定 Context，确保客户端断开时同步停止上游请求
	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[Relay] 创建请求失败: %v", err)
		requestLog.HttpCode = 0
		writeLog()
		return 0, nil, nil, err
	}

	// 应用预过滤的请求头
	for k, v := range clientHeaders {
		req.Header.Set(k, v)
	}
	
	// 设置必要的请求头
	req.Header.Set("Content-Type", "application/json")
	// 同时设置两种认证头，兼容不同的 API 服务
	// - Authorization: Bearer xxx (标准 OAuth2 格式，大多数云服务使用)
	// - x-api-key: xxx (Anthropic 官方格式，本地代理如 gcli2api 使用)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey))
	req.Header.Set("x-api-key", provider.APIKey)
	// 强制设置 anthropic-version，仅针对 Claude 平台
	if kind == "claude" && req.Header.Get("anthropic-version") == "" {
		req.Header.Set("anthropic-version", "2023-06-01")
	}

	log.Printf("[Relay] 转发请求到 %s, model=%s, stream=%v", targetURL, model, isStream)
	log.Printf("[Relay] 请求头: anthropic-version=%s, x-api-key=%s...", 
		req.Header.Get("anthropic-version"), 
		func() string {
			key := req.Header.Get("x-api-key")
			if len(key) > 8 {
				return key[:8]
			}
			return key
		}())
	
	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[Relay] 请求失败: %v", err)
		requestLog.HttpCode = 0
		writeLog()
		return 0, nil, nil, err
	}

	status := resp.StatusCode
	requestLog.HttpCode = status
	log.Printf("[Relay] 收到响应, status=%d, content-type=%s", status, resp.Header.Get("Content-Type"))

	// 检测是否为 SSE 流式响应
	contentType := resp.Header.Get("Content-Type")
	isSSE := strings.Contains(contentType, "text/event-stream")

	// 如果请求是流式的，或者响应是 SSE，都使用流式转发
	// 但只有成功时才流式转发，失败时需要尝试下一个 provider
	shouldStream := isStream || isSSE
	
	if shouldStream && status >= 200 && status < 300 {
		log.Printf("[Relay] 使用流式转发模式 (请求stream=%v, 响应SSE=%v, status=%d)", isStream, isSSE, status)
		// 使用更加透传的方案：直接设置响应头并使用自定义 Writer 进行转发
		for k, vv := range resp.Header {
			for _, v := range vv {
				c.Writer.Header().Add(k, v)
			}
		}
		c.Writer.WriteHeader(status)

		// 创建一个带有 Token 统计功能的 Writer
		hook := RequestLogHook(c, kind, requestLog)
		parser := &streamParser{
			writer: c.Writer,
			hook:   hook,
		}

		// 使用 io.Copy 实现高性能透传，避免 bufio.Scanner 的行缓冲区造成的延迟
		if _, err := io.Copy(parser, resp.Body); err != nil {
			log.Printf("[Relay] 流式转发中断: %v", err)
		}
		// 必须调用 Flush 处理最后一行的解析
		parser.Flush()
		resp.Body.Close()

		log.Printf("[Relay] 流式转发完成, tokens: in=%d, out=%d",
			requestLog.InputTokens, requestLog.OutputTokens)

		writeLog()
		go prs.checkAndAutoDisable(kind, provider.Name)

		return -1, nil, nil, nil
	}

	// 非流式响应：读取完整响应体
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Relay] 读取响应体失败: %v", err)
		requestLog.HttpCode = 0
		writeLog()
		return 0, nil, nil, err
	}

	// 解析 token 用量
	// 为非流式响应解析 token 用量
	parserFn := getTokenParser(kind)
	bodyStr := string(body)
	if isSSE {
		parseEventPayload(bodyStr, parserFn, requestLog)
	} else {
		// 非 SSE 响应直接解析完整 JSON
		parserFn(bodyStr, requestLog)
	}

	log.Printf("[Relay] 转发完成, status=%d, body_size=%d, tokens: in=%d, out=%d", 
		status, len(body), requestLog.InputTokens, requestLog.OutputTokens)

	writeLog()
	
	// 异步检查成功率，如果低于阈值则自动禁用
	go prs.checkAndAutoDisable(kind, provider.Name)
	
	// 返回响应数据，由调用者决定如何处理
	return status, resp.Header, body, nil
}

func getTokenParser(kind string) func(string, *RequestLog) {
	if kind == "codex" || kind == "gemini" {
		return CodexParseTokenUsageFromResponse
	}
	return ClaudeCodeParseTokenUsageFromResponse
}

// checkAndAutoDisable 检查供应商成功率，如果低于阈值则自动禁用
// 逻辑：只有当自上次检查后新增了至少5个请求，才重新检查成功率
func (prs *ProviderRelayService) checkAndAutoDisable(kind string, providerName string) {
	if prs.logService == nil || prs.providerService == nil {
		return
	}

	// 获取成功率和总请求数
	successRate, totalRequests, err := prs.logService.GetProviderSuccessRate(kind, providerName)
	if err != nil {
		log.Printf("[Relay] 获取供应商 %s 成功率失败: %v", providerName, err)
		return
	}

	// 构建 key
	key := kind + ":" + providerName
	
	prs.lastCheckMu.Lock()
	lastCheck, exists := prs.lastCheckRequests[key]
	// 如果是第一次检查这个供应商，以当前请求数为基准
	// 这样重启后不会立即禁用，需要等新增5个请求
	if !exists {
		prs.lastCheckRequests[key] = totalRequests
		prs.lastCheckMu.Unlock()
		log.Printf("[Relay] 初始化供应商 %s 检查基准: %d 个请求", providerName, totalRequests)
		return
	}
	prs.lastCheckMu.Unlock()

	// 计算自上次检查后的新请求数
	newRequests := totalRequests - lastCheck
	
	// 新请求数不足，不检查
	if newRequests < AutoDisableMinNewRequests {
		return
	}

	// 成功率高于阈值，更新检查点但不禁用
	if successRate >= AutoDisableSuccessRateThreshold {
		prs.lastCheckMu.Lock()
		prs.lastCheckRequests[key] = totalRequests
		prs.lastCheckMu.Unlock()
		return
	}

	// 成功率低于阈值，自动禁用
	log.Printf("[Relay] ⚠️ 供应商 %s 成功率 %.1f%% 低于阈值 %.0f%%（新请求数: %d），自动禁用", 
		providerName, successRate*100, AutoDisableSuccessRateThreshold*100, newRequests)
	
	if err := prs.providerService.DisableProvider(kind, providerName); err != nil {
		log.Printf("[Relay] 自动禁用供应商 %s 失败: %v", providerName, err)
	} else {
		log.Printf("[Relay] ✅ 供应商 %s 已自动禁用，请手动检查后重新启用", providerName)
		// 记录禁用时的请求数，下次启用后需要再累积5个新请求才会重新检查
		prs.lastCheckMu.Lock()
		prs.lastCheckRequests[key] = totalRequests
		prs.lastCheckMu.Unlock()
	}
}

var allowedForwardHeaders = map[string]bool{
	"accept":                      true,
	"user-agent":                  true,
	"x-request-id":                true,
	"x-stainless-arch":            true,
	"x-stainless-lang":            true,
	"x-stainless-os":              true,
	"x-stainless-package-version": true,
	"x-stainless-runtime":         true,
	"x-stainless-runtime-version": true,
	"anthropic-version":           true,
	"anthropic-beta":              true,
}

func filterHeaders(header http.Header) map[string]string {
	filtered := make(map[string]string)
	for key, values := range header {
		if len(values) > 0 && allowedForwardHeaders[strings.ToLower(key)] {
			filtered[key] = values[len(values)-1]
		}
	}
	return filtered
}

func flattenQuery(values map[string][]string) map[string]string {
	query := make(map[string]string, len(values))
	for key, items := range values {
		if len(items) > 0 {
			query[key] = items[len(items)-1]
		}
	}
	return query
}

func joinURL(base string, endpoint string) string {
	base = strings.TrimSuffix(base, "/")
	endpoint = "/" + strings.TrimPrefix(endpoint, "/")
	return base + endpoint
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func ensureRequestLogColumn(db *sql.DB, column string, definition string) error {
	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('request_log') WHERE name = '%s'", column)
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		alter := fmt.Sprintf("ALTER TABLE request_log ADD COLUMN %s %s", column, definition)
		if _, err := db.Exec(alter); err != nil {
			return err
		}
	}
	return nil
}

func ensureRequestLogTable() error {
	db, err := xdb.DB("default")
	if err != nil {
		return err
	}
	return ensureRequestLogTableWithDB(db)
}

func ensureRequestLogTableWithDB(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return err
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return err
	}

	const createTableSQL = `CREATE TABLE IF NOT EXISTS request_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		platform TEXT,
		model TEXT,
		provider TEXT,
		http_code INTEGER,
		input_tokens INTEGER,
		output_tokens INTEGER,
		cache_create_tokens INTEGER,
		cache_read_tokens INTEGER,
		reasoning_tokens INTEGER,
		is_stream INTEGER DEFAULT 0,
		duration_sec REAL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(createTableSQL); err != nil {
		return err
	}

	if err := ensureRequestLogColumn(db, "created_at", "DATETIME DEFAULT CURRENT_TIMESTAMP"); err != nil {
		return err
	}
	if err := ensureRequestLogColumn(db, "is_stream", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureRequestLogColumn(db, "duration_sec", "REAL DEFAULT 0"); err != nil {
		return err
	}

	return nil
}

// streamParser 是一个自定义 Writer，它在透传原始字节流的同时
// 使用钩子函数（hook）来解析和记录 Token 用量，不会造成行级缓冲延迟
type streamParser struct {
	writer gin.ResponseWriter
	hook   func([]byte) (bool, []byte)
	buffer []byte // 用于暂存不完整的 SSE 线
}

func (s *streamParser) Write(p []byte) (n int, err error) {
	// 1. 立即透传数据给客户端，确保最低延迟 (TTFT)
	n, err = s.writer.Write(p)
	if flusher, ok := s.writer.(http.Flusher); ok {
		flusher.Flush()
	}

	// 2. 将数据交给 hook 处理（用于异步解析 token）
	// hook 会处理 SSE 行的分解和解析
	s.hook(p)

	return
}

func (s *streamParser) Flush() {
	// 调用 hook 处理 buffer 中剩余的最后一行（可能没有换行符）
	s.hook(nil)
}

func RequestLogHook(c *gin.Context, kind string, usage *RequestLog) func(data []byte) (bool, []byte) {
	// 使用 stateful buffer 记录未关闭的行，防止 chunk 截断导致解析失败
	var rowBuf bytes.Buffer
	
	parserFn := ClaudeCodeParseTokenUsageFromResponse
	if kind == "codex" {
		parserFn = CodexParseTokenUsageFromResponse
	}

	return func(data []byte) (bool, []byte) {
		if data == nil {
			// Flush 逻辑：处理最后一块可能不完整的数据
			if rowBuf.Len() > 0 {
				trimmed := strings.TrimSpace(rowBuf.String())
				if strings.HasPrefix(trimmed, "data:") {
					payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
					parserFn(payload, usage)
				}
				rowBuf.Reset()
			}
			return true, nil
		}

		rowBuf.Write(data)
		
		for {
			line, err := rowBuf.ReadString('\n')
			if err != nil {
				// 如果没有换行符，将读取到的部分存回 buffer 底部
				rowBuf.Write([]byte(line))
				break
			}
			
			// 处理完整的一行 SSE 数据
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "data:") {
				payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				parserFn(payload, usage)
			}
		}

		return true, data
	}
}

func parseEventPayload(payload string, parser func(string, *RequestLog), usage *RequestLog) {
	lines := strings.Split(payload, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data:") {
			// 统一使用更鲁棒的解析逻辑，兼容 data: { 和 data:{
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			parser(data, usage)
		}
	}
}

type RequestLog struct {
	ID                int64   `json:"id"`
	Platform          string  `json:"platform"` // claude code or codex
	Model             string  `json:"model"`
	Provider          string  `json:"provider"` // provider name
	HttpCode          int     `json:"http_code"`
	InputTokens       int     `json:"input_tokens"`
	OutputTokens      int     `json:"output_tokens"`
	CacheCreateTokens int     `json:"cache_create_tokens"`
	CacheReadTokens   int     `json:"cache_read_tokens"`
	ReasoningTokens   int     `json:"reasoning_tokens"`
	IsStream          bool    `json:"is_stream"`
	DurationSec       float64 `json:"duration_sec"`
	CreatedAt         string  `json:"created_at"`
	InputCost         float64 `json:"input_cost"`
	OutputCost        float64 `json:"output_cost"`
	CacheCreateCost   float64 `json:"cache_create_cost"`
	CacheReadCost     float64 `json:"cache_read_cost"`
	Ephemeral5mCost   float64 `json:"ephemeral_5m_cost"`
	Ephemeral1hCost   float64 `json:"ephemeral_1h_cost"`
	TotalCost         float64 `json:"total_cost"`
	HasPricing        bool    `json:"has_pricing"`
}

// claude code usage parser
func ClaudeCodeParseTokenUsageFromResponse(data string, usage *RequestLog) {
	// 1. 处理流式 chunk 中的 message 嵌套格式
	usage.InputTokens += int(gjson.Get(data, "message.usage.input_tokens").Int())
	usage.OutputTokens += int(gjson.Get(data, "message.usage.output_tokens").Int())
	usage.CacheCreateTokens += int(gjson.Get(data, "message.usage.cache_creation_input_tokens").Int())
	usage.CacheReadTokens += int(gjson.Get(data, "message.usage.cache_read_input_tokens").Int())

	// 2. 处理标准响应或某些 chunk 中的根级 usage 格式
	u := gjson.Get(data, "usage")
	if u.Exists() {
		usage.InputTokens += int(u.Get("input_tokens").Int())
		usage.OutputTokens += int(u.Get("output_tokens").Int())
		usage.CacheCreateTokens += int(u.Get("cache_creation_input_tokens").Int())
		usage.CacheReadTokens += int(u.Get("cache_read_input_tokens").Int())
	}
}

// codex usage parser
func CodexParseTokenUsageFromResponse(data string, usage *RequestLog) {
	// 尝试解析 response.usage (部分代理格式) 或根级 usage (标准格式)
	u := gjson.Get(data, "response.usage")
	if !u.Exists() {
		u = gjson.Get(data, "usage")
	}

	if u.Exists() {
		usage.InputTokens += int(u.Get("input_tokens").Int())
		usage.OutputTokens += int(u.Get("output_tokens").Int())
		// 增加对 prompt_tokens/completion_tokens 的兼容
		if usage.InputTokens == 0 {
			usage.InputTokens = int(u.Get("prompt_tokens").Int())
		}
		if usage.OutputTokens == 0 {
			usage.OutputTokens = int(u.Get("completion_tokens").Int())
		}
		usage.CacheReadTokens += int(u.Get("input_tokens_details.cached_tokens").Int())
		usage.ReasoningTokens += int(u.Get("output_tokens_details.reasoning_tokens").Int())
	}
}

// ReplaceModelInRequestBody 替换请求体中的模型名
// 使用 gjson + sjson 实现高性能 JSON 操作，避免完整反序列化
func ReplaceModelInRequestBody(bodyBytes []byte, newModel string) ([]byte, error) {
	// 检查请求体中是否存在 model 字段
	result := gjson.GetBytes(bodyBytes, "model")
	if !result.Exists() {
		return bodyBytes, fmt.Errorf("请求体中未找到 model 字段")
	}

	// 使用 sjson.SetBytes 替换模型名（高性能操作）
	modified, err := sjson.SetBytes(bodyBytes, "model", newModel)
	if err != nil {
		return bodyBytes, fmt.Errorf("替换模型名失败: %w", err)
	}

	return modified, nil
}
