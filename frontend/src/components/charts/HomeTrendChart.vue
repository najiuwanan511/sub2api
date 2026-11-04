<template>
  <!-- Home trend chart (DESIGN_1, new style only).
       Tokens = straight bars in chart-bar colour (70% of the grid slot),
       cost = 3px chart-line, no points except a 4px hover point, straight
       segments (tension 0), no area fill, rule gridlines, 11px ink-3 ticks,
       no legend, inverse solid tooltip. Colours are read live from the
       --c-* CSS variables so a theme switch redraws the chart. -->
  <div class="tile p-4">
    <h3 v-if="title" class="mb-4 text-sm font-semibold text-ink">{{ title }}</h3>
    <div v-if="loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="chartData" class="h-48">
      <Chart type="bar" :data="chartData" :options="chartOptions" />
    </div>
    <div v-else class="flex h-48 items-center justify-center text-sm text-ink-3">
      {{ emptyText }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  BarController,
  PointElement,
  LineElement,
  LineController,
  Title,
  Tooltip,
  Legend
} from 'chart.js'
import { Chart } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { TrendDataPoint } from '@/types'

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  BarController,
  PointElement,
  LineElement,
  LineController,
  Title,
  Tooltip,
  Legend
)

const props = defineProps<{
  trendData: TrendDataPoint[]
  loading?: boolean
  title?: string
  emptyText?: string
}>()

function cssVar(name: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback
  const value = getComputedStyle(document.documentElement).getPropertyValue(name).trim()
  if (!value) return fallback
  // --c-* variables are RGB triplets ("0 112 60"); convert to rgb().
  return value.includes(' ') && !value.startsWith('#') ? `rgb(${value})` : value
}

const chartData = computed(() => {
  if (!props.trendData?.length) return null
  const bar = cssVar('--c-chart-bar', '#0a0b0a')
  const line = cssVar('--c-chart-line', '#00a85a')
  return {
    labels: props.trendData.map((d) => d.date),
    datasets: [
      {
        type: 'bar' as const,
        label: 'Tokens',
        data: props.trendData.map((d) => d.total_tokens),
        backgroundColor: bar,
        borderRadius: 0,
        barPercentage: 0.7,
        categoryPercentage: 1,
        yAxisID: 'y'
      },
      {
        type: 'line' as const,
        label: 'Cost',
        data: props.trendData.map((d) => d.actual_cost),
        borderColor: line,
        borderWidth: 3,
        pointRadius: 0,
        pointHoverRadius: 4,
        pointBackgroundColor: line,
        tension: 0,
        fill: false,
        yAxisID: 'y1'
      }
    ]
  }
})

const chartOptions = computed(() => {
  const ink3 = cssVar('--c-ink-3', '#6a6f6a')
  const rule = cssVar('--c-rule', '#dfe1dd')
  const isDark = document.documentElement.classList.contains('dark')
  return {
    responsive: true,
    maintainAspectRatio: false,
    animation: { duration: 500, easing: 'easeOutQuart' as const },
    plugins: {
      legend: { display: false },
      tooltip: {
        backgroundColor: isDark ? '#ffffff' : '#000000',
        titleColor: isDark ? '#000000' : '#ffffff',
        bodyColor: isDark ? '#000000' : '#ffffff',
        borderWidth: 0,
        cornerRadius: 0,
        padding: 10,
        displayColors: false
      }
    },
    scales: {
      x: { grid: { color: rule }, ticks: { color: ink3, font: { size: 11 } } },
      y: { grid: { color: rule }, ticks: { color: ink3, font: { size: 11 } } },
      y1: { position: 'right' as const, grid: { drawOnChartArea: false }, ticks: { color: ink3, font: { size: 11 } } }
    }
  }
})
</script>
