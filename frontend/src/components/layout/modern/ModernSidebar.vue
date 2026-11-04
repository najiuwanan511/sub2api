<template>
  <!-- DESIGN_1 modern sidebar (user shell): grouped navigation on the same
       background as the page, no border. Current item is bold money-text with
       an 8px square marker before it. Bottom row: dark mode, classic style,
       collapse. Admin users keep the upstream sidebar in ModernLayout. -->
  <aside
    class="fixed inset-y-0 left-0 z-40 flex flex-col bg-surface transition-transform duration-300 lg:translate-x-0"
    :class="[collapsed ? 'w-[72px]' : 'w-64', appStore.mobileOpen ? 'translate-x-0' : '-translate-x-full']"
    :aria-label="isZh ? '主导航' : 'Main navigation'"
  >
    <!-- Brand -->
    <div class="flex h-16 shrink-0 items-center gap-3 px-4" :class="collapsed ? 'justify-center px-0' : ''">
      <div class="flex h-9 w-9 shrink-0 items-center justify-center bg-money text-money-ink">
        <img v-if="siteLogo" :src="siteLogo" class="h-full w-full object-contain" :alt="siteName">
        <Icon v-else name="bolt" size="md" />
      </div>
      <span v-if="!collapsed" class="truncate text-lg font-semibold text-ink">{{ siteName }}</span>
    </div>

    <!-- Grouped navigation -->
    <nav class="flex-1 overflow-y-auto px-3 pb-4">
      <div v-for="group in visibleGroups" :key="group.label" class="mt-6 first:mt-0">
        <p v-if="!collapsed" class="mb-1.5 px-[17px] text-[13px] leading-none text-ink-3">{{ group.label }}</p>
        <div v-else class="mx-2 mb-2 h-px bg-rule" />
        <router-link
          v-for="item in group.items"
          :key="item.path"
          :to="item.path"
          class="nav-item flex items-center gap-3 py-2 pl-[17px] pr-3 text-[15px] text-ink-2 transition-colors hover:text-ink"
          :class="{ 'nav-item-active': isActive(item.path), 'justify-center px-0': collapsed }"
          :title="collapsed ? item.label : undefined"
        >
          <Icon :name="(item.icon as any)" size="md" class="shrink-0" />
          <span v-if="!collapsed" class="truncate">{{ item.label }}</span>
        </router-link>
      </div>
    </nav>

    <!-- Bottom: dark mode / classic style / collapse -->
    <div class="shrink-0 px-3 pb-4">
      <div class="mx-2 mb-2 h-px bg-rule" />
      <button
        type="button"
        class="nav-item flex w-full items-center gap-3 py-2 pl-[17px] pr-3 text-[15px] text-ink-2 transition-colors hover:text-ink"
        :class="collapsed ? 'justify-center px-0' : ''"
        :title="collapsed ? darkLabel : undefined"
        @click="toggleDark"
      >
        <Icon :name="isDark ? 'sun' : 'moon'" size="md" class="shrink-0" />
        <span v-if="!collapsed">{{ darkLabel }}</span>
      </button>
      <button
        type="button"
        class="nav-item flex w-full items-center gap-3 py-2 pl-[17px] pr-3 text-[15px] text-ink-2 transition-colors hover:text-ink"
        :class="collapsed ? 'justify-center px-0' : ''"
        :title="collapsed ? classicLabel : undefined"
        @click="toggleUiStyle()"
      >
        <Icon name="sparkles" size="md" class="shrink-0" />
        <span v-if="!collapsed">{{ classicLabel }}</span>
      </button>
      <button
        type="button"
        class="nav-item flex w-full items-center gap-3 py-2 pl-[17px] pr-3 text-[15px] text-ink-2 transition-colors hover:text-ink"
        :class="collapsed ? 'justify-center px-0' : ''"
        :title="collapsed ? collapseLabel : undefined"
        @click="appStore.toggleSidebar()"
      >
        <Icon :name="collapsed ? 'chevronRight' : 'chevronLeft'" size="md" class="shrink-0" />
        <span v-if="!collapsed">{{ collapseLabel }}</span>
      </button>
    </div>
  </aside>
  <!-- Mobile drawer overlay: 50% black, no blur (spec) -->
  <div
    v-if="appStore.mobileOpen"
    class="fixed inset-0 z-30 bg-black/50 lg:hidden"
    @click="appStore.setMobileOpen(false)"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { toggleUiStyle } from '@/utils/uiStyle'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { sanitizeUrl } from '@/utils/url'

interface NavEntry {
  path: string
  label: string
  icon: string
}

const route = useRoute()
const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const isZh = computed(() => locale.value.startsWith('zh'))
const collapsed = computed(() => appStore.sidebarCollapsed)
const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))

const isDark = ref(false)
onMounted(() => {
  isDark.value = document.documentElement.classList.contains('dark')
})

function toggleDark() {
  const next = !isDark.value
  isDark.value = next
  document.documentElement.classList.toggle('dark', next)
  try {
    localStorage.setItem('theme', next ? 'dark' : 'light')
  } catch {
    // Best effort only (private mode).
  }
}

const darkLabel = computed(() => (isDark.value ? (isZh.value ? '浅色模式' : 'Light mode') : isZh.value ? '深色模式' : 'Dark mode'))
const classicLabel = computed(() => (isZh.value ? '经典风格' : 'Classic style'))
const collapseLabel = computed(() => (collapsed.value ? (isZh.value ? '展开' : 'Expand') : isZh.value ? '收起' : 'Collapse'))

const flagPayment = computed(() => isFeatureFlagEnabled(FeatureFlags.payment))
const flagSubscription = computed(() => isFeatureFlagEnabled(FeatureFlags.subscription))
const flagChannelMonitor = computed(() => isFeatureFlagEnabled(FeatureFlags.channelMonitor))
const flagAvailableChannels = computed(() => isFeatureFlagEnabled(FeatureFlags.availableChannels))
const simpleMode = computed(() => authStore.isSimpleMode)

// Grouped items mirror the upstream AppSidebar routes (user side only),
// grouped per DESIGN_1: 常用 / 钱 / 查看 / 账户.
const visibleGroups = computed(() => {
  const common: NavEntry[] = [
    { path: '/dashboard', label: t('nav.dashboard'), icon: 'home' },
    { path: '/keys', label: t('nav.apiKeys'), icon: 'key' },
    { path: '/usage', label: t('nav.usage'), icon: 'chart' }
  ]
  const money: NavEntry[] = [
    { path: '/purchase', label: isZh.value ? '充值' : 'Recharge', icon: 'dollar' },
    { path: '/subscriptions', label: t('nav.mySubscriptions'), icon: 'creditCard' },
    { path: '/orders', label: t('nav.myOrders'), icon: 'clipboard' },
    { path: '/redeem', label: t('nav.redeem'), icon: 'gift' }
  ]
  const view: NavEntry[] = [
    { path: '/available-channels', label: t('nav.availableChannels'), icon: 'cloud' },
    { path: '/monitor', label: t('nav.channelStatus'), icon: 'server' }
  ]
  const account: NavEntry[] = [{ path: '/profile', label: t('nav.profile'), icon: 'user' }]

  return [
    { label: isZh.value ? '常用' : 'Common', items: common.filter((item) => item.path !== '/usage' || !simpleMode.value) },
    {
      label: isZh.value ? '钱' : 'Money',
      items: money.filter((item) => {
        if (simpleMode.value) return false
        if (item.path === '/purchase' || item.path === '/orders') return flagPayment.value
        if (item.path === '/subscriptions') return flagSubscription.value
        return true
      })
    },
    {
      label: isZh.value ? '查看' : 'View',
      items: view.filter((item) => {
        if (item.path === '/available-channels') return flagAvailableChannels.value && !simpleMode.value
        if (item.path === '/monitor') return flagChannelMonitor.value
        return true
      })
    },
    { label: isZh.value ? '账户' : 'Account', items: account }
  ].filter((group) => group.items.length > 0)
})

function isActive(path: string): boolean {
  if (path === '/dashboard') return route.path === '/dashboard'
  return route.path === path || route.path.startsWith(`${path}/`)
}
</script>

<style scoped>
.nav-item-active {
  color: rgb(var(--c-money-text));
  font-weight: 600;
  position: relative;
  padding-left: 26px;
}

/* Current item marker: 8px solid square, 3px from the left (spec). */
.nav-item-active::before {
  content: '';
  position: absolute;
  left: 3px;
  top: 50%;
  transform: translateY(-50%);
  width: 8px;
  height: 8px;
  background: currentColor;
}

/* Collapsed: no room for the square marker, keep the item centred. */
.nav-item-active.justify-center {
  padding-left: 0;
}
.nav-item-active.justify-center::before {
  content: none;
}
</style>
