<template>
  <!-- Wallet Money Tile (DESIGN_1 signature component).
       One solid money-green tile: balance (left, spans both rows),
       today spend (top right), estimated remaining days (bottom right).
       Clicking opens the payment/recharge page — upstream has no /wallet
       route, so /purchase (PaymentView) is the closest equivalent. Amounts are exact
       (never abbreviated) and tabular. -->
  <button
    type="button"
    class="tile-money tile-glide block w-full text-left transition-colors hover:bg-money-hover"
    :aria-label="balanceLabel"
    @click="goToPayment"
  >
    <div v-if="loading" class="grid grid-cols-[1.3fr_1fr] grid-rows-2 gap-0">
      <div class="row-span-2 border-r border-white/30 pr-5">
        <div class="h-4 w-16 animate-pulse bg-white/20" />
        <div class="mt-3 h-10 w-40 animate-pulse bg-white/20" />
      </div>
      <div class="pl-5"><div class="h-4 w-20 animate-pulse bg-white/20" /></div>
      <div class="pl-5"><div class="h-4 w-20 animate-pulse bg-white/20" /></div>
    </div>
    <div v-else class="grid min-h-[172px] grid-cols-[1.3fr_1fr] grid-rows-2 lg:min-h-[172px]">
      <!-- Balance: left cell, spans both rows -->
      <div class="row-span-2 flex flex-col justify-center border-r border-white/30 pr-5">
        <p class="text-sm font-medium text-white/80">{{ balanceLabel }}</p>
        <p
          class="mt-2 whitespace-nowrap leading-none tracking-tight"
          :class="balanceSizeClass"
          style="font-variant-numeric: tabular-nums"
        >
          ${{ formattedBalance }}
        </p>
      </div>
      <!-- Today spend: top right -->
      <div class="flex flex-col justify-center pl-5">
        <p class="text-sm font-medium text-white/80">{{ todayLabel }}</p>
        <p class="mt-1 text-2xl font-medium leading-none sm:text-[32px]" style="font-variant-numeric: tabular-nums">
          ${{ formattedTodaySpend }}
        </p>
      </div>
      <!-- Estimated remaining days: bottom right, hidden when unknown -->
      <div class="flex flex-col justify-center border-t border-white/30 pl-5">
        <template v-if="estimatedDays != null">
          <p class="text-sm font-medium text-white/80">{{ estimatedLabel }}</p>
          <p class="mt-1 text-[17px] leading-none sm:text-xl">
            {{ daysText }}
          </p>
          <p v-if="estimateNote" class="mt-1 hidden text-xs text-white/70 sm:block">{{ estimateNote }}</p>
        </template>
        <p v-else class="text-sm text-white/70">{{ noUsageLabel }}</p>
      </div>
    </div>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  balance: number
  todaySpend: number
  /** null = no recent usage, so no day estimate is shown (spec). */
  estimatedDays?: number | null
  /** Short basis note, e.g. "Based on today's spend". */
  estimateNote?: string
  loading?: boolean
}>()

const router = useRouter()
const { t, locale } = useI18n()

const balanceLabel = computed(() => t('dashboard.balance'))
const todayLabel = computed(() => t('dashboard.todayCost'))
// No new i18n keys are introduced (the build checks locale completeness):
// the estimated-days strings fall back to locale-aware literals.
const isZh = computed(() => locale.value.startsWith('zh'))
const estimatedLabel = computed(() => (isZh.value ? '预计还能用' : 'Estimated remaining'))
const noUsageLabel = computed(() => (isZh.value ? '暂无近期用量' : 'No recent usage'))
const daysText = computed(() =>
  isZh.value ? `约 ${props.estimatedDays} 天` : `About ${props.estimatedDays} days`
)

function formatMoney(value: number): string {
  const fixed = Number(value || 0).toFixed(2)
  // Strip a trailing ".00" for integers (spec: no redundant ".0"), keep
  // cents otherwise. Thousands separators are added for readability.
  const [int, frac] = fixed.split('.')
  const grouped = Number(int).toLocaleString('en-US')
  return frac === '00' ? grouped : `${grouped}.${frac}`
}

const formattedBalance = computed(() => formatMoney(props.balance))
const formattedTodaySpend = computed(() => formatMoney(props.todaySpend))

// Fit-the-Tile rule: shrink the balance font as the digit count grows,
// so it never wraps or overflows (never truncate a money amount).
const balanceSizeClass = computed(() => {
  const digits = formattedBalance.value.replace(/[^0-9]/g, '').length
  if (digits <= 7) return 'text-[40px] sm:text-[44px] 2xl:text-[56px] font-normal'
  if (digits <= 9) return 'text-[34px] sm:text-[38px] 2xl:text-[46px] font-normal'
  return 'text-[28px] sm:text-[32px] 2xl:text-[38px] font-normal'
})

function goToPayment() {
  void router.push('/purchase')
}
</script>
