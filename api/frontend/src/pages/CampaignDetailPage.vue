<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Activity,
  AlertTriangle,
  Eye,
  Info,
  Mail,
  MousePointerClick,
  Send,
  UserCircle,
} from 'lucide-vue-next'
import CampaignFunnelChart from '../components/charts/CampaignFunnelChart.vue'
import CampaignRiskGauge from '../components/charts/CampaignRiskGauge.vue'
import CampaignScheduleTimelineChart from '../components/charts/CampaignScheduleTimelineChart.vue'
import EventDistributionDonutChart from '../components/charts/EventDistributionDonutChart.vue'
import EventsOverTimeChart from '../components/charts/EventsOverTimeChart.vue'
import AnalyticsSlot from '../components/dashboard/AnalyticsSlot.vue'
import ConfirmDialog from '../components/ui/ConfirmDialog.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import IconButton from '../components/ui/IconButton.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import PreviewDrawer from '../components/ui/PreviewDrawer.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import { useNotifications } from '../composables/useNotifications'
import { useSession } from '../composables/useSession'
import { translateCampaignStatus, translateEventType, translateTargetStatus } from '../i18n'
import {
  cancelCampaign,
  deleteCampaign,
  getCampaign,
  isCampaignAuthError,
  isCampaignNotFoundError,
  pauseCampaign,
  startCampaign,
} from '../resources/campaigns'
import {
  EVENT_LABELS,
  EVENT_TYPES,
  TARGET_STATUSES,
  isVTargetAuthError,
  listVTargets,
} from '../resources/vtargets'
import {
  formatDate as formatLocalizedDate,
  formatDateRange as formatLocalizedDateRange,
  formatDateTime as formatLocalizedDateTime,
  formatMonthDay,
  formatNullable as formatLocalizedNullable,
  nullableId,
  formatTechnicalId as formatSharedTechnicalId,
} from '../utils/resourceFormat'

const route = useRoute()
const router = useRouter()
const { csrfToken, loadCurrentSession } = useSession()
const { notifySuccess, notifyError } = useNotifications()
const { t, locale } = useI18n()

const emailFunnelStagesConfig = [
  {
    eventType: 'MESSAGE_SENT',
    reachedBy: ['MESSAGE_SENT', 'EMAIL_OPENED', 'LINK_OPENED', 'DATA_SENT', 'MESSAGE_READ', 'MESSAGE_REPLIED'],
  },
  {
    eventType: 'EMAIL_OPENED',
    reachedBy: ['EMAIL_OPENED', 'LINK_OPENED', 'DATA_SENT', 'MESSAGE_READ', 'MESSAGE_REPLIED'],
  },
  {
    eventType: 'LINK_OPENED',
    reachedBy: ['LINK_OPENED', 'DATA_SENT'],
  },
  {
    eventType: 'DATA_SENT',
    reachedBy: ['DATA_SENT'],
  },
]
const maxBaseFunnelStagesConfig = [
  {
    eventType: 'MESSAGE_SENT',
    reachedBy: ['MESSAGE_SENT', 'MESSAGE_READ', 'MESSAGE_REPLIED', 'LINK_OPENED', 'DATA_SENT'],
  },
  {
    eventType: 'MESSAGE_READ',
    reachedBy: ['MESSAGE_READ', 'MESSAGE_REPLIED', 'LINK_OPENED', 'DATA_SENT'],
  },
  {
    eventType: 'MESSAGE_REPLIED',
    reachedBy: ['MESSAGE_REPLIED'],
  },
]
const maxLandingFunnelStagesConfig = [
  ...maxBaseFunnelStagesConfig,
  {
    eventType: 'LINK_OPENED',
    reachedBy: ['LINK_OPENED', 'DATA_SENT'],
  },
  {
    eventType: 'DATA_SENT',
    reachedBy: ['DATA_SENT'],
  },
]
const riskScores = {
  MESSAGE_SENT: 0,
  DELIVERY_FAILED: 0,
  EMAIL_OPENED: 15,
  MESSAGE_READ: 15,
  LINK_OPENED: 45,
  MESSAGE_REPLIED: 60,
  DATA_SENT: 100,
}

const tabs = computed(() => [
  { id: 'overview', label: t('pages.campaignDetail.tabs.overview') },
  { id: 'targets', label: t('pages.campaignDetail.tabs.targets') },
  { id: 'settings', label: t('pages.campaignDetail.tabs.settings') },
])
const activeTab = ref('overview')

const targetColumns = computed(() => [
  t('pages.campaignDetail.targets.columns.employee'),
  t('pages.campaignDetail.targets.columns.campaign'),
  t('pages.campaignDetail.targets.columns.status'),
  t('pages.campaignDetail.targets.columns.scheduledAt'),
  t('pages.campaignDetail.targets.columns.createdAt'),
  t('pages.campaignDetail.targets.columns.token'),
  t('pages.campaignDetail.targets.columns.actions'),
])

const campaign = ref(null)
const isLoading = ref(false)
const notFound = ref(false)
const error = ref('')
const authSuppressed = ref(false)
const vtargets = ref([])
const isVTargetsLoading = ref(false)
const hasLoadedVTargets = ref(false)
const vTargetsError = ref('')
const vTargetsPartial = ref(false)
const vTargetsAuthSuppressed = ref(false)
const pendingAction = ref('')
const confirmationDialog = ref(null)
const selectedTargetId = ref('')
const targetStatusFilter = ref('')
const targetDrawerOpen = ref(false)
const scheduleDrawerOpen = ref(false)
const scheduleTargetId = ref('')
const scheduleValue = ref('')
const scheduleError = ref('')
const addTargetsDrawerOpen = ref(false)
const employees = ref([])
const departments = ref([])
const referenceLoading = ref(false)
const referenceLoaded = ref(false)
const referenceError = ref('')
const employeeSearch = ref('')
const departmentFilter = ref('')
const selectedAddEmployeeIds = ref(new Set())
const targetMutationError = ref('')
const pendingTargetAction = ref('')
const targetDeleteCandidate = ref(null)

let requestSequence = 0
let activeController = null
let vTargetRequestSequence = 0
let activeVTargetController = null
let activeLifecycleController = null
let activeReferenceController = null
let activeTargetMutationController = null

const campaignId = computed(() => String(route.params.id ?? ''))
const activeTabPanelId = computed(() => `campaign-detail-${activeTab.value.toLowerCase()}`)
const campaignTitle = computed(() => campaign.value?.label || '—')
const isMaxCampaign = computed(() => String(campaign.value?.type || '').toUpperCase() === 'MAX')
const hasCampaignLanding = computed(() => Boolean(nullableId(campaign.value?.landing_id)))
const activeFunnelStagesConfig = computed(() => {
  if (!isMaxCampaign.value) {
    return emailFunnelStagesConfig
  }

  return hasCampaignLanding.value
    ? maxLandingFunnelStagesConfig
    : maxBaseFunnelStagesConfig
})
const supportedEvents = computed(() => flattenSupportedEvents(vtargets.value))
const hasSupportedEvents = computed(() => supportedEvents.value.length > 0)
const isLifecyclePending = computed(() => Boolean(pendingAction.value))

const actionAvailability = computed(() => {
  switch (campaign.value?.status) {
    case 'DRAFT':
      return {
        start: true,
        pause: false,
        cancel: true,
        delete: true,
      }
    case 'ACTIVE':
      return {
        start: false,
        pause: true,
        cancel: true,
        delete: false,
      }
    case 'PAUSED':
      return {
        start: true,
        pause: false,
        cancel: true,
        delete: true,
      }
    case 'COMPLETED':
    case 'CANCELED':
      return {
        start: false,
        pause: false,
        cancel: false,
        delete: true,
      }
    default:
      return {
        start: false,
        pause: false,
        cancel: false,
        delete: false,
      }
  }
})

const canEditCampaign = computed(() => ['DRAFT', 'PAUSED'].includes(campaign.value?.status))
const campaignEditDisabledReason = computed(() => {
  if (!campaign.value?.status) {
    return t('campaign.actions.unknownStatus')
  }

  if (campaign.value.status === 'CANCELED') {
    return t('campaign.actions.editUnavailableCanceled')
  }

  return canEditCampaign.value ? '' : t('campaign.actions.editUnavailableStatus')
})

const canMutateTargets = computed(() => ['DRAFT', 'PAUSED'].includes(campaign.value?.status))
const targetMutationDisabledReason = computed(() => {
  if (!campaign.value?.status) {
    return t('campaign.actions.unknownStatus')
  }

  return canMutateTargets.value
    ? ''
    : t('campaign.actions.targetMutationUnavailable')
})

const attributes = computed(() => {
  const raw = campaign.value?.attributes
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) {
    return []
  }

  return Object.entries(raw)
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([key, value]) => ({
      key,
      value: formatNullable(value),
    }))
})

const settings = computed(() => [
  {
    label: t('common.labels.type'),
    value: campaign.value?.type || 'EMAIL',
  },
  {
    label: t('pages.campaignDetail.labels.campaignLabel'),
    value: campaign.value?.label || '—',
  },
  {
    label: t('common.labels.domain'),
    value: campaign.value?.domain || '—',
  },
  {
    label: t('common.labels.message'),
    value: formatTechnicalId(campaign.value?.message_id),
    title: campaign.value?.message_id || '',
    technical: true,
  },
  {
    label: t('common.labels.landingPage'),
    value: formatTechnicalId(campaign.value?.landing_id),
    title: campaign.value?.landing_id || '',
    technical: true,
  },
  {
    label: t('common.labels.educationAsset'),
    value: campaign.value?.type === 'MAX'
      ? formatTechnicalId(campaign.value?.max_education_text_id)
      : formatTechnicalId(campaign.value?.education_id),
    title: campaign.value?.type === 'MAX'
      ? nullableId(campaign.value?.max_education_text_id)
      : nullableId(campaign.value?.education_id),
    technical: true,
  },
  {
    label: t('common.labels.dateFrom'),
    value: formatDate(campaign.value?.date_from),
  },
  {
    label: t('common.labels.dateTo'),
    value: formatDate(campaign.value?.date_to),
  },
])

const campaignRiskIndex = computed(() => {
  if (!vtargets.value.length || !hasSupportedEvents.value) {
    return '—'
  }

  const totalRisk = vtargets.value.reduce((sum, target) => {
    const targetRisk = supportedEventsForTarget(target).reduce(
      (max, event) => Math.max(max, riskScores[event.type] ?? 0),
      0,
    )
    return sum + targetRisk
  }, 0)

  return String(Math.round(totalRisk / vtargets.value.length))
})

const funnelStages = computed(() => {
  locale.value

  return activeFunnelStagesConfig.value.map((stage) => ({
    eventType: stage.eventType,
    label: translateEventType(stage.eventType),
    count: countTargetsReachingStage(stage.reachedBy),
  }))
})

const eventsOverTimeBuckets = computed(() => {
  locale.value

  const groups = new Map()

  supportedEvents.value.forEach(({ event }) => {
    const date = new Date(event.occurred_at)
    if (Number.isNaN(date.getTime())) {
      return
    }

    const key = [
      date.getFullYear(),
      String(date.getMonth() + 1).padStart(2, '0'),
      String(date.getDate()).padStart(2, '0'),
    ].join('-')

    if (!groups.has(key)) {
      groups.set(key, {
        key,
        label: formatDayLabel(key),
        counts: Object.fromEntries(EVENT_TYPES.map((eventType) => [eventType, 0])),
      })
    }

    const bucket = groups.get(key)
    bucket.counts[event.type] += 1
  })

  return Array.from(groups.values())
    .sort((left, right) => left.key.localeCompare(right.key))
})

const targetRows = computed(() => vtargets.value.map((target) => ({
  ...target,
  employeeName: formatEmployeeName(target),
})))

const filteredTargetRows = computed(() => {
  if (!targetStatusFilter.value) {
    return targetRows.value
  }

  return targetRows.value.filter((target) => target.status === targetStatusFilter.value)
})

const scheduledTargetRows = computed(() => targetRows.value.filter((target) => target.scheduled_at))
const unscheduledTargetCount = computed(() => targetRows.value.length - scheduledTargetRows.value.length)
const hasCampaignDateRange = computed(() => Boolean(campaign.value?.date_from && campaign.value?.date_to))
const canDistributeTargets = computed(() => (
  canMutateTargets.value
    && hasCampaignDateRange.value
    && targetRows.value.length > 0
    && !Boolean(pendingTargetAction.value)
))
const distributeDisabledReason = computed(() => {
  if (pendingTargetAction.value) {
    return t('common.states.working')
  }

  if (!canMutateTargets.value) {
    return targetMutationDisabledReason.value
  }

  if (!hasCampaignDateRange.value) {
    return t('common.labels.dateRange')
  }

  if (!targetRows.value.length) {
    return t('pages.campaignDetail.targets.noTargets')
  }

  return ''
})

const selectedTarget = computed(() => {
  return targetRows.value.find((target) => target.id === selectedTargetId.value) || null
})

const selectedScheduleTarget = computed(() => {
  return targetRows.value.find((target) => target.id === scheduleTargetId.value) || null
})

const selectedTargetEvents = computed(() => {
  return supportedEventsForTarget(selectedTarget.value)
    .slice()
    .sort((left, right) => eventTime(right) - eventTime(left))
})

const targetSummary = computed(() => ({
  total: targetRows.value.length,
  visible: filteredTargetRows.value.length,
  activity: supportedEvents.value.length,
  scheduled: scheduledTargetRows.value.length,
  unscheduled: unscheduledTargetCount.value,
}))

const existingEmployeeIds = computed(() => new Set(targetRows.value.map((target) => String(target.employee_id))))
const departmentById = computed(() => Object.fromEntries(
  departments.value.map((department) => [String(department.id), department]),
))
const availableEmployees = computed(() => employees.value.filter((employee) => (
  employee?.id && !existingEmployeeIds.value.has(String(employee.id))
)))
const filteredAvailableEmployees = computed(() => {
  const query = employeeSearch.value.trim().toLowerCase()
  const selectedDepartment = departmentFilter.value

  return availableEmployees.value.filter((employee) => {
    if (selectedDepartment && String(employee.department_id || '') !== selectedDepartment) {
      return false
    }

    if (!query) {
      return true
    }

    return [
      employeeFullName(employee),
      employee.email,
      employee.phone,
      departmentLabel(employee.department_id),
    ]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(query))
  })
})
const selectedAddEmployeeCount = computed(() => selectedAddEmployeeIds.value.size)

const campaignRiskGaugeValue = computed(() => {
  const value = Number(campaignRiskIndex.value)
  return Number.isFinite(value) ? value : null
})

const eventDistributionItems = computed(() => EVENT_TYPES.map((eventType) => ({
  eventType,
  label: eventDistributionLabel(eventType),
  count: countTargetsWithEvent(eventType),
})))

const eventsOverTimeEmptyText = computed(() => {
  if (vTargetsError.value && !hasLoadedVTargets.value) {
    return t('errors.monitoringLoad')
  }

  return hasLoadedVTargets.value
    ? t('charts.eventsOverTime.empty')
    : t('common.states.loadingEllipsis')
})

function resetVTargetState(options = {}) {
  const preserveExisting = Boolean(options.preserveExisting)

  if (!preserveExisting) {
    vtargets.value = []
    hasLoadedVTargets.value = false
  }

  isVTargetsLoading.value = false
  vTargetsError.value = ''
  vTargetsPartial.value = false
  vTargetsAuthSuppressed.value = false
}

function abortActiveRequest() {
  if (activeController) {
    activeController.abort()
    activeController = null
  }
}

function abortActiveVTargetRequest() {
  if (activeVTargetController) {
    activeVTargetController.abort()
    activeVTargetController = null
  }
}

function abortActiveLifecycleRequest() {
  if (activeLifecycleController) {
    activeLifecycleController.abort()
    activeLifecycleController = null
  }
}

function abortActiveReferenceRequest() {
  if (activeReferenceController) {
    activeReferenceController.abort()
    activeReferenceController = null
  }
}

function abortActiveTargetMutationRequest() {
  if (activeTargetMutationController) {
    activeTargetMutationController.abort()
    activeTargetMutationController = null
  }
}

async function loadCampaignDetail(options = {}) {
  const preserveExisting = Boolean(options.preserveExisting)
  const id = campaignId.value
  const requestId = requestSequence + 1
  requestSequence = requestId
  vTargetRequestSequence += 1
  abortActiveRequest()
  abortActiveVTargetRequest()
  resetVTargetState({ preserveExisting })

  if (!id) {
    campaign.value = null
    notFound.value = true
    error.value = ''
    authSuppressed.value = false
    isLoading.value = false
    return
  }

  const controller = new AbortController()
  activeController = controller
  if (!preserveExisting) {
    campaign.value = null
  }
  isLoading.value = !preserveExisting
  notFound.value = false
  error.value = ''
  authSuppressed.value = false

  try {
    const response = await getCampaign(id, {
      signal: controller.signal,
    })

    if (requestId !== requestSequence) {
      return
    }

    campaign.value = response || null

    if (!campaign.value) {
      notFound.value = true
      return
    }

    loadVTargets(id, { preserveExisting })
  } catch (err) {
    if (err?.name === 'AbortError' || requestId !== requestSequence) {
      return
    }

    if (!preserveExisting) {
      campaign.value = null
    }

    if (isCampaignAuthError(err)) {
      authSuppressed.value = true
      error.value = ''
      await loadCurrentSession()
      return
    }

    if (isCampaignNotFoundError(err)) {
      notFound.value = true
      error.value = ''
      return
    }

    error.value = err?.message || t('errors.campaignLoad')
  } finally {
    if (requestId === requestSequence) {
      isLoading.value = false
      activeController = null
    }
  }
}

async function loadVTargets(id = campaignId.value, options = {}) {
  const preserveExisting = Boolean(options.preserveExisting)
  const requestId = vTargetRequestSequence + 1
  vTargetRequestSequence = requestId
  abortActiveVTargetRequest()

  const controller = new AbortController()
  activeVTargetController = controller
  const loadedTargets = []
  isVTargetsLoading.value = true
  if (!preserveExisting) {
    hasLoadedVTargets.value = false
  }
  vTargetsError.value = ''
  vTargetsPartial.value = false
  vTargetsAuthSuppressed.value = false

  try {
    let page = 1
    let total = 0
    const rows = 100

    do {
      const response = await listVTargets({
        campaignId: id,
        page,
        rows,
      }, {
        signal: controller.signal,
      })

      if (requestId !== vTargetRequestSequence) {
        return
      }

      const items = Array.isArray(response?.items) ? response.items : []
      total = Number.isFinite(response?.total) ? response.total : loadedTargets.length + items.length
      loadedTargets.push(...items)

      if (!items.length) {
        break
      }

      page += 1
    } while (loadedTargets.length < total)

    vtargets.value = loadedTargets
    hasLoadedVTargets.value = true
  } catch (err) {
    if (err?.name === 'AbortError' || requestId !== vTargetRequestSequence) {
      return
    }

    if (loadedTargets.length) {
      vtargets.value = loadedTargets
      hasLoadedVTargets.value = true
      vTargetsPartial.value = true
    } else {
      if (!preserveExisting) {
        vtargets.value = []
        hasLoadedVTargets.value = false
      }
    }

    if (isVTargetAuthError(err)) {
      vTargetsAuthSuppressed.value = true
      vTargetsError.value = ''
      await loadCurrentSession()
      return
    }

    vTargetsError.value = err?.message || t('errors.monitoringLoad')
  } finally {
    if (requestId === vTargetRequestSequence) {
      isVTargetsLoading.value = false
      activeVTargetController = null
    }
  }
}

function retryLoad() {
  loadCampaignDetail()
}

function retryVTargets() {
  loadVTargets()
}

function actionRequestOptions(controller) {
  const headers = csrfToken.value
    ? {
        'X-CSRF-Token': csrfToken.value,
      }
    : {}

  return {
    signal: controller.signal,
    headers,
  }
}

function actionTitle(action) {
  switch (action) {
    case 'start':
    return t('campaign.actions.start')
    case 'pause':
      return t('campaign.actions.pause')
    case 'cancel':
      return t('common.actions.cancel')
    case 'delete':
      return t('common.actions.delete')
    default:
      return t('common.actions.confirm')
  }
}

function actionLoadingLabel(action) {
  switch (action) {
    case 'start':
      return t('common.states.starting')
    case 'pause':
      return t('common.states.pausing')
    case 'cancel':
      return t('common.states.canceling')
    case 'delete':
      return t('common.states.deleting')
    default:
      return t('common.states.working')
  }
}

function actionSuccessMessage(action) {
  switch (action) {
    case 'start':
      return t('campaign.actions.startSuccess')
    case 'pause':
      return t('campaign.actions.pauseSuccess')
    case 'cancel':
      return t('campaign.actions.cancelSuccess')
    default:
      return t('campaign.actions.updateSuccess')
  }
}

function isActionLoading(action) {
  return pendingAction.value === action
}

function isActionDisabled(action) {
  return isLifecyclePending.value || !actionAvailability.value[action]
}

function actionButtonTitle(action) {
  if (isLifecyclePending.value && !isActionLoading(action)) {
    return t('campaign.actions.otherActionPending')
  }

  if (!actionAvailability.value[action]) {
    return t('campaign.actions.notAvailableForStatus')
  }

  return ''
}

function openConfirmation(action) {
  if (isActionDisabled(action)) {
    return
  }

  if (action === 'cancel') {
    confirmationDialog.value = {
      action,
        title: t('campaign.actions.cancelDialogTitle'),
        text: t('campaign.actions.cancelDialogMessage'),
        confirmLabel: t('campaign.actions.cancelDialogTitle'),
        loadingLabel: t('common.states.canceling'),
        variant: 'danger',
    }
    return
  }

  if (action === 'delete') {
    confirmationDialog.value = {
      action,
        title: t('campaign.actions.deleteDialogTitle'),
        text: t('campaign.actions.deleteDialogMessage'),
        confirmLabel: t('campaign.actions.deleteDialogConfirm'),
        loadingLabel: t('common.states.deleting'),
        variant: 'danger',
    }
  }
}

function closeConfirmation() {
  if (isLifecyclePending.value) {
    return
  }

  confirmationDialog.value = null
}

function runConfirmedLifecycleAction() {
  const action = confirmationDialog.value?.action
  if (action) {
    runLifecycleAction(action)
  }
}

async function handleLifecycleAuthFailure() {
  const refreshedSession = await loadCurrentSession()

  if (!refreshedSession) {
    notifyError(t('errors.auth.sessionExpired'), t('errors.auth.sessionExpired'))
    await router.push({
      name: 'login',
      query: {
        redirect: route.fullPath,
      },
    })
    return
  }

  notifyError(t('errors.sessionCheck'), t('errors.sessionCheck'))
}

async function ensureCsrfToken() {
  if (csrfToken.value) {
    return true
  }

  return Boolean(await loadCurrentSession())
}

async function handleTargetAuthFailure() {
  const refreshedSession = await loadCurrentSession()

  if (!refreshedSession) {
    notifyError(t('errors.auth.sessionExpired'), t('errors.auth.sessionExpired'))
    await router.push({
      name: 'login',
      query: {
        redirect: route.fullPath,
      },
    })
    return
  }

  notifyError(t('errors.sessionCheck'), t('errors.sessionCheck'))
}

async function runLifecycleAction(action) {
  if (isActionDisabled(action)) {
    return
  }

  const id = campaignId.value
  if (!id) {
    return
  }

  abortActiveLifecycleRequest()

  const controller = new AbortController()
  activeLifecycleController = controller
  pendingAction.value = action

  try {
    if (!csrfToken.value) {
      await loadCurrentSession()
    }

    const options = actionRequestOptions(controller)

    switch (action) {
      case 'start':
        await startCampaign(id, options)
        break
      case 'pause':
        await pauseCampaign(id, options)
        break
      case 'cancel':
        await cancelCampaign(id, options)
        break
      case 'delete':
        await deleteCampaign(id, options)
        notifySuccess(t('campaign.actions.deleteSuccess'), t('campaign.actions.deleteDescription'))
        await router.push('/campaigns')
        return
      default:
        return
    }

    await loadCampaignDetail({ preserveExisting: true })
    notifySuccess(t('campaign.actions.updateSuccess'), actionSuccessMessage(action))
  } catch (err) {
    if (err?.name === 'AbortError') {
      return
    }

    if (isCampaignAuthError(err)) {
      await handleLifecycleAuthFailure()
      return
    }

    notifyError(`${actionTitle(action)} failed`, err?.message || t('errors.requestFailed'))
  } finally {
    if (activeLifecycleController === controller) {
      activeLifecycleController = null
    }

    if (pendingAction.value === action) {
      pendingAction.value = ''
    }

    if (confirmationDialog.value?.action === action) {
      confirmationDialog.value = null
    }
  }
}

function statusVariant(status) {
  switch (status) {
    case 'ACTIVE':
      return 'success'
    case 'PAUSED':
      return 'warning'
    case 'COMPLETED':
      return 'info'
    case 'CANCELED':
      return 'danger'
    default:
      return 'neutral'
  }
}

function targetStatusVariant(status) {
  switch (status) {
    case 'SENT':
    case 'OPENED':
    case 'REPLIED':
      return 'info'
    case 'CLICKED':
      return 'warning'
    case 'SUBMITTED':
    case 'FAILED':
      return 'danger'
    default:
      return 'neutral'
  }
}

function displayStatus(status) {
  return status ? translateCampaignStatus(status) : t('common.placeholder')
}

function displayTargetStatus(status) {
  return TARGET_STATUSES.includes(status) ? translateTargetStatus(status) : (status || t('common.placeholder'))
}

function displayEventType(type) {
  return EVENT_LABELS[type] || translateEventType(type) || t('common.placeholder')
}

function eventDistributionLabel(type) {
  return displayEventType(type)
}

function eventIcon(type) {
  switch (type) {
    case 'MESSAGE_SENT':
      return Mail
    case 'DELIVERY_FAILED':
      return AlertTriangle
    case 'EMAIL_OPENED':
    case 'MESSAGE_READ':
      return Eye
    case 'LINK_OPENED':
      return MousePointerClick
    case 'MESSAGE_REPLIED':
    case 'DATA_SENT':
      return Send
    default:
      return Activity
  }
}

function inspectTarget(target) {
  selectedTargetId.value = target?.id || ''
  targetDrawerOpen.value = Boolean(target?.id)
}

function closeTargetDrawer() {
  targetDrawerOpen.value = false
}

function supportedEventsForTarget(target) {
  return Array.isArray(target?.events)
    ? target.events.filter((event) => EVENT_TYPES.includes(event?.type))
    : []
}

function flattenSupportedEvents(targets) {
  return targets.flatMap((target) => supportedEventsForTarget(target).map((event, eventIndex) => ({
    event,
    target,
    employeeName: formatEmployeeName(target),
    rowKey: [
      target?.id || target?.token || 'target',
      event?.type || 'event',
      event?.occurred_at || 'time',
      eventIndex,
    ].join(':'),
  })))
}

function countTargetsWithEvent(eventType) {
  return vtargets.value.reduce((count, target) => {
    const hasEvent = supportedEventsForTarget(target).some((event) => event.type === eventType)
    return hasEvent ? count + 1 : count
  }, 0)
}

function countTargetsReachingStage(eventTypes) {
  return vtargets.value.reduce((count, target) => {
    const hasReachedStage = supportedEventsForTarget(target)
      .some((event) => eventTypes.includes(event.type))
    return hasReachedStage ? count + 1 : count
  }, 0)
}

function eventTime(event) {
  const date = new Date(event?.occurred_at)
  return Number.isNaN(date.getTime()) ? 0 : date.getTime()
}

function formatEmployeeName(target) {
  const name = [target?.first_name, target?.last_name].filter(Boolean).join(' ').trim()
  return name || '—'
}

function employeeFullName(employee) {
  return [employee?.first_name, employee?.last_name].filter(Boolean).join(' ').trim() || '—'
}

function departmentLabel(departmentId) {
  return departmentById.value[String(departmentId || '')]?.label || '—'
}

function formatEmployeeContact(target) {
  return target?.email || target?.email_address || target?.employee_email || `${t('common.labels.employee')} ${formatTechnicalId(target?.employee_id)}`
}

function formatTechnicalId(value) {
  return formatSharedTechnicalId(value)
}

function formatNullable(value) {
  return formatLocalizedNullable(value)
}

function formatDate(value) {
  return formatLocalizedDate(value)
}

function formatDateTime(value) {
  return formatLocalizedDateTime(value)
}

function toDatetimeLocalValue(value) {
  if (!value) {
    return ''
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }

  return [
    [
      date.getFullYear(),
      String(date.getMonth() + 1).padStart(2, '0'),
      String(date.getDate()).padStart(2, '0'),
    ].join('-'),
    [
      String(date.getHours()).padStart(2, '0'),
      String(date.getMinutes()).padStart(2, '0'),
    ].join(':'),
  ].join('T')
}

function formatDayLabel(key) {
  return formatMonthDay(`${key}T00:00:00`)
}

function formatDateRange(cmp) {
  return formatLocalizedDateRange(cmp?.date_from, cmp?.date_to)
}

watch(campaignId, () => {
  abortActiveLifecycleRequest()
  pendingAction.value = ''
  confirmationDialog.value = null
  selectedTargetId.value = ''
  targetStatusFilter.value = ''
  targetDrawerOpen.value = false
  scheduleDrawerOpen.value = false
  addTargetsDrawerOpen.value = false
  scheduleTargetId.value = ''
  targetDeleteCandidate.value = null
  targetMutationError.value = ''
  selectedAddEmployeeIds.value = new Set()
  loadCampaignDetail()
})

watch(targetRows, (rows) => {
  if (!rows.length) {
    selectedTargetId.value = ''
    targetDrawerOpen.value = false
    return
  }

  if (selectedTargetId.value && !rows.some((target) => target.id === selectedTargetId.value)) {
    selectedTargetId.value = ''
    targetDrawerOpen.value = false
  }

  if (scheduleTargetId.value && !rows.some((target) => target.id === scheduleTargetId.value)) {
    closeScheduleDrawer()
  }
}, { immediate: true })

onMounted(() => {
  loadCampaignDetail()
})

onBeforeUnmount(() => {
  requestSequence += 1
  vTargetRequestSequence += 1
  abortActiveRequest()
  abortActiveVTargetRequest()
  abortActiveLifecycleRequest()
  abortActiveReferenceRequest()
  abortActiveTargetMutationRequest()
})
</script>

<template>
  <div class="campaign-detail-page">
    <template v-if="isLoading">
      <PageHeader
        :eyebrow="t('sections.campaigns')"
        :title="t('pages.campaignDetail.loadingTitle')"
        :description="t('pages.campaignDetail.loadingDescription')"
      />
      <section class="campaign-detail-loading" :aria-label="t('pages.campaignDetail.loadingDescription')">
        <SkeletonBlock :rows="6" />
      </section>
    </template>

    <template v-else-if="notFound">
      <EmptyState
        :title="t('pages.campaignDetail.notFoundTitle')"
        :description="t('pages.campaignDetail.notFoundDescription')"
      >
        <RouterLink class="ui-button ui-button--secondary" to="/campaigns">
          {{ t('pages.campaignDetail.backToCampaigns') }}
        </RouterLink>
      </EmptyState>
    </template>

    <template v-else-if="authSuppressed">
      <PageHeader
        :eyebrow="t('sections.campaigns')"
        :title="t('pages.campaignDetail.loadingTitle')"
        :description="t('pages.campaignDetail.sessionDescription')"
      />
      <section class="campaign-detail-loading" :aria-label="t('pages.campaignDetail.sessionDescription')">
        <SkeletonBlock :rows="4" />
      </section>
    </template>

    <template v-else-if="error">
      <PageHeader
        :eyebrow="t('sections.campaigns')"
        :title="t('pages.campaignDetail.loadingTitle')"
        :description="t('pages.campaignDetail.loadDescription')"
      />
      <section class="campaign-detail-error" role="status">
        <div>
          <h2>{{ t('pages.campaignDetail.loadTitle') }}</h2>
          <p>{{ error }}</p>
        </div>
        <UiButton variant="secondary" @click="retryLoad">
          {{ t('pages.campaignDetail.retry') }}
        </UiButton>
      </section>
    </template>

    <template v-else>
      <PageHeader
        :eyebrow="t('sections.campaigns')"
        :title="campaignTitle"
      >
        <template #actions>
          <UiButton
            variant="secondary"
            class="campaign-detail-action"
            :loading="isActionLoading('start')"
            :disabled="isActionDisabled('start')"
            :title="actionButtonTitle('start')"
            @click="runLifecycleAction('start')"
          >
            {{ isActionLoading('start') ? actionLoadingLabel('start') : t('campaign.actions.start') }}
          </UiButton>
          <UiButton
            variant="secondary"
            class="campaign-detail-action"
            :loading="isActionLoading('pause')"
            :disabled="isActionDisabled('pause')"
            :title="actionButtonTitle('pause')"
            @click="runLifecycleAction('pause')"
          >
            {{ isActionLoading('pause') ? actionLoadingLabel('pause') : t('campaign.actions.pause') }}
          </UiButton>
          <UiButton
            variant="secondary"
            class="campaign-detail-action"
            :loading="isActionLoading('cancel')"
            :disabled="isActionDisabled('cancel')"
            :title="actionButtonTitle('cancel')"
            @click="openConfirmation('cancel')"
          >
            {{ isActionLoading('cancel') ? actionLoadingLabel('cancel') : t('common.actions.cancel') }}
          </UiButton>
          <RouterLink
            v-if="canEditCampaign"
            class="ui-button ui-button--secondary campaign-detail-action"
            :to="`/campaigns/${campaignId}/edit`"
          >
            {{ t('common.actions.edit') }}
          </RouterLink>
          <UiButton
            v-else
            variant="ghost"
            class="campaign-detail-action"
            disabled
            aria-disabled="true"
            :title="campaignEditDisabledReason"
          >
            {{ t('common.actions.edit') }}
          </UiButton>
          <UiButton
            variant="danger"
            class="campaign-detail-action"
            :loading="isActionLoading('delete')"
            :disabled="isActionDisabled('delete')"
            :title="actionButtonTitle('delete')"
            @click="openConfirmation('delete')"
          >
            {{ isActionLoading('delete') ? actionLoadingLabel('delete') : t('common.actions.delete') }}
          </UiButton>
        </template>
      </PageHeader>

      <section class="campaign-detail-summary" :aria-label="t('pages.campaignDetail.summaryAria')">
        <div class="campaign-detail-meta">
          <span class="campaign-detail-meta-label">{{ t('common.labels.status') }}</span>
          <UiBadge :variant="statusVariant(campaign?.status)">
            {{ displayStatus(campaign?.status) }}
          </UiBadge>
        </div>
        <div class="campaign-detail-meta">
          <span class="campaign-detail-meta-label">{{ t('common.labels.domain') }}</span>
          <span class="campaign-detail-placeholder">{{ campaign?.domain || '—' }}</span>
        </div>
        <div class="campaign-detail-meta">
          <span class="campaign-detail-meta-label">{{ t('common.labels.dateRange') }}</span>
          <span class="campaign-detail-placeholder">{{ formatDateRange(campaign) }}</span>
        </div>
      </section>

      <section
        v-if="isVTargetsLoading && !hasLoadedVTargets"
        class="campaign-monitoring-state"
        :aria-label="t('pages.campaignDetail.monitoringLoadingState')"
      >
        <SkeletonBlock :rows="3" />
      </section>

      <section
        v-else-if="vTargetsAuthSuppressed"
        class="campaign-monitoring-state"
        :aria-label="t('pages.campaignDetail.monitoringSessionState')"
      >
        <SkeletonBlock :rows="2" />
      </section>

      <section
        v-else-if="vTargetsError"
        class="campaign-detail-error campaign-monitoring-error"
        role="status"
      >
        <div>
          <h2>{{ vTargetsPartial ? t('pages.campaignDetail.partialMonitoringData') : t('pages.campaignDetail.monitoringUnavailable') }}</h2>
          <p>{{ vTargetsError }}</p>
        </div>
        <UiButton variant="secondary" @click="retryVTargets">
          {{ t('common.actions.retry') }}
        </UiButton>
      </section>

      <section class="campaign-detail-tabs" :aria-label="t('pages.campaignDetail.tabsAria')">
          <div class="campaign-detail-tab-list" role="tablist" :aria-label="t('pages.campaignDetail.tabsAria')">
            <button
              v-for="tab in tabs"
              :id="`campaign-detail-tab-${tab.id}`"
              :key="tab.id"
              type="button"
              class="campaign-detail-tab"
              :class="{ 'is-active': activeTab === tab.id }"
              role="tab"
              :aria-selected="activeTab === tab.id"
              :aria-controls="`campaign-detail-${tab.id}`"
              @click="activeTab = tab.id"
            >
              {{ tab.label }}
            </button>
          </div>

        <div
          class="campaign-detail-tab-panel"
          role="tabpanel"
          :id="activeTabPanelId"
          :aria-labelledby="`campaign-detail-tab-${activeTab.toLowerCase()}`"
        >
          <template v-if="activeTab === 'overview'">
            <section class="campaign-overview-grid" :aria-label="t('pages.campaignDetail.tabs.overview')">
              <AnalyticsSlot :title="t('pages.campaignDetail.overview.riskTitle')" :description="t('pages.campaignDetail.overview.riskDescription')">
                <div class="campaign-risk-chart-shell">
                  <span class="campaign-risk-help campaign-risk-help--chart">
                    <button
                      type="button"
                      class="campaign-risk-help-button"
                      :aria-label="t('pages.campaignDetail.overview.riskFormula')"
                    >
                      <Info :size="14" stroke-width="1.8" aria-hidden="true" />
                    </button>
                    <span class="campaign-risk-tooltip" role="tooltip">
                      {{ t('pages.campaignDetail.overview.riskFormulaText') }}
                    </span>
                  </span>
                  <CampaignRiskGauge
                    :value="campaignRiskGaugeValue"
                    :empty-text="t('charts.risk.empty')"
                  />
                </div>
              </AnalyticsSlot>

              <AnalyticsSlot :title="t('pages.campaignDetail.overview.eventDistributionTitle')" :description="t('pages.campaignDetail.overview.eventDistributionDescription')">
                <EventDistributionDonutChart
                  :items="eventDistributionItems"
                  :empty-text="t('charts.eventDistribution.empty')"
                />
              </AnalyticsSlot>

              <section class="campaign-funnel-panel" aria-labelledby="campaign-funnel-title">
                <header class="campaign-detail-panel-header">
                  <div>
                    <h2 id="campaign-funnel-title">{{ t('pages.campaignDetail.overview.funnelTitle') }}</h2>
                    <p>{{ t('pages.campaignDetail.overview.funnelDescription') }}</p>
                  </div>
                </header>

                <div v-if="isVTargetsLoading && !hasLoadedVTargets" class="campaign-inline-state">
                  <SkeletonBlock :rows="3" />
                </div>
                <div
                  v-else-if="vTargetsError && !hasLoadedVTargets"
                  class="campaign-detail-empty-panel campaign-detail-empty-panel--error"
                >
                  {{ t('pages.campaignDetail.overview.monitoringLoadFailed') }}
                </div>
                <div v-else-if="hasLoadedVTargets && !vtargets.length" class="campaign-detail-empty-panel">
                  {{ t('pages.campaignDetail.overview.noTargets') }}
                </div>

                <div v-else class="campaign-chart-frame">
                  <CampaignFunnelChart
                    :stages="funnelStages"
                    :empty-text="t('charts.funnel.empty')"
                  />
                </div>
              </section>

              <AnalyticsSlot
                class="campaign-overview-wide"
                :title="t('pages.campaignDetail.overview.eventsOverTimeTitle')"
                :description="t('pages.campaignDetail.overview.eventsOverTimeDescription')"
              >
                <EventsOverTimeChart
                  :buckets="eventsOverTimeBuckets"
                  :empty-text="eventsOverTimeEmptyText"
                />
              </AnalyticsSlot>
            </section>
          </template>

          <template v-else-if="activeTab === 'targets'">
            <section class="campaign-monitoring-console" aria-labelledby="campaign-targets-title">
              <section class="campaign-detail-panel campaign-targets-pane">
                <header class="campaign-detail-panel-header">
                  <div>
                    <h2 id="campaign-targets-title">{{ t('pages.campaignDetail.targets.title') }}</h2>
                    <p>{{ t('pages.campaignDetail.targets.description') }}</p>
                  </div>
                </header>

                <div class="campaign-target-toolbar" :aria-label="t('pages.campaignDetail.targets.filtersAria')">
                  <div class="campaign-target-summary">
                    <span><strong>{{ targetSummary.total }}</strong> {{ t('pages.campaignDetail.targets.summary.targets') }}</span>
                    <span><strong>{{ targetSummary.scheduled }}</strong> {{ t('pages.campaignDetail.targets.summary.scheduled') }}</span>
                    <span><strong>{{ targetSummary.unscheduled }}</strong> {{ t('pages.campaignDetail.targets.summary.unscheduled') }}</span>
                    <span><strong>{{ targetSummary.activity }}</strong> {{ t('pages.campaignDetail.targets.summary.activityRecords') }}</span>
                    <span><strong>{{ targetSummary.visible }}</strong> {{ t('pages.campaignDetail.targets.summary.visible') }}</span>
                  </div>
                  <label class="campaign-target-filter">
                    <span>{{ t('common.labels.status') }}</span>
                    <select v-model="targetStatusFilter" class="ui-select">
                      <option value="">{{ t('pages.campaignDetail.targets.allStatuses') }}</option>
                      <option
                        v-for="status in TARGET_STATUSES"
                        :key="status"
                        :value="status"
                      >
                        {{ displayTargetStatus(status) }}
                      </option>
                    </select>
                  </label>
                </div>

                <section class="campaign-schedule-panel" aria-labelledby="campaign-schedule-title">
                  <header class="campaign-schedule-header">
                    <div>
                      <h3 id="campaign-schedule-title">{{ t('pages.campaignDetail.targets.scheduleTitle') }}</h3>
                      <p>{{ t('pages.campaignDetail.targets.scheduleDescription') }}</p>
                    </div>
                    <UiBadge :variant="scheduledTargetRows.length ? 'primary' : 'neutral'">
                      {{ scheduledTargetRows.length }} {{ t('common.labels.scheduled') }}
                    </UiBadge>
                  </header>
                  <CampaignScheduleTimelineChart
                    :targets="targetRows"
                    :empty-text="t('charts.schedule.empty')"
                  />
                </section>

                <div v-if="isVTargetsLoading && !hasLoadedVTargets" class="campaign-inline-state">
                  <SkeletonBlock :rows="4" />
                </div>

                <div
                  v-else-if="vTargetsError && !hasLoadedVTargets"
                  class="campaign-detail-empty-panel campaign-detail-empty-panel--error"
                >
                  {{ t('pages.campaignDetail.targets.monitoringTargetLoadFailed') }}
                </div>

                <div v-else-if="hasLoadedVTargets && !targetRows.length" class="campaign-detail-empty-panel">
                  {{ t('pages.campaignDetail.targets.noTargets') }}
                </div>

                <div v-else-if="!filteredTargetRows.length" class="campaign-detail-empty-panel">
                  {{ t('pages.campaignDetail.targets.noMatchingTargets') }}
                </div>

                <div v-else class="campaign-detail-table" role="table" :aria-label="t('pages.campaignDetail.targets.title')">
                  <div class="campaign-detail-table-row campaign-detail-table-head campaign-detail-target-head" role="row">
                    <span v-for="column in targetColumns" :key="column" role="columnheader">
                      {{ column }}
                    </span>
                  </div>

                  <div
                    v-for="target in filteredTargetRows"
                    :key="target.id"
                    class="campaign-detail-table-row campaign-detail-target-row campaign-detail-target-button"
                    :class="{ 'is-selected': target.id === selectedTargetId && targetDrawerOpen }"
                    role="row"
                  >
                    <span role="cell">
                      <span class="campaign-detail-table-value campaign-detail-table-value--strong">
                        {{ target.employeeName }}
                      </span>
                    </span>
                    <span role="cell">
                      <span class="campaign-detail-table-value" :title="campaignTitle">
                        {{ campaignTitle }}
                      </span>
                    </span>
                    <span role="cell">
                      <UiBadge :variant="targetStatusVariant(target.status)">
                        {{ displayTargetStatus(target.status) }}
                      </UiBadge>
                    </span>
                    <span role="cell">{{ formatDateTime(target.scheduled_at) }}</span>
                    <span role="cell">{{ formatDateTime(target.created_at) }}</span>
                    <span role="cell">
                      <code class="campaign-detail-technical" :title="target.token || undefined">
                        {{ formatTechnicalId(target.token) }}
                      </code>
                    </span>
                    <span role="cell">
                      <span class="campaign-target-row-actions">
                          <IconButton :label="t('pages.campaignDetail.targets.inspect')" variant="secondary" @click="inspectTarget(target)">
                            <Eye :size="16" stroke-width="1.8" aria-hidden="true" />
                          </IconButton>
                      </span>
                    </span>
                  </div>
                </div>
              </section>
            </section>
          </template>

          <template v-else>
            <section class="campaign-detail-panel" aria-labelledby="campaign-settings-title">
              <header class="campaign-detail-panel-header">
                <div>
                  <h2 id="campaign-settings-title">{{ t('pages.campaignDetail.tabs.settings') }}</h2>
                  <p>{{ t('pages.campaignDetail.labels.settingsDescription') }}</p>
                </div>
              </header>

              <div class="campaign-settings-grid" :aria-label="t('pages.campaignDetail.tabs.settings')">
                <div
                  v-for="item in settings"
                  :key="item.label"
                  class="campaign-setting-field"
                >
                  <span>{{ item.label }}</span>
                  <code
                    v-if="item.technical"
                    class="campaign-setting-code"
                    :title="item.title || undefined"
                  >
                    {{ item.value }}
                  </code>
                  <span v-else class="campaign-setting-value">
                    {{ item.value }}
                  </span>
                </div>
              </div>

              <section class="campaign-attributes-panel" :aria-label="t('pages.campaignDetail.labels.attributes')">
                <header class="campaign-attributes-header">
                  <h3>{{ t('common.labels.attributes') }}</h3>
                  <p>{{ t('pages.campaignDetail.labels.attributesDescription') }}</p>
                </header>

                <div v-if="attributes.length" class="campaign-attributes-list">
                  <div
                    v-for="attribute in attributes"
                    :key="attribute.key"
                    class="campaign-attribute-row"
                  >
                    <code>{{ attribute.key }}</code>
                    <span>{{ attribute.value }}</span>
                  </div>
                </div>
                <div v-else class="campaign-attributes-empty">—</div>
              </section>
            </section>
          </template>
        </div>
      </section>

      <PreviewDrawer
        :open="targetDrawerOpen && Boolean(selectedTarget)"
        :title="t('pages.campaignDetail.labels.targetActivity')"
        :subtitle="t('pages.campaignDetail.targets.drawerDescription')"
        :initial-width="620"
        :min-width="440"
        @close="closeTargetDrawer"
      >
        <template v-if="selectedTarget">
          <header class="campaign-target-card">
            <span class="campaign-target-avatar" aria-hidden="true">
              <UserCircle :size="28" stroke-width="1.6" />
            </span>
            <div>
              <h2>{{ selectedTarget.employeeName }}</h2>
              <p>{{ formatEmployeeContact(selectedTarget) }}</p>
            </div>
            <UiBadge :variant="targetStatusVariant(selectedTarget.status)">
              {{ displayTargetStatus(selectedTarget.status) }}
            </UiBadge>
          </header>

          <dl class="campaign-target-details">
            <div>
              <dt>{{ t('pages.campaignDetail.labels.campaign') }}</dt>
              <dd>{{ campaignTitle }}</dd>
            </div>
            <div>
              <dt>{{ t('common.labels.token') }}</dt>
              <dd>
                <code :title="selectedTarget.token || undefined">
                  {{ formatTechnicalId(selectedTarget.token) }}
                </code>
              </dd>
            </div>
            <div>
              <dt>{{ t('pages.campaignDetail.labels.scheduledAt') }}</dt>
              <dd>{{ formatDateTime(selectedTarget.scheduled_at) }}</dd>
            </div>
            <div>
              <dt>{{ t('pages.campaignDetail.labels.createdAt') }}</dt>
              <dd>{{ formatDateTime(selectedTarget.created_at) }}</dd>
            </div>
          </dl>

          <section class="campaign-activity-panel" aria-labelledby="target-activity-title">
            <header>
              <h3 id="target-activity-title">{{ t('pages.campaignDetail.labels.activityHistory') }}</h3>
              <p>{{ t('pages.campaignDetail.targets.activityDescription') }}</p>
            </header>

            <div v-if="selectedTargetEvents.length" class="campaign-activity-timeline">
              <article
                v-for="(event, eventIndex) in selectedTargetEvents"
                :key="`${selectedTarget.id}-${event.type}-${event.occurred_at}-${eventIndex}`"
                class="campaign-activity-item"
              >
                <span class="campaign-activity-icon" aria-hidden="true">
                  <component :is="eventIcon(event.type)" :size="16" stroke-width="1.8" />
                </span>
                <div class="campaign-activity-content">
                  <div class="campaign-activity-heading">
                    <strong>{{ displayEventType(event.type) }}</strong>
                    <time>{{ formatDateTime(event.occurred_at) }}</time>
                  </div>
                  <dl>
                    <div v-if="event.ip_address">
                      <dt>IP</dt>
                      <dd>{{ formatNullable(event.ip_address) }}</dd>
                    </div>
                    <div v-if="event.user_agent">
                      <dt>User Agent</dt>
                      <dd :title="event.user_agent">{{ event.user_agent }}</dd>
                    </div>
                    <div v-if="event.referer">
                      <dt>Referer</dt>
                      <dd :title="event.referer">{{ event.referer }}</dd>
                    </div>
                  </dl>
                </div>
              </article>
            </div>
            <div v-else class="campaign-detail-empty-panel campaign-target-empty-panel">
              {{ t('pages.campaignDetail.targets.noActivity') }}
            </div>
          </section>
        </template>
      </PreviewDrawer>

      <ConfirmDialog
        :open="Boolean(confirmationDialog)"
        :title="confirmationDialog?.title || ''"
        :message="confirmationDialog?.text || ''"
        :confirm-label="confirmationDialog?.confirmLabel || t('common.actions.confirm')"
        :loading-label="confirmationDialog?.loadingLabel || ''"
        :cancel-label="t('common.actions.keepCampaign')"
        :variant="confirmationDialog?.variant || 'primary'"
        :loading="Boolean(confirmationDialog && isActionLoading(confirmationDialog.action))"
        @cancel="closeConfirmation"
        @confirm="runConfirmedLifecycleAction"
      />

    </template>
  </div>
</template>
