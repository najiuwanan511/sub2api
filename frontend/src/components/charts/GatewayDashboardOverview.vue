<template>
  <section class="gateway-dashboard-overview" :aria-label="t('nav.dashboard')">
    <div class="gateway-stat-grid">
      <article v-for="metric in metrics" :key="metric.key" class="card gateway-stat" :data-metric="metric.key">
        <div class="gateway-stat-label"><span :data-tone="metric.tone" class="gateway-stat-icon"><Icon :name="metric.icon" size="sm" /></span>{{ metric.label }}</div>
        <strong>{{ metric.value }}</strong>
        <p>
          <span v-if="metric.delta != null" :class="{ 'gateway-positive': metric.delta > 0, 'gateway-negative': metric.delta < 0 }">{{ metric.delta > 0 ? '↗ +' : metric.delta < 0 ? '↘ ' : '' }}{{ metric.delta.toFixed(1) }}% </span>{{ metric.note }}
        </p>
      </article>
    </div>

    <div class="gateway-dashboard-grid gateway-dashboard-primary">
      <article class="card gateway-panel">
        <header class="gateway-panel-heading">
          <div><h2>{{ t('gatewayUi.trend') }}</h2><p>{{ t('gatewayUi.trendPeriod', { unit: t(`gatewayUi.${granularity}`), start, end }) }}</p></div>
          <div class="gateway-segment" role="group" :aria-label="t('gatewayUi.trend')">
            <button v-for="item in trendMetrics" :key="item.key" type="button" :aria-pressed="selectedMetric === item.key" @click="selectedMetric = item.key">{{ t(item.label) }}</button>
          </div>
        </header>
        <div class="gateway-panel-body gateway-trend-body">
          <div v-if="loading || !trend.length" class="gateway-chart-empty"><LoadingSpinner v-if="loading" /><span v-else>{{ emptyMessage }}</span></div>
          <Line v-else :data="trendChart" :options="trendOptions" :aria-label="t('gatewayUi.trend')" />
        </div>
        <footer v-if="!loading && trend.length" class="gateway-chart-legend"><i></i>{{ t(trendMetrics.find(item => item.key === selectedMetric)!.label) }}</footer>
      </article>
      <article class="card gateway-panel">
        <header class="gateway-panel-heading"><div><h2>{{ t('gatewayUi.groupShare') }}</h2><p>{{ t('gatewayUi.actualCost') }}</p></div></header>
        <div class="gateway-panel-body">
          <div v-if="loading || !groupTotal" class="gateway-chart-empty"><LoadingSpinner v-if="loading" /><span v-else>{{ emptyMessage }}</span></div>
          <div v-else class="gateway-group-layout">
            <div class="gateway-group-ring">
              <Doughnut :data="groupChart" :options="groupOptions" :aria-label="t('gatewayUi.groupShare')" />
              <div class="gateway-ring-value"><strong>{{ money(groupTotal) }}</strong><small>{{ t('gatewayUi.actualCost') }}</small></div>
            </div>
            <ul class="gateway-group-legend">
              <li v-for="(group, i) in groupDisplay" :key="group.id">
                <span class="truncate" :title="group.name"><i :style="{ background: colors[i] }"></i>{{ group.name }}</span>
                <strong>{{ money(group.cost) }}</strong><small>{{ percent(group.cost / groupTotal) }}</small>
              </li>
            </ul>
          </div>
        </div>
      </article>
    </div>

    <div class="gateway-dashboard-grid gateway-dashboard-secondary">
      <article class="card gateway-panel">
        <header class="gateway-panel-heading"><div><h2>{{ t('gatewayUi.modelRank') }}</h2><p>{{ t('gatewayUi.modelPeriod') }}</p></div></header>
        <div class="gateway-panel-body gateway-model-body">
          <div v-if="modelsLoading || !modelDisplay.length" class="gateway-chart-empty"><LoadingSpinner v-if="modelsLoading" /><span v-else>{{ modelsFailed ? t('gatewayUi.dataError') : t('admin.dashboard.noDataAvailable') }}</span></div>
          <Bar v-else :data="modelChart" :options="modelOptions" :aria-label="t('gatewayUi.modelRank')" />
        </div>
      </article>
      <article class="card gateway-panel">
        <header class="gateway-panel-heading"><div><h2>{{ t('gatewayUi.realtime') }}</h2><p>{{ t('gatewayUi.realtimeDesc') }}</p></div></header>
        <div class="gateway-panel-body">
          <p v-if="opsDisabled || opsPartial" class="gateway-ops-note">{{ t(opsDisabled ? 'gatewayUi.opsDisabled' : 'gatewayUi.opsUnavailable') }}</p>
          <div class="gateway-runtime-grid">
            <div v-for="metric in runtime" :key="metric.key" class="gateway-runtime-item">
              <div class="gateway-stat-label"><span :data-tone="metric.tone" class="gateway-stat-icon"><Icon :name="metric.icon" size="sm" /></span>{{ t(metric.label) }}</div>
              <strong>{{ metric.value }}</strong>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, BarElement, ArcElement, Tooltip, Filler, type ScriptableContext, type ChartOptions } from 'chart.js'
import { Line, Bar, Doughnut } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import type { DashboardStats, TrendDataPoint, GroupStat, ModelStat } from '@/types'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, ArcElement, Tooltip, Filler)
const props = defineProps<{
  stats: DashboardStats | null
  trend: TrendDataPoint[]
  groups: GroupStat[]
  models: ModelStat[]
  granularity: 'day' | 'hour'
  start: string
  end: string
  loading: boolean
  failed: boolean
  modelsLoading: boolean
  modelsFailed: boolean
  opsDisabled: boolean
  opsPartial: boolean
  concurrency: number | null
  rpm: number | null
  duration: number | null
  errorRate: number | null
}>()
const { t, locale } = useI18n()
const emptyMessage = computed(() => t(props.failed ? 'gatewayUi.dataError' : 'admin.dashboard.noDataAvailable'))
const trendMetrics = [
  { key: 'requests', label: 'gatewayUi.requests' }, { key: 'total_tokens', label: 'admin.dashboard.tokens' },
  { key: 'actual_cost', label: 'gatewayUi.cost' }
] as const
const selectedMetric = ref<typeof trendMetrics[number]['key']>('requests')
const colors = ['#5b8cff', '#32b96d', '#e4a13c', '#9966f5', '#429ecf', '#e36b80']
const isDark = ref(document.documentElement.classList.contains('dark'))
let themeObserver: MutationObserver | undefined
onMounted(() => {
  themeObserver = new MutationObserver(() => { isDark.value = document.documentElement.classList.contains('dark') })
  themeObserver.observe(document.documentElement, { attributeFilter: ['class'] })
})
onBeforeUnmount(() => themeObserver?.disconnect())
const chartGrid = computed(() => isDark.value ? '#262d38' : '#e3e8ef')
const chartText = computed(() => isDark.value ? '#a0acbc' : '#526075')
const valid = (value: number | null | undefined): value is number => typeof value === 'number' && Number.isFinite(value)
const number = (value: number | null | undefined) => valid(value) ? value.toLocaleString(locale.value) : '—'
const compact = (value: number | undefined) => !valid(value) ? '—' : new Intl.NumberFormat('en', { notation: 'compact', maximumFractionDigits: 2 }).format(value)
const money = (value: number | undefined) => valid(value) ? `$${value.toLocaleString(locale.value, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}` : '—'
const percent = (value: number | null | undefined) => valid(value) ? `${(value * 100).toFixed(1)}%` : '—'
const dates = new Date()
dates.setDate(dates.getDate() - 1)
const yesterday = `${dates.getFullYear()}-${String(dates.getMonth()+1).padStart(2, '0')}-${String(dates.getDate()).padStart(2, '0')}`
const previous = computed(() => props.granularity === 'day' ? props.trend.find(point => point.date === yesterday) : undefined)
function change(today: number | undefined, before: number | undefined) {
  if (!valid(today) || !valid(before) || before <= 0) return null
  return (today - before) * 100 / before
}
const metrics = computed(() => {
  const s = props.stats
  const deltaRequests = s?.stats_stale ? null : change(s?.today_requests, previous.value?.requests)
  const deltaTokens = s?.stats_stale ? null : change(s?.today_tokens, previous.value?.total_tokens)
  const availability = s && s.total_accounts > 0 && valid(s.normal_accounts) ? s.normal_accounts / s.total_accounts : null
  return [
    { key: 'requests', label: t('admin.dashboard.todayRequests'), icon: 'bolt', tone: 'blue', value: number(s?.today_requests), delta: deltaRequests, note: deltaRequests == null ? `${t('common.total')}: ${number(s?.total_requests)}` : t('gatewayUi.dayChange') },
    { key: 'tokens', label: t('admin.dashboard.todayTokens'), icon: 'cube', tone: 'cyan', value: compact(s?.today_tokens), delta: deltaTokens, note: deltaTokens == null ? `${t('common.total')}: ${compact(s?.total_tokens)}` : t('gatewayUi.dayChange') },
    { key: 'cost', label: t('gatewayUi.actualToday'), icon: 'dollar', tone: 'green', value: money(s?.today_actual_cost), delta: null, note: `${t('admin.dashboard.standard')} ${money(s?.today_cost)}` },
    { key: 'pool', label: t('gatewayUi.poolAvailable'), icon: 'shield', tone: 'amber', value: percent(availability), delta: null, note: s ? t('gatewayUi.normalRange', { normal: number(s.normal_accounts), total: number(s.total_accounts) }) : '—' }
  ] as const
})
const runtime = computed(() => [
  { key: 'concurrency', label: 'gatewayUi.concurrent', value: number(props.concurrency), icon: 'bolt', tone: 'blue' },
  { key: 'rpm', label: 'gatewayUi.rpm', value: number(props.rpm), icon: 'clock', tone: 'cyan' },
  { key: 'duration', label: 'admin.dashboard.avgResponse', value: valid(props.duration) ? `${number(Math.round(props.duration))} ms` : '—', icon: 'clock', tone: 'cyan' },
  { key: 'errors', label: 'gatewayUi.errorRate', value: valid(props.errorRate) ? `${(props.errorRate * 100).toFixed(2)}%` : '—', icon: 'shield', tone: 'green' }
] as const)
const trendChart = computed(() => ({
  labels: props.trend.map(point => point.date),
  datasets: [{ data: props.trend.map(point => point[selectedMetric.value]), borderColor: colors[0], backgroundColor: ({ chart }: ScriptableContext<'line'>) => {
    const area = chart.chartArea
    if (!area) return '#5b8cff26'
    const gradient = chart.ctx.createLinearGradient(0, area.top, 0, area.bottom)
    gradient.addColorStop(0, '#5b8cff45')
    gradient.addColorStop(1, '#5b8cff02')
    return gradient
  }, fill: true, borderWidth: 2, tension: 0, pointRadius: 0, pointHoverRadius: 4 }]
}))
const trendOptions = computed<ChartOptions<'line'>>(() => ({
  responsive: true, maintainAspectRatio: false, animation: false,
  interaction: { mode: 'index', intersect: false },
  plugins: { legend: { display: false }, tooltip: { callbacks: { label: item => selectedMetric.value === 'actual_cost' ? money(item.parsed.y ?? undefined) : number(item.parsed.y) } } },
  scales: {
    x: { grid: { display: false }, ticks: { color: chartText.value, maxTicksLimit: 8, callback(value) { const label = this.getLabelForValue(Number(value)); return props.granularity === 'day' && /^\d{4}-\d{2}-\d{2}$/.test(label) ? label.slice(5) : label } } },
    y: { beginAtZero: true, grid: { color: chartGrid.value }, ticks: { color: chartText.value, maxTicksLimit: 5, callback: value => selectedMetric.value === 'actual_cost' ? `$${compact(Number(value))}` : compact(Number(value)) } }
  }
}))
const groupDisplay = computed(() => {
  const rows = props.groups.filter(row => valid(row.actual_cost) && row.actual_cost > 0)
    .map(row => ({ id: String(row.group_id), name: row.group_name || `#${row.group_id}`, cost: row.actual_cost }))
    .sort((a,b) => b.cost-a.cost)
  if (rows.length <= 5) return rows
  return [...rows.slice(0, 5), { id: 'other', name: t('gatewayUi.otherGroups'), cost: rows.slice(5).reduce((sum, row) => sum + row.cost, 0) }]
})
const groupTotal = computed(() => groupDisplay.value.reduce((sum, group) => sum + group.cost, 0))
const groupChart = computed(() => ({ labels: groupDisplay.value.map(group => group.name), datasets: [{ data: groupDisplay.value.map(group => group.cost), backgroundColor: colors, borderWidth: 0 }] }))
const groupOptions: ChartOptions<'doughnut'> = { cutout: '77%', responsive: true, maintainAspectRatio: false, animation: false, plugins: { legend: { display: false }, tooltip: { callbacks: { label: item => `${item.label}: ${money(item.parsed)}` } } } }
const modelDisplay = computed(() => [...props.models].filter(model => valid(model.requests) && model.requests > 0).sort((a,b) => b.requests-a.requests).slice(0, 8))
const modelChart = computed(() => ({ labels: modelDisplay.value.map(model => model.model), datasets: [{ data: modelDisplay.value.map(model => model.requests), backgroundColor: '#5b8cffce', borderRadius: 3, maxBarThickness: 36 }] }))
const modelOptions = computed<ChartOptions<'bar'>>(() => ({
  responsive: true, maintainAspectRatio: false, animation: false,
  plugins: { legend: { display: false }, tooltip: { callbacks: { label: item => number(item.parsed.y) } } },
  scales: {
    x: { grid: { display: false }, ticks: { color: chartText.value, maxRotation: 0, callback(value) { const text = this.getLabelForValue(Number(value)); return text.length > 13 ? `${text.slice(0, 12)}…` : text } } },
    y: { beginAtZero: true, grid: { color: chartGrid.value }, ticks: { color: chartText.value, maxTicksLimit: 5, callback: value => compact(Number(value)) } }
  }
}))
</script>

