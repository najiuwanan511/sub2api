<template>
  <!-- DESIGN_1 Quick Action Tile: a whole clickable button tile, min-height
       104px, title 17px/600, hint 13px at 80% opacity, icon 24px. Variants
       mirror btn-primary / btn-solid / btn-secondary. Press scales to 0.985. -->
  <button
    type="button"
    class="quick-tile tile-glide flex min-h-[104px] w-full flex-col justify-center p-[18px] px-[20px] text-left transition-colors active:scale-[0.985]"
    :class="variantClass"
    :style="{ '--n': glideIndex }"
    :aria-label="title"
    @click="$emit('activate')"
  >
    <Icon :name="(icon as any)" size="lg" />
    <span class="mt-3 text-[17px] font-semibold leading-tight">{{ title }}</span>
    <span v-if="hint" class="mt-1 text-[13px] opacity-80 leading-snug">{{ hint }}</span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(
  defineProps<{
    title: string
    hint?: string
    icon?: string
    variant?: 'money' | 'solid' | 'tile'
    /** Stagger index for the tile-glide entrance (55ms apart, spec). */
    glideIndex?: number
  }>(),
  { icon: 'arrowRight', variant: 'tile', glideIndex: 0 }
)

defineEmits<{ (e: 'activate'): void }>()

const variantClass = computed(() => {
  switch (props.variant) {
    case 'money':
      return 'bg-money text-money-ink hover:bg-money-hover'
    case 'solid':
      return 'bg-ink text-surface hover:opacity-90'
    default:
      return 'bg-tile text-ink hover:bg-tile-2'
  }
})
</script>
