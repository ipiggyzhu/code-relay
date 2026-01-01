package services

import (
	"errors"
	"log"
	"sort"
	"strings"
	"time"

	modelpricing "coderelay/resources/model-pricing"

	"github.com/daodao97/xgo/xdb"
)

const timeLayout = "2006-01-02 15:04:05"

type LogService struct {
	pricing *modelpricing.Service
}

func NewLogService() *LogService {
	svc, err := modelpricing.DefaultService()
	if err != nil {
		log.Printf("pricing service init failed: %v", err)
	}
	return &LogService{pricing: svc}
}

func (ls *LogService) ListRequestLogs(platform string, provider string, limit int) ([]RequestLog, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	
	model := xdb.New("request_log")
	options := []xdb.Option{
		xdb.OrderByDesc("id"),
		xdb.Limit(limit),
	}
	if platform != "" {
		options = append(options, xdb.WhereEq("platform", platform))
	}
	if provider != "" {
		options = append(options, xdb.WhereEq("provider", provider))
	}
	records, err := model.Selects(options...)
	if err != nil {
		// 处理表不存在或其他错误
		if errors.Is(err, xdb.ErrNotFound) || isNoSuchTableErr(err) {
			return []RequestLog{}, nil
		}
		return nil, err
	}
	logs := make([]RequestLog, 0, len(records))
	for _, record := range records {
		logEntry := RequestLog{
			ID:                record.GetInt64("id"),
			Platform:          record.GetString("platform"),
			Model:             record.GetString("model"),
			Provider:          record.GetString("provider"),
			HttpCode:          record.GetInt("http_code"),
			InputTokens:       record.GetInt("input_tokens"),
			OutputTokens:      record.GetInt("output_tokens"),
			CacheCreateTokens: record.GetInt("cache_create_tokens"),
			CacheReadTokens:   record.GetInt("cache_read_tokens"),
			ReasoningTokens:   record.GetInt("reasoning_tokens"),
			CreatedAt:         record.GetString("created_at"),
			IsStream:          record.GetBool("is_stream"),
			DurationSec:       record.GetFloat64("duration_sec"),
		}
		ls.decorateCost(&logEntry)
		logs = append(logs, logEntry)
	}
	return logs, nil
}

func (ls *LogService) ListProviders(platform string) ([]string, error) {
	model := xdb.New("request_log")
	options := []xdb.Option{
		xdb.Field("DISTINCT provider as provider"),
		xdb.WhereNotEq("provider", ""),
		xdb.OrderByAsc("provider"),
	}
	if platform != "" {
		options = append(options, xdb.WhereEq("platform", platform))
	}
	records, err := model.Selects(options...)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) || isNoSuchTableErr(err) {
			return []string{}, nil
		}
		return nil, err
	}
	providers := make([]string, 0, len(records))
	for _, record := range records {
		name := strings.TrimSpace(record.GetString("provider"))
		if name != "" {
			providers = append(providers, name)
		}
	}
	return providers, nil
}

func (ls *LogService) HeatmapStats(days int) ([]HeatmapStat, error) {
	if days <= 0 {
		days = 30
	}
	totalHours := days * 24
	if totalHours <= 0 {
		totalHours = 24
	}
	rangeStart := startOfHour(time.Now())
	if totalHours > 1 {
		rangeStart = rangeStart.Add(-time.Duration(totalHours-1) * time.Hour)
	}
	model := xdb.New("request_log")
	options := []xdb.Option{
		xdb.WhereGe("created_at", rangeStart.Format(timeLayout)),
		xdb.Field(
			"model",
			"input_tokens",
			"output_tokens",
			"reasoning_tokens",
			"cache_create_tokens",
			"cache_read_tokens",
			"created_at",
		),
		xdb.OrderByDesc("created_at"),
	}
	records, err := model.Selects(options...)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) || isNoSuchTableErr(err) {
			return []HeatmapStat{}, nil
		}
		return nil, err
	}
	hourBuckets := map[int64]*HeatmapStat{}
	for _, record := range records {
		createdAt, _ := parseCreatedAt(record)
		if createdAt.IsZero() {
			continue
		}
		hourStart := startOfHour(createdAt)
		hourKey := hourStart.Unix()
		bucket := hourBuckets[hourKey]
		if bucket == nil {
			bucket = &HeatmapStat{Day: hourStart.Format("2006-01-02 15")}
			hourBuckets[hourKey] = bucket
		}
		bucket.TotalRequests++
		input := record.GetInt("input_tokens")
		output := record.GetInt("output_tokens")
		reasoning := record.GetInt("reasoning_tokens")
		cacheCreate := record.GetInt("cache_create_tokens")
		cacheRead := record.GetInt("cache_read_tokens")
		bucket.InputTokens += int64(input)
		bucket.OutputTokens += int64(output)
		bucket.ReasoningTokens += int64(reasoning)
		usage := modelpricing.UsageSnapshot{
			InputTokens:       input,
			OutputTokens:      output,
			CacheCreateTokens: cacheCreate,
			CacheReadTokens:   cacheRead,
		}
		cost := ls.calculateCost(record.GetString("model"), usage)
		bucket.TotalCost += cost.TotalCost
	}
	if len(hourBuckets) == 0 {
		return []HeatmapStat{}, nil
	}
	hourKeys := make([]int64, 0, len(hourBuckets))
	for key := range hourBuckets {
		hourKeys = append(hourKeys, key)
	}
	sort.Slice(hourKeys, func(i, j int) bool {
		return hourKeys[i] < hourKeys[j]
	})
	stats := make([]HeatmapStat, 0, min(len(hourKeys), totalHours))
	for i := len(hourKeys) - 1; i >= 0 && len(stats) < totalHours; i-- {
		stats = append(stats, *hourBuckets[hourKeys[i]])
	}
	return stats, nil
}

func (ls *LogService) StatsSince(platform string) (LogStats, error) {
	const seriesHours = 24

	stats := LogStats{
		Series: make([]LogStatsSeries, 0, seriesHours),
	}
	now := time.Now()
	model := xdb.New("request_log")
	seriesStart := startOfDay(now)
	seriesEnd := seriesStart.Add(seriesHours * time.Hour)
	// 只查询今天的数据，不再查询昨天的
	queryStart := seriesStart
	summaryStart := seriesStart
	options := []xdb.Option{
		xdb.WhereGte("created_at", queryStart.Format(timeLayout)),
		xdb.Field(
			"model",
			"input_tokens",
			"output_tokens",
			"reasoning_tokens",
			"cache_create_tokens",
			"cache_read_tokens",
			"created_at",
		),
		xdb.OrderByAsc("created_at"),
	}
	if platform != "" {
		options = append(options, xdb.WhereEq("platform", platform))
	}
	records, err := model.Selects(options...)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) || isNoSuchTableErr(err) {
			return stats, nil
		}
		return stats, err
	}

	seriesBuckets := make([]*LogStatsSeries, seriesHours)
	for i := 0; i < seriesHours; i++ {
		bucketTime := seriesStart.Add(time.Duration(i) * time.Hour)
		seriesBuckets[i] = &LogStatsSeries{
			Day: bucketTime.Format(timeLayout),
		}
	}

	for _, record := range records {
		createdAt, hasTime := parseCreatedAt(record)
		dayKey := dayFromTimestamp(record.GetString("created_at"))
		isToday := dayKey == seriesStart.Format("2006-01-02")

		if hasTime {
			if createdAt.Before(seriesStart) || !createdAt.Before(seriesEnd) {
				continue
			}
		} else {
			if !isToday {
				continue
			}
			createdAt = seriesStart
		}

		bucketIndex := 0
		if hasTime {
			bucketIndex = int(createdAt.Sub(seriesStart) / time.Hour)
			if bucketIndex < 0 {
				bucketIndex = 0
			}
			if bucketIndex >= seriesHours {
				bucketIndex = seriesHours - 1
			}
		}
		bucket := seriesBuckets[bucketIndex]
		input := record.GetInt("input_tokens")
		output := record.GetInt("output_tokens")
		reasoning := record.GetInt("reasoning_tokens")
		cacheCreate := record.GetInt("cache_create_tokens")
		cacheRead := record.GetInt("cache_read_tokens")
		usage := modelpricing.UsageSnapshot{
			InputTokens:       input,
			OutputTokens:      output,
			CacheCreateTokens: cacheCreate,
			CacheReadTokens:   cacheRead,
		}
		cost := ls.calculateCost(record.GetString("model"), usage)

		bucket.TotalRequests++
		bucket.InputTokens += int64(input)
		bucket.OutputTokens += int64(output)
		bucket.ReasoningTokens += int64(reasoning)
		bucket.CacheCreateTokens += int64(cacheCreate)
		bucket.CacheReadTokens += int64(cacheRead)
		bucket.TotalCost += cost.TotalCost

		if createdAt.IsZero() || createdAt.Before(summaryStart) {
			continue
		}
		stats.TotalRequests++
		stats.InputTokens += int64(input)
		stats.OutputTokens += int64(output)
		stats.ReasoningTokens += int64(reasoning)
		stats.CacheCreateTokens += int64(cacheCreate)
		stats.CacheReadTokens += int64(cacheRead)
		stats.CostInput += cost.InputCost
		stats.CostOutput += cost.OutputCost
		stats.CostCacheCreate += cost.CacheCreateCost
		stats.CostCacheRead += cost.CacheReadCost
		stats.CostTotal += cost.TotalCost
	}

	for i := 0; i < seriesHours; i++ {
		if bucket := seriesBuckets[i]; bucket != nil {
			stats.Series = append(stats.Series, *bucket)
		} else {
			bucketTime := seriesStart.Add(time.Duration(i) * time.Hour)
			stats.Series = append(stats.Series, LogStatsSeries{
				Day: bucketTime.Format(timeLayout),
			})
		}
	}

	return stats, nil
}

func (ls *LogService) ProviderDailyStats(platform string) ([]ProviderDailyStat, error) {
	start := startOfDay(time.Now())
	end := start.Add(24 * time.Hour)
	queryStart := start.Add(-24 * time.Hour)

	model := xdb.New("request_log")
	options := []xdb.Option{
		xdb.WhereGte("created_at", queryStart.Format(timeLayout)),
		xdb.Field(
			"provider",
			"model",
			"http_code",
			"input_tokens",
			"output_tokens",
			"reasoning_tokens",
			"cache_create_tokens",
			"cache_read_tokens",
			"duration_sec",
			"created_at",
		),
	}
	if platform != "" {
		options = append(options, xdb.WhereEq("platform", platform))
	}
	records, err := model.Selects(options...)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) || isNoSuchTableErr(err) {
			return []ProviderDailyStat{}, nil
		}
		return nil, err
	}

	statMap := map[string]*ProviderDailyStat{}
	// 临时存储每个 provider 的响应时间列表，用于计算 min/max/avg
	durationMap := map[string][]float64{}
	for _, record := range records {
		provider := strings.TrimSpace(record.GetString("provider"))
		if provider == "" {
			provider = "(unknown)"
		}
		createdAt, hasTime := parseCreatedAt(record)

		if hasTime {
			if createdAt.Before(start) || !createdAt.Before(end) {
				continue
			}
		} else {
			dayKey := dayFromTimestamp(record.GetString("created_at"))
			if dayKey != start.Format("2006-01-02") {
				continue
			}
		}
		stat := statMap[provider]
		if stat == nil {
			stat = &ProviderDailyStat{Provider: provider}
			statMap[provider] = stat
			durationMap[provider] = []float64{}
		}
		httpCode := record.GetInt("http_code")
		input := record.GetInt("input_tokens")
		output := record.GetInt("output_tokens")
		reasoning := record.GetInt("reasoning_tokens")
		cacheCreate := record.GetInt("cache_create_tokens")
		cacheRead := record.GetInt("cache_read_tokens")
		durationSec := record.GetFloat64("duration_sec")
		usage := modelpricing.UsageSnapshot{
			InputTokens:       input,
			OutputTokens:      output,
			CacheCreateTokens: cacheCreate,
			CacheReadTokens:   cacheRead,
		}
		cost := ls.calculateCost(record.GetString("model"), usage)
		stat.TotalRequests++
		// 只有 HTTP 200-299 才算成功，其他（包括 0）都算失败
		if httpCode >= 200 && httpCode < 300 {
			stat.SuccessfulRequests++
		} else {
			stat.FailedRequests++
		}
		stat.InputTokens += int64(input)
		stat.OutputTokens += int64(output)
		stat.ReasoningTokens += int64(reasoning)
		stat.CacheCreateTokens += int64(cacheCreate)
		stat.CacheReadTokens += int64(cacheRead)
		stat.CostTotal += cost.TotalCost
		// 记录响应时间（只记录有效的响应时间）
		if durationSec > 0 {
			durationMap[provider] = append(durationMap[provider], durationSec)
		}
	}
	
	stats := make([]ProviderDailyStat, 0, len(statMap))
	for _, stat := range statMap {
		if stat.TotalRequests > 0 {
			stat.SuccessRate = float64(stat.SuccessfulRequests) / float64(stat.TotalRequests)
		}
		// 计算响应时间统计
		durations := durationMap[stat.Provider]
		if len(durations) > 0 {
			var sum float64
			minDur := durations[0]
			maxDur := durations[0]
			for _, d := range durations {
				sum += d
				if d < minDur {
					minDur = d
				}
				if d > maxDur {
					maxDur = d
				}
			}
			stat.AvgDurationSec = sum / float64(len(durations))
			stat.MinDurationSec = minDur
			stat.MaxDurationSec = maxDur
		}
		stats = append(stats, *stat)
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].TotalRequests == stats[j].TotalRequests {
			return stats[i].Provider < stats[j].Provider
		}
		return stats[i].TotalRequests > stats[j].TotalRequests
	})

	return stats, nil
}

func (ls *LogService) decorateCost(logEntry *RequestLog) {
	if ls == nil || ls.pricing == nil || logEntry == nil {
		return
	}
	usage := modelpricing.UsageSnapshot{
		InputTokens:       logEntry.InputTokens,
		OutputTokens:      logEntry.OutputTokens,
		CacheCreateTokens: logEntry.CacheCreateTokens,
		CacheReadTokens:   logEntry.CacheReadTokens,
	}
	cost := ls.pricing.CalculateCost(logEntry.Model, usage)
	logEntry.HasPricing = cost.HasPricing
	logEntry.InputCost = cost.InputCost
	logEntry.OutputCost = cost.OutputCost
	logEntry.CacheCreateCost = cost.CacheCreateCost
	logEntry.CacheReadCost = cost.CacheReadCost
	logEntry.Ephemeral5mCost = cost.Ephemeral5mCost
	logEntry.Ephemeral1hCost = cost.Ephemeral1hCost
	logEntry.TotalCost = cost.TotalCost
}

func (ls *LogService) calculateCost(model string, usage modelpricing.UsageSnapshot) modelpricing.CostBreakdown {
	if ls == nil || ls.pricing == nil {
		return modelpricing.CostBreakdown{}
	}
	return ls.pricing.CalculateCost(model, usage)
}

func parseCreatedAt(record xdb.Record) (time.Time, bool) {
	raw := strings.TrimSpace(record.GetString("created_at"))
	if raw == "" {
		return time.Time{}, false
	}

	// 提取日期时间部分（去掉任何时区后缀）
	// xdb/SQLite 返回的格式可能是：
	// - "2025-12-19 23:49:00" (纯本地时间)
	// - "2025-12-19 23:49:00 +0000 UTC" (xdb 自动添加的 UTC 后缀，但实际是本地时间)
	dateTimePart := raw

	// 移除 " +0000 UTC" 或类似的时区后缀
	if idx := strings.Index(raw, " +"); idx > 0 {
		dateTimePart = raw[:idx]
	} else if idx := strings.Index(raw, " -"); idx > 0 && idx > 10 {
		// 确保不是日期中的 "-"，只匹配时区偏移
		dateTimePart = raw[:idx]
	}

	// 尝试解析标准格式 "2006-01-02 15:04:05"
	if parsed, err := time.ParseInLocation(timeLayout, dateTimePart, time.Local); err == nil {
		return parsed, true
	}

	// 尝试 ISO 格式 "2006-01-02T15:04:05"
	if parsed, err := time.ParseInLocation("2006-01-02T15:04:05", dateTimePart, time.Local); err == nil {
		return parsed, true
	}

	// 尝试只有日期的格式
	if len(dateTimePart) >= 10 {
		if parsed, err := time.ParseInLocation("2006-01-02", dateTimePart[:10], time.Local); err == nil {
			return parsed, false
		}
	}

	return time.Time{}, false
}

func dayFromTimestamp(value string) string {
	if len(value) >= len("2006-01-02") {
		if t, err := time.ParseInLocation(timeLayout, value, time.Local); err == nil {
			return t.Format("2006-01-02")
		}
		return value[:10]
	}
	return value
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func startOfHour(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), 0, 0, 0, t.Location())
}


func isNoSuchTableErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "no such table")
}

// GetProviderSuccessRate 获取指定供应商的成功率
// 返回成功率（0-1）和总请求数
func (ls *LogService) GetProviderSuccessRate(platform string, providerName string) (float64, int64, error) {
	if providerName == "" {
		return 0, 0, nil
	}

	// 查询今天的数据
	start := startOfDay(time.Now())
	
	model := xdb.New("request_log")
	options := []xdb.Option{
		xdb.WhereGte("created_at", start.Format(timeLayout)),
		xdb.WhereEq("provider", providerName),
		xdb.Field("http_code"),
	}
	if platform != "" {
		options = append(options, xdb.WhereEq("platform", platform))
	}
	
	records, err := model.Selects(options...)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) || isNoSuchTableErr(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	if len(records) == 0 {
		return 0, 0, nil
	}

	var totalRequests int64
	var successfulRequests int64
	
	for _, record := range records {
		totalRequests++
		httpCode := record.GetInt("http_code")
		if httpCode >= 200 && httpCode < 300 {
			successfulRequests++
		}
	}

	if totalRequests == 0 {
		return 0, 0, nil
	}

	successRate := float64(successfulRequests) / float64(totalRequests)
	return successRate, totalRequests, nil
}

type HeatmapStat struct {
	Day             string  `json:"day"`
	TotalRequests   int64   `json:"total_requests"`
	InputTokens     int64   `json:"input_tokens"`
	OutputTokens    int64   `json:"output_tokens"`
	ReasoningTokens int64   `json:"reasoning_tokens"`
	TotalCost       float64 `json:"total_cost"`
}

type LogStats struct {
	TotalRequests     int64            `json:"total_requests"`
	InputTokens       int64            `json:"input_tokens"`
	OutputTokens      int64            `json:"output_tokens"`
	ReasoningTokens   int64            `json:"reasoning_tokens"`
	CacheCreateTokens int64            `json:"cache_create_tokens"`
	CacheReadTokens   int64            `json:"cache_read_tokens"`
	CostTotal         float64          `json:"cost_total"`
	CostInput         float64          `json:"cost_input"`
	CostOutput        float64          `json:"cost_output"`
	CostCacheCreate   float64          `json:"cost_cache_create"`
	CostCacheRead     float64          `json:"cost_cache_read"`
	Series            []LogStatsSeries `json:"series"`
}

type ProviderDailyStat struct {
	Provider           string  `json:"provider"`
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	SuccessRate        float64 `json:"success_rate"`
	InputTokens        int64   `json:"input_tokens"`
	OutputTokens       int64   `json:"output_tokens"`
	ReasoningTokens    int64   `json:"reasoning_tokens"`
	CacheCreateTokens  int64   `json:"cache_create_tokens"`
	CacheReadTokens    int64   `json:"cache_read_tokens"`
	CostTotal          float64 `json:"cost_total"`
	// 响应时间统计
	AvgDurationSec float64 `json:"avg_duration_sec"` // 平均响应时间（秒）
	MinDurationSec float64 `json:"min_duration_sec"` // 最小响应时间（秒）
	MaxDurationSec float64 `json:"max_duration_sec"` // 最大响应时间（秒）
}

type LogStatsSeries struct {
	Day               string  `json:"day"`
	TotalRequests     int64   `json:"total_requests"`
	InputTokens       int64   `json:"input_tokens"`
	OutputTokens      int64   `json:"output_tokens"`
	ReasoningTokens   int64   `json:"reasoning_tokens"`
	CacheCreateTokens int64   `json:"cache_create_tokens"`
	CacheReadTokens   int64   `json:"cache_read_tokens"`
	TotalCost         float64 `json:"total_cost"`
}
