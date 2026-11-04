<template>
  <button
    type="button"
    class="btn btn-ghost"
    :title="isModern ? 'Switch to classic style' : 'Switch to modern style'"
    @click="onToggle"
  >
    {{ isModern ? 'Classic style' : 'Modern style' }}
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { getUiStyle, toggleUiStyle } from '@/utils/uiStyle'

// Read once at setup; toggling reloads the whole page (spec behaviour),
// so no reactive store is needed here.
const isModern = computed(() => getUiStyle() === 'modern')

function onToggle() {
  toggleUiStyle()
}
</script>
