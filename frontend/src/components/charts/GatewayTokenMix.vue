<template>
  <section class="card gateway-panel">
    <header class="gateway-panel-heading"><div><h2>{{ t('gatewayUi.tokenMix') }}</h2><p>{{ t('usage.inSelectedRange') }}</p></div></header>
    <div v-if="total <= 0" class="gateway-chart-empty gateway-mix-empty">{{ t('admin.dashboard.noDataAvailable') }}</div>
    <div v-else class="gateway-panel-body gateway-group-layout gateway-token-mix">
      <div class="gateway-group-ring">
        <Doughnut :data="chart" :options="options" aria-hidden="true" />
        <div class="gateway-ring-value"><strong>{{ formatCompactNumber(total) }}</strong><small>Token</small></div>
      </div>
      <ul class="gateway-group-legend">
        <li v-for="(item, i) in items" :key="item.key">
          <span class="truncate" :title="t(item.key)"><i :style="{ background: colors[i] }"></i>{{ t(item.key) }}</span>
          <strong>{{ formatCompactNumber(item.value) }}</strong><small>{{ (item.value / total * 100).toFixed(1) }}%</small>
        </li>
      </ul>
    </div>
  </section>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, ArcElement, Tooltip, type ChartOptions } from 'chart.js'
import { Doughnut } from 'vue-chartjs'
import type { AdminUsageStatsResponse } from '@/api/admin/usage'
import { formatCompactNumber } from '@/utils/format'

ChartJS.register(ArcElement, Tooltip)
const { t } = useI18n()
const props = defineProps<{ stats: AdminUsageStatsResponse | null }>()
const colors = ['#6387ff', '#48bc9f', '#f3b34f', '#9a70eb']
const items = computed(() => {
  const s = props.stats
  const count = (n?: number) => typeof n === 'number' && Number.isFinite(n) ? Math.max(0, n) : 0
  return [
    { key: 'gatewayUi.tokenInput', value: count(s?.total_input_tokens) },
    { key: 'gatewayUi.tokenOutput', value: count(s?.total_output_tokens) },
    { key: 'usage.cacheCreationTokensLabel', value: count(s?.total_cache_creation_tokens) },
    { key: 'usage.cacheReadTokensLabel', value: count(s?.total_cache_read_tokens) },
  ]
})
const total = computed(() => items.value.reduce((sum, item) => sum + item.value, 0))
const chart = computed(() => ({ labels: items.value.map(item => t(item.key)), datasets: [{ data: items.value.map(item => item.value), backgroundColor: colors, borderWidth: 0 }] }))
const options: ChartOptions<'doughnut'> = { responsive: true, maintainAspectRatio: false, animation: false, cutout: '75%', plugins: { legend: { display: false } } }
</script>
