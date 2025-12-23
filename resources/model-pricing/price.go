package modelpricing

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed model_prices_and_context_window.json
var pricingFile []byte

var (
	defaultOnce    sync.Once
	defaultService *Service
	defaultErr     error
	nameReplacer   = strings.NewReplacer("-", "", "_", "", ".", "", ":", "", "/", "", " ", "")
)

// Service 提供模型价格相关的计算能力。
type Service struct {
	pricingMap   map[string]*PricingEntry
	normalized   map[string]string
	ephemeral1h  map[string]float64
	longContexts map[string]LongContextPricing
}

// PricingEntry 映射 JSON 内的字段。
type PricingEntry struct {
	InputCostPerToken                   float64 `json:"input_cost_per_token"`
	OutputCostPerToken                  float64 `json:"output_cost_per_token"`
	CacheCreationInputTokenCost         float64 `json:"cache_creation_input_token_cost"`
	CacheCreationInputTokenCostAbove1Hr float64 `json:"cache_creation_input_token_cost_above_1hr"`
	CacheCreationInputTokenCostAbove200 float64 `json:"cache_creation_input_token_cost_above_200k_tokens"`
	CacheReadInputTokenCost             float64 `json:"cache_read_input_token_cost"`
	InputCostPerTokenAbove200k          float64 `json:"input_cost_per_token_above_200k_tokens"`
	InputCostPerTokenAbove128k          float64 `json:"input_cost_per_token_above_128k_tokens"`
	OutputCostPerTokenAbove200k         float64 `json:"output_cost_per_token_above_200k_tokens"`
}

// UsageSnapshot 描述一次请求的 token 用量。
type UsageSnapshot struct {
	InputTokens       int
	OutputTokens      int
	CacheCreateTokens int
	CacheReadTokens   int
	CacheCreation     *CacheCreationDetail
}

// CacheCreationDetail 细分缓存创建 tokens。
type CacheCreationDetail struct {
	Ephemeral5mTokens int
	Ephemeral1hTokens int
}

// CostBreakdown 表示一次费用计算的结果。
type CostBreakdown struct {
	InputCost       float64 `json:"input_cost"`
	OutputCost      float64 `json:"output_cost"`
	CacheCreateCost float64 `json:"cache_create_cost"`
	CacheReadCost   float64 `json:"cache_read_cost"`
	Ephemeral5mCost float64 `json:"ephemeral_5m_cost"`
	Ephemeral1hCost float64 `json:"ephemeral_1h_cost"`
	TotalCost       float64 `json:"total_cost"`
	HasPricing      bool    `json:"has_pricing"`
	IsLongContext   bool    `json:"is_long_context"`
}

// LongContextPricing 描述 1M 上下文模型的单价。
type LongContextPricing struct {
	Input  float64
	Output float64
}

// DefaultService 返回单例。
func DefaultService() (*Service, error) {
	defaultOnce.Do(func() {
		defaultService, defaultErr = NewService()
	})
	return defaultService, defaultErr
}

// NewService 从嵌入的 JSON 创建服务实例。
func NewService() (*Service, error) {
	raw := make(map[string]PricingEntry)
	if err := json.Unmarshal(pricingFile, &raw); err != nil {
		return nil, fmt.Errorf("parse pricing file: %w", err)
	}
	pricing := make(map[string]*PricingEntry, len(raw))
	normalized := make(map[string]string, len(raw))
	for key, entry := range raw {
		item := entry
		ensureCachePricing(&item)
		pricing[key] = &item
		norm := normalizeName(key)
		if _, exists := normalized[norm]; !exists {
			normalized[norm] = key
		}
	}
	return &Service{
		pricingMap:   pricing,
		normalized:   normalized,
		ephemeral1h:  buildEphemeral1hPricing(),
		longContexts: buildLongContextPricing(),
	}, nil
}

// CalculateCost 根据模型与 token 用量返回费用明细（美元）。
func (s *Service) CalculateCost(model string, usage UsageSnapshot) CostBreakdown {
	if s == nil || model == "" {
		return CostBreakdown{}
	}
	entry, hasPricing := s.getPricing(model)
	breakdown := CostBreakdown{HasPricing: hasPricing}
	if entry == nil && !strings.Contains(strings.ToLower(model), "[1m]") {
		return breakdown
	}
	longTier, useLong := s.longContextTier(model, usage)
	if entry == nil {
		entry = &PricingEntry{}
	}
	if useLong {
		breakdown.IsLongContext = true
		breakdown.InputCost = float64(usage.InputTokens) * longTier.Input
		breakdown.OutputCost = float64(usage.OutputTokens) * longTier.Output
	} else {
		breakdown.InputCost = float64(usage.InputTokens) * entry.InputCostPerToken
		breakdown.OutputCost = float64(usage.OutputTokens) * entry.OutputCostPerToken
	}
	cacheCreateTokens, cache1hTokens := resolveCacheTokens(usage)
	cache5mCost := float64(cacheCreateTokens) * entry.CacheCreationInputTokenCost
	cache1hCost := float64(cache1hTokens) * s.getEphemeral1hPricing(model)
	breakdown.Ephemeral5mCost = cache5mCost
	breakdown.Ephemeral1hCost = cache1hCost
	breakdown.CacheCreateCost = cache5mCost + cache1hCost
	breakdown.CacheReadCost = float64(usage.CacheReadTokens) * entry.CacheReadInputTokenCost
	breakdown.TotalCost = breakdown.InputCost + breakdown.OutputCost + breakdown.CacheCreateCost + breakdown.CacheReadCost
	if breakdown.TotalCost > 0 {
		breakdown.HasPricing = true
	}
	return breakdown
}

func (s *Service) getPricing(model string) (*PricingEntry, bool) {
	if model == "" {
		return nil, false
	}
	if entry, ok := s.pricingMap[model]; ok {
		return entry, true
	}
	if model == "gpt-5-codex" {
		if entry, ok := s.pricingMap["gpt-5"]; ok {
			return entry, true
		}
	}

	// 检查硬编码的热门模型价格
	if entry := getHardcodedPricing(model); entry != nil {
		return entry, true
	}

	withoutRegion := stripRegionPrefix(model)
	if entry, ok := s.pricingMap[withoutRegion]; ok {
		return entry, true
	}
	withoutProvider := strings.TrimPrefix(withoutRegion, "anthropic.")
	if entry, ok := s.pricingMap[withoutProvider]; ok {
		return entry, true
	}
	normalizedTarget := normalizeName(model)
	if key, ok := s.normalized[normalizedTarget]; ok {
		return s.pricingMap[key], true
	}
	for key, entry := range s.pricingMap {
		normKey := normalizeName(key)
		if strings.Contains(normKey, normalizedTarget) || strings.Contains(normalizedTarget, normKey) {
			return entry, true
		}
	}
	return nil, false
}

// getHardcodedPricing 返回硬编码的热门模型价格
// 这些价格会覆盖 JSON 文件中的价格，确保最新
func getHardcodedPricing(model string) *PricingEntry {
	lowerModel := strings.ToLower(model)

	// ==================== Anthropic Claude 系列 ====================

	// Claude Opus 4.5 (Input: $5/MTok, Output: $25/MTok)
	if strings.Contains(lowerModel, "opus-4-5") || strings.Contains(lowerModel, "opus-4.5") ||
		strings.Contains(lowerModel, "opus4.5") || strings.Contains(lowerModel, "opus45") {
		return &PricingEntry{
			InputCostPerToken:           0.000005,   // $5/MTok
			OutputCostPerToken:          0.000025,   // $25/MTok
			CacheCreationInputTokenCost: 0.00000625, // $6.25/MTok
			CacheReadInputTokenCost:     0.0000005,  // $0.50/MTok
		}
	}

	// Claude Sonnet 4.5 (Input: $3/MTok, Output: $15/MTok)
	if strings.Contains(lowerModel, "sonnet-4-5") || strings.Contains(lowerModel, "sonnet-4.5") ||
		strings.Contains(lowerModel, "sonnet4.5") || strings.Contains(lowerModel, "sonnet45") {
		return &PricingEntry{
			InputCostPerToken:           0.000003,   // $3/MTok
			OutputCostPerToken:          0.000015,   // $15/MTok
			CacheCreationInputTokenCost: 0.00000375, // $3.75/MTok
			CacheReadInputTokenCost:     0.0000003,  // $0.30/MTok
		}
	}

	// Claude Haiku 4.5 (Input: $1/MTok, Output: $5/MTok)
	if strings.Contains(lowerModel, "haiku-4-5") || strings.Contains(lowerModel, "haiku-4.5") ||
		strings.Contains(lowerModel, "haiku4.5") || strings.Contains(lowerModel, "haiku45") {
		return &PricingEntry{
			InputCostPerToken:           0.000001,   // $1/MTok
			OutputCostPerToken:          0.000005,   // $5/MTok
			CacheCreationInputTokenCost: 0.00000125, // $1.25/MTok
			CacheReadInputTokenCost:     0.0000001,  // $0.10/MTok
		}
	}

	// ==================== OpenAI GPT 系列 ====================

	// GPT-5 (Input: $1.25/MTok, Output: $10/MTok) - 从 JSON 文件获取
	if strings.Contains(lowerModel, "gpt-5") && !strings.Contains(lowerModel, "mini") && !strings.Contains(lowerModel, "nano") {
		return &PricingEntry{
			InputCostPerToken:       0.00000125, // $1.25/MTok
			OutputCostPerToken:      0.00001,    // $10/MTok
			CacheReadInputTokenCost: 0.000000125, // $0.125/MTok
		}
	}

	// GPT-5-mini (Input: $0.30/MTok, Output: $1.25/MTok)
	if strings.Contains(lowerModel, "gpt-5-mini") || strings.Contains(lowerModel, "gpt5-mini") {
		return &PricingEntry{
			InputCostPerToken:  0.0000003,  // $0.30/MTok
			OutputCostPerToken: 0.00000125, // $1.25/MTok
		}
	}

	// GPT-5-nano (Input: $0.10/MTok, Output: $0.40/MTok)
	if strings.Contains(lowerModel, "gpt-5-nano") || strings.Contains(lowerModel, "gpt5-nano") {
		return &PricingEntry{
			InputCostPerToken:  0.0000001, // $0.10/MTok
			OutputCostPerToken: 0.0000004, // $0.40/MTok
		}
	}

	// GPT-4o (Input: $2.50/MTok, Output: $10/MTok) - 2024年底降价后
	if strings.Contains(lowerModel, "gpt-4o") && !strings.Contains(lowerModel, "mini") {
		return &PricingEntry{
			InputCostPerToken:  0.0000025, // $2.50/MTok
			OutputCostPerToken: 0.00001,   // $10/MTok
		}
	}

	// GPT-4o-mini (Input: $0.15/MTok, Output: $0.60/MTok)
	if strings.Contains(lowerModel, "gpt-4o-mini") || strings.Contains(lowerModel, "gpt4o-mini") {
		return &PricingEntry{
			InputCostPerToken:  0.00000015, // $0.15/MTok
			OutputCostPerToken: 0.0000006,  // $0.60/MTok
		}
	}

	// o1 (Input: $15/MTok, Output: $60/MTok)
	if (lowerModel == "o1" || strings.HasPrefix(lowerModel, "o1-") || strings.Contains(lowerModel, "/o1")) &&
		!strings.Contains(lowerModel, "mini") && !strings.Contains(lowerModel, "pro") {
		return &PricingEntry{
			InputCostPerToken:  0.000015, // $15/MTok
			OutputCostPerToken: 0.00006,  // $60/MTok
		}
	}

	// o1-mini (Input: $3/MTok, Output: $12/MTok)
	if strings.Contains(lowerModel, "o1-mini") || strings.Contains(lowerModel, "o1mini") {
		return &PricingEntry{
			InputCostPerToken:  0.000003, // $3/MTok
			OutputCostPerToken: 0.000012, // $12/MTok
		}
	}

	// o1-pro (Input: $150/MTok, Output: $600/MTok)
	if strings.Contains(lowerModel, "o1-pro") || strings.Contains(lowerModel, "o1pro") {
		return &PricingEntry{
			InputCostPerToken:  0.00015, // $150/MTok
			OutputCostPerToken: 0.0006,  // $600/MTok
		}
	}

	// o3 (Input: $10/MTok, Output: $40/MTok) - 估算
	if (lowerModel == "o3" || strings.HasPrefix(lowerModel, "o3-") || strings.Contains(lowerModel, "/o3")) &&
		!strings.Contains(lowerModel, "mini") {
		return &PricingEntry{
			InputCostPerToken:  0.00001, // $10/MTok
			OutputCostPerToken: 0.00004, // $40/MTok
		}
	}

	// o3-mini (Input: $1.10/MTok, Output: $4.40/MTok)
	if strings.Contains(lowerModel, "o3-mini") || strings.Contains(lowerModel, "o3mini") {
		return &PricingEntry{
			InputCostPerToken:  0.0000011, // $1.10/MTok
			OutputCostPerToken: 0.0000044, // $4.40/MTok
		}
	}

	// ==================== Google Gemini 系列 ====================

	// Gemini 3 Pro (Input: $2/MTok, Output: $12/MTok)
	if strings.Contains(lowerModel, "gemini-3-pro") || strings.Contains(lowerModel, "gemini3pro") ||
		strings.Contains(lowerModel, "gemini-3.0-pro") {
		return &PricingEntry{
			InputCostPerToken:  0.000002, // $2/MTok
			OutputCostPerToken: 0.000012, // $12/MTok
		}
	}

	// Gemini 3 Flash (Input: $0.50/MTok, Output: $3/MTok)
	if strings.Contains(lowerModel, "gemini-3-flash") || strings.Contains(lowerModel, "gemini3flash") ||
		strings.Contains(lowerModel, "gemini-3.0-flash") {
		return &PricingEntry{
			InputCostPerToken:  0.0000005, // $0.50/MTok
			OutputCostPerToken: 0.000003,  // $3/MTok
		}
	}

	// Gemini 2.5 Pro (Input: $1.25/MTok, Output: $10/MTok)
	if strings.Contains(lowerModel, "gemini-2.5-pro") || strings.Contains(lowerModel, "gemini2.5pro") ||
		strings.Contains(lowerModel, "gemini-2-5-pro") {
		return &PricingEntry{
			InputCostPerToken:  0.00000125, // $1.25/MTok
			OutputCostPerToken: 0.00001,    // $10/MTok
		}
	}

	// Gemini 2.5 Flash (Input: $0.30/MTok, Output: $2.50/MTok)
	if strings.Contains(lowerModel, "gemini-2.5-flash") || strings.Contains(lowerModel, "gemini2.5flash") ||
		strings.Contains(lowerModel, "gemini-2-5-flash") {
		return &PricingEntry{
			InputCostPerToken:  0.0000003,  // $0.30/MTok
			OutputCostPerToken: 0.0000025,  // $2.50/MTok
		}
	}

	// Gemini 2.0 Flash (Input: $0.10/MTok, Output: $0.40/MTok)
	if strings.Contains(lowerModel, "gemini-2.0-flash") || strings.Contains(lowerModel, "gemini2.0flash") ||
		strings.Contains(lowerModel, "gemini-2-0-flash") {
		return &PricingEntry{
			InputCostPerToken:  0.0000001, // $0.10/MTok
			OutputCostPerToken: 0.0000004, // $0.40/MTok
		}
	}

	// Gemini 1.5 Pro (Input: $1.25/MTok, Output: $5/MTok)
	if strings.Contains(lowerModel, "gemini-1.5-pro") || strings.Contains(lowerModel, "gemini1.5pro") ||
		strings.Contains(lowerModel, "gemini-1-5-pro") {
		return &PricingEntry{
			InputCostPerToken:  0.00000125, // $1.25/MTok
			OutputCostPerToken: 0.000005,   // $5/MTok
		}
	}

	// Gemini 1.5 Flash (Input: $0.075/MTok, Output: $0.30/MTok)
	if strings.Contains(lowerModel, "gemini-1.5-flash") || strings.Contains(lowerModel, "gemini1.5flash") ||
		strings.Contains(lowerModel, "gemini-1-5-flash") {
		return &PricingEntry{
			InputCostPerToken:  0.000000075, // $0.075/MTok
			OutputCostPerToken: 0.0000003,   // $0.30/MTok
		}
	}

	// ==================== DeepSeek 系列 ====================

	// DeepSeek-V3 / DeepSeek-Chat (Input: $0.28/MTok, Output: $0.42/MTok)
	if strings.Contains(lowerModel, "deepseek-v3") || strings.Contains(lowerModel, "deepseek-chat") ||
		strings.Contains(lowerModel, "deepseekv3") || strings.Contains(lowerModel, "deepseekchat") {
		return &PricingEntry{
			InputCostPerToken:           0.00000028, // $0.28/MTok
			OutputCostPerToken:          0.00000042, // $0.42/MTok
			CacheReadInputTokenCost:     0.000000028, // $0.028/MTok (cache hit)
		}
	}

	// DeepSeek-R1 / DeepSeek-Reasoner (Input: $0.55/MTok, Output: $2.19/MTok)
	if strings.Contains(lowerModel, "deepseek-r1") || strings.Contains(lowerModel, "deepseek-reasoner") ||
		strings.Contains(lowerModel, "deepseekr1") || strings.Contains(lowerModel, "deepseekreasoner") {
		return &PricingEntry{
			InputCostPerToken:           0.00000055, // $0.55/MTok
			OutputCostPerToken:          0.00000219, // $2.19/MTok
			CacheReadInputTokenCost:     0.000000055, // $0.055/MTok (cache hit)
		}
	}

	// DeepSeek-Coder (Input: $0.14/MTok, Output: $0.28/MTok)
	if strings.Contains(lowerModel, "deepseek-coder") || strings.Contains(lowerModel, "deepseekcoder") {
		return &PricingEntry{
			InputCostPerToken:  0.00000014, // $0.14/MTok
			OutputCostPerToken: 0.00000028, // $0.28/MTok
		}
	}

	// ==================== 阿里 Qwen 系列 (人民币转美元，按 7.2 汇率) ====================

	// Qwen-Max (Input: ¥0.0032/1K ≈ $0.44/MTok, Output: ¥0.0128/1K ≈ $1.78/MTok)
	if strings.Contains(lowerModel, "qwen-max") || strings.Contains(lowerModel, "qwenmax") {
		return &PricingEntry{
			InputCostPerToken:  0.00000044, // ~$0.44/MTok
			OutputCostPerToken: 0.00000178, // ~$1.78/MTok
		}
	}

	// Qwen-Plus (Input: ¥0.0008/1K ≈ $0.11/MTok, Output: ¥0.002/1K ≈ $0.28/MTok)
	if strings.Contains(lowerModel, "qwen-plus") || strings.Contains(lowerModel, "qwenplus") {
		return &PricingEntry{
			InputCostPerToken:  0.00000011, // ~$0.11/MTok
			OutputCostPerToken: 0.00000028, // ~$0.28/MTok
		}
	}

	// Qwen-Turbo / Qwen-Flash (Input: ¥0.00015/1K ≈ $0.02/MTok, Output: ¥0.0015/1K ≈ $0.21/MTok)
	if strings.Contains(lowerModel, "qwen-turbo") || strings.Contains(lowerModel, "qwen-flash") ||
		strings.Contains(lowerModel, "qwenturbo") || strings.Contains(lowerModel, "qwenflash") {
		return &PricingEntry{
			InputCostPerToken:  0.00000002, // ~$0.02/MTok
			OutputCostPerToken: 0.00000021, // ~$0.21/MTok
		}
	}

	// ==================== 智谱 GLM 系列 (人民币转美元，按 7.2 汇率) ====================

	// GLM-4 / GLM-4-Plus (Input: ¥0.05/1K ≈ $6.94/MTok, Output: ¥0.05/1K ≈ $6.94/MTok)
	if strings.Contains(lowerModel, "glm-4-plus") || strings.Contains(lowerModel, "glm4plus") ||
		strings.Contains(lowerModel, "glm-4plus") {
		return &PricingEntry{
			InputCostPerToken:  0.00000694, // ~$6.94/MTok
			OutputCostPerToken: 0.00000694, // ~$6.94/MTok
		}
	}

	// GLM-4 (Input: ¥0.1/1K ≈ $13.89/MTok, Output: ¥0.1/1K ≈ $13.89/MTok)
	if (strings.Contains(lowerModel, "glm-4") || strings.Contains(lowerModel, "glm4")) &&
		!strings.Contains(lowerModel, "plus") && !strings.Contains(lowerModel, "flash") &&
		!strings.Contains(lowerModel, "air") && !strings.Contains(lowerModel, "v") {
		return &PricingEntry{
			InputCostPerToken:  0.00001389, // ~$13.89/MTok
			OutputCostPerToken: 0.00001389, // ~$13.89/MTok
		}
	}

	// GLM-4-Flash / GLM-4-Air (免费或极低价)
	if strings.Contains(lowerModel, "glm-4-flash") || strings.Contains(lowerModel, "glm-4-air") ||
		strings.Contains(lowerModel, "glm4flash") || strings.Contains(lowerModel, "glm4air") {
		return &PricingEntry{
			InputCostPerToken:  0.000000014, // ~$0.014/MTok (几乎免费)
			OutputCostPerToken: 0.000000014, // ~$0.014/MTok
		}
	}

	// GLM-4.5 / GLM-4-5 系列
	if strings.Contains(lowerModel, "glm-4.5") || strings.Contains(lowerModel, "glm-4-5") ||
		strings.Contains(lowerModel, "glm4.5") || strings.Contains(lowerModel, "glm45") {
		return &PricingEntry{
			InputCostPerToken:  0.00000694, // ~$6.94/MTok (估算)
			OutputCostPerToken: 0.00000694, // ~$6.94/MTok
		}
	}

	// ==================== Mistral 系列 ====================

	// Mistral Large (Input: $2/MTok, Output: $6/MTok)
	if strings.Contains(lowerModel, "mistral-large") || strings.Contains(lowerModel, "mistrallarge") {
		return &PricingEntry{
			InputCostPerToken:  0.000002, // $2/MTok
			OutputCostPerToken: 0.000006, // $6/MTok
		}
	}

	// Mistral Medium (Input: $2.7/MTok, Output: $8.1/MTok)
	if strings.Contains(lowerModel, "mistral-medium") || strings.Contains(lowerModel, "mistralmedium") {
		return &PricingEntry{
			InputCostPerToken:  0.0000027, // $2.7/MTok
			OutputCostPerToken: 0.0000081, // $8.1/MTok
		}
	}

	// Mistral Small (Input: $0.2/MTok, Output: $0.6/MTok)
	if strings.Contains(lowerModel, "mistral-small") || strings.Contains(lowerModel, "mistralsmall") {
		return &PricingEntry{
			InputCostPerToken:  0.0000002, // $0.2/MTok
			OutputCostPerToken: 0.0000006, // $0.6/MTok
		}
	}

	// Codestral (Input: $0.2/MTok, Output: $0.6/MTok)
	if strings.Contains(lowerModel, "codestral") {
		return &PricingEntry{
			InputCostPerToken:  0.0000002, // $0.2/MTok
			OutputCostPerToken: 0.0000006, // $0.6/MTok
		}
	}

	// ==================== Meta Llama 系列 (通过云服务商) ====================

	// Llama 3.1 405B (Input: $3/MTok, Output: $3/MTok via Together/Fireworks)
	if strings.Contains(lowerModel, "llama-3.1-405b") || strings.Contains(lowerModel, "llama3.1-405b") ||
		strings.Contains(lowerModel, "llama-3-1-405b") {
		return &PricingEntry{
			InputCostPerToken:  0.000003, // $3/MTok
			OutputCostPerToken: 0.000003, // $3/MTok
		}
	}

	// Llama 3.1 70B (Input: $0.88/MTok, Output: $0.88/MTok)
	if strings.Contains(lowerModel, "llama-3.1-70b") || strings.Contains(lowerModel, "llama3.1-70b") ||
		strings.Contains(lowerModel, "llama-3-1-70b") {
		return &PricingEntry{
			InputCostPerToken:  0.00000088, // $0.88/MTok
			OutputCostPerToken: 0.00000088, // $0.88/MTok
		}
	}

	// Llama 3.1 8B (Input: $0.18/MTok, Output: $0.18/MTok)
	if strings.Contains(lowerModel, "llama-3.1-8b") || strings.Contains(lowerModel, "llama3.1-8b") ||
		strings.Contains(lowerModel, "llama-3-1-8b") {
		return &PricingEntry{
			InputCostPerToken:  0.00000018, // $0.18/MTok
			OutputCostPerToken: 0.00000018, // $0.18/MTok
		}
	}

	return nil
}

func (s *Service) longContextTier(model string, usage UsageSnapshot) (LongContextPricing, bool) {
	totalInput := usage.InputTokens + usage.CacheCreateTokens + usage.CacheReadTokens
	if strings.Contains(strings.ToLower(model), "[1m]") && totalInput > 200000 && len(s.longContexts) > 0 {
		if tier, ok := s.longContexts[model]; ok {
			return tier, true
		}
		for _, tier := range s.longContexts {
			return tier, true
		}
	}
	return LongContextPricing{}, false
}

func (s *Service) getEphemeral1hPricing(model string) float64 {
	if price, ok := s.ephemeral1h[model]; ok {
		return price
	}
	name := strings.ToLower(model)
	switch {
	case strings.Contains(name, "opus"):
		return 0.00003
	case strings.Contains(name, "sonnet"):
		return 0.000006
	case strings.Contains(name, "haiku"):
		return 0.0000016
	default:
		return 0
	}
}

func ensureCachePricing(entry *PricingEntry) {
	if entry == nil {
		return
	}
	if entry.CacheCreationInputTokenCost == 0 && entry.InputCostPerToken > 0 {
		entry.CacheCreationInputTokenCost = entry.InputCostPerToken * 1.25
	}
	if entry.CacheReadInputTokenCost == 0 && entry.InputCostPerToken > 0 {
		entry.CacheReadInputTokenCost = entry.InputCostPerToken * 0.1
	}
}

func stripRegionPrefix(name string) string {
	for _, prefix := range []string{"us.", "eu.", "apac."} {
		if strings.HasPrefix(strings.ToLower(name), prefix) {
			return name[len(prefix):]
		}
	}
	return name
}

func normalizeName(name string) string {
	return nameReplacer.Replace(strings.ToLower(name))
}

func resolveCacheTokens(usage UsageSnapshot) (fiveMin int, oneHour int) {
	if usage.CacheCreation == nil {
		return usage.CacheCreateTokens, 0
	}
	five := usage.CacheCreation.Ephemeral5mTokens
	one := usage.CacheCreation.Ephemeral1hTokens
	remaining := usage.CacheCreateTokens - five - one
	if remaining > 0 {
		five += remaining
	}
	if five < 0 {
		five = 0
	}
	if one < 0 {
		one = 0
	}
	return five, one
}

func buildEphemeral1hPricing() map[string]float64 {
	return map[string]float64{
		"claude-opus-4-1":            0.00003,
		"claude-opus-4-1-20250805":   0.00003,
		"claude-opus-4":              0.00003,
		"claude-opus-4-20250514":     0.00003,
		"claude-3-opus":              0.00003,
		"claude-3-opus-latest":       0.00003,
		"claude-3-opus-20240229":     0.00003,
		"claude-3-5-sonnet":          0.000006,
		"claude-3-5-sonnet-latest":   0.000006,
		"claude-3-5-sonnet-20241022": 0.000006,
		"claude-3-5-sonnet-20240620": 0.000006,
		"claude-3-sonnet":            0.000006,
		"claude-3-sonnet-20240307":   0.000006,
		"claude-sonnet-3":            0.000006,
		"claude-sonnet-3-5":          0.000006,
		"claude-sonnet-3-7":          0.000006,
		"claude-sonnet-4":            0.000006,
		"claude-sonnet-4-20250514":   0.000006,
		"claude-3-5-haiku":           0.0000016,
		"claude-3-5-haiku-latest":    0.0000016,
		"claude-3-5-haiku-20241022":  0.0000016,
		"claude-3-haiku":             0.0000016,
		"claude-3-haiku-20240307":    0.0000016,
		"claude-haiku-3":             0.0000016,
		"claude-haiku-3-5":           0.0000016,
	}
}

func buildLongContextPricing() map[string]LongContextPricing {
	return map[string]LongContextPricing{
		"claude-sonnet-4-20250514[1m]": {
			Input:  0.000006,
			Output: 0.0000225,
		},
	}
}
