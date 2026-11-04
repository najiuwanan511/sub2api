<template>
  <button
    type="button"
    :class="pinned ? 'btn btn-primary shadow-lg' : 'btn btn-ghost'"
    :title="isModern ? labelSwitchClassic : labelSwitchModern"
    @click="onToggle"
  >
    {{ isModern ? labelClassic : labelModern }}
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getUiStyle, toggleUiStyle } from '@/utils/uiStyle'

// "pinned" renders a prominent solid button (used for the fixed switch that
// the classic shell mounts, where a ghost button is too easy to miss).
withDefaults(defineProps<{ pinned?: boolean }>(), { pinned: false })

const { locale } = useI18n()

// Read once at setup; toggling reloads the whole page (spec behaviour),
// so no reactive store is needed here.
const isModern = computed(() => getUiStyle() === 'modern')
// The app uses locale codes 'en' and 'zh' (src/i18n/index.ts), so match by
// prefix instead of an exact 'zh-CN' that never occurs.
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))

const labelClassic = computed(() => (isZh.value ? '经典风格' : 'Classic style'))
const labelModern = computed(() => (isZh.value ? '新版风格' : 'Modern style'))
const labelSwitchClassic = computed(() =>
  isZh.value ? '切换到经典风格' : 'Switch to classic style'
)
const labelSwitchModern = computed(() =>
  isZh.value ? '切回新版风格' : 'Switch to modern style'
)

function onToggle() {
  toggleUiStyle()
}
</script>
