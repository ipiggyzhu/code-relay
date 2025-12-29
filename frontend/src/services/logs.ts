import { Call } from '@wailsio/runtime'

export type RequestLog = {
  id: number
  platform: string
  model: string
  provider: string
  http_code: number
  input_tokens: number
  output_tokens: number
  cache_create_tokens: number
  cache_read_tokens: number
  reasoning_tokens: number
  is_stream?: boolean | number
  duration_sec?: number
  created_at: string
  total_cost?: number
  input_cost?: number
  output_cost?: number
  cache_create_cost?: number
  cache_read_cost?: number
  ephemeral_5m_cost?: number
  ephemeral_1h_cost?: number
  has_pricing?: boolean
}

type RequestLogQuery = {
  platform?: string
  provider?: string
  limit?: number
}

export const fetchRequestLogs = async (query: RequestLogQuery = {}): Promise<RequestLog[]> => {
  const platform = query.platform ?? ''
  const provider = query.provider ?? ''
  const limit = query.limit ?? 100
  return Call.ByName('coderelay/services.LogService.ListRequestLogs', platform, provider, limit)
}

export const fetchLogProviders = async (platform = ''): Promise<string[]> => {
  return Call.ByName('coderelay/services.LogService.ListProviders', platform)
}

export type LogStatsSeries = {
  day: string
  total_requests: number
  input_tokens: number
  output_tokens: number
  reasoning_tokens: number
  cache_create_tokens: number
  cache_read_tokens: number
  total_cost: number
}

export type LogStats = {
  total_requests: number
  input_tokens: number
  output_tokens: number
  reasoning_tokens: number
  cache_create_tokens: number
  cache_read_tokens: number
  cost_total: number
  cost_input: number
  cost_output: number
  cost_cache_create: number
  cost_cache_read: number
  series: LogStatsSeries[]
}

export const fetchLogStats = async (platform = ''): Promise<LogStats> => {
  return Call.ByName('coderelay/services.LogService.StatsSince', platform)
}

export type ProviderDailyStat = {
  provider: string
  total_requests: number
  successful_requests: number
  failed_requests: number
  success_rate: number
  input_tokens: number
  output_tokens: number
  reasoning_tokens: number
  cache_create_tokens: number
  cache_read_tokens: number
  cost_total: number
  // 响应时间统计
  avg_duration_sec: number
  min_duration_sec: number
  max_duration_sec: number
}

export const fetchProviderDailyStats = async (
  platform = '',
): Promise<ProviderDailyStat[]> => {
  return Call.ByName('coderelay/services.LogService.ProviderDailyStats', platform)
}

export type HeatmapStat = {
  day: string
  total_requests: number
  input_tokens: number
  output_tokens: number
  reasoning_tokens: number
  total_cost: number
}

export const fetchHeatmapStats = async (days: number): Promise<HeatmapStat[]> => {
  const range = Number.isFinite(days) && days > 0 ? Math.floor(days) : 30
  return Call.ByName('coderelay/services.LogService.HeatmapStats', range)
}
