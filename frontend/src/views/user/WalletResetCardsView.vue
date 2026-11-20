<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl space-y-5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('walletReset.title') }}</h1>
        <RouterLink to="/redeem" class="btn btn-secondary">{{ t('nav.redeem') }}</RouterLink>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="card p-5"><p>{{ t('walletReset.regular') }}</p><p class="mt-2 text-2xl font-semibold tabular-nums">{{ money(auth.user?.balance ?? 0) }}</p></div>
        <div class="card p-5"><p>{{ t('walletReset.temporary') }}</p><p class="mt-2 text-2xl font-semibold tabular-nums">{{ creditsLoaded ? money(activeCredits) : '—' }}</p></div>
      </div>
      <section class="card space-y-4 p-5">
        <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('walletReset.hint') }}</p>
        <form class="flex flex-wrap gap-2" @submit.prevent="previewCard">
          <input id="reset-card-code" v-model.trim="code" :aria-label="t('walletReset.code')" :placeholder="t('walletReset.code')" class="input min-w-48 flex-1" required maxlength="128" autocomplete="off" :disabled="busy" @input="preview = null; message = ''" />
          <button type="submit" class="btn btn-secondary" :disabled="busy || !code">{{ t('walletReset.preview') }}</button>
        </form>
        <div v-if="preview" class="flex flex-wrap items-center justify-between gap-3 rounded-lg bg-primary-50 p-4 dark:bg-primary-900/20">
          <div>
            <p class="font-medium">{{ t('walletReset.estimate', { amount: money(preview.credit_usd), group: preview.group_name }) }}</p>
            <p class="mt-1 text-sm">{{ t('walletReset.detail', { spend: money(preview.eligible_usd), rate: preview.percent, days: preview.validity_days }) }}</p>
          </div>
          <button type="button" class="btn btn-primary" :disabled="busy || preview.credit_usd <= 0" @click="applyCard">{{ t('walletReset.apply') }}</button>
        </div>
        <p v-if="message" role="status" class="text-sm" :class="failed ? 'text-red-600 dark:text-red-400' : 'text-emerald-700 dark:text-emerald-300'">{{ message }}</p>
      </section>
      <section class="card space-y-3 p-5">
        <h2 class="font-semibold">{{ t('walletReset.history') }}</h2>
        <p v-if="loadFailed" role="status" class="text-sm text-red-600 dark:text-red-400">{{ t('walletReset.loadFailed') }}</p>
        <p v-else-if="!creditsLoaded" class="text-sm">{{ t('common.loading') }}</p>
        <p v-else-if="!credits.length" class="text-sm text-gray-500">{{ t('walletReset.empty') }}</p>
        <div v-for="credit in credits" :key="credit.redeem_code_id" class="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-gray-200 p-3 dark:border-dark-700">
          <div>
            <p class="font-medium">{{ credit.group_name || '—' }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('walletReset.expiry', { date: formatDateTime(credit.expires_at) }) }}</p>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('walletReset.creditUsage', { used: money(usedCredit(credit)), total: money(credit.amount_usd) }) }}</p>
            <progress :value="usedCredit(credit)" :max="credit.amount_usd" :aria-label="t('walletReset.creditUsage', { used: money(usedCredit(credit)), total: money(credit.amount_usd) })" class="mt-1 block h-1 w-full min-w-28 accent-primary-600" />
          </div>
          <div class="text-right tabular-nums">
            <p>{{ money(credit.status === 'active' ? credit.remaining_usd : 0) }} / {{ money(credit.amount_usd) }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t(`walletReset.status.${credit.status}`) }}</p>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { redeemAPI, type WalletResetCredit, type WalletResetPreview } from '@/api/redeem'
import { useAuthStore } from '@/stores/auth'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const auth = useAuthStore()
const code = ref('')
const preview = ref<WalletResetPreview | null>(null)
const credits = ref<WalletResetCredit[]>([])
const creditsLoaded = ref(false)
const loadFailed = ref(false)
const busy = ref(false)
const failed = ref(false)
const message = ref('')
const activeCredits = ref(0)
const usedCredit = (credit: WalletResetCredit) => Math.max(0, credit.amount_usd - credit.remaining_usd)
const money = (amount: number) => `$${amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 8 })}`

async function loadCredits() {
  try {
    const result = await redeemAPI.getResetCredits()
    credits.value = result.credits
    activeCredits.value = result.active_usd
    creditsLoaded.value = true
    loadFailed.value = false
  } catch {
    loadFailed.value = true
  }
}

function showError(error: unknown, fallback: string) {
  failed.value = true
  const reason = typeof error === 'object' && error && 'reason' in error ? String(error.reason) : ''
  const key = ({
    REDEEM_CODE_NOT_FOUND: 'notFound', REDEEM_CODE_EXPIRED: 'expired', REDEEM_CODE_USED: 'used',
    WALLET_RESET_NO_SPEND: 'noSpend', REDEEM_CODE_UNSUPPORTED_TYPE: 'notCard', WALLET_RESET_CARD_GROUP: 'groupUnavailable',
  } as Record<string, string>)[reason]
  message.value = t(`walletReset.${key || fallback}`)
}

async function previewCard() {
  if (!code.value || busy.value) return
  busy.value = true
  message.value = ''
  preview.value = null
  try {
    preview.value = await redeemAPI.previewResetCard(code.value)
  } catch (error) {
    showError(error, 'previewFailed')
  } finally {
    busy.value = false
  }
}

async function applyCard() {
  if (!preview.value || preview.value.credit_usd <= 0 || busy.value) return
  busy.value = true
  message.value = ''
  try {
    const result = await redeemAPI.applyResetCard(code.value)
    failed.value = false
    message.value = t('walletReset.applied', { amount: money(result.amount_usd) })
    code.value = ''
    preview.value = null
    await loadCredits()
  } catch (error) {
    preview.value = null
    showError(error, 'applyFailed')
  } finally {
    busy.value = false
  }
}

onMounted(loadCredits)
</script>
