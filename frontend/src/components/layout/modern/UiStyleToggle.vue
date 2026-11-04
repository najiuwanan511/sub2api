<template>
  <button
    type="button"
    class="btn btn-ghost"
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

const { locale } = useI18n()

// Read once at setup; toggling reloads the whole page (spec behaviour),
// so no reactive store is needed here.
const isModern = computed(() => getUiStyle() === 'modern')
const isZh = computed(() => locale.value === 'zh-CN')

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
