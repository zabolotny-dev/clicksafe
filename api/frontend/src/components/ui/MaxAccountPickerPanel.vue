<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Plus, RefreshCw } from 'lucide-vue-next'
import EmptyState from './EmptyState.vue'
import IconButton from './IconButton.vue'
import ResourcePagination from './ResourcePagination.vue'
import SkeletonBlock from './SkeletonBlock.vue'
import UiAlert from './UiAlert.vue'
import UiButton from './UiButton.vue'
import UiInput from './UiInput.vue'
import { useResourceActions } from '../../composables/useResourceActions'
import { listMaxAccounts } from '../../resources/maxAccounts'
import { errorMessage, formatTechnicalId, nullableId } from '../../utils/resourceFormat'

const ROW_OPTIONS = [5, 10, 25]

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  title: {
    type: String,
    required: true,
  },
  selectedTitle: {
    type: String,
    default: '',
  },
  selectedEmptyText: {
    type: String,
    default: '',
  },
  emptyTitle: {
    type: String,
    default: '',
  },
  emptyDescription: {
    type: String,
    default: '',
  },
  createTo: {
    type: Object,
    default: null,
  },
  createLabel: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['update:modelValue', 'selection-change'])

const { t } = useI18n()
const { handleAuthError } = useResourceActions()

const accounts = ref([])
const page = ref(1)
const rows = ref(5)
const total = ref(0)
const isLoading = ref(false)
const loadError = ref('')
const pickerName = `max-account-picker-${Math.random().toString(36).slice(2)}`
const filters = reactive({
  label: '',
  phone: '',
})

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / rows.value)))
const selectedAccount = computed(() => accounts.value.find((account) => accountId(account) === props.modelValue) || null)
const selectedSummary = computed(() => {
  if (selectedAccount.value) {
    return accountDisplay(selectedAccount.value)
  }

  return props.modelValue
    ? formatTechnicalId(props.modelValue)
    : (props.selectedEmptyText || t('common.selection.none'))
})
const resolvedCreateLabel = computed(() => props.createLabel || t('pages.maxAccounts.addTitle'))

function accountId(account) {
  return nullableId(account?.id) || (account?.id ? String(account.id) : '')
}

function accountDisplay(account) {
  return [account?.label, account?.phone].filter(Boolean).join(' · ') || formatTechnicalId(account?.id)
}

function emitSelectedFromAccounts() {
  emit('selection-change', selectedAccount.value)
}

function selectAccount(account) {
  const id = accountId(account)
  emit('update:modelValue', id)
  emit('selection-change', account)
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

    if (total.value > 0 && !accounts.value.length && page.value > 1) {
      page.value = Math.max(1, Math.ceil(total.value / rows.value))
      await loadAccounts()
      return
    }

    emitSelectedFromAccounts()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

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

watch(() => props.modelValue, emitSelectedFromAccounts)

onMounted(loadAccounts)
</script>

<template>
  <section class="attachment-picker-panel max-account-picker" :aria-label="title">
    <header class="attachment-picker-header">
      <div>
        <h3>{{ title }}</h3>
        <p>{{ t('common.loadedCount', { count: accounts.length }) }}</p>
      </div>
      <div class="attachment-picker-header-actions">
        <RouterLink
          v-if="createTo"
          class="ui-button ui-button--secondary"
          :to="createTo"
        >
          <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ resolvedCreateLabel }}
        </RouterLink>
        <IconButton :label="t('common.actions.refresh')" :disabled="isLoading" @click="loadAccounts">
          <RefreshCw :size="16" stroke-width="1.8" aria-hidden="true" />
        </IconButton>
      </div>
    </header>

    <section class="attachment-picker-toolbar max-account-picker-toolbar">
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

    <section class="attachment-picker-selection-bar">
      <div>
        <span>{{ selectedTitle || t('pages.emailTemplates.selectedMaxAccount') }}</span>
        <strong>{{ selectedSummary }}</strong>
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

    <div class="attachment-picker-table max-account-picker-table" role="table" :aria-label="title">
      <div class="attachment-picker-row attachment-picker-head max-account-picker-row" role="row">
        <span>{{ t('common.actions.select') }}</span>
        <span>{{ t('common.labels.name') }}</span>
        <span>{{ t('common.labels.phone') }}</span>
      </div>

      <template v-if="isLoading">
        <div
          v-for="index in 5"
          :key="index"
          class="attachment-picker-row max-account-picker-row"
          role="row"
        >
          <span v-for="cell in 3" :key="cell"><SkeletonBlock :rows="1" /></span>
        </div>
      </template>

      <template v-else-if="accounts.length">
        <div
          v-for="account in accounts"
          :key="accountId(account)"
          class="attachment-picker-row max-account-picker-row"
          :class="{ 'is-selected': modelValue === accountId(account) }"
          role="row"
          @click="selectAccount(account)"
        >
          <span>
            <input
              :name="pickerName"
              type="radio"
              :checked="modelValue === accountId(account)"
              :aria-label="`${t('common.actions.select')} ${accountDisplay(account)}`"
              @click.stop
              @change="selectAccount(account)"
            />
          </span>
          <span><strong>{{ account.label || t('common.placeholder') }}</strong></span>
          <span>{{ account.phone || t('common.placeholder') }}</span>
        </div>
      </template>

      <EmptyState
        v-else
        :title="emptyTitle || t('pages.maxAccounts.emptyFilteredTitle')"
        :description="emptyDescription || t('pages.maxAccounts.emptyFilteredDescription')"
      />
    </div>

    <ResourcePagination
      :page="page"
      :rows="rows"
      :total="total"
      :loaded-count="accounts.length"
      :loading="isLoading"
      :rows-options="ROW_OPTIONS"
      @update:page="setPage"
      @update:rows="setRows"
    />
  </section>
</template>
