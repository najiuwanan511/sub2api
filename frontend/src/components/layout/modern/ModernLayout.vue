<template>
  <!-- Modern shell (DESIGN_1): flat, no mesh gradient, no blur, no radius.
       Regular users get the modern grouped sidebar + flat header; admins keep
       the upstream shell (AppSidebar/AppHeader) restyled by the modern CSS
       layer, per the design doc ("管理后台沿用上游的布局"). -->
  <div class="min-h-screen bg-surface text-ink">
    <template v-if="useModernShell">
      <ModernSidebar />
      <div
        class="relative min-h-screen transition-all duration-300"
        :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
      >
        <ModernHeader />
        <main class="mx-auto w-full max-w-[1320px] p-4 md:p-6 lg:p-8">
          <slot />
        </main>
      </div>
    </template>
    <template v-else>
      <AppSidebar />
      <div
        class="relative min-h-screen transition-all duration-300"
        :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
      >
        <AppHeader />
        <main class="mx-auto w-full max-w-[1320px] p-4 md:p-6 lg:p-8">
          <slot />
        </main>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import AppSidebar from '../AppSidebar.vue'
import AppHeader from '../AppHeader.vue'
import ModernSidebar from './ModernSidebar.vue'
import ModernHeader from './ModernHeader.vue'

const appStore = useAppStore()
const authStore = useAuthStore()

const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)

// The modern shell is the user-facing one; admins keep the upstream shell so
// no admin navigation is lost (upstream layout, modern colours/radius/fonts).
const useModernShell = computed(() => authStore.user?.role !== 'admin')
</script>
