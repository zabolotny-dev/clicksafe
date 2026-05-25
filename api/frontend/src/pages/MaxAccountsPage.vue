<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { Cable, CircleOff, KeyRound, Pencil, Plus, RefreshCw, Trash2 } from 'lucide-vue-next'
import IconButton from '../components/ui/IconButton.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import ResourceDrawer from '../components/ui/ResourceDrawer.vue'
import ResourcePagination from '../components/ui/ResourcePagination.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiInput from '../components/ui/UiInput.vue'
import { useNotifications } from '../composables/useNotifications'
import { useResourceActions } from '../composables/useResourceActions'
import {
  beginMaxLogin,
  confirmMaxLogin,
  confirmMaxPassword,
  connectMaxAccount,
  deleteMaxAccount,
  disconnectMaxAccount,
  listMaxAccounts,
  updateMaxAccount,
} from '../resources/maxAccounts'
import { errorMessage, formatDateTime } from '../utils/resourceFormat'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()

const STATUS_LABEL_KEYS = {
  ACTIVE: 'pages.maxAccounts.status.active',
  CONNECTED: 'pages.maxAccounts.status.connected',
  DISCONNECTED: 'pages.maxAccounts.status.disconnected',
  ERROR: 'pages.maxAccounts.status.error',
}

const accounts = ref([])
const page = ref(1)
const rows = ref(10)
const total = ref(0)
const isLoading = ref(false)
const loadError = ref('')
const pendingAction = ref('')
const formError = ref('')
const editDrawerOpen = ref(false)
const isSaving = ref(false)
const selectedAccountRecord = ref(null)
const attempt = ref(null)
const codeDigits = ref(Array.from({ length: 6 }, () => ''))
const codeInputRefs = ref([])
const remainingSeconds = ref(null)
let countdownTimer = null
const loginForm = reactive({
  phone: '',
  label: '',
  code: '',
  password: '',
})
const filters = reactive({
  label: '',
  phone: '',
})
const editForm = reactive({
  label: '',
})

const isCreateMode = computed(() => route.name === 'max-accounts-new')
const hasAccounts = computed(() => accounts.value.length > 0)
const hasActiveFilters = computed(() => Boolean(filters.label.trim() || filters.phone.trim()))
const emptyTableText = computed(() => hasActiveFilters.value
  ? t('pages.maxAccounts.emptyFilteredDescription')
  : t('pages.maxAccounts.emptyDescription'))
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / rows.value)))
const needsPassword = computed(() => attempt.value?.status === 'PASSWORD_REQUIRED')
const needsCode = computed(() => attempt.value?.status === 'CODE_REQUIRED')
const codeValue = computed(() => codeDigits.value.join(''))
const isCodeComplete = computed(() => codeDigits.value.every(Boolean))
const isCodeExpired = computed(() => remainingSeconds.value === 0)
const countdownLabel = computed(() => {
  if (remainingSeconds.value == null) {
    return ''
  }

  if (remainingSeconds.value <= 0) {
    return t('pages.maxAccounts.codeExpired')
  }

  const minutes = Math.floor(remainingSeconds.value / 60)
  const seconds = remainingSeconds.value % 60
  const value = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  return t('pages.maxAccounts.expiresCountdown', { value })
})
const createPanelTitle = computed(() => {
  if (needsCode.value) return t('pages.maxAccounts.codeTitle')
  if (needsPassword.value) return t('pages.maxAccounts.passwordTitle')
  return t('pages.maxAccounts.detailsTitle')
})
const createPanelDescription = computed(() => {
  if (needsCode.value) return t('pages.maxAccounts.codeDescription')
  if (needsPassword.value) return t('pages.maxAccounts.passwordDescription')
  return t('pages.maxAccounts.detailsDescription')
})

function statusVariant(status) {
  if (status === 'CONNECTED') return 'success'
  if (status === 'ACTIVE') return 'info'
  if (status === 'ERROR') return 'danger'
  if (status === 'DISCONNECTED') return 'warning'
  return 'neutral'
}

function displayStatus(status) {
  const key = STATUS_LABEL_KEYS[status]
  return key ? t(key) : (status || t('common.placeholder'))
}

function resetLoginSecretFields() {
  loginForm.code = ''
  loginForm.password = ''
  codeDigits.value = Array.from({ length: 6 }, () => '')
  codeInputRefs.value = []
}

function resetCreateFlow() {
  stopCountdown()
  pendingAction.value = ''
  formError.value = ''
  attempt.value = null
  loginForm.phone = ''
  loginForm.label = ''
  resetLoginSecretFields()
}

function setCodeInputRef(element, index) {
  if (element) {
    codeInputRefs.value[index] = element
  }
}

function syncCodeValue() {
  loginForm.code = codeValue.value
}

function focusCodeInput(index) {
  nextTick(() => {
    codeInputRefs.value[index]?.focus()
    codeInputRefs.value[index]?.select()
  })
}

function applyCodeDigits(value, startIndex = 0) {
  const digits = String(value || '').replace(/\D/g, '').slice(0, codeDigits.value.length - startIndex).split('')

  digits.forEach((digit, offset) => {
    codeDigits.value[startIndex + offset] = digit
  })

  syncCodeValue()

  const nextIndex = Math.min(startIndex + digits.length, codeDigits.value.length - 1)
  if (digits.length && nextIndex < codeDigits.value.length - 1) {
    focusCodeInput(nextIndex + 1)
  } else if (digits.length) {
    focusCodeInput(nextIndex)
  }
}

function handleCodeInput(event, index) {
  const value = event.target.value
  const digits = String(value || '').replace(/\D/g, '')

  if (digits.length > 1) {
    applyCodeDigits(digits, index)
    event.target.value = codeDigits.value[index]
    return
  }

  codeDigits.value[index] = digits.slice(0, 1)
  event.target.value = codeDigits.value[index]
  syncCodeValue()

  if (codeDigits.value[index] && index < codeDigits.value.length - 1) {
    focusCodeInput(index + 1)
  }
}

function handleCodeKeydown(event, index) {
  if (event.key === 'Backspace' && !codeDigits.value[index] && index > 0) {
    focusCodeInput(index - 1)
  }
}

function handleCodePaste(event, index) {
  const text = event.clipboardData?.getData('text')
  if (!text) {
    return
  }

  event.preventDefault()
  applyCodeDigits(text, index)
}

function stopCountdown() {
  if (countdownTimer) {
    window.clearInterval(countdownTimer)
    countdownTimer = null
  }
  remainingSeconds.value = null
}

function updateCountdown(expiresAt) {
  const expiresAtTime = new Date(expiresAt).getTime()
  if (Number.isNaN(expiresAtTime)) {
    remainingSeconds.value = null
    return
  }

  remainingSeconds.value = Math.max(0, Math.ceil((expiresAtTime - Date.now()) / 1000))

  if (remainingSeconds.value === 0 && countdownTimer) {
    window.clearInterval(countdownTimer)
    countdownTimer = null
  }
}

function startCountdown(expiresAt) {
  stopCountdown()

  if (!expiresAt) {
    return
  }

  updateCountdown(expiresAt)
  countdownTimer = window.setInterval(() => updateCountdown(expiresAt), 1000)
}

async function loadAccounts() {
  isLoading.value = true
  loadError.value = ''

  try {
    const response = await listMaxAccounts({
      ...filters,
      page: page.value,
      rows: rows.value,
    })
    accounts.value = Array.isArray(response?.items) ? response.items : []
    total.value = Number.isFinite(response?.total) ? response.total : accounts.value.length
    page.value = Number.isFinite(response?.page) ? response.page : page.value
    rows.value = Number.isFinite(response?.rowsPerPage) ? response.rowsPerPage : rows.value
  } catch (error) {
    if (await handleAuthError(error)) return
    loadError.value = errorMessage(error, 'errors.maxAccountsLoad')
  } finally {
    isLoading.value = false
  }
}

function applyFilters() {
  page.value = 1
  loadAccounts()
}

function clearFilters() {
  filters.label = ''
  filters.phone = ''
  page.value = 1
  loadAccounts()
}

function setPage(nextPage) {
  page.value = Math.min(Math.max(1, nextPage), totalPages.value)
  loadAccounts()
}

function setRows(nextRows) {
  rows.value = nextRows
  page.value = 1
  loadAccounts()
}

async function beginLogin() {
  if (!(await ensureCsrfToken())) return

  pendingAction.value = 'begin'
  formError.value = ''
  attempt.value = null
  resetLoginSecretFields()

  try {
    attempt.value = await beginMaxLogin({
      phone: loginForm.phone.trim(),
      label: loginForm.label.trim(),
    }, mutationOptions())
    loginForm.label = attempt.value?.label || loginForm.label.trim()
    notifySuccess(t('pages.maxAccounts.notifications.codeRequested'))
    if (attempt.value?.status === 'CODE_REQUIRED') {
      focusCodeInput(0)
    }
  } catch (error) {
    if (await handleAuthError(error)) return
    formError.value = errorMessage(error, 'errors.maxAccountLogin')
    notifyError(t('pages.maxAccounts.notifications.loginFailedTitle'), formError.value)
  } finally {
    pendingAction.value = ''
  }
}

async function confirmCode() {
  if (!attempt.value?.id || !(await ensureCsrfToken())) return

  pendingAction.value = 'confirm-code'
  formError.value = ''
  syncCodeValue()

  try {
    const result = await confirmMaxLogin(attempt.value.id, {
      code: codeValue.value,
      label: loginForm.label.trim(),
    }, mutationOptions())
    attempt.value = result.attempt
    if (result.account) {
      notifySuccess(t('pages.maxAccounts.notifications.accountAdded'))
      resetCreateFlow()
      await router.push({ name: 'max-accounts' })
    }
  } catch (error) {
    if (await handleAuthError(error)) return
    formError.value = errorMessage(error, 'errors.maxAccountLogin')
    notifyError(t('pages.maxAccounts.notifications.loginFailedTitle'), formError.value)
  } finally {
    pendingAction.value = ''
  }
}

async function confirmPassword() {
  if (!attempt.value?.id || !(await ensureCsrfToken())) return

  pendingAction.value = 'confirm-password'
  formError.value = ''

  try {
    const result = await confirmMaxPassword(attempt.value.id, {
      password: loginForm.password,
      label: loginForm.label.trim(),
    }, mutationOptions())
    attempt.value = result.attempt
    if (result.account) {
      notifySuccess(t('pages.maxAccounts.notifications.accountAdded'))
      resetCreateFlow()
      await router.push({ name: 'max-accounts' })
    }
  } catch (error) {
    if (await handleAuthError(error)) return
    formError.value = errorMessage(error, 'errors.maxAccountLogin')
    notifyError(t('pages.maxAccounts.notifications.loginFailedTitle'), formError.value)
  } finally {
    pendingAction.value = ''
  }
}

function openEditDrawer(account) {
  selectedAccountRecord.value = account
  editForm.label = account.label || ''
  editDrawerOpen.value = true
}

function closeEditDrawer(force = false) {
  const forced = force === true
  if (isSaving.value && !forced) {
    return
  }

  editDrawerOpen.value = false
  selectedAccountRecord.value = null
  editForm.label = ''
}

async function submitEdit() {
  if (!selectedAccountRecord.value?.id || !(await ensureCsrfToken())) return

  isSaving.value = true

  try {
    await updateMaxAccount(selectedAccountRecord.value.id, {
      label: editForm.label.trim(),
    }, mutationOptions())
    notifySuccess(t('pages.maxAccounts.notifications.updated'))
    closeEditDrawer(true)
    await loadAccounts()
  } catch (error) {
    if (await handleAuthError(error)) return
    notifyError(t('pages.maxAccounts.notifications.saveFailedTitle'), errorMessage(error, 'errors.maxAccountUpdate'))
  } finally {
    isSaving.value = false
  }
}

async function mutateAccount(action, account) {
  if (!(await ensureCsrfToken())) return

  pendingAction.value = `${action}:${account.id}`

  try {
    if (action === 'connect') {
      await connectMaxAccount(account.id, mutationOptions())
      notifySuccess(t('pages.maxAccounts.notifications.connected'))
    } else if (action === 'disconnect') {
      await disconnectMaxAccount(account.id, mutationOptions())
      notifySuccess(t('pages.maxAccounts.notifications.disconnected'))
    } else if (action === 'delete') {
      await deleteMaxAccount(account.id, mutationOptions())
      notifySuccess(t('pages.maxAccounts.notifications.deleted'))
    }
    await loadAccounts()
  } catch (error) {
    if (await handleAuthError(error)) return
    notifyError(t('pages.maxAccounts.notifications.actionFailedTitle'), errorMessage(error, 'errors.maxAccountAction'))
  } finally {
    pendingAction.value = ''
  }
}

watch(() => route.name, (name) => {
  if (name === 'max-accounts-new') {
    resetCreateFlow()
  } else if (name === 'max-accounts') {
    resetCreateFlow()
    loadAccounts()
  }
})

watch(() => attempt.value?.expires_at, (expiresAt) => {
  startCountdown(expiresAt)
})

onMounted(() => {
  if (isCreateMode.value) {
    resetCreateFlow()
  } else {
    loadAccounts()
  }
})

onBeforeUnmount(() => {
  stopCountdown()
})
</script>

<template>
  <section class="resource-page max-accounts-page">
    <template v-if="isCreateMode">
      <PageHeader
        :eyebrow="t('sections.settings')"
        :title="t('pages.maxAccounts.createTitle')"
        :description="t('pages.maxAccounts.createDescription')"
      >
        <template #actions>
          <RouterLink class="ui-button ui-button--secondary" :to="{ name: 'max-accounts' }">
            {{ t('pages.maxAccounts.backToList') }}
          </RouterLink>
        </template>
      </PageHeader>

      <section class="resource-editor-shell max-account-wizard">
        <section class="resource-editor-panel max-account-wizard-panel">
          <header class="resource-panel-header resource-panel-header--plain">
            <div>
              <h2>{{ createPanelTitle }}</h2>
              <p>{{ createPanelDescription }}</p>
            </div>
          </header>

          <section v-if="!attempt" class="max-account-wizard-step">
            <div class="resource-form-grid">
              <UiInput v-model="loginForm.phone" :label="t('common.labels.phone')" placeholder="+79990000000" autocomplete="tel" />
              <UiInput v-model="loginForm.label" :label="t('common.labels.name')" :placeholder="t('pages.maxAccounts.labelPlaceholder')" />
            </div>
          </section>

          <section v-else-if="needsCode" class="max-account-wizard-step max-account-code-step">
            <div class="max-account-code-panel">
              <span class="max-account-code-label">{{ t('pages.maxAccounts.code') }}</span>
              <div class="max-account-code-inputs" role="group" :aria-label="t('pages.maxAccounts.codeAria')">
                <input
                  v-for="(_, index) in codeDigits"
                  :key="index"
                  :ref="(element) => setCodeInputRef(element, index)"
                  class="max-account-code-input"
                  :value="codeDigits[index]"
                  type="text"
                  inputmode="numeric"
                  pattern="[0-9]*"
                  :maxlength="index === 0 ? 6 : 1"
                  :autocomplete="index === 0 ? 'one-time-code' : 'off'"
                  :aria-label="t('pages.maxAccounts.codeDigitAria', { number: index + 1 })"
                  @input="handleCodeInput($event, index)"
                  @keydown="handleCodeKeydown($event, index)"
                  @paste="handleCodePaste($event, index)"
                />
              </div>
              <p v-if="countdownLabel" class="max-account-countdown" :class="{ 'is-expired': isCodeExpired }">
                {{ countdownLabel }}
              </p>
            </div>
          </section>

          <section v-else-if="needsPassword" class="max-account-wizard-step">
            <div class="resource-form-grid resource-form-grid--single">
              <UiInput v-model="loginForm.password" :label="t('pages.maxAccounts.password')" type="password" autocomplete="current-password" />
            </div>
          </section>

        </section>

        <footer class="resource-editor-actions max-account-wizard-actions">
          <RouterLink class="ui-button ui-button--ghost" :to="{ name: 'max-accounts' }">
            {{ t('common.actions.cancel') }}
          </RouterLink>
          <UiButton v-if="!attempt" :loading="pendingAction === 'begin'" @click="beginLogin">
            <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('pages.maxAccounts.requestCode') }}
          </UiButton>
          <UiButton
            v-else-if="needsCode"
            :loading="pendingAction === 'confirm-code'"
            :disabled="!isCodeComplete || Boolean(pendingAction) || isCodeExpired"
            @click="confirmCode"
          >
            <KeyRound :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('common.actions.confirm') }}
          </UiButton>
          <UiButton
            v-else-if="needsPassword"
            :loading="pendingAction === 'confirm-password'"
            :disabled="!loginForm.password || Boolean(pendingAction)"
            @click="confirmPassword"
          >
            <KeyRound :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('common.actions.confirm') }}
          </UiButton>
        </footer>
      </section>
    </template>

    <template v-else>
      <PageHeader
        :eyebrow="t('sections.settings')"
        :title="t('pages.maxAccounts.title')"
        :description="t('pages.maxAccounts.description')"
      >
        <template #actions>
          <RouterLink class="ui-button ui-button--primary" :to="{ name: 'max-accounts-new' }">
            <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('pages.maxAccounts.addTitle') }}
          </RouterLink>
        </template>
      </PageHeader>

      <section class="resource-filter-panel resource-filter-panel--compact" :aria-label="t('pages.maxAccounts.filtersAria')">
        <UiInput v-model="filters.label" :label="t('common.labels.name')" :placeholder="t('pages.maxAccounts.searchPlaceholder')" />
        <UiInput v-model="filters.phone" :label="t('common.labels.phone')" :placeholder="t('common.labels.phone')" />
        <div class="resource-filter-actions">
          <UiButton variant="secondary" :loading="isLoading" @click="applyFilters">
            {{ t('common.filters.apply') }}
          </UiButton>
          <UiButton variant="ghost" :disabled="isLoading" @click="clearFilters">
            {{ t('common.actions.clear') }}
          </UiButton>
        </div>
      </section>

      <UiAlert
        v-if="loadError"
        variant="error"
        :title="t('pages.maxAccounts.unavailable')"
        :message="loadError"
      >
        <UiButton variant="secondary" @click="loadAccounts">{{ t('common.actions.retry') }}</UiButton>
      </UiAlert>

      <section class="resource-panel resource-panel--table">
        <header class="resource-panel-header">
          <div>
            <h2>{{ t('pages.maxAccounts.accountsTitle') }}</h2>
            <p>{{ t('common.loadedCount', { count: accounts.length }) }}</p>
          </div>
          <UiButton variant="secondary" :loading="isLoading" @click="loadAccounts">
            <RefreshCw :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('common.actions.refresh') }}
          </UiButton>
        </header>

        <div class="resource-table resource-table--compact resource-max-account-table" role="table" :aria-label="t('pages.maxAccounts.accountsTitle')">
          <div class="resource-table-row resource-table-head" role="row">
            <span>{{ t('common.labels.name') }}</span>
            <span>{{ t('common.labels.phone') }}</span>
            <span>{{ t('common.labels.status') }}</span>
            <span>{{ t('common.labels.createdAt') }}</span>
            <span>{{ t('pages.campaigns.columns.actions') }}</span>
          </div>
          <template v-if="isLoading">
            <div v-for="index in 4" :key="index" class="resource-table-row" role="row">
              <span v-for="cell in 5" :key="cell"><SkeletonBlock :rows="1" /></span>
            </div>
          </template>
          <template v-else-if="hasAccounts">
            <div
              v-for="account in accounts"
              :key="account.id"
              class="resource-table-row"
              role="row"
            >
              <span><strong>{{ account.label || t('common.placeholder') }}</strong></span>
              <span>{{ account.phone || t('common.placeholder') }}</span>
              <span>
                <UiBadge :variant="statusVariant(account.status)">{{ displayStatus(account.status) }}</UiBadge>
                <small v-if="account.last_error">{{ account.last_error }}</small>
              </span>
              <span>{{ formatDateTime(account.created_at) }}</span>
              <span class="resource-row-actions">
                <IconButton
                  variant="secondary"
                  :label="t('common.actions.edit')"
                  :disabled="!!pendingAction"
                  @click="openEditDrawer(account)"
                >
                  <Pencil :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
                <IconButton
                  v-if="account.status !== 'CONNECTED'"
                  variant="secondary"
                  :label="t('pages.maxAccounts.connect')"
                  :disabled="!!pendingAction"
                  @click="mutateAccount('connect', account)"
                >
                  <Cable :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
                <IconButton
                  v-else
                  variant="secondary"
                  :label="t('pages.maxAccounts.disconnect')"
                  :disabled="!!pendingAction"
                  @click="mutateAccount('disconnect', account)"
                >
                  <CircleOff :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
                <IconButton
                  variant="danger"
                  :label="t('common.actions.delete')"
                  :disabled="!!pendingAction"
                  @click="mutateAccount('delete', account)"
                >
                  <Trash2 :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
              </span>
            </div>
          </template>
          <div v-else class="resource-empty-row">
            {{ emptyTableText }}
          </div>
        </div>

        <ResourcePagination
          :page="page"
          :rows="rows"
          :total="total"
          :loaded-count="accounts.length"
          :loading="isLoading"
          @update:page="setPage"
          @update:rows="setRows"
        />
      </section>

      <ResourceDrawer
        :open="editDrawerOpen"
        :title="t('pages.maxAccounts.editTitle')"
        :subtitle="t('pages.maxAccounts.editDescription')"
        @close="closeEditDrawer"
      >
        <div class="resource-form-grid resource-form-grid--single">
          <UiInput
            v-model="editForm.label"
            :label="t('common.labels.name')"
            :placeholder="t('pages.maxAccounts.labelPlaceholder')"
          />
        </div>

        <template #footer>
          <UiButton variant="secondary" :disabled="isSaving" @click="closeEditDrawer">
            {{ t('common.actions.cancel') }}
          </UiButton>
          <UiButton :loading="isSaving" :disabled="!editForm.label.trim()" @click="submitEdit">
            {{ t('common.actions.saveChanges') }}
          </UiButton>
        </template>
      </ResourceDrawer>
    </template>
  </section>
</template>
