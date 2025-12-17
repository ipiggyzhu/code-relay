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
  LegendComponent, // We implement custom legend but keep this for internal logic if needed
  DataZoomComponent
])

provide(THEME_KEY, 'light') // or dynamic based on app theme

const props = defineProps<{
  data: UsageHeatmapWeek[]
}>()

const { t } = useI18n()

// Color palette matching the app's aesthetic (Apple-like/Modern)
const colors = {
  requests: '#3b82f6', // Blue
  inputTokens: '#10b981', // Emerald
  outputTokens: '#f59e0b', // Amber
  reasoningTokens: '#8b5cf6', // Violet
  cost: '#ef4444' // Red
}

const metrics = [
  { key: 'requests', label: t('components.main.heatmap.metrics.requests'), color: colors.requests },
  { key: 'inputTokens', label: t('components.main.heatmap.metrics.inputTokens'), color: colors.inputTokens },
  { key: 'outputTokens', label: t('components.main.heatmap.metrics.outputTokens'), color: colors.outputTokens },
  { key: 'reasoningTokens', label: t('components.main.heatmap.metrics.reasoningTokens'), color: colors.reasoningTokens },
  { key: 'cost', label: t('components.main.heatmap.metrics.cost'), color: colors.cost }
]

// Default active metrics
const activeMetrics = ref(['inputTokens', 'outputTokens', 'cost'])

const toggleMetric = (key: string) => {
  if (activeMetrics.value.includes(key)) {
    // Prevent unselecting the last one
    if (activeMetrics.value.length > 1) {
      activeMetrics.value = activeMetrics.value.filter(k => k !== key)
    }
  } else {
    activeMetrics.value.push(key)
  }
}

// Flatten data: standardizing the weeks structure into a flat daily array
const flatData = computed(() => {
  const result: any[] = []
  if (!props.data) return result

  // The data comes as weeks (arrays of days/buckets), we just want to sort them by date flatly
  // Iterate strictly through the structure
  for (const week of props.data) {
    for (const day of week) {
       // Filter out future dates or placeholder dates if necessary, 
       // but assuming data is clean or we act on valid dates
       if (day.dateKey) {
           result.push({
               date: day.dateKey,
               requests: day.requests,
               inputTokens: day.inputTokens,
               outputTokens: day.outputTokens,
               reasoningTokens: day.reasoningTokens,
               cost: day.cost
           })
       }
    }
  }
  // Ensure sorted by date
  return result.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
})

const chartOption = computed(() => {
  const dates = flatData.value.map(d => d.date)
  
  const series = metrics
    .filter(m => activeMetrics.value.includes(m.key))
    .map(m => ({
      name: m.label,
      type: 'line',
      data: flatData.value.map(d => d[m.key]),
      smooth: true,
      showSymbol: false,
      symbol: 'circle',
      symbolSize: 6,
      // Use secondary axis for cost to avoid scale issues
      yAxisIndex: m.key === 'cost' ? 1 : 0,
      itemStyle: {
        color: m.color
      },
      lineStyle: {
        width: 1.5
      },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: m.color + '33' }, // 20% opacity (lighter area)
            { offset: 1, color: m.color + '00' }  // 0% opacity
          ]
        }
      }
    }))

  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e5e7eb',
      textStyle: {
        color: '#374151',
        fontSize: 12
      },
      padding: [8, 12],
      axisPointer: {
        lineStyle: {
          color: '#9ca3af',
          type: 'dashed',
          width: 1
        }
      }
    },
    grid: {
      top: 30, 
      right: 40, // Increased for right Y-axis
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
          // Format date as MM-DD
          const d = new Date(value)
          return `${d.getMonth() + 1}-${d.getDate()}`
        }
      }
    },
    yAxis: [
      {
        type: 'value',
        name: t('components.main.heatmap.metrics.requests') + '/' + t('components.main.providers.tokens'),
        nameTextStyle: {
           color: '#9ca3af',
           padding: [0, 0, 0, 10]
        },
        splitLine: {
          lineStyle: {
            color: '#f3f4f6'
          }
        },
        axisLabel: {
          color: '#9ca3af',
          fontSize: 11
        }
      },
      {
        type: 'value',
        name: t('components.main.heatmap.metrics.cost'),
        position: 'right',
        nameTextStyle: {
           color: '#ef4444',
           padding: [0, 10, 0, 0]
        },
        splitLine: { show: false }, // Hide grid lines for secondary axis
        axisLabel: {
          color: '#ef4444', // Match cost color
          fontSize: 11,
          formatter: '${value}'
        }
      }
    ],
    series
  }
})
</script>

<style scoped>
.usage-chart-container {
  width: 100%;
  height: 100%;
  padding-top: 1rem;
  display: flex;
  flex-direction: column;
}

.chart-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 0.5rem;
  padding: 0 1rem;
}

.chart-legend {
  display: flex;
  gap: 1.5rem;
  flex-wrap: nowrap; /* Prevent wrapping */
  overflow-x: auto; /* Allow scroll if screen is tiny */
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  color: #9ca3af; /* Lighter text by default */
  cursor: pointer;
  background: none;
  border: none;
  padding: 4px 8px;
  border-radius: 999px; /* Pill shape */
  transition: all 0.2s;
  opacity: 1; /* Handle opacity via color state */
  white-space: nowrap; /* Prevent text wrapping */
}

.legend-item:hover {
  background-color: #f3f4f6;
  color: #4b5563;
}

.legend-item.active {
  background-color: #f3f4f6;
  color: #111827;
  font-weight: 500;
}

.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}


.chart {
  /* Ensure it fills the container height */
  flex: 1;
  width: 100%;
  min-height: 250px;
}
</style>
