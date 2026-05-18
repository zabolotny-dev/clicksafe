<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import IconButton from '../components/ui/IconButton.vue'
import UiButton from '../components/ui/UiButton.vue'
import { useSession } from '../composables/useSession'
import { setLocale, supportedLocales } from '../i18n'
import { navigationGroups } from '../router/navigation'

const SIDEBAR_COLLAPSED_KEY = 'clicksafe.sidebar.collapsed'

const route = useRoute()
const router = useRouter()
const { session, isLoading, logout } = useSession()
const { t, locale } = useI18n()
const sidebarCollapsed = ref(false)

const currentTitle = computed(() => (
  route.meta.titleKey ? t(route.meta.titleKey) : t('app.adminPanel')
))
const currentDescription = computed(() => (
  route.meta.descriptionKey ? t(route.meta.descriptionKey) : t('app.securitySimulationConsole')
))
const currentSection = computed(() => (
  route.meta.sectionKey ? t(route.meta.sectionKey) : t('sections.settings')
))
const userName = computed(() => {
  const firstName = session.value?.first_name ?? ''
  const lastName = session.value?.last_name ?? ''
  return `${firstName} ${lastName}`.trim() || t('layout.administrator')
})
const userInitials = computed(() => {
  const names = userName.value.split(' ').filter(Boolean)
  return names.slice(0, 2).map((name) => name[0]).join('').toUpperCase() || 'A'
})
const sidebarToggleLabel = computed(() => (
  sidebarCollapsed.value ? t('layout.expandSidebar') : t('layout.collapseSidebar')
))
const currentLocale = computed(() => supportedLocales.includes(locale.value) ? locale.value : 'ru')
const nextLocale = computed(() => currentLocale.value === 'ru' ? 'en' : 'ru')
const collapsedLanguageLabel = computed(() => (
  nextLocale.value === 'en'
    ? t('layout.language.switchToEnglish')
    : t('layout.language.switchToRussian')
))

function languageLabel(value) {
  return String(value).toUpperCase()
}

function readSidebarPreference() {
  if (typeof window === 'undefined') {
    return
  }

  sidebarCollapsed.value = window.localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true'
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

function changeLocale(value) {
  setLocale(value)
}

function toggleLocale() {
  changeLocale(nextLocale.value)
}

async function signOut() {
  await logout()
  await router.push({ name: 'login' })
}

onMounted(() => {
  readSidebarPreference()
})

watch(sidebarCollapsed, (value) => {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(value))
})
</script>

<template>
  <div class="admin-shell" :class="{ 'admin-shell--sidebar-collapsed': sidebarCollapsed }">
    <aside class="admin-sidebar" :aria-label="t('layout.mainNavigation')">
      <div class="sidebar-header">
        <RouterLink
          class="brand-lockup"
          to="/dashboard"
          :aria-label="t('layout.dashboardTitle')"
          :title="sidebarCollapsed ? t('layout.dashboardTitle') : undefined"
        >
          <span class="brand-copy">
            <strong>ClickSafe</strong>
          </span>
        </RouterLink>
        <IconButton
          class="sidebar-collapse-toggle"
          :label="sidebarToggleLabel"
          @click="toggleSidebar"
        >
          <ChevronRight v-if="sidebarCollapsed" :size="16" stroke-width="1.9" aria-hidden="true" />
          <ChevronLeft v-else :size="16" stroke-width="1.9" aria-hidden="true" />
        </IconButton>
      </div>

      <nav class="sidebar-nav">
        <section
          v-for="(group, groupIndex) in navigationGroups"
          :key="`${group.labelKey || 'main'}-${groupIndex}`"
          class="nav-group"
        >
          <p v-if="group.labelKey" class="nav-group-label">
            {{ t(group.labelKey) }}
          </p>

          <RouterLink
            v-for="item in group.items"
            :key="item.to"
            class="nav-link"
            :to="item.to"
            :title="sidebarCollapsed ? t(item.labelKey) : undefined"
            :aria-label="sidebarCollapsed ? t(item.labelKey) : undefined"
          >
            <component :is="item.icon" class="nav-link-icon" aria-hidden="true" />
            <span class="nav-link-label">{{ t(item.labelKey) }}</span>
          </RouterLink>
        </section>
      </nav>

      <div class="sidebar-footer">
        <div v-if="!sidebarCollapsed" class="sidebar-language">
          <span class="sidebar-language-label">{{ t('layout.language.label') }}</span>
          <div class="sidebar-language-segmented" :aria-label="t('layout.language.aria')">
            <button
              v-for="item in supportedLocales"
              :key="item"
              type="button"
              class="sidebar-language-option"
              :class="{ 'is-active': item === currentLocale }"
              :aria-pressed="item === currentLocale"
              @click="changeLocale(item)"
            >
              {{ languageLabel(item) }}
            </button>
          </div>
        </div>
        <button
          v-else
          type="button"
          class="sidebar-language-collapsed"
          :title="collapsedLanguageLabel"
          :aria-label="collapsedLanguageLabel"
          @click="toggleLocale"
        >
          {{ languageLabel(currentLocale) }}
        </button>
      </div>
    </aside>

    <div class="admin-workspace">
      <header class="topbar">
        <div class="topbar-copy">
          <span>{{ currentSection }}</span>
          <strong>{{ currentTitle }}</strong>
          <p>{{ currentDescription }}</p>
        </div>
        <div class="account-area" :aria-label="t('layout.currentSession')">
          <span class="account-avatar" aria-hidden="true">{{ userInitials }}</span>
          <span class="account-copy">
            <span>{{ t('layout.online') }}</span>
            <strong>{{ userName }}</strong>
          </span>
          <UiButton variant="ghost" :loading="isLoading" @click="signOut">
            {{ t('common.actions.signOut') }}
          </UiButton>
        </div>
      </header>

      <main class="page-canvas" aria-live="polite">
        <RouterView />
      </main>
    </div>
  </div>
</template>
