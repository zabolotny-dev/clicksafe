<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import ConfirmDialog from '../components/ui/ConfirmDialog.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiInput from '../components/ui/UiInput.vue'
import { useNotifications } from '../composables/useNotifications'
import { useResourceActions } from '../composables/useResourceActions'
import { translateEventType, translateTargetStatus } from '../i18n'
import { listCampaigns } from '../resources/campaigns'
import { listEmployees } from '../resources/employees'
import { deleteTarget } from '../resources/targets'
import {
  EVENT_TYPES,
  TARGET_STATUSES,
  listVTargets,
} from '../resources/vtargets'
import {
  employeeName,
  errorMessage,
  formatDateTime,
  formatNullable,
  formatTechnicalId,
} from '../utils/resourceFormat'

const targets = ref([])
const campaigns = ref([])
const employees = ref([])
const selectedId = ref('')
const isLoading = ref(false)
const isDeleting = ref(false)
const loadError = ref('')
const deleteDialogOpen = ref(false)
const filters = reactive({
  fullName: '',
  campaignId: '',
  employeeId: '',
  status: '',
})

const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()
const { t } = useI18n()

const selectedTarget = computed(() => targets.value.find((target) => target.id === selectedId.value) || null)
const campaignById = computed(() => campaigns.value.reduce((acc, campaign) => {
  acc[campaign.id] = campaign
  return acc
}, {}))
const employeeById = computed(() => employees.value.reduce((acc, employee) => {
  acc[employee.id] = employee
  return acc
}, {}))
const supportedEvents = computed(() => {
  if (!selectedTarget.value?.events) {
    return []
  }

  return selectedTarget.value.events
    .filter((event) => EVENT_TYPES.includes(event.type))
    .slice()
    .sort((left, right) => new Date(right.occurred_at) - new Date(left.occurred_at))
})
const summaryCards = computed(() => {
  const counts = targets.value.reduce((acc, target) => {
    acc[target.status] = (acc[target.status] || 0) + 1
    return acc
  }, {})

  return [
    [t('common.labels.targets'), targets.value.length],
    [translateTargetStatus('PENDING'), counts.PENDING || 0],
    [translateTargetStatus('SENT'), counts.SENT || 0],
    [translateTargetStatus('OPENED'), counts.OPENED || 0],
    [translateTargetStatus('CLICKED'), counts.CLICKED || 0],
    [translateTargetStatus('SUBMITTED'), counts.SUBMITTED || 0],
    [translateTargetStatus('REPLIED'), counts.REPLIED || 0],
    [translateTargetStatus('FAILED'), counts.FAILED || 0],
  ]
})

function targetName(target) {
  const employee = employeeById.value[target.employee_id]
  return employee ? employeeName(employee) : [target.first_name, target.last_name].filter(Boolean).join(' ').trim() || t('common.placeholder')
}

function employeeOptionLabel(employee) {
  return `${employeeName(employee)} · ${employee.email || t('common.placeholder')}`
}

function campaignLabel(id) {
  return campaignById.value[id]?.label || formatTechnicalId(id)
}

function lastEvent(target) {
  const events = Array.isArray(target.events)
    ? target.events.filter((event) => EVENT_TYPES.includes(event.type))
    : []

  if (!events.length) {
    return t('common.placeholder')
  }

  const latest = events.slice().sort((left, right) => new Date(right.occurred_at) - new Date(left.occurred_at))[0]
  return `${translateEventType(latest.type) || latest.type} · ${formatDateTime(latest.occurred_at)}`
}

async function loadData() {
  isLoading.value = true
  loadError.value = ''

  try {
    const [targetResponse, campaignResponse, employeeResponse] = await Promise.all([
      listVTargets({ ...filters, rows: 100 }),
      listCampaigns({ rows: 100 }),
      listEmployees({ rows: 100 }),
    ])

    targets.value = Array.isArray(targetResponse?.items) ? targetResponse.items : []
    campaigns.value = Array.isArray(campaignResponse?.items) ? campaignResponse.items : []
    employees.value = Array.isArray(employeeResponse?.items) ? employeeResponse.items : []

    if (!targets.value.find((target) => target.id === selectedId.value)) {
      selectedId.value = targets.value[0]?.id || ''
    }
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    loadError.value = errorMessage(error, 'errors.targetsEventsLoad')
  } finally {
    isLoading.value = false
  }
}

function selectTarget(target) {
  selectedId.value = target.id
}

function openDeleteDialog(target = selectedTarget.value) {
  if (!target) {
    return
  }

  selectedId.value = target.id
  deleteDialogOpen.value = true
}

async function confirmDelete() {
  if (!selectedTarget.value || !(await ensureCsrfToken())) {
    return
  }

  isDeleting.value = true

  try {
    await deleteTarget(selectedTarget.value.id, mutationOptions())
    notifySuccess(t('common.actions.delete'))
    deleteDialogOpen.value = false
    selectedId.value = ''
    await loadData()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    notifyError(t('common.actions.delete'), errorMessage(error, 'errors.targetMutation'))
  } finally {
    isDeleting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <section class="resource-page">
    <PageHeader
      :eyebrow="t('sections.operations')"
      :title="t('pages.targetsEvents.title')"
      :description="t('pages.targetsEvents.description')"
    />

    <section class="resource-summary-grid">
      <article
        v-for="[label, value] in summaryCards"
        :key="label"
        class="resource-summary-card"
      >
        <span>{{ label }}</span>
        <strong>{{ value }}</strong>
      </article>
    </section>

    <section class="resource-filter-panel">
      <UiInput v-model="filters.fullName" :label="t('common.labels.fullName')" :placeholder="t('common.labels.employee')" />
      <label class="ui-field">
        <span class="ui-field-label">{{ t('common.labels.campaign') }}</span>
        <select v-model="filters.campaignId" class="ui-select">
          <option value="">{{ t('common.filters.allCampaigns') }}</option>
          <option v-for="campaign in campaigns" :key="campaign.id" :value="campaign.id">
            {{ campaign.label }}
          </option>
        </select>
      </label>
      <label class="ui-field">
        <span class="ui-field-label">{{ t('common.labels.employee') }}</span>
        <select v-model="filters.employeeId" class="ui-select">
          <option value="">{{ t('common.filters.allEmployees') }}</option>
          <option v-for="employee in employees" :key="employee.id" :value="employee.id">
            {{ employeeOptionLabel(employee) }}
          </option>
        </select>
      </label>
      <label class="ui-field">
        <span class="ui-field-label">{{ t('common.labels.status') }}</span>
        <select v-model="filters.status" class="ui-select">
          <option value="">{{ t('common.filters.allStatuses') }}</option>
          <option v-for="status in TARGET_STATUSES" :key="status" :value="status">
            {{ translateTargetStatus(status) }}
          </option>
        </select>
      </label>
      <UiButton variant="secondary" :loading="isLoading" @click="loadData">
        {{ t('common.filters.apply') }}
      </UiButton>
    </section>

    <UiAlert
      v-if="loadError"
      variant="error"
      :title="t('pages.targetsEvents.unavailable')"
      :message="loadError"
    >
      <UiButton variant="secondary" @click="loadData">{{ t('common.actions.retry') }}</UiButton>
    </UiAlert>

    <div class="resource-two-column resource-two-column--wide-left">
      <section class="resource-panel resource-panel--table">
        <header class="resource-panel-header">
          <div>
            <h2>{{ t('common.labels.targets') }}</h2>
            <p>{{ t('common.loadedCount', { count: targets.length }) }}</p>
          </div>
        </header>

        <div class="resource-table resource-target-table">
          <div class="resource-table-row resource-table-head">
            <span>{{ t('common.labels.employee') }}</span>
            <span>{{ t('common.labels.campaign') }}</span>
            <span>{{ t('common.labels.status') }}</span>
            <span>{{ t('common.labels.scheduledAt') }}</span>
            <span>{{ t('common.labels.createdAt') }}</span>
            <span>{{ t('common.labels.events') }}</span>
            <span>{{ t('common.labels.token') }}</span>
            <span>{{ t('pages.campaigns.columns.actions') }}</span>
          </div>
          <template v-if="isLoading">
            <div v-for="index in 5" :key="index" class="resource-table-row">
              <span v-for="cell in 8" :key="cell"><SkeletonBlock :rows="1" /></span>
            </div>
          </template>
          <template v-else-if="targets.length">
            <button
              v-for="target in targets"
              :key="target.id"
              type="button"
              class="resource-table-row resource-table-button"
              :class="{ 'is-selected': target.id === selectedId }"
              @click="selectTarget(target)"
            >
              <span><strong>{{ targetName(target) }}</strong></span>
              <span>{{ campaignLabel(target.campaign_id) }}</span>
              <span><UiBadge variant="info">{{ translateTargetStatus(target.status) }}</UiBadge></span>
              <span>{{ formatDateTime(target.scheduled_at) }}</span>
              <span>{{ formatDateTime(target.created_at) }}</span>
              <span>{{ lastEvent(target) }}</span>
              <span><code>{{ formatTechnicalId(target.token) }}</code></span>
              <span class="resource-row-actions">
                <span class="action-link">{{ t('common.actions.view') }}</span>
                <span class="action-placeholder" @click.stop="openDeleteDialog(target)">{{ t('common.actions.delete') }}</span>
              </span>
            </button>
          </template>
          <div v-else class="resource-empty-row">
            {{ t('common.empty.noMatchingRecords') }}
          </div>
        </div>
      </section>

      <aside class="resource-panel resource-detail-panel">
        <header class="resource-panel-header">
          <div>
            <h2>{{ t('common.labels.events') }}</h2>
            <p>{{ t('pages.targetsEvents.description') }}</p>
          </div>
        </header>

        <p v-if="!selectedTarget" class="resource-muted">{{ t('pages.targetsEvents.noSelection') }}</p>
        <template v-else>
          <dl class="resource-description-list">
            <div><dt>{{ t('common.labels.target') }}</dt><dd>{{ targetName(selectedTarget) }}</dd></div>
            <div><dt>{{ t('common.labels.campaign') }}</dt><dd>{{ campaignLabel(selectedTarget.campaign_id) }}</dd></div>
            <div><dt>{{ t('common.labels.status') }}</dt><dd>{{ translateTargetStatus(selectedTarget.status) }}</dd></div>
            <div><dt>{{ t('common.labels.token') }}</dt><dd><code>{{ formatTechnicalId(selectedTarget.token) }}</code></dd></div>
          </dl>

          <section class="resource-event-list">
            <article
              v-for="event in supportedEvents"
              :key="`${event.type}-${event.occurred_at}`"
              class="resource-event-item"
            >
              <div>
                <strong>{{ translateEventType(event.type) || event.type }}</strong>
                <span>{{ formatDateTime(event.occurred_at) }}</span>
              </div>
              <dl>
                <div><dt>IP</dt><dd>{{ formatNullable(event.ip_address) }}</dd></div>
                <div><dt>User Agent</dt><dd>{{ formatNullable(event.user_agent) }}</dd></div>
                <div><dt>Referer</dt><dd>{{ formatNullable(event.referer) }}</dd></div>
              </dl>
            </article>
            <p v-if="!supportedEvents.length" class="resource-muted">
              {{ t('pages.targetsEvents.noEvents') }}
            </p>
          </section>

          <footer class="resource-form-actions">
            <UiButton variant="danger" :loading="isDeleting" @click="openDeleteDialog()">
              {{ isDeleting ? t('common.states.deleting') : t('common.actions.delete') }}
            </UiButton>
          </footer>
        </template>
      </aside>
    </div>

    <ConfirmDialog
      :open="deleteDialogOpen"
      :title="t('common.actions.delete')"
      :message="t('pages.targetsEvents.deleteMessage')"
      :confirm-label="t('common.actions.delete')"
      :loading-label="t('common.states.deleting')"
      :cancel-label="t('common.actions.cancel')"
      variant="danger"
      :loading="isDeleting"
      @cancel="deleteDialogOpen = false"
      @confirm="confirmDelete"
    />
  </section>
</template>
