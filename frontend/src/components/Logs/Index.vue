<template>
  <div class="logs-page">
    <PageHeader :title="t('components.main.logs.view')">
      <template #actions>
        <div class="header-right-actions">
          <BaseButton size="sm" :disabled="loading" @click="loadDashboard">
            {{ t('components.logs.refresh') }}
          </BaseButton>
        </div>
      </template>
    </PageHeader>

    <section class="logs-summary" v-if="statsCards.length">
      <article v-for="card in statsCards" :key="card.key" class="summary-card">
        <div class="summary-card__label">{{ card.label }}</div>
        <div class="summary-card__value">{{ card.value }}</div>
        <div class="summary-card__hint">{{ card.hint }}</div>
      </article>
    </section>

    <section class="logs-chart">
      <div class="chart-header">
        <div class="chart-legend">
          <button 
            v-for="ds in visibleDatasets" 
            :key="ds.label"
            class="legend-item"
            :class="{ active: !ds.hidden }"
            @click="toggleDataset(ds)"
          >
            <span class="dot" :style="{ backgroundColor: ds.borderColor }"></span>
            {{ translateProvider(ds.label) }}
          </button>
        </div>
      </div>
      <div class="chart-content">
        <Line :data="chartData" :options="chartOptions" />
      </div>
    </section>

    <form class="logs-filter-row" @submit.prevent="applyFilters">
      <div class="filter-fields">
        <label class="filter-field">
          <span>{{ t('components.logs.filters.platform') }}</span>
          <select v-model="filters.platform" class="mac-select">
            <option value="">{{ t('components.logs.filters.allPlatforms') }}</option>
            <option value="claude">Claude</option>
            <option value="codex">Codex</option>
          </select>
        </label>
        <label class="filter-field">
          <span>{{ t('components.logs.filters.provider') }}</span>
          <select v-model="filters.provider" class="mac-select">
            <option value="">{{ t('components.logs.filters.allProviders') }}</option>
            <option v-for="provider in providerOptions" :key="provider" :value="provider">
              {{ translateProvider(provider) }}
            </option>
          </select>
        </label>
      </div>
      <div class="filter-actions">
        <BaseButton type="submit" :disabled="loading">
          {{ t('components.logs.query') }}
        </BaseButton>
      </div>
    </form>

    <section class="logs-table-wrapper">
      <table class="logs-table">
        <thead>
          <tr>
            <th class="col-time">{{ t('components.logs.table.time') }}</th>
            <th class="col-platform">{{ t('components.logs.table.platform') }}</th>
            <th class="col-provider">{{ t('components.logs.table.provider') }}</th>
            <th class="col-model">{{ t('components.logs.table.model') }}</th>
            <th class="col-http">{{ t('components.logs.table.httpCode') }}</th>
            <th class="col-stream">{{ t('components.logs.table.stream') }}</th>
            <th class="col-duration">{{ t('components.logs.table.duration') }}</th>
            <th class="col-tokens">{{ t('components.logs.table.tokens') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in pagedLogs" :key="item.id">
            <td>{{ formatTime(item.created_at) }}</td>
            <td>{{ item.platform || '—' }}</td>
            <td>{{ translateProvider(item.provider) || '—' }}</td>
            <td>{{ item.model || '—' }}</td>
            <td :class="['code', httpCodeClass(item.http_code)]">{{ item.http_code }}</td>
            <td><span :class="['stream-tag', item.is_stream ? 'on' : 'off']">{{ formatStream(item.is_stream) }}</span></td>
            <td><span :class="['duration-tag', durationColor(item.duration_sec)]">{{ formatDuration(item.duration_sec) }}</span></td>
            <td class="token-cell">
              <div>
                <span class="token-label">{{ t('components.logs.tokenLabels.input') }}</span>
                <span class="token-value">{{ formatNumber(item.input_tokens) }}</span>
              </div>
              <div>
                <span class="token-label">{{ t('components.logs.tokenLabels.output') }}</span>
                <span class="token-value">{{ formatNumber(item.output_tokens) }}</span>
              </div>
              <div>
                <span class="token-label">{{ t('components.logs.tokenLabels.reasoning') }}</span>
                <span class="token-value">{{ formatNumber(item.reasoning_tokens) }}</span>
              </div>
              <div>
                <span class="token-label">{{ t('components.logs.tokenLabels.cacheWrite') }}</span>
                <span class="token-value">{{ formatNumber(item.cache_create_tokens) }}</span>
              </div>
              <div>
                <span class="token-label">{{ t('components.logs.tokenLabels.cacheRead') }}</span>
                <span class="token-value">{{ formatNumber(item.cache_read_tokens) }}</span>
              </div>
            </td>
          </tr>
          <tr v-if="!pagedLogs.length && !loading">
            <td colspan="8" class="empty">{{ t('components.logs.empty') }}</td>
          </tr>
        </tbody>
      </table>
      <p v-if="loading" class="empty">{{ t('components.logs.loading') }}</p>
    </section>

    <div class="logs-pagination">
      <BaseButton variant="outline" size="sm" :disabled="page === 1 || loading" @click="prevPage">
        ‹
      </BaseButton>
      <span class="pagination-info">{{ page }} / {{ totalPages }}</span>
      <BaseButton variant="outline" size="sm" :disabled="page >= totalPages || loading" @click="nextPage">
        ›
      </BaseButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '../Navigation/PageHeader.vue'
import { useI18n } from 'vue-i18n'
import BaseButton from '../common/BaseButton.vue'
import {
  fetchRequestLogs,
  fetchLogProviders,
  fetchLogStats,
  type RequestLog,
  type LogStats,
  type LogStatsSeries,
} from '../../services/logs'
import {
  Chart,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
} from 'chart.js'
import type { ChartOptions } from 'chart.js'
import { Line } from 'vue-chartjs'

Chart.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend)

const visibleDatasets = ref<any[]>([])

const toggleDataset = (ds: any) => {
  ds.hidden = !ds.hidden
  // 强制触发图表更新
  chartData.value.datasets = [...chartData.value.datasets]
}

const { t, te } = useI18n()

const translateProvider = (name: string) => {
  if (!name) return name
  const key = `components.main.providers.names.${name.trim()}`
  return te(key) ? t(key) : name
}
const router = useRouter()

const logs = ref<RequestLog[]>([])
const stats = ref<LogStats | null>(null)
const loading = ref(false)
const filters = reactive({ platform: '', provider: '' })
const page = ref(1)
const PAGE_SIZE = 15
const providerOptions = ref<string[]>([])
const statsSeries = computed<LogStatsSeries[]>(() => stats.value?.series ?? [])

const isBrowser = typeof window !== 'undefined' && typeof document !== 'undefined'
const readDarkMode = () => (isBrowser ? document.documentElement.classList.contains('dark') : false)
const isDarkMode = ref(readDarkMode())
let themeObserver: MutationObserver | null = null

const getCssVarValue = (name: string, fallback: string) => {
  if (!isBrowser) return fallback
  const value = getComputedStyle(document.documentElement).getPropertyValue(name)
  return value?.trim() || fallback
}

const syncThemeState = () => {
  isDarkMode.value = readDarkMode()
}

const setupThemeObserver = () => {
  if (!isBrowser || themeObserver) return
  syncThemeState()
  themeObserver = new MutationObserver((mutations) => {
    if (mutations.some((mutation) => mutation.attributeName === 'class')) {
      syncThemeState()
    }
  })
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class'],
  })
}

const teardownThemeObserver = () => {
  if (!themeObserver) return
  themeObserver.disconnect()
  themeObserver = null
}

const parseLogDate = (value?: string) => {
  if (!value) return null
  
  // 提取日期时间部分（去掉任何时区后缀）
  // 后端返回的格式可能是：
  // - "2025-12-19 23:49:00" (纯本地时间)
  // - "2025-12-19 23:49:00 +0000 UTC" (xdb 自动添加的 UTC 后缀，但实际是本地时间)
  let dateTimePart = value.trim()
  
  // 移除 " +0000 UTC" 或类似的时区后缀
  const plusIdx = dateTimePart.indexOf(' +')
  if (plusIdx > 0) {
    dateTimePart = dateTimePart.substring(0, plusIdx)
  } else {
    // 检查是否有 " -" 时区偏移（但要确保不是日期中的 "-"）
    const minusIdx = dateTimePart.lastIndexOf(' -')
    if (minusIdx > 10) {
      dateTimePart = dateTimePart.substring(0, minusIdx)
    }
  }
  
  // 处理标准格式 "2025-12-19 23:49:44"（当作本地时间）
  const localMatch = dateTimePart.match(/^(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2})$/)
  if (localMatch) {
    const [, day, time] = localMatch
    // 使用 new Date(year, month, day, hour, min, sec) 构造本地时间
    const [year, month, dayNum] = day.split('-').map(Number)
    const [hour, min, sec] = time.split(':').map(Number)
    return new Date(year, month - 1, dayNum, hour, min, sec)
  }
  
  // 尝试只有日期的格式
  const dateOnlyMatch = dateTimePart.match(/^(\d{4}-\d{2}-\d{2})$/)
  if (dateOnlyMatch) {
    const [year, month, dayNum] = dateTimePart.split('-').map(Number)
    return new Date(year, month - 1, dayNum, 0, 0, 0)
  }
  
  return null
}

const chartData = computed(() => {
  const series = statsSeries.value
  return {
    labels: series.map((item) => formatSeriesLabel(item.day)),
    datasets: [
      {
        label: t('components.logs.tokenLabels.cost'),
        data: series.map((item) => Number(((item.total_cost ?? 0)).toFixed(4))),
        borderColor: '#f97316',
        backgroundColor: 'rgba(249, 115, 22, 0.2)',
        tension: 0.3,
        fill: false,
        yAxisID: 'yCost',
      },
      {
        label: t('components.logs.tokenLabels.input'),
        data: series.map((item) => item.input_tokens ?? 0),
        borderColor: '#34d399',
        backgroundColor: 'rgba(52, 211, 153, 0.25)',
        tension: 0.35,
        fill: true,
      },
      {
        label: t('components.logs.tokenLabels.output'),
        data: series.map((item) => item.output_tokens ?? 0),
        borderColor: '#60a5fa',
        backgroundColor: 'rgba(96, 165, 250, 0.2)',
        tension: 0.35,
        fill: true,
      },
      {
        label: t('components.logs.tokenLabels.reasoning'),
        data: series.map((item) => item.reasoning_tokens ?? 0),
        borderColor: '#f472b6',
        backgroundColor: 'rgba(244, 114, 182, 0.2)',
        tension: 0.35,
        fill: true,
      },
      {
        label: t('components.logs.tokenLabels.cacheWrite'),
        data: series.map((item) => item.cache_create_tokens ?? 0),
        borderColor: '#fbbf24',
        backgroundColor: 'rgba(251, 191, 36, 0.2)',
        tension: 0.35,
        fill: false,
      },
      {
        label: t('components.logs.tokenLabels.cacheRead'),
        data: series.map((item) => item.cache_read_tokens ?? 0),
        borderColor: '#38bdf8',
        backgroundColor: 'rgba(56, 189, 248, 0.15)',
        tension: 0.35,
        fill: false,
      },
    ],
  }
})

watch(chartData, (newData) => {
  if (newData.datasets) {
    visibleDatasets.value = newData.datasets
  }
}, { immediate: true })

const chartOptions = computed<ChartOptions<'line'>>(() => {
  const legendColor = getCssVarValue('--mac-text', isDarkMode.value ? '#f8fafc' : '#0f172a')
  const axisColor = getCssVarValue(
    '--mac-text-secondary',
    isDarkMode.value ? '#cbd5f5' : '#94a3b8',
  )
  const axisStrongColor = getCssVarValue('--mac-text', isDarkMode.value ? '#e2e8f0' : '#475569')
  const gridColor = isDarkMode.value ? 'rgba(148, 163, 184, 0.35)' : 'rgba(148, 163, 184, 0.2)'

  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      mode: 'index',
      intersect: false,
    },
    plugins: {
      legend: {
        display: false,
      },
    },
    scales: {
      x: {
        grid: { display: false },
        ticks: { 
          color: axisColor,
          maxRotation: 0,
          autoSkip: true,
          maxTicksLimit: 15,
          font: { size: 10 }
        },
      },
      y: {
        beginAtZero: true,
        ticks: { color: axisColor },
        grid: { color: gridColor },
      },
      yCost: {
        position: 'right',
        beginAtZero: true,
        grid: { drawOnChartArea: false },
        ticks: {
          color: axisStrongColor,
          callback: (value: string | number) => {
            const numeric = typeof value === 'number' ? value : Number(value)
            if (Number.isNaN(numeric)) return '$0'
            if (numeric >= 1) return `$${numeric.toFixed(2)}`
            return `$${numeric.toFixed(4)}`
          },
        },
      },
    },
  }
})
const formatSeriesLabel = (value?: string) => {
  if (!value) return ''
  const parsed = parseLogDate(value)
  if (parsed) {
    return `${padHour(parsed.getHours())}:00`
  }
  const match = value.match(/(\d{2}):(\d{2})/)
  if (match) {
    return `${match[1]}:${match[2]}`
  }
  return value
}


const loadLogs = async () => {
  loading.value = true
  try {
    const data = await fetchRequestLogs({
      platform: filters.platform,
      provider: filters.provider,
      limit: 200,
    })
    logs.value = data ?? []
    page.value = Math.min(page.value, totalPages.value)
  } catch (error) {
    console.error('failed to load request logs', error)
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const data = await fetchLogStats(filters.platform)
    stats.value = data ?? null
  } catch (error) {
    console.error('failed to load log stats', error)
  }
}

const loadDashboard = async () => {
  await Promise.all([loadLogs(), loadStats()])
}

const pagedLogs = computed(() => {
  const start = (page.value - 1) * PAGE_SIZE
  return logs.value.slice(start, start + PAGE_SIZE)
})

const totalPages = computed(() => Math.max(1, Math.ceil(logs.value.length / PAGE_SIZE)))

const applyFilters = async () => {
  page.value = 1
  await loadDashboard()
}


const nextPage = () => {
  if (page.value < totalPages.value) {
    page.value += 1
  }
}

const prevPage = () => {
  if (page.value > 1) {
    page.value -= 1
  }
}


const padHour = (num: number) => num.toString().padStart(2, '0')

const formatTime = (value?: string) => {
  const date = parseLogDate(value)
  if (!date) return value || '—'
  return `${date.getFullYear()}-${padHour(date.getMonth() + 1)}-${padHour(date.getDate())} ${padHour(date.getHours())}:${padHour(date.getMinutes())}:${padHour(date.getSeconds())}`
}

const formatStream = (value?: boolean | number) => {
  const isOn = value === true || value === 1
  return isOn ? t('components.logs.streamOn') : t('components.logs.streamOff')
}

const formatDuration = (value?: number) => {
  if (!value || Number.isNaN(value)) return '—'
  return `${value.toFixed(2)}s`
}

const httpCodeClass = (code: number) => {
  if (code >= 500) return 'http-server-error'
  if (code >= 400) return 'http-client-error'
  if (code >= 300) return 'http-redirect'
  if (code >= 200) return 'http-success'
  return 'http-info'
}

const durationColor = (value?: number) => {
  if (!value || Number.isNaN(value)) return 'neutral'
  if (value < 2) return 'fast'
  if (value < 5) return 'medium'
  return 'slow'
}

const formatNumber = (value?: number) => {
  if (value === undefined || value === null) return '—'
  return value.toLocaleString()
}

const formatCurrency = (value?: number) => {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return '$0.0000'
  }
  if (value >= 1) {
    return `$${value.toFixed(2)}`
  }
  if (value >= 0.01) {
    return `$${value.toFixed(3)}`
  }
  return `$${value.toFixed(4)}`
}

const startOfTodayLocal = () => {
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  return now
}

const statsCards = computed(() => {
  const data = stats.value
  const summaryDate = summaryDateLabel.value
  const totalTokens =
    (data?.input_tokens ?? 0) + (data?.output_tokens ?? 0) + (data?.reasoning_tokens ?? 0)
  return [
    {
      key: 'requests',
      label: t('components.logs.summary.total'),
      hint: t('components.logs.summary.requests'),
      value: data ? formatNumber(data.total_requests) : '—',
    },
    {
      key: 'tokens',
      label: t('components.logs.summary.tokens'),
      hint: t('components.logs.summary.tokenHint'),
      value: data ? formatNumber(totalTokens) : '—',
    },
    {
      key: 'cacheReads',
      label: t('components.logs.summary.cache'),
      hint: t('components.logs.summary.cacheHint'),
      value: data ? formatNumber(data.cache_read_tokens) : '—',
    },
    {
      key: 'cost',
      label: t('components.logs.tokenLabels.cost'),
      hint: summaryDate ? t('components.logs.summary.todayScope', { date: summaryDate }) : '',
      value: formatCurrency(data?.cost_total ?? 0),
    },
  ]
})

const summaryDateLabel = computed(() => {
  const firstBucket = statsSeries.value.find((item) => item.day)
  const parsed = parseLogDate(firstBucket?.day ?? '')
  const date = parsed ?? startOfTodayLocal()
  return `${date.getFullYear()}-${padHour(date.getMonth() + 1)}-${padHour(date.getDate())}`
})

const loadProviderOptions = async () => {
  try {
    const list = await fetchLogProviders(filters.platform)
    providerOptions.value = list ?? []
    if (filters.provider && !providerOptions.value.includes(filters.provider)) {
      filters.provider = ''
    }
  } catch (error) {
    console.error('failed to load provider options', error)
  }
}

watch(
  () => filters.platform,
  async () => {
    await loadProviderOptions()
  },
)

onMounted(async () => {
  await Promise.all([loadDashboard(), loadProviderOptions()])
  setupThemeObserver()
})

onUnmounted(() => {
  teardownThemeObserver()
})
</script>

<style scoped>
.logs-page {
  max-width: 100%;
}
.logs-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.header-right-actions {
  display: flex;
  align-items: center;
}

.summary-meta {
  grid-column: 1 / -1;
  font-size: 0.85rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #64748b;
}

.summary-card {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 16px;
  padding: 1rem 1.25rem;
  background: radial-gradient(circle at top, rgba(148, 163, 184, 0.1), rgba(15, 23, 42, 0));
  backdrop-filter: blur(6px);
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.summary-card__label {
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #475569;
}

.summary-card__value {
  font-size: 1.85rem;
  font-weight: 600;
  color: #0f172a;
}

.summary-card__hint {
  font-size: 0.85rem;
  color: #94a3b8;
}

html.dark .summary-card {
  border-color: rgba(255, 255, 255, 0.12);
  background: radial-gradient(circle at top, rgba(148, 163, 184, 0.2), rgba(15, 23, 42, 0.35));
}

html.dark .summary-card__label {
  color: rgba(248, 250, 252, 0.75);
}

html.dark .summary-card__value {
  color: rgba(248, 250, 252, 0.95);
}

html.dark .summary-card__hint {
  color: rgba(186, 194, 210, 0.8);
}

.logs-chart {
  margin-bottom: 2rem;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(15, 23, 42, 0.05);
  border-radius: 16px;
  padding: 1.5rem 1rem;
  height: auto;
  min-height: 440px;
  display: flex;
  flex-direction: column;
}

html.dark .logs-chart {
  border-color: rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.2);
}

.chart-header {
  display: flex;
  justify-content: center;
  margin-bottom: 1.5rem;
  width: 100%;
  padding: 0 10px;
  box-sizing: border-box;
}

.chart-legend {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 0.85rem;
  line-height: 1;
  color: var(--mac-text-secondary, #64748b);
  cursor: pointer;
  background: rgba(148, 163, 184, 0.08);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 20px;
  padding: 8px 16px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  flex: 0 0 auto;
  min-width: max-content;
  box-sizing: border-box;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.02);
}

.legend-item:hover {
  background: rgba(148, 163, 184, 0.1);
  border-color: rgba(148, 163, 184, 0.2);
  transform: translateY(-1px);
}

.legend-item.active {
  color: var(--mac-text, #0f172a);
  background: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  border-color: rgba(148, 163, 184, 0.3);
}

html.dark .legend-item {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.08);
  color: #94a3b8;
}

html.dark .legend-item.active {
  background: rgba(255, 255, 255, 0.1);
  color: #f8fafc;
  border-color: rgba(255, 255, 255, 0.2);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.8);
}

html.dark .dot {
  box-shadow: 0 0 0 2px rgba(15, 23, 42, 0.5);
}

.chart-content {
  height: 300px;
  width: 100%;
}

@media (max-width: 768px) {
  .logs-summary {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  }
}

.logs-table-wrapper {
  overflow-x: auto !important;
  overflow-y: hidden;
  display: block;
  width: 100%;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.02);
  border: 1px solid rgba(15, 23, 42, 0.05);
  margin-top: 1rem;
  /* Override global scrollbar-width: none */
  scrollbar-width: thin !important;
  scrollbar-color: var(--mac-accent, #3b82f6) rgba(148, 163, 184, 0.1);
  -ms-overflow-style: auto !important;
}

html.dark .logs-table-wrapper {
  background: rgba(15, 23, 42, 0.2);
  border-color: rgba(255, 255, 255, 0.08);
  scrollbar-color: rgba(255, 255, 255, 0.3) rgba(255, 255, 255, 0.05);
}

.logs-table {
  width: 100%;
  min-width: 1200px;
  border-collapse: collapse;
}

/* Liquid Glass Scrollbar - WebKit browsers */
.logs-table-wrapper::-webkit-scrollbar {
  height: 8px;
  display: block !important;
}

.logs-table-wrapper::-webkit-scrollbar-track {
  background: rgba(148, 163, 184, 0.08);
  border-radius: 8px;
  margin: 0 4px;
}

.logs-table-wrapper::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, var(--mac-accent, #3b82f6) 0%, #60a5fa 100%);
  border-radius: 8px;
  border: 2px solid transparent;
  background-clip: padding-box;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.logs-table-wrapper::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #2563eb 0%, var(--mac-accent, #3b82f6) 100%);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.5);
}

html.dark .logs-table-wrapper::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.03);
}

html.dark .logs-table-wrapper::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.2) 0%, rgba(255, 255, 255, 0.35) 100%);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

html.dark .logs-table-wrapper::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.35) 0%, rgba(255, 255, 255, 0.5) 100%);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
}
</style>
