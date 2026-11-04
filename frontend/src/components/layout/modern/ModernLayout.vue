<template>
  <!-- Modern shell (DESIGN_1): flat, no mesh gradient, no blur, no radius.
       Navigation itself is still AppSidebar/AppHeader, so menu items, auth
       state and onboarding behaviour are unchanged; the modern CSS layer
       flattens their rounded/shadow/gradient classes. -->
  <div class="min-h-screen bg-surface text-ink">
    <AppSidebar />
    <div
      class="relative min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <AppHeader />
      <main class="mx-auto w-full max-w-[1320px] p-4 md:p-6 lg:p-8">
        <div class="mb-4 flex justify-end">
          <UiStyleToggle />
        </div>
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores'
import AppSidebar from '../AppSidebar.vue'
import AppHeader from '../AppHeader.vue'
import UiStyleToggle from './UiStyleToggle.vue'

const appStore = useAppStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
</script>
