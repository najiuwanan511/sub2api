<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AccountChannelProbeResult } from '@/api/admin/accounts'
import type { AccountListItem } from '@/types'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ account: AccountListItem }>()
const { t } = useI18n()
const prefix = 'admin.accounts.channelProbe'
const result = ref<AccountChannelProbeResult | null>(null)
const pending = ref(false)
const failed = ref(false)
const base = computed(() => props.account.credentials?.base_url)
const supported = computed(() =>
  (props.account.type === 'apikey' || props.account.type === 'upstream') &&
  typeof base.value === 'string' && base.value.trim() !== '',
)
let generation = 0
watch([() => props.account.id, () => props.account.type, base], () => {
  generation++
  result.value = null
  pending.value = false
  failed.value = false
})

async function run() {
  if (!supported.value || pending.value) return
  const request = ++generation
  pending.value = true
  failed.value = false
  result.value = null
  try {
    const data = await adminAPI.accounts.probeChannel(props.account.id)
    if (request === generation) result.value = data
  } catch {
    if (request === generation) failed.value = true
  } finally {
    if (request === generation) pending.value = false
  }
}

const label = computed(() => {
  if (pending.value) return t(`${prefix}.running`)
  if (failed.value) return t(`${prefix}.requestFailed`)
  if (!result.value) return t(`${prefix}.run`)
  if (result.value.status !== 'reachable') return t(`${prefix}.${result.value.status}`)
  return t(`${prefix}.reachable`, { ms: result.value.latency_ms })
})
const detail = computed(() => [
  t(`${prefix}.description`),
  result.value?.http_status ? `HTTP ${result.value.http_status}` : '',
  result.value ? formatDateTime(result.value.checked_at) : '',
].filter(Boolean).join(' · '))
</script>

<template>
  <div v-if="supported" class="flex items-center gap-1.5 whitespace-nowrap text-xs">
    <button
      type="button"
      class="rounded-md px-1.5 py-1 text-primary-700 hover:bg-primary-50 disabled:cursor-wait disabled:opacity-60 dark:text-primary-300 dark:hover:bg-primary-900/20"
      :class="{
        '!text-emerald-700 dark:!text-emerald-300': result?.status === 'reachable',
        '!text-amber-700 dark:!text-amber-300': failed || (result && result.status !== 'reachable'),
      }"
      :aria-label="result || failed ? t(`${prefix}.retry`, { label }) : label"
      :title="detail"
      :disabled="pending"
      @click="run"
    >
      {{ label }}
    </button>
    <span v-if="result?.http_status" class="text-gray-500 dark:text-gray-400" :title="detail">HTTP {{ result.http_status }}</span>
  </div>
  <span v-else class="text-xs text-gray-400" :title="t(`${prefix}.unsupported`)">—</span>
</template>
