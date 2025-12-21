<template>
  <div class="usage-chart-container">
    <div class="chart-header">
      <div class="chart-legend">
        <button 
          v-for="metric in metrics" 
          :key="metric.key"
          class="legend-item"
          :class="{ active: activeMetrics.includes(metric.key) }"
          @click="toggleMetric(metric.key)"
        >
          <span class="dot" :style="{ backgroundColor: metric.color }"></span>
          {{ metric.label }}
        </button>
      </div>
    </div>
    <v-chart class="chart" :option="chartOption" autoresize />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, provide, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import VChart, { THEME_KEY } from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent
} from 'echarts/components'
import type { UsageHeatmapWeek } from '../../data/usageHeatmap'

use([
  CanvasRenderer,
  LineChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent
])
const props = defineProps<{
  data: UsageHeatmapWeek[]
}>()

const { t } = useI18n()

// 注入主题 key，根据 html.dark 类动态切换
const isDark = ref(document.documentElement.classList.contains('dark'))
const theme = computed(() => isDark.value ? 'dark' : 'light')
provide(THEME_KEY, theme)

// 监听主题变化
const observer = new MutationObserver(() => {
  isDark.value = document.documentElement.classList.contains('dark')
})
onMounted(() => {
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})
onUnmounted(() => {
  observer.disconnect()
})

const colors = {
  requests: '#3b82f6',
  inputTokens: '#10b981',
  outputTokens: '#f59e0b',
  reasoningTokens: '#8b5cf6',
  cost: '#ef4444'
}

const metrics = [
  { key: 'requests', label: t('components.main.heatmap.metrics.requests'), color: colors.requests },
  { key: 'inputTokens', label: t('components.main.heatmap.metrics.inputTokens'), color: colors.inputTokens },
  { key: 'outputTokens', label: t('components.main.heatmap.metrics.outputTokens'), color: colors.outputTokens },
  { key: 'reasoningTokens', label: t('components.main.heatmap.metrics.reasoningTokens'), color: colors.reasoningTokens },
  { key: 'cost', label: t('components.main.heatmap.metrics.cost'), color: colors.cost }
]

const activeMetrics = ref(['inputTokens', 'outputTokens', 'cost'])

const toggleMetric = (key: string) => {
  if (activeMetrics.value.includes(key)) {
    if (activeMetrics.value.length > 1) {
      activeMetrics.value = activeMetrics.value.filter(k => k !== key)
    }
  } else {
    activeMetrics.value.push(key)
  }
}

const flatData = computed(() => {
  // 生成完整的14天日期范围（从13天前到今天）
  const days = 14
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  
  const allDates: string[] = []
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(d.getDate() - i)
    const dateKey = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    allDates.push(dateKey)
  }

  // 从 props.data 中收集每天的数据
  const dailyMap = new Map<string, {
    date: string
    requests: number
    inputTokens: number
    outputTokens: number
    reasoningTokens: number
    cost: number
  }>()

  if (props.data && props.data.length > 0) {
    for (const week of props.data) {
      for (const day of week) {
        if (!day.dateKey) continue
        
        const d = new Date(day.dateKey)
        if (isNaN(d.getTime())) continue
        
        const dateKey = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
        
        const existing = dailyMap.get(dateKey)
        if (existing) {
          existing.requests += day.requests || 0
          existing.inputTokens += day.inputTokens || 0
          existing.outputTokens += day.outputTokens || 0
          existing.reasoningTokens += day.reasoningTokens || 0
          existing.cost += day.cost || 0
        } else {
          dailyMap.set(dateKey, {
            date: dateKey,
            requests: day.requests || 0,
            inputTokens: day.inputTokens || 0,
            outputTokens: day.outputTokens || 0,
            reasoningTokens: day.reasoningTokens || 0,
            cost: day.cost || 0
          })
        }
      }
    }
  }

  // 返回完整的14天数据，没有数据的天用0填充
  return allDates.map(dateKey => {
    const existing = dailyMap.get(dateKey)
    if (existing) {
      return existing
    }
    return {
      date: dateKey,
      requests: 0,
      inputTokens: 0,
      outputTokens: 0,
      reasoningTokens: 0,
      cost: 0
    }
  })
})

const costLabel = t('components.main.heatmap.metrics.cost')

const chartOption = computed(() => {
  const dates = flatData.value.map(d => d.date)
  
  const getYAxisIndex = (key: string) => {
    if (key === 'cost') return 1
    if (key === 'requests') return 2
    return 0
  }

  const series = metrics
    .filter(m => activeMetrics.value.includes(m.key))
    .map(m => ({
      name: m.label,
      type: 'line',
      data: flatData.value.map(d => d[m.key as keyof typeof d]),
      smooth: 0.3,
      showSymbol: true,
      symbol: 'circle',
      symbolSize: 8,
      yAxisIndex: getYAxisIndex(m.key),
      itemStyle: {
        color: m.color,
        borderColor: '#fff',
        borderWidth: 2
      },
      lineStyle: {
        width: 2.5
      },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: m.color + '40' },
            { offset: 1, color: m.color + '05' }
          ]
        }
      },
      emphasis: {
        focus: 'series',
        itemStyle: {
          borderWidth: 3,
          shadowBlur: 10,
          shadowColor: m.color + '80'
        }
      }
    }))

  return {
    tooltip: {
      show: true,
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
        snap: true,
        crossStyle: { color: '#d1d5db' },
        lineStyle: { color: '#d1d5db', type: 'dashed', width: 1 }
      },
      triggerOn: 'mousemove|click',
      confine: true,
      backgroundColor: 'rgba(255, 255, 255, 0.98)',
      borderColor: '#e5e7eb',
      borderWidth: 1,
      borderRadius: 8,
      textStyle: { color: '#374151', fontSize: 12 },
      padding: [10, 14],
      formatter: (params: any) => {
        // 处理单个 item 触发的情况
        const items = Array.isArray(params) ? params : [params]
        if (!items || items.length === 0) return ''
        const date = items[0].axisValue || items[0].name
        let html = '<div style="font-weight:500;margin-bottom:6px;color:#111827">' + date + '</div>'
        items.forEach((item: any) => {
          const isCost = item.seriesName?.includes(costLabel)
          const value = isCost
            ? '\u0024' + (item.value || 0).toFixed(4)
            : (item.value || 0).toLocaleString()
          html += '<div style="display:flex;align-items:center;gap:6px;margin:3px 0">'
          html += '<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:' + item.color + '"></span>'
          html += '<span style="color:#6b7280">' + (item.seriesName || '') + '</span>'
          html += '<span style="margin-left:auto;font-weight:500;color:#111827">' + value + '</span>'
          html += '</div>'
        })
        return html
      }
    },
    grid: {
      top: 30,
      right: 60,
      bottom: 30,
      left: 60,
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dates,
      boundaryGap: false,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: {
        color: '#9ca3af',
        fontSize: 11,
        formatter: (value: string) => {
          const parts = value.split('-')
          if (parts.length >= 3) {
            return parseInt(parts[1]) + '-' + parseInt(parts[2])
          }
          return value
        }
      },
      axisPointer: {
        label: {
          backgroundColor: '#6b7280'
        }
      }
    },
    yAxis: [
      {
        type: 'value',
        name: t('components.main.providers.tokens'),
        nameTextStyle: { color: '#9ca3af', padding: [0, 40, 0, 0] },
        nameGap: 15,
        splitLine: { lineStyle: { color: '#f3f4f6' } },
        axisLabel: {
          color: '#9ca3af',
          fontSize: 11,
          formatter: (value: number) => {
            if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M'
            if (value >= 1000) return (value / 1000).toFixed(0) + 'K'
            return value.toString()
          }
        }
      },
      {
        type: 'value',
        name: t('components.main.heatmap.metrics.cost'),
        position: 'right',
        nameTextStyle: { color: '#9ca3af', padding: [0, 10, 0, 0] },
        splitLine: { show: false },
        axisLabel: {
          color: '#9ca3af',
          fontSize: 11,
          formatter: (value: number) => '\u0024' + value.toFixed(value >= 1 ? 0 : 2)
        }
      },
      {
        type: 'value',
        show: false,
        splitLine: { show: false }
      }
    ],
    series,
    backgroundColor: 'transparent'
  }
})
</script>

<style scoped>
.usage-chart-container {
  width: 100%;
  height: 320px;
  padding-top: 1rem;
  display: flex;
  flex-direction: column;
}

.chart-header {
  display: flex;
  justify-content: center;
  margin-bottom: 0.5rem;
  padding: 0 0.5rem;
}

.chart-legend {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  justify-content: center;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8rem;
  color: var(--mac-text-secondary, #86868b);
  cursor: pointer;
  background: transparent;
  border: none;
  padding: 4px 8px;
  transition: color 0.2s;
  white-space: nowrap;
  opacity: 0.6;
}

.legend-item:hover {
  opacity: 0.85;
}

.legend-item.active {
  color: var(--mac-text, #f5f5f7);
  font-weight: 500;
  opacity: 1;
}

.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.chart {
  flex: 1;
  width: 100%;
  min-height: 250px;
}
</style>
