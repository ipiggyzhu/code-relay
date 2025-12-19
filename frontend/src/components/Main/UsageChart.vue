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
import { computed, ref, provide } from 'vue'
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

provide(THEME_KEY, 'light')

const props = defineProps<{
  data: UsageHeatmapWeek[]
}>()

const { t } = useI18n()

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
  if (!props.data || props.data.length === 0) return []

  const dailyMap = new Map<string, {
    date: string
    requests: number
    inputTokens: number
    outputTokens: number
    reasoningTokens: number
    cost: number
  }>()

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

  return Array.from(dailyMap.values())
    .filter(d => d.requests > 0 || d.inputTokens > 0 || d.outputTokens > 0 || d.cost > 0)
    .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
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
      trigger: 'axis',
      axisPointer: {
        type: 'line',
        lineStyle: { color: '#d1d5db', type: 'dashed', width: 1 }
      },
      backgroundColor: 'rgba(255, 255, 255, 0.98)',
      borderColor: '#e5e7eb',
      borderWidth: 1,
      borderRadius: 8,
      textStyle: { color: '#374151', fontSize: 12 },
      padding: [10, 14],
      formatter: (params: any) => {
        if (!params || params.length === 0) return ''
        const date = params[0].axisValue
        let html = '<div style="font-weight:500;margin-bottom:6px;color:#111827">' + date + '</div>'
        params.forEach((item: any) => {
          const isCost = item.seriesName.includes(costLabel)
          const value = isCost
            ? '\u0024' + (item.value || 0).toFixed(4)
            : (item.value || 0).toLocaleString()
          html += '<div style="display:flex;align-items:center;gap:6px;margin:3px 0">'
          html += '<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:' + item.color + '"></span>'
          html += '<span style="color:#6b7280">' + item.seriesName + '</span>'
          html += '<span style="margin-left:auto;font-weight:500;color:#111827">' + value + '</span>'
          html += '</div>'
        })
        return html
      }
    },
    grid: {
      top: 30,
      right: 50,
      bottom: 20,
      left: 10,
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
        nameTextStyle: { color: '#9ca3af', padding: [0, 0, 0, 10] },
        splitLine: { lineStyle: { color: '#f3f4f6' } },
        axisLabel: { color: '#9ca3af', fontSize: 11 }
      },
      {
        type: 'value',
        name: t('components.main.heatmap.metrics.cost'),
        position: 'right',
        min: 0,
        max: 10,
        nameTextStyle: { color: '#9ca3af', padding: [0, 10, 0, 0] },
        splitLine: { show: false },
        axisLabel: {
          color: '#9ca3af',
          fontSize: 11,
          formatter: (value: number) => '\u0024' + value
        }
      },
      {
        type: 'value',
        show: false,
        splitLine: { show: false }
      }
    ],
    series
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
