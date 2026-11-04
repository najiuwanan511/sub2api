<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-12"><LoadingSpinner /></div>
      <template v-else-if="stats">
        <!-- DESIGN_1 modern mode: greeting, signature tile row (wallet +
             quick actions), stat tiles, then the new-style trend chart.
             Classic mode keeps the original dashboard untouched. -->
        <template v-if="isModern">
          <!-- Greeting: display type, light weight, poster scale -->
          <section class="tile-glide" style="--n: 0">
            <h1 class="display">{{ greeting }}</h1>
            <p class="mt-2 text-[15px] font-light text-ink-3 lg:text-[17px]">{{ greetingSub }}</p>
          </section>

          <!-- Signature first row: wallet tile + quick action tiles -->
          <div class="tile-grid mt-6 grid-cols-2 lg:grid-cols-[2fr_1fr_1fr_1fr]">
            <div class="col-span-2 min-h-[150px] lg:col-span-1 lg:min-h-[172px]">
              <MoneyTile
                :balance="user?.balance || 0"
                :today-spend="stats.today_actual_cost || 0"
                :estimated-days="estimatedDays"
                :estimate-note="estimateNote"
                :loading="loading"
              />
            </div>
            <QuickActionTile
              :title="rechargeTitle"
              :hint="rechargeHint"
              icon="plus"
              variant="money"
              :glide-index="1"
              @activate="goTo('/purchase')"
            />
            <QuickActionTile
              :title="createKeyTitle"
              :hint="createKeyHint"
              icon="key"
              variant="solid"
              :glide-index="2"
              @activate="goTo('/keys')"
            />
            <QuickActionTile
              :title="copyTitle"
              :hint="apiBaseUrl || copyHint"
              icon="link"
              :variant="copied ? 'money' : 'tile'"
              :glide-index="3"
              @activate="copyEndpoint"
            />
          </div>

          <!-- Stat tiles: neutral tiles, metric numbers -->
          <div class="tile-grid mt-8 grid-cols-2 lg:grid-cols-4">
            <div v-for="(stat, i) in statTiles" :key="stat.label" class="tile tile-glide" :style="{ '--n': i }">
              <p class="text-[13px] text-ink-3">{{ stat.label }}</p>
              <p class="mt-2 flex items-baseline gap-1 leading-none">
                <span class="metric">{{ stat.value }}</span>
                <span v-if="stat.unit" class="text-[15px] text-ink-3">{{ stat.unit }}</span>
              </p>
            </div>
          </div>

          <div class="mt-8">
            <HomeTrendChart
              :trend-data="trendData"
              :loading="loadingCharts"
              :title="$t('dashboard.tokenUsageTrend')"
              :empty-text="$t('dashboard.noDataAvailable')"
            />
          </div>

          <div class="mt-8">
            <UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" />
          </div>
        </template>
        <template v-else>
          <UserDashboardStats :stats="stats" :balance="user?.balance || 0" :is-simple="authStore.isSimpleMode" :platform-quotas="platformQuotas" />
          <UserDashboardCharts v-model:startDate="startDate" v-model:endDate="endDate" v-model:granularity="granularity" :loading="loadingCharts" :trend="trendData" :models="modelStats" @dateRangeChange="loadCharts" @granularityChange="loadCharts" @refresh="refreshAll" />
          <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
            <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" /></div>
            <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
          </div>
        </template>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth';
import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'; import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'; import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'; import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import MoneyTile from '@/components/user/MoneyTile.vue'; import HomeTrendChart from '@/components/charts/HomeTrendChart.vue'
import QuickActionTile from '@/components/user/QuickActionTile.vue'
import { getUiStyle } from '@/utils/uiStyle'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'

const isModern = getUiStyle() === 'modern'
const { locale } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const isZh = computed(() => locale.value.startsWith('zh'))

// Estimated remaining days, honestly labelled: balance / today's actual
// spend. No 7-day average is available in this view's stats, so the note
// says so instead of inventing one; zero spend means no estimate (spec).
const estimatedDays = computed(() => {
  const today = stats.value?.today_actual_cost || 0
  const balance = user.value?.balance || 0
  if (today <= 0 || balance <= 0) return null
  return Math.floor(balance / today)
})
const estimateNote = computed(() =>
  locale.value.startsWith('zh') ? '按今日实际花费估算' : 'Estimated from today\'s actual spend'
)
import type { UsageLog, TrendDataPoint, ModelStat, PlatformQuotaItem } from '@/types'
import { getMyPlatformQuotas } from '@/api/user'
import { formatDateLocalInput } from '@/utils/format'

const authStore = useAuthStore(); const user = computed(() => authStore.user)
const stats = ref<UserStatsType | null>(null); const loading = ref(false); const loadingUsage = ref(false); const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const recentUsage = ref<UsageLog[]>([])
const platformQuotas = ref<PlatformQuotaItem[] | null>(null)

const startDate = ref(formatDateLocalInput(new Date(Date.now() - 6 * 86400000))); const endDate = ref(formatDateLocalInput(new Date())); const granularity = ref('day')

const loadStats = async () => { loading.value = true; try { await authStore.refreshUser(); stats.value = await usageAPI.getDashboardStats() } catch (error) { console.error('Failed to load dashboard stats:', error) } finally { loading.value = false } }
const loadCharts = async () => { loadingCharts.value = true; try { const res = await Promise.all([usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }), usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })]); trendData.value = res[0].trend || []; modelStats.value = res[1].models || [] } catch (error) { console.error('Failed to load charts:', error) } finally { loadingCharts.value = false } }
const loadRecent = async () => { loadingUsage.value = true; try { const res = await usageAPI.getByDateRange(startDate.value, endDate.value); recentUsage.value = res.items.slice(0, 5) } catch (error) { console.error('Failed to load recent usage:', error) } finally { loadingUsage.value = false } }
const loadPlatformQuotas = async () => { try { const data = await getMyPlatformQuotas(); platformQuotas.value = data.platform_quotas ?? [] } catch (error) { console.warn('Failed to load platform quotas:', error); platformQuotas.value = [] } }
const refreshAll = () => { loadStats(); loadCharts(); loadRecent(); loadPlatformQuotas() }

onMounted(() => { refreshAll() })

/* ---------- Modern mode: greeting, quick actions, stat tiles ---------- */

function goTo(path: string) {
  void router.push(path)
}

const greeting = computed(() => {
  const hour = new Date().getHours()
  const word = isZh.value
    ? hour < 12 ? '早上好' : hour < 18 ? '下午好' : '晚上好'
    : hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening'
  const name = user.value?.username || user.value?.email?.split('@')[0] || ''
  return `${word}${name ? (isZh.value ? '，' : ', ') + name : ''}`
})

const greetingSub = computed(() => {
  if (!stats.value) return ''
  const requests = Number(stats.value.today_requests || 0).toLocaleString('en-US')
  const cost = money(stats.value.today_actual_cost || 0)
  return isZh.value
    ? `今天已调用 ${requests} 次，花费 ${cost}`
    : `${requests} requests today, ${cost} spent`
})

function money(value: number): string {
  const fixed = Number(value || 0).toFixed(2)
  const [int, frac] = fixed.split('.')
  const grouped = Number(int).toLocaleString('en-US')
  // Screenshots and MoneyTile both show $-denominated balances; keep one
  // currency symbol everywhere so tiles never mix.
  return `$${frac === '00' ? grouped : `${grouped}.${frac}`}`
}

/** Compact big numbers: 1.2K / 3.4M (spec: The Tabular Rule). */
function compact(value: number): string {
  const n = Number(value || 0)
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1).replace(/\.0$/, '')}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, '')}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1).replace(/\.0$/, '')}K`
  return n.toLocaleString('en-US')
}

const rechargeTitle = computed(() => (isZh.value ? '充值' : 'Recharge'))
const rechargeHint = computed(() => (isZh.value ? '进入钱包充值' : 'Open the wallet'))
const createKeyTitle = computed(() => (isZh.value ? '创建 Key' : 'Create Key'))
const createKeyHint = computed(() => (isZh.value ? '选分组生成' : 'Pick a group'))
const copyTitle = computed(() => (isZh.value ? '复制接口地址' : 'Copy API endpoint'))
const copyHint = computed(() => (isZh.value ? '一键复制' : 'Copy in one click'))

const apiBaseUrl = computed(() => appStore.apiBaseUrl || '')

const copied = ref(false)
let copyTimer: ReturnType<typeof setTimeout> | null = null
async function copyEndpoint() {
  const url = apiBaseUrl.value
  if (!url) return
  try {
    await navigator.clipboard.writeText(url)
    copied.value = true
    if (copyTimer) clearTimeout(copyTimer)
    copyTimer = setTimeout(() => { copied.value = false }, 1500)
  } catch (error) {
    console.error('Failed to copy endpoint:', error)
  }
}

const statTiles = computed(() => {
  const s = stats.value
  if (!s) return []
  return [
    { label: isZh.value ? '请求数' : 'Requests', value: Number(s.today_requests || 0).toLocaleString('en-US'), unit: isZh.value ? '次' : '' },
    { label: 'Token', value: compact(s.today_tokens || 0), unit: '' },
    { label: isZh.value ? '花费' : 'Spend', value: money(s.today_actual_cost || 0), unit: '' },
    { label: 'RPM', value: Number(s.rpm || 0).toLocaleString('en-US'), unit: isZh.value ? '次/分' : 'req/min' }
  ]
})
</script>
