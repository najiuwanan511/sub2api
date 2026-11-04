<template>
  <!-- DESIGN_1 modern header (user shell): flat, same background as the page,
       no glass/blur. Left: page name. Right: announcements, language, money
       balance pill (inverse hover card), square avatar. -->
  <header class="sticky top-0 z-30 bg-surface">
    <div class="flex h-16 items-center justify-between gap-2 px-4 lg:h-[76px] lg:px-6">
      <!-- Left: mobile menu + page title -->
      <div class="flex min-w-0 shrink-0 items-center gap-3">
        <button
          type="button"
          class="btn-ghost btn-icon lg:hidden"
          :aria-label="t('common.toggleMenu')"
          @click="appStore.toggleMobileSidebar()"
        >
          <Icon name="menu" size="md" />
        </button>
        <div class="min-w-0">
          <h1 class="truncate text-[15px] leading-tight text-ink-3">{{ pageTitle }}</h1>
          <p v-if="pageDescription" class="truncate text-[13px] leading-tight text-ink-3">{{ pageDescription }}</p>
        </div>
      </div>

      <!-- Right: bell, language, balance pill, avatar -->
      <div class="flex min-w-0 items-center gap-2 sm:gap-3">
        <AnnouncementBell v-if="user" />

        <LocaleSwitcher />

        <!-- Balance pill: money solid block, inverse hover card (spec) -->
        <div v-if="user" class="group relative hidden items-center sm:flex">
          <router-link
            to="/purchase"
            class="flex h-10 items-center px-4 text-sm font-semibold text-money-ink transition-colors hover:bg-money-hover"
            :aria-label="balanceAriaLabel"
          >
            {{ formattedBalance }}
          </router-link>
          <div
            class="pointer-events-none absolute right-0 top-full z-40 mt-2 hidden w-56 p-4 text-xs group-hover:block"
            style="background-color: rgb(var(--c-ink)); color: rgb(var(--c-surface))"
          >
            <div class="flex items-center justify-between">
              <span class="opacity-70">{{ balanceAvailableText }}</span>
              <span>{{ formattedBalance }}</span>
            </div>
            <div v-if="frozenBalance > 0" class="mt-2 flex items-center justify-between">
              <span class="opacity-70">{{ balanceFrozenText }}</span>
              <span>{{ formatMoney(frozenBalance) }}</span>
            </div>
            <div class="mt-2 border-t pt-2" style="border-color: rgb(var(--c-surface) / 0.2)">
              <div class="flex items-center justify-between">
                <span class="opacity-70">{{ balanceTotalText }}</span>
                <span class="font-semibold">{{ formatMoney(totalBalance) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Avatar block (square, inverse) + dropdown -->
        <div v-if="user" ref="dropdownRef" class="relative">
          <button
            type="button"
            class="flex items-center gap-2 p-1.5 transition-colors hover:bg-tile"
            :aria-label="t('common.userMenu')"
            @click="dropdownOpen = !dropdownOpen"
          >
            <div class="flex h-10 w-10 items-center justify-center overflow-hidden text-sm font-semibold" style="background-color: rgb(var(--c-ink)); color: rgb(var(--c-surface))">
              <img v-if="avatarUrl" :src="avatarUrl" :alt="displayName" class="h-full w-full object-cover">
              <span v-else>{{ userInitials }}</span>
            </div>
            <div class="hidden text-left md:block">
              <div class="text-sm font-medium text-ink">{{ displayName }}</div>
              <div class="text-xs text-ink-3">{{ roleLabel }}</div>
            </div>
            <Icon name="chevronDown" size="sm" class="hidden text-ink-3 md:block" />
          </button>

          <!-- Dropdown: page background + 1px rule outline (interactive overlay, spec) -->
          <transition name="dropdown">
            <div
              v-if="dropdownOpen"
              class="absolute right-0 z-40 mt-2 w-56 py-1"
              style="background-color: rgb(var(--c-surface)); box-shadow: 0 0 0 1px rgb(var(--c-rule))"
            >
              <div class="px-4 py-3" style="box-shadow: 0 1px 0 rgb(var(--c-rule))">
                <div class="text-sm font-medium text-ink">{{ displayName }}</div>
                <div class="text-xs text-ink-3">{{ user.email }}</div>
              </div>
              <router-link to="/profile" class="modern-dropdown-item" @click="dropdownOpen = false">
                <Icon name="user" size="sm" />
                {{ t('nav.profile') }}
              </router-link>
              <router-link to="/keys" class="modern-dropdown-item" @click="dropdownOpen = false">
                <Icon name="key" size="sm" />
                {{ t('nav.apiKeys') }}
              </router-link>
              <div class="mt-1" style="box-shadow: 0 -1px 0 rgb(var(--c-rule))"></div>
              <button type="button" class="modern-dropdown-item w-full !text-bad" @click="handleLogout">
                <Icon name="login" size="sm" />
                {{ t('nav.logout') }}
              </button>
            </div>
          </transition>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { resolveRouteMetaKeys } from '@/router/title'
import { resolveSiteBillingMode } from '@/utils/siteBillingMode'

const route = useRoute()
const router = useRouter()
const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const user = computed(() => authStore.user)
const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const isZh = computed(() => locale.value.startsWith('zh'))

const availableBalance = computed(() => Number(user.value?.balance || 0))
const frozenBalance = computed(() => Number(user.value?.frozen_balance || 0))
const totalBalance = computed(() => availableBalance.value + frozenBalance.value)

const balanceAriaLabel = computed(() => t('common.balance'))
const balanceAvailableText = computed(() => t('common.availableBalance') === 'common.availableBalance' ? (isZh.value ? '可用余额' : 'Available') : t('common.availableBalance'))
const balanceFrozenText = computed(() => t('common.frozenBalance') === 'common.frozenBalance' ? (isZh.value ? '冻结金额' : 'Frozen') : t('common.frozenBalance'))
const balanceTotalText = computed(() => t('common.totalBalance') === 'common.totalBalance' ? (isZh.value ? '总余额' : 'Total') : t('common.totalBalance'))

function formatMoney(value: number): string {
  const fixed = Number(value || 0).toFixed(2)
  const [int, frac] = fixed.split('.')
  const grouped = Number(int).toLocaleString('en-US')
  return frac === '00' ? `$${grouped}` : `$${grouped}.${frac}`
}

const formattedBalance = computed(() => formatMoney(availableBalance.value))

const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')

const userInitials = computed(() => {
  if (!user.value) return ''
  if (user.value.username) return user.value.username.substring(0, 2).toUpperCase()
  if (user.value.email) return user.value.email.split('@')[0].substring(0, 2).toUpperCase()
  return ''
})

const displayName = computed(() => user.value?.username || user.value?.email?.split('@')[0] || '')

const roleLabel = computed(() => (isZh.value ? '用户' : 'User'))

// Page title/description follow the same resolution as the classic header,
// including the /purchase billing-mode-aware wording.
const routeMetaKeys = computed(() =>
  resolveRouteMetaKeys(route, {
    billingMode: resolveSiteBillingMode(appStore.cachedPublicSettings),
  })
)

const pageTitle = computed(() => {
  const titleKey = routeMetaKeys.value.titleKey
  if (titleKey) return t(titleKey)
  return (route.meta.title as string) || ''
})

const pageDescription = computed(() => {
  const descKey = routeMetaKeys.value.descriptionKey
  if (descKey) return t(descKey)
  return (route.meta.description as string) || ''
})

function onClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    dropdownOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))

async function handleLogout() {
  dropdownOpen.value = false
  try {
    await authStore.logout()
  } catch (error) {
    console.error('Logout error:', error)
  }
  void router.push('/login')
}
</script>

<style scoped>
.modern-dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px 16px;
  font-size: 14px;
  color: rgb(var(--c-ink-2));
  transition: background-color 150ms, color 150ms;
}
.modern-dropdown-item:hover {
  background-color: rgb(var(--c-tile));
  color: rgb(var(--c-ink));
}
</style>
