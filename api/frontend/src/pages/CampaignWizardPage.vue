<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { CalendarClock, Plus, Shuffle, Trash2 } from 'lucide-vue-next'
import CampaignScheduleTimelineChart from '../components/charts/CampaignScheduleTimelineChart.vue'
import AttachmentPickerPanel from '../components/ui/AttachmentPickerPanel.vue'
import CampaignResourcePickerPanel from '../components/ui/CampaignResourcePickerPanel.vue'
import ConfirmDialog from '../components/ui/ConfirmDialog.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import HtmlAttachmentPreview from '../components/ui/HtmlAttachmentPreview.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import PreviewDrawer from '../components/ui/PreviewDrawer.vue'
import ResourcePagination from '../components/ui/ResourcePagination.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiDatePicker from '../components/ui/UiDatePicker.vue'
import UiInput from '../components/ui/UiInput.vue'
import UiTimePicker from '../components/ui/UiTimePicker.vue'
import WizardStepper from '../components/ui/WizardStepper.vue'
import { useNotifications } from '../composables/useNotifications'
import { useSession } from '../composables/useSession'
import { translateTargetStatus } from '../i18n'
import { listAttachments } from '../resources/attachments'
import {
  CAMPAIGN_TYPES,
  createCampaign,
  getCampaign,
  isCampaignAuthError,
  isCampaignNotFoundError,
  updateCampaign,
} from '../resources/campaigns'
import { listDepartments } from '../resources/departments'
import { listEmployees } from '../resources/employees'
import { listLandings } from '../resources/landings'
import { listMessages, MESSAGE_TYPES } from '../resources/messages'
import {
  createTarget,
  deleteTarget,
  distributeCampaignTargets,
  scheduleTarget,
} from '../resources/targets'
import { listVTargets } from '../resources/vtargets'
import {
  formatDate as formatLocalizedDate,
  formatDateTime as formatLocalizedDateTime,
  formatNullable as formatSharedNullable,
  formatTechnicalId as formatSharedTechnicalId,
} from '../utils/resourceFormat'

const route = useRoute()
const router = useRouter()
const { csrfToken, loadCurrentSession } = useSession()
const { notifySuccess, notifyError, notifyWarning } = useNotifications()
const { t } = useI18n()

let attributeRowSequence = 0
let referenceController = null
let submitController = null
let editCampaignController = null
let editTargetsController = null
let targetMutationController = null

const activeStepIndex = ref(0)
const editCampaign = ref(null)
const editLoading = ref(false)
const editNotFound = ref(false)
const editLoadError = ref('')
const editAuthSuppressed = ref(false)
const messages = ref([])
const landings = ref([])
const attachments = ref([])
const employees = ref([])
const departments = ref([])
const referenceLoading = ref(false)
const referenceLoaded = ref(false)
const referenceErrors = ref(referenceErrorState())
const selectedMessageId = ref('')
const selectedLandingId = ref('')
const selectedEducationId = ref('')
const selectedMaxEducationTextId = ref('')
const selectedEmployeeIds = ref(new Set())
const messageSearch = ref('')
const landingSearch = ref('')
const educationSearch = ref('')
const employeeSearch = ref('')
const departmentFilter = ref('')
const targetPage = ref(1)
const targetRows = ref(10)
const formErrors = ref({})
const submitError = ref('')
const pendingSubmit = ref('')
const submitStage = ref('')
const createdCampaign = ref(null)
const partialFailure = ref(null)
const discardDialogOpen = ref(false)
const suppressDirtyGuard = ref(false)
const cleanSnapshot = ref('')
const editTargets = ref([])
const editTargetsLoading = ref(false)
const editTargetsError = ref('')
const editTargetsLoaded = ref(false)
const scheduleDrafts = ref({})
const targetMutationError = ref('')
const pendingTargetAction = ref('')
const focusedScheduleTargetId = ref('')
const targetDeleteCandidate = ref(null)
const resourcePreviewOpen = ref(false)
const resourcePreviewTitle = ref('Resource preview')
const resourcePreviewKind = ref('')
const resourcePreviewAttachment = ref(null)
const resourcePreviewMeta = ref([])

const form = ref({
  type: 'EMAIL',
  label: '',
  domain: '',
  includeSite: true,
  dateFrom: '',
  dateTo: '',
  attributes: [newAttributeRow()],
})

const isEditMode = computed(() => route.name === 'campaign-edit')
const campaignId = computed(() => String(route.params.id ?? ''))
const isEmailCampaign = computed(() => form.value.type === 'EMAIL')
const isMaxCampaign = computed(() => form.value.type === 'MAX')
const campaignUsesSite = computed(() => isEmailCampaign.value || Boolean(form.value.includeSite))
const stepDefinitions = computed(() => ([
  { key: 'basic', label: t('pages.campaignWizard.steps.basicInfo') },
  { key: 'message', label: t('pages.campaignWizard.steps.message') },
  ...(campaignUsesSite.value ? [{ key: 'landing', label: t('pages.campaignWizard.steps.landing') }] : []),
  { key: 'education', label: t('pages.campaignWizard.steps.education') },
  { key: 'targets', label: t('pages.campaignWizard.steps.targets') },
  { key: 'review', label: t('pages.campaignWizard.steps.review') },
]))
const steps = computed(() => stepDefinitions.value.map((step) => step.label))
const activeStepDefinition = computed(() => stepDefinitions.value[activeStepIndex.value] || stepDefinitions.value[0])
const activeStep = computed(() => activeStepDefinition.value?.label || '')
const activeStepKey = computed(() => activeStepDefinition.value?.key || 'basic')
const showPanelHeader = computed(() => !['message', 'landing', 'education', 'targets'].includes(activeStepKey.value))
const isFirstStep = computed(() => activeStepIndex.value === 0)
const isFinalStep = computed(() => activeStepIndex.value === steps.value.length - 1)
const isSubmitting = computed(() => Boolean(pendingSubmit.value))
const canShowWizard = computed(() => (
  !isEditMode.value
    || (!editLoading.value && !editNotFound.value && !editLoadError.value && !editAuthSuppressed.value)
))
const pageTitle = computed(() => (isEditMode.value ? t('pages.campaignWizard.editTitle') : t('pages.campaignWizard.createTitle')))
const pageDescription = computed(() => (
  isEditMode.value
    ? t('pages.campaignWizard.editDescription')
    : t('pages.campaignWizard.createDescription')
))
const canMutateTargetsInWizard = computed(() => (
  isEditMode.value ? isCampaignEditableForUpdate(editCampaign.value) : true
))
const targetMutationDisabledReason = computed(() => {
  if (!isEditMode.value || canMutateTargetsInWizard.value) {
    return ''
  }

  return t('pages.campaignWizard.targetMutationDisabled')
})
const canAutoDistributeSchedule = computed(() => {
  return canMutateTargetsInWizard.value
    && Boolean(form.value.dateFrom && form.value.dateTo)
    && allScheduleRows.value.length > 0
})
const autoDistributeDisabledReason = computed(() => {
  if (!canMutateTargetsInWizard.value) {
    return targetMutationDisabledReason.value
  }

  if (!form.value.dateFrom || !form.value.dateTo) {
    return t('pages.campaignWizard.validation.dateRangeRequiredForDistribution')
  }

  if (!allScheduleRows.value.length) {
    return isEditMode.value ? t('pages.campaignWizard.validation.addTargetBeforeDistribution') : t('pages.campaignWizard.validation.selectEmployee')
  }

  return ''
})

const selectedMessage = computed(() => messages.value.find((item) => itemId(item) === selectedMessageId.value) || null)
const selectedLanding = computed(() => landings.value.find((item) => itemId(item) === selectedLandingId.value) || null)
const selectedEducation = computed(() => landings.value.find((item) => itemId(item) === selectedEducationId.value) || null)
const selectedEducationAttachment = computed(() => attachmentById.value[nullableId(selectedEducation.value?.html_body_id)] || null)
const selectedMessageTextBodyAttachment = computed(() => attachmentById.value[nullableId(selectedMessage.value?.text_body_id)] || null)
const selectedMaxEducationAttachment = computed(() => attachmentById.value[selectedMaxEducationTextId.value] || null)
const maxMessageUsesLink = computed(() => attachmentUsesTargetLink(selectedMessageTextBodyAttachment.value))
const maxMessageRequiresSite = computed(() => isMaxCampaign.value && maxMessageUsesLink.value && !campaignUsesSite.value)
const campaignTypeOptions = computed(() => CAMPAIGN_TYPES.map((type) => ({
  value: type,
  label: type === 'MAX' ? t('pages.emailTemplates.tabs.max') : t('pages.emailTemplates.tabs.email'),
})))
const selectedEmployees = computed(() => employees.value.filter((employee) => selectedEmployeeIds.value.has(itemId(employee))))
const selectedTargetCount = computed(() => selectedEmployeeIds.value.size)
const existingTargetEmployeeIds = computed(() => {
  return new Set(editTargets.value.map((target) => nullableId(target.employee_id) || itemId(target.employee_id)).filter(Boolean))
})
const editTargetRows = computed(() => {
  return editTargets.value.map((target) => normalizeVTargetRow(target))
})
const allScheduleRows = computed(() => (isEditMode.value ? editTargetRows.value : selectedEmployees.value.map((employee) => ({
  id: itemId(employee),
  employee_id: itemId(employee),
  employeeName: employeeName(employee),
  email: employee.email || '',
  status: 'PENDING',
  scheduled_at: '',
  created_at: '',
  token: '',
}))))
const scheduleTimelineRows = computed(() => allScheduleRows.value.map((row) => ({
  ...row,
  scheduled_at: scheduleDraftToIso(scheduleRowKey(row)) || row.scheduled_at,
})))
const scheduledScheduleCount = computed(() => scheduleTimelineRows.value.filter((target) => target.scheduled_at).length)
const unscheduledScheduleCount = computed(() => Math.max(allScheduleRows.value.length - scheduledScheduleCount.value, 0))
const hasAttributeDraft = computed(() => {
  return form.value.attributes.some((row) => row.key.trim() || String(row.value).trim())
})
const isWizardDirty = computed(() => {
  if (suppressDirtyGuard.value) {
    return false
  }

  if (isEditMode.value) {
    return Boolean(cleanSnapshot.value && cleanSnapshot.value !== wizardSnapshot())
  }

  return Boolean(
    form.value.label.trim()
      || form.value.domain.trim()
      || form.value.includeSite !== true
      || form.value.type !== 'EMAIL'
      || form.value.dateFrom
      || form.value.dateTo
      || hasAttributeDraft.value
      || selectedMessageId.value
      || selectedLandingId.value
      || selectedEducationId.value
      || selectedMaxEducationTextId.value
      || selectedEmployeeIds.value.size
      || Object.values(scheduleDrafts.value).some(Boolean),
  )
})

const departmentById = computed(() => {
  return Object.fromEntries(departments.value.map((department) => [itemId(department), department]))
})

const attachmentById = computed(() => {
  return Object.fromEntries(attachments.value.map((attachment) => [itemId(attachment), attachment]))
})

const normalizedAttributes = computed(() => {
  return form.value.attributes.reduce((result, row) => {
    const key = row.key.trim()
    if (key) {
      result[key] = row.value
    }
    return result
  }, {})
})

const filteredMessages = computed(() => {
  const query = messageSearch.value.trim().toLowerCase()
  const typedMessages = messages.value.filter((message) => {
    const type = MESSAGE_TYPES.includes(message?.type) ? message.type : 'EMAIL'
    return type === form.value.type
  })

  if (!query) {
    return typedMessages
  }

  return typedMessages.filter((message) => {
    return [
      message.label,
      message.from_email,
      message.from_name,
      message.subject,
      message.max_account_id,
      message.text_body_id,
    ]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(query))
  })
})

const filteredLandings = computed(() => filterLandingsByLabel(landings.value, landingSearch.value))
const filteredEducationLandings = computed(() => filterLandingsByLabel(landings.value, educationSearch.value))

const filteredEmployees = computed(() => {
  const query = employeeSearch.value.trim().toLowerCase()
  const departmentId = departmentFilter.value

  return employees.value.filter((employee) => {
    if (isEditMode.value && existingTargetEmployeeIds.value.has(itemId(employee))) {
      return false
    }

    if (departmentId && itemId(employee.department_id) !== departmentId) {
      return false
    }

    if (!query) {
      return true
    }

    return [
      employeeName(employee),
      employee.email,
      employee.phone,
      departmentLabel(employee.department_id),
    ]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(query))
  })
})
const targetTotalPages = computed(() => Math.max(1, Math.ceil(filteredEmployees.value.length / targetRows.value)))
const pagedEmployees = computed(() => {
  const start = (targetPage.value - 1) * targetRows.value
  return filteredEmployees.value.slice(start, start + targetRows.value)
})

const reviewWarnings = computed(() => {
  const warnings = []

  if (!form.value.label.trim()) {
    warnings.push(t('pages.campaignWizard.validation.labelMissing'))
  }

  if (maxMessageRequiresSite.value) {
    warnings.push(t('pages.campaignWizard.validation.messageRequiresSite'))
  }

  if (campaignUsesSite.value && !form.value.domain.trim()) {
    warnings.push(t('pages.campaignWizard.validation.domainMissing'))
  } else if (form.value.domain.trim() && !isValidDomain(form.value.domain)) {
    warnings.push(t('pages.campaignWizard.validation.domainInvalid'))
  }

  const hasDateFrom = Boolean(form.value.dateFrom)
  const hasDateTo = Boolean(form.value.dateTo)
  if (!hasDateFrom || !hasDateTo) {
    warnings.push(t('pages.campaignWizard.validation.dateRangeMissing'))
  } else if (!isValidDateRange()) {
    warnings.push(t('pages.campaignWizard.validation.dateRangeInvalid'))
  }

  if (!selectedMessage.value) {
    warnings.push(t('pages.campaignWizard.validation.messageMissing'))
  }

  if (campaignUsesSite.value && !selectedLanding.value) {
    warnings.push(t('pages.campaignWizard.validation.landingMissing'))
  }

  if (campaignUsesSite.value && !selectedEducation.value) {
    warnings.push(t('pages.campaignWizard.validation.educationMissing'))
  }

  if (isMaxCampaign.value && !selectedMaxEducationAttachment.value) {
    warnings.push(t('pages.campaignWizard.validation.maxEducationMissing'))
  }

  if (!isEditMode.value && !selectedTargetCount.value) {
    warnings.push(t('pages.campaignWizard.validation.targetsMissing'))
  }

  return warnings
})

function referenceErrorState() {
  return {
    messages: '',
    landings: '',
    attachments: '',
    employees: '',
    departments: '',
  }
}

function newAttributeRow() {
  attributeRowSequence += 1
  return {
    id: `attribute-${attributeRowSequence}`,
    key: '',
    value: '',
  }
}

function itemId(item) {
  if (!item) {
    return ''
  }

  if (typeof item === 'string') {
    return item
  }

  if (item.UUID || item.uuid) {
    return String(item.UUID || item.uuid)
  }

  return item.id ? String(item.id) : ''
}

function nullableId(value) {
  if (!value) {
    return ''
  }

  if (typeof value === 'string') {
    return value
  }

  if (value.Valid === false || value.valid === false) {
    return ''
  }

  return String(value.UUID || value.uuid || value.id || '')
}

function formatTechnicalId(value) {
  return formatSharedTechnicalId(nullableId(value) || itemId(value))
}

function formatNullable(value) {
  return formatSharedNullable(value)
}

function employeeName(employee) {
  return [employee?.first_name, employee?.last_name].filter(Boolean).join(' ').trim() || '—'
}

function vTargetEmployeeName(target) {
  return [target?.first_name, target?.last_name].filter(Boolean).join(' ').trim() || '—'
}

function normalizeVTargetRow(target) {
  const id = itemId(target)
  return {
    ...target,
    id,
    employee_id: nullableId(target?.employee_id) || itemId(target?.employee_id),
    employeeName: vTargetEmployeeName(target),
    scheduled_at: target?.scheduled_at || '',
    created_at: target?.created_at || '',
    status: target?.status || 'PENDING',
    token: target?.token || '',
  }
}

function departmentLabel(departmentId) {
  return departmentById.value[nullableId(departmentId) || itemId(departmentId)]?.label || '—'
}

function formatDateTime(value) {
  return formatLocalizedDate(value)
}

function formatReviewDateRange() {
  const from = form.value.dateFrom ? formatDateTime(form.value.dateFrom) : '—'
  const to = form.value.dateTo ? formatDateTime(form.value.dateTo) : '—'

  if (from === '—' && to === '—') {
    return '—'
  }

  return `${from} — ${to}`
}

function formatFullDateTime(value) {
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

  const datePart = [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, '0'),
    String(date.getDate()).padStart(2, '0'),
  ].join('-')
  const timePart = [
    String(date.getHours()).padStart(2, '0'),
    String(date.getMinutes()).padStart(2, '0'),
  ].join(':')

  return `${datePart}T${timePart}`
}

function datetimeLocalToIso(value) {
  if (!value) {
    return ''
  }

  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : date.toISOString()
}

function isCompleteScheduleDraft(value) {
  const match = /^(\d{4}-\d{2}-\d{2})T(([01]\d|2[0-3]):[0-5]\d)$/.exec(String(value || ''))
  if (!match) {
    return false
  }

  const [year, month, day] = match[1].split('-').map(Number)
  const date = new Date(year, month - 1, day)

  return date.getFullYear() === year
    && date.getMonth() === month - 1
    && date.getDate() === day
}

function scheduleRowKey(row) {
  return itemId(row?.id) || nullableId(row?.employee_id) || itemId(row?.employee_id)
}

function getScheduleDraftDate(key) {
  return scheduleDrafts.value[key]?.split('T')[0] || ''
}

function getScheduleDraftTime(key) {
  return scheduleDrafts.value[key]?.split('T')[1] || ''
}

function scheduleDraftToIso(key) {
  const draft = scheduleDrafts.value[key]
  return isCompleteScheduleDraft(draft) ? datetimeLocalToIso(draft) : ''
}

function scheduleDraftInCampaignRange(value) {
  if (!value || !isCompleteScheduleDraft(value) || !form.value.dateFrom || !form.value.dateTo) {
    return true
  }

  const timestamp = new Date(value).getTime()
  const start = new Date(`${form.value.dateFrom}T00:00:00.000`).getTime()
  const end = new Date(`${form.value.dateTo}T23:59:59.999`).getTime()

  return timestamp >= start && timestamp <= end
}

function toDateInputValue(value) {
  if (!value) {
    return ''
  }

  const text = String(value)
  const datePart = text.slice(0, 10)
  if (/^\d{4}-\d{2}-\d{2}$/.test(datePart)) {
    return datePart
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }

  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, '0'),
    String(date.getDate()).padStart(2, '0'),
  ].join('-')
}

function filterLandingsByLabel(items, searchValue) {
  const query = searchValue.trim().toLowerCase()

  if (!query) {
    return items
  }

  return items.filter((landing) => String(landing.label || '').toLowerCase().includes(query))
}

function attachmentUsesTargetLink(attachment) {
  return Array.isArray(attachment?.required_vars)
    && attachment.required_vars.includes('Target.Link')
}

function isValidDomain(value) {
  try {
    const parsed = new URL(value.trim())
    return (parsed.protocol === 'http:' || parsed.protocol === 'https:') && Boolean(parsed.host)
  } catch {
    return false
  }
}

function isValidDateRange() {
  if (!form.value.dateFrom || !form.value.dateTo) {
    return true
  }

  return new Date(`${form.value.dateTo}T23:59:59.999`).getTime()
    >= new Date(`${form.value.dateFrom}T00:00:00.000`).getTime()
}

function toRFC3339Date(value, edge) {
  if (!value) {
    return undefined
  }

  const [year, month, day] = value.split('-').map(Number)
  if (!year || !month || !day) {
    return undefined
  }

  const date = edge === 'end'
    ? new Date(year, month - 1, day, 23, 59, 59, 999)
    : new Date(year, month - 1, day, 0, 0, 0, 0)

  return date.toISOString()
}

function attributesToRows(attributes) {
  const entries = Object.entries(attributes || {})

  if (!entries.length) {
    return [newAttributeRow()]
  }

  return entries
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([key, value]) => ({
      id: newAttributeRow().id,
      key,
      value: value == null ? '' : String(value),
    }))
}

function wizardSnapshot() {
  return JSON.stringify({
    type: form.value.type,
    label: form.value.label.trim(),
    domain: form.value.domain.trim(),
    includeSite: form.value.includeSite,
    dateFrom: form.value.dateFrom,
    dateTo: form.value.dateTo,
    attributes: normalizedAttributes.value,
    messageId: selectedMessageId.value,
    landingId: selectedLandingId.value,
    educationId: selectedEducationId.value,
    maxEducationTextId: selectedMaxEducationTextId.value,
    selectedEmployeeIds: Array.from(selectedEmployeeIds.value).sort(),
    scheduleDrafts: scheduleDrafts.value,
  })
}

function markWizardSnapshotClean() {
  cleanSnapshot.value = wizardSnapshot()
}

function goToStep(index) {
  activeStepIndex.value = index
}

function goBack() {
  if (!isFirstStep.value) {
    activeStepIndex.value -= 1
  }
}

function goNext() {
  if (!isFinalStep.value) {
    activeStepIndex.value += 1
  }
}

function addAttribute() {
  form.value.attributes.push(newAttributeRow())
}

function removeAttribute(id) {
  if (form.value.attributes.length === 1) {
    form.value.attributes = [newAttributeRow()]
    return
  }

  form.value.attributes = form.value.attributes.filter((row) => row.id !== id)
}

function selectMessage(message) {
  selectedMessageId.value = itemId(message)
  clearSubmitState()
}

function handleMessageSelection() {
  clearSubmitState()
}

function selectLanding(landing) {
  selectedLandingId.value = itemId(landing)
  clearSubmitState()
}

function handleLandingSelection() {
  clearSubmitState()
}

function selectEducation(landing) {
  selectedEducationId.value = itemId(landing)
  clearSubmitState()
}

function handleEducationSelection() {
  clearSubmitState()
}

function handleMaxEducationSelection() {
  clearSubmitState()
}

function closeResourcePreview() {
  resourcePreviewOpen.value = false
  resourcePreviewAttachment.value = null
  resourcePreviewMeta.value = []
  resourcePreviewKind.value = ''
}

function openResourcePreview({ title, kind, attachmentId, attachmentLabel, attachmentType = '.html', meta }) {
  if (!attachmentId) {
    notifyError(t('common.preview.unavailable'), t('common.preview.selectHtmlOrText'))
    return
  }

  resourcePreviewTitle.value = title
  resourcePreviewKind.value = kind
  resourcePreviewAttachment.value = {
    id: attachmentId,
    label: attachmentLabel,
    type: attachmentType,
  }
  resourcePreviewMeta.value = meta
  resourcePreviewOpen.value = true
}

function previewMessageResource(message) {
  const isMax = (message?.type || 'EMAIL') === 'MAX'
  const bodyId = nullableId(isMax ? message?.text_body_id : message?.html_body_id)
  const bodyAttachment = attachmentById.value[bodyId]

  openResourcePreview({
    title: isMax ? t('pages.emailTemplates.maxTextPreview') : t('pages.campaignWizard.preview.message'),
    kind: t('common.labels.message'),
    attachmentId: bodyId,
    attachmentLabel: message?.label || (isMax ? t('common.labels.textBody') : t('pages.campaignWizard.preview.messageHtmlBody')),
    attachmentType: bodyAttachment?.type || (isMax ? '.txt' : '.html'),
    meta: [
      [t('common.labels.name'), message?.label || t('common.placeholder')],
      isMax
        ? [t('pages.emailTemplates.maxAccount'), formatTechnicalId(message?.max_account_id)]
        : [t('common.labels.fromEmail'), message?.from_email || t('common.placeholder')],
      isMax
        ? [t('common.labels.textBody'), formatTechnicalId(message?.text_body_id)]
        : [t('common.labels.fromName'), message?.from_name || t('common.placeholder')],
      isMax
        ? [t('common.labels.attachments'), Array.isArray(message?.attachment_ids) ? String(message.attachment_ids.length) : '0']
        : [t('common.labels.subject'), message?.subject || t('common.placeholder')],
    ],
  })
}

function previewLandingResource(landing, kind = t('common.labels.landingPage')) {
  openResourcePreview({
    title: kind === t('common.labels.educationAsset') ? t('pages.campaignWizard.preview.education') : t('pages.campaignWizard.preview.landing'),
    kind,
    attachmentId: nullableId(landing?.html_body_id),
    attachmentLabel: landing?.label || t('pages.campaignWizard.preview.landingHtmlBody'),
    meta: [
      [t('common.labels.name'), landing?.label || t('common.placeholder')],
    ],
  })
}

function toggleEmployee(employee) {
  const id = itemId(employee)
  const next = new Set(selectedEmployeeIds.value)
  if (next.has(id)) {
    next.delete(id)
    if (!isEditMode.value) {
      const nextDrafts = { ...scheduleDrafts.value }
      delete nextDrafts[id]
      scheduleDrafts.value = nextDrafts
    }
  } else {
    next.add(id)
  }
  selectedEmployeeIds.value = next
  clearSubmitState()
}

function selectVisibleEmployees() {
  const next = new Set(selectedEmployeeIds.value)
  pagedEmployees.value.forEach((employee) => {
    const id = itemId(employee)
    if (id) {
      next.add(id)
    }
  })
  selectedEmployeeIds.value = next
  clearSubmitState()
}

function setTargetPage(nextPage) {
  targetPage.value = Math.min(Math.max(1, nextPage), targetTotalPages.value)
}

function setTargetRows(nextRows) {
  targetRows.value = nextRows
  targetPage.value = 1
}

function clearSelectedEmployees() {
  if (!isEditMode.value) {
    const nextDrafts = { ...scheduleDrafts.value }
    selectedEmployeeIds.value.forEach((id) => {
      delete nextDrafts[id]
    })
    scheduleDrafts.value = nextDrafts
  }
  selectedEmployeeIds.value = new Set()
  clearSubmitState()
}

function isEmployeeSelected(employee) {
  return selectedEmployeeIds.value.has(itemId(employee))
}

function clearSubmitState() {
  submitError.value = ''
  partialFailure.value = null
  targetMutationError.value = ''
}

function targetStatusVariant(status) {
  switch (status) {
    case 'SENT':
    case 'OPENED':
    case 'REPLIED':
      return 'primary'
    case 'CLICKED':
      return 'warning'
    case 'FAILED':
    case 'SUBMITTED':
      return 'danger'
    case 'PENDING':
    default:
      return 'neutral'
  }
}

function displayTargetStatus(status) {
  return status ? translateTargetStatus(status) : t('common.placeholder')
}

function canScheduleRow(row) {
  return canMutateTargetsInWizard.value && (!isEditMode.value || row.status === 'PENDING')
}

function scheduleDisabledReason(row) {
  if (!canMutateTargetsInWizard.value) {
    return targetMutationDisabledReason.value
  }

  if (isEditMode.value && row?.status !== 'PENDING') {
    return t('pages.campaignWizard.validation.onlyPendingCanSchedule')
  }

  return ''
}

function setScheduleDraft(key, value) {
  scheduleDrafts.value = {
    ...scheduleDrafts.value,
    [key]: value,
  }
  clearSubmitState()
}

function updateScheduleDraftDate(key, date) {
  const time = getScheduleDraftTime(key)
  setScheduleDraft(key, date || time ? `${date || ''}T${time || ''}` : '')
}

function updateScheduleDraftTime(key, time) {
  const date = getScheduleDraftDate(key)
  setScheduleDraft(key, date || time ? `${date || ''}T${time || ''}` : '')
}

function clearScheduleDraft(row) {
  setScheduleDraft(scheduleRowKey(row), '')
}

function focusScheduleTarget(target) {
  focusedScheduleTargetId.value = scheduleRowKey(target)
}

function buildLocalDistributionValues(rows) {
  if (!rows.length || !form.value.dateFrom || !form.value.dateTo) {
    return {}
  }

  const start = new Date(`${form.value.dateFrom}T00:00:00.000`).getTime()
  const end = new Date(`${form.value.dateTo}T23:59:59.999`).getTime()

  if (!Number.isFinite(start) || !Number.isFinite(end) || end < start) {
    return {}
  }

  const span = end - start
  const step = rows.length > 1 ? span / (rows.length - 1) : span / 2

  return Object.fromEntries(rows.map((row, index) => {
    const timestamp = rows.length > 1 ? start + (step * index) : start + step
    return [scheduleRowKey(row), toDatetimeLocalValue(new Date(timestamp).toISOString())]
  }))
}

function distributeLocalSchedule() {
  const values = buildLocalDistributionValues(allScheduleRows.value)

  if (!Object.keys(values).length) {
    targetMutationError.value = autoDistributeDisabledReason.value || t('pages.campaignWizard.errors.scheduleDistributionFailed')
    return false
  }

  scheduleDrafts.value = {
    ...scheduleDrafts.value,
    ...values,
  }
  focusedScheduleTargetId.value = ''
  notifySuccess(t('pages.campaignWizard.notifications.scheduleDistributedTitle'), t('pages.campaignWizard.notifications.scheduleDistributedMessage'))
  return true
}

async function runWizardAutoDistribute() {
  targetMutationError.value = ''

  if (!canAutoDistributeSchedule.value) {
    targetMutationError.value = autoDistributeDisabledReason.value
    return
  }

  if (!isEditMode.value) {
    distributeLocalSchedule()
    return
  }

  if (!(await ensureCsrfToken())) {
    await handleAuthFailure()
    return
  }

  abortTargetMutationRequest()
  const controller = new AbortController()
  targetMutationController = controller
  pendingTargetAction.value = 'distribute'

  try {
    await distributeCampaignTargets(campaignId.value, actionRequestOptions(controller))
    notifySuccess(t('pages.campaignWizard.notifications.targetsDistributedTitle'), t('pages.campaignWizard.notifications.targetsDistributedMessage'))
    await loadEditTargets()
    markWizardSnapshotClean()
  } catch (error) {
    if (error?.name === 'AbortError' || controller.signal.aborted) {
      return
    }

    if (isAuthError(error)) {
      await handleAuthFailure()
      return
    }

    targetMutationError.value = error?.message || t('pages.campaignWizard.errors.targetsDistributionFailed')
    notifyError(t('pages.campaignWizard.notifications.autoDistributeFailedTitle'), targetMutationError.value)
  } finally {
    if (targetMutationController === controller) {
      targetMutationController = null
    }
    pendingTargetAction.value = ''
  }
}

async function addSelectedTargetsInEdit() {
  targetMutationError.value = ''

  if (!isEditMode.value || !selectedTargetCount.value || pendingTargetAction.value) {
    return
  }

  if (!(await ensureCsrfToken())) {
    await handleAuthFailure()
    return
  }

  abortTargetMutationRequest()
  const controller = new AbortController()
  targetMutationController = controller
  pendingTargetAction.value = 'add-targets'
  const options = actionRequestOptions(controller)
  const failures = []
  let createdCount = 0

  try {
    for (const employee of selectedEmployees.value) {
      try {
        await createTarget({
          employee_id: itemId(employee),
          campaign_id: campaignId.value,
        }, options)
        createdCount += 1
      } catch (error) {
        if (isAuthError(error)) {
          throw error
        }
        failures.push({
          employee,
          message: error?.message || t('pages.campaignWizard.errors.targetCreateFailed'),
        })
      }
    }

    selectedEmployeeIds.value = new Set()
    await loadEditTargets()
    markWizardSnapshotClean()

    if (failures.length) {
      targetMutationError.value = `${createdCount} targets added. ${failures.length} target requests failed.`
      notifyWarning(t('pages.campaignWizard.notifications.targetsPartiallyAddedTitle'), targetMutationError.value)
      return
    }

    notifySuccess(t('pages.campaignWizard.notifications.targetsAddedTitle'), t('pages.campaignWizard.notifications.targetsAddedMessage', { count: createdCount }))
  } catch (error) {
    if (error?.name === 'AbortError' || controller.signal.aborted) {
      return
    }

    if (isAuthError(error)) {
      await handleAuthFailure()
      return
    }

    targetMutationError.value = error?.message || t('pages.campaignWizard.errors.targetsAddFailed')
    notifyError(t('pages.campaignWizard.notifications.addTargetsFailedTitle'), targetMutationError.value)
  } finally {
    if (targetMutationController === controller) {
      targetMutationController = null
    }
    pendingTargetAction.value = ''
  }
}

async function saveEditTargetSchedule(row) {
  const key = scheduleRowKey(row)
  const draft = scheduleDrafts.value[key]
  const scheduledAt = scheduleDraftToIso(key)
  targetMutationError.value = ''

  if (!draft) {
    targetMutationError.value = t('pages.campaignWizard.validation.scheduleRequired')
    return
  }

  if (!scheduledAt) {
    targetMutationError.value = t('pages.campaignWizard.validation.scheduleInvalid')
    return
  }

  if (!scheduleDraftInCampaignRange(draft)) {
    targetMutationError.value = t('pages.campaignWizard.validation.scheduleInsideRange')
    return
  }

  if (!canScheduleRow(row) || pendingTargetAction.value) {
    targetMutationError.value = scheduleDisabledReason(row)
    return
  }

  if (!(await ensureCsrfToken())) {
    await handleAuthFailure()
    return
  }

  abortTargetMutationRequest()
  const controller = new AbortController()
  targetMutationController = controller
  pendingTargetAction.value = `schedule-${key}`

  try {
    await scheduleTarget(row.id, scheduledAt, actionRequestOptions(controller))
    notifySuccess(t('pages.campaignWizard.notifications.scheduleSavedTitle'), t('pages.campaignWizard.notifications.scheduleSavedMessage', { employee: row.employeeName }))
    await loadEditTargets()
    focusedScheduleTargetId.value = key
    markWizardSnapshotClean()
  } catch (error) {
    if (error?.name === 'AbortError' || controller.signal.aborted) {
      return
    }

    if (isAuthError(error)) {
      await handleAuthFailure()
      return
    }

    targetMutationError.value = error?.message || t('pages.campaignWizard.errors.scheduleSaveFailed')
    notifyError(t('pages.campaignWizard.notifications.scheduleFailedTitle'), targetMutationError.value)
  } finally {
    if (targetMutationController === controller) {
      targetMutationController = null
    }
    pendingTargetAction.value = ''
  }
}

function openDeleteTargetDialog(row) {
  if (!canMutateTargetsInWizard.value || pendingTargetAction.value) {
    return
  }

  targetDeleteCandidate.value = row
}

function closeDeleteTargetDialog() {
  if (pendingTargetAction.value === 'delete-target') {
    return
  }

  targetDeleteCandidate.value = null
}

async function confirmDeleteTarget() {
  const target = targetDeleteCandidate.value

  if (!target || pendingTargetAction.value) {
    return
  }

  if (!(await ensureCsrfToken())) {
    await handleAuthFailure()
    return
  }

  abortTargetMutationRequest()
  const controller = new AbortController()
  targetMutationController = controller
  pendingTargetAction.value = 'delete-target'

  try {
    await deleteTarget(target.id, actionRequestOptions(controller))
    notifySuccess(t('pages.campaignWizard.notifications.targetDeletedTitle'), t('pages.campaignWizard.notifications.targetDeletedMessage', { employee: target.employeeName }))
    targetDeleteCandidate.value = null
    await loadEditTargets()
    markWizardSnapshotClean()
  } catch (error) {
    if (error?.name === 'AbortError' || controller.signal.aborted) {
      return
    }

    if (isAuthError(error)) {
      await handleAuthFailure()
      return
    }

    targetMutationError.value = error?.message || t('pages.campaignWizard.errors.targetDeleteFailed')
    notifyError(t('pages.campaignWizard.notifications.targetDeleteFailedTitle'), targetMutationError.value)
  } finally {
    if (targetMutationController === controller) {
      targetMutationController = null
    }
    pendingTargetAction.value = ''
  }
}

async function saveChangedEditSchedules(options) {
  const changedRows = editTargetRows.value.filter((row) => {
    const key = scheduleRowKey(row)
    const draft = scheduleDrafts.value[key] || ''
    return draft && draft !== toDatetimeLocalValue(row.scheduled_at)
  })

  if (!changedRows.length) {
    return []
  }

  const failures = []

  for (const row of changedRows) {
    if (!canScheduleRow(row)) {
      failures.push(`${row.employeeName}: ${scheduleDisabledReason(row)}`)
      continue
    }

    const key = scheduleRowKey(row)
    const scheduledAt = scheduleDraftToIso(key)
    if (!scheduledAt) {
      failures.push(`${row.employeeName}: ${t('pages.campaignWizard.validation.scheduleInvalid')}`)
      continue
    }

    try {
      await scheduleTarget(row.id, scheduledAt, options)
    } catch (error) {
      if (isAuthError(error)) {
        throw error
      }

      failures.push(`${row.employeeName}: ${error?.message || t('pages.campaignWizard.errors.scheduleSaveFailed')}`)
    }
  }

  return failures
}

function markWizardClean() {
  suppressDirtyGuard.value = true
}

function campaignDetailPath() {
  return campaignId.value ? `/campaigns/${campaignId.value}` : '/campaigns'
}

function requestCancelWizard() {
  if (!isWizardDirty.value) {
    markWizardClean()
    router.push(isEditMode.value ? campaignDetailPath() : '/campaigns')
    return
  }

  discardDialogOpen.value = true
}

async function discardWizardDraft() {
  markWizardClean()
  discardDialogOpen.value = false
  await router.push(isEditMode.value ? campaignDetailPath() : '/campaigns')
}

function closeDiscardDialog() {
  discardDialogOpen.value = false
}

function abortReferenceRequest() {
  if (referenceController) {
    referenceController.abort()
    referenceController = null
  }
}

function abortSubmitRequest() {
  if (submitController) {
    submitController.abort()
    submitController = null
  }
}

function abortEditCampaignRequest() {
  if (editCampaignController) {
    editCampaignController.abort()
    editCampaignController = null
  }
}

function abortEditTargetsRequest() {
  if (editTargetsController) {
    editTargetsController.abort()
    editTargetsController = null
  }
}

function abortTargetMutationRequest() {
  if (targetMutationController) {
    targetMutationController.abort()
    targetMutationController = null
  }
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

function isAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

async function handleAuthFailure() {
  const refreshedSession = await loadCurrentSession()

  if (!refreshedSession) {
    await router.push({
      name: 'login',
      query: {
        redirect: route.fullPath,
      },
    })
    return true
  }

  return false
}

async function loadAllPages(fetcher, signal) {
  const items = []
  let page = 1
  let total = 0

  do {
    const response = await fetcher({
      page,
      rows: 100,
    }, {
      signal,
    })

    const pageItems = Array.isArray(response?.items) ? response.items : []
    total = Number.isFinite(response?.total) ? response.total : items.length + pageItems.length
    items.push(...pageItems)

    if (!pageItems.length) {
      break
    }

    page += 1
  } while (items.length < total)

  return items
}

function prefillCampaignForEdit(campaign) {
  editCampaign.value = campaign
  form.value = {
    type: CAMPAIGN_TYPES.includes(campaign?.type) ? campaign.type : 'EMAIL',
    label: campaign?.label || '',
    domain: campaign?.domain || '',
    includeSite: campaign?.type === 'EMAIL' || Boolean(campaign?.landing_id || campaign?.education_id || campaign?.domain),
    dateFrom: toDateInputValue(campaign?.date_from),
    dateTo: toDateInputValue(campaign?.date_to),
    attributes: attributesToRows(campaign?.attributes),
  }
  selectedMessageId.value = nullableId(campaign?.message_id)
  selectedLandingId.value = nullableId(campaign?.landing_id)
  selectedEducationId.value = nullableId(campaign?.education_id)
  selectedMaxEducationTextId.value = nullableId(campaign?.max_education_text_id)
  selectedEmployeeIds.value = new Set()
  scheduleDrafts.value = {}
  editTargets.value = []
  editTargetsLoaded.value = false
  editTargetsError.value = ''
  formErrors.value = {}
  clearSubmitState()
  activeStepIndex.value = 0
  markWizardSnapshotClean()
}

function isCampaignEditableForUpdate(campaign) {
  return ['DRAFT', 'PAUSED'].includes(campaign?.status)
}

async function loadEditCampaign() {
  abortEditCampaignRequest()

  if (!campaignId.value) {
    editCampaign.value = null
    editNotFound.value = true
    editLoadError.value = ''
    editAuthSuppressed.value = false
    editLoading.value = false
    return
  }

  const controller = new AbortController()
  editCampaignController = controller
  editLoading.value = true
  editNotFound.value = false
  editLoadError.value = ''
  editAuthSuppressed.value = false
  editCampaign.value = null

  try {
    const campaign = await getCampaign(campaignId.value, {
      signal: controller.signal,
    })

    if (controller.signal.aborted) {
      return
    }

    if (!campaign) {
      editNotFound.value = true
      prefillCampaignForEdit(null)
      return
    }

    if (!isCampaignEditableForUpdate(campaign)) {
      editCampaign.value = campaign
      editLoadError.value = t('pages.campaignWizard.errors.editUnavailableStatus')
      return
    }

    prefillCampaignForEdit(campaign)
  } catch (error) {
    if (error?.name === 'AbortError' || controller.signal.aborted) {
      return
    }

    editCampaign.value = null

    if (isCampaignAuthError(error)) {
      editAuthSuppressed.value = true
      await loadCurrentSession()
      return
    }

    if (isCampaignNotFoundError(error)) {
      editNotFound.value = true
      return
    }

    editLoadError.value = error?.message || t('errors.campaignEditLoad')
  } finally {
    if (editCampaignController === controller) {
      editCampaignController = null
    }

    if (!controller.signal.aborted) {
      editLoading.value = false
    }
  }
}

async function loadReferenceData() {
  abortReferenceRequest()

  const controller = new AbortController()
  referenceController = controller
  referenceLoading.value = true
  referenceLoaded.value = false
  referenceErrors.value = referenceErrorState()

  const loaders = [
    ['messages', listMessages, messages],
    ['landings', listLandings, landings],
    ['attachments', listAttachments, attachments],
    ['employees', listEmployees, employees],
    ['departments', listDepartments, departments],
  ]

  const results = await Promise.all(loaders.map(async ([key, fetcher, target]) => {
    try {
      return {
        key,
        target,
        items: await loadAllPages(fetcher, controller.signal),
      }
    } catch (error) {
      return {
        key,
        target,
        error,
      }
    }
  }))

  if (controller.signal.aborted) {
    return
  }

  const nextErrors = referenceErrorState()

  for (const result of results) {
    if (!result.error) {
      result.target.value = result.items
      continue
    }

    if (isAuthError(result.error)) {
      await handleAuthFailure()
      continue
    }

    nextErrors[result.key] = result.error?.message || t('errors.referenceLoad')
  }

  referenceErrors.value = nextErrors
  referenceLoaded.value = true
  referenceLoading.value = false
  referenceController = null
}

async function loadEditTargets() {
  abortEditTargetsRequest()

  if (!isEditMode.value || !campaignId.value) {
    editTargets.value = []
    editTargetsLoaded.value = false
    editTargetsError.value = ''
    return
  }

  const controller = new AbortController()
  editTargetsController = controller
  editTargetsLoading.value = true
  editTargetsError.value = ''

  try {
    const targets = await loadAllPages((filters, options) => listVTargets({
      ...filters,
      campaignId: campaignId.value,
    }, options), controller.signal)

    if (controller.signal.aborted) {
      return
    }

    editTargets.value = targets
    scheduleDrafts.value = Object.fromEntries(targets.map((target) => [
      itemId(target),
      toDatetimeLocalValue(target.scheduled_at),
    ]))
    editTargetsLoaded.value = true
  } catch (error) {
    if (error?.name === 'AbortError' || controller.signal.aborted) {
      return
    }

    if (isAuthError(error)) {
      await handleAuthFailure()
      return
    }

    editTargetsError.value = error?.message || t('errors.targetsLoad')
  } finally {
    if (editTargetsController === controller) {
      editTargetsController = null
    }

    if (!controller.signal.aborted) {
      editTargetsLoading.value = false
    }
  }
}

async function loadInitialData() {
  if (isEditMode.value) {
    await loadEditCampaign()
    if (!editCampaign.value || editLoadError.value || editAuthSuppressed.value || editNotFound.value) {
      return
    }
  }

  await loadReferenceData()

  if (isEditMode.value && editCampaign.value) {
    await loadEditTargets()
    markWizardSnapshotClean()
  }
}

function validateSubmit(mode) {
  const errors = {}
  const label = form.value.label.trim()
  const domain = form.value.domain.trim()
  const hasDateFrom = Boolean(form.value.dateFrom)
  const hasDateTo = Boolean(form.value.dateTo)

  if (!label) {
    errors.label = t('pages.campaignWizard.validation.labelRequired')
  }

  if (maxMessageRequiresSite.value) {
    errors.message = t('pages.campaignWizard.validation.messageRequiresSite')
  }

  if (campaignUsesSite.value && !domain) {
    errors.domain = t('pages.campaignWizard.validation.domainRequired')
  } else if (domain && !isValidDomain(domain)) {
    errors.domain = t('pages.campaignWizard.validation.domainInvalid')
  }

  if (hasDateFrom !== hasDateTo) {
    const message = t('pages.campaignWizard.validation.bothDatesRequired')
    if (!hasDateFrom) {
      errors.dateFrom = message
    }
    if (!hasDateTo) {
      errors.dateTo = message
    }
  }

  if (hasDateFrom && hasDateTo && !isValidDateRange()) {
    errors.dateTo = t('pages.campaignWizard.validation.dateRangeInvalid')
  }

  if (mode === 'distribute' && (!hasDateFrom || !hasDateTo)) {
    errors.dateFrom = errors.dateFrom || t('pages.campaignWizard.validation.dateRangeRequiredForDistribution')
    errors.dateTo = errors.dateTo || t('pages.campaignWizard.validation.dateRangeRequiredForDistribution')
  }

  for (const [key, value] of Object.entries(scheduleDrafts.value)) {
    if (!value) {
      continue
    }

    if (!scheduleDraftToIso(key)) {
      errors[`schedule-${key}`] = t('pages.campaignWizard.validation.scheduleInvalid')
    } else if (!scheduleDraftInCampaignRange(value)) {
      errors[`schedule-${key}`] = t('pages.campaignWizard.validation.scheduleInsideRange')
    }
  }

  if (isEditMode.value) {
    editTargetRows.value.forEach((row) => {
      const key = scheduleRowKey(row)
      if (row.scheduled_at && !scheduleDrafts.value[key]) {
        errors[`schedule-${key}`] = t('pages.campaignWizard.validation.clearingScheduleUnsupported')
      }
    })
  }

  if (mode !== 'draft' && mode !== 'edit') {
    if (!selectedMessage.value) {
      errors.message = t('pages.campaignWizard.validation.selectMessage')
    }

    if (campaignUsesSite.value && !selectedLanding.value) {
      errors.landing = t('pages.campaignWizard.validation.selectLanding')
    }

    if (campaignUsesSite.value && !selectedEducation.value) {
      errors.education = t('pages.campaignWizard.validation.selectEducation')
    }

    if (isMaxCampaign.value && !selectedMaxEducationAttachment.value) {
      errors.maxEducationText = t('pages.campaignWizard.validation.selectMaxEducation')
    }

    if (!selectedTargetCount.value) {
      errors.targets = t('pages.campaignWizard.validation.selectEmployee')
    }
  }

  return errors
}

function buildCampaignPayload() {
  const payload = {
    type: form.value.type,
    label: form.value.label.trim(),
    domain: campaignUsesSite.value ? form.value.domain.trim() : '',
    attributes: normalizedAttributes.value,
  }

  if (selectedMessageId.value) {
    payload.message_id = selectedMessageId.value
  }

  if (campaignUsesSite.value && selectedLandingId.value) {
    payload.landing_id = selectedLandingId.value
  }

  if (campaignUsesSite.value && selectedEducationId.value) {
    payload.education_id = selectedEducationId.value
  }

  if (isMaxCampaign.value && selectedMaxEducationTextId.value) {
    payload.max_education_text_id = selectedMaxEducationTextId.value
  }

  if (form.value.dateFrom && form.value.dateTo) {
    payload.date_from = toRFC3339Date(form.value.dateFrom, 'start')
    payload.date_to = toRFC3339Date(form.value.dateTo, 'end')
  }

  return payload
}

function submitLabel(mode) {
  switch (mode) {
    case 'edit':
      return t('common.actions.saveChanges')
    case 'draft':
      return t('common.actions.saveDraft')
    case 'create':
      return t('common.actions.createCampaign')
    case 'distribute':
      return t('common.actions.createAndDistribute')
    default:
      return t('common.actions.confirm')
  }
}

function submitLoadingLabel(mode) {
  switch (mode) {
    case 'edit':
      return t('common.states.saving')
    case 'draft':
      return t('common.states.saving')
    case 'create':
      return t('common.states.creating')
    case 'distribute':
      return t('common.states.distributing')
    default:
      return t('common.states.submitting')
  }
}

async function ensureCsrfToken() {
  if (csrfToken.value) {
    return true
  }

  return Boolean(await loadCurrentSession())
}

async function submitWizard(mode) {
  if (isSubmitting.value) {
    return
  }

  clearSubmitState()
  const submitMode = isEditMode.value ? 'edit' : mode
  const errors = validateSubmit(submitMode)
  formErrors.value = errors

  if (Object.keys(errors).length) {
    activeStepIndex.value = steps.value.length - 1
    submitError.value = t('pages.campaignWizard.completeRequiredBeforeSubmit')
    return
  }

  if (!isEditMode.value && mode === 'distribute') {
    const distributedDrafts = buildLocalDistributionValues(allScheduleRows.value)
    scheduleDrafts.value = {
      ...scheduleDrafts.value,
      ...distributedDrafts,
    }
  }

  if (!(await ensureCsrfToken())) {
    await handleAuthFailure()
    return
  }

  abortSubmitRequest()
  const controller = new AbortController()
  submitController = controller
  pendingSubmit.value = submitMode
  submitStage.value = isEditMode.value ? t('pages.campaignWizard.submitStages.savingCampaign') : t('pages.campaignWizard.submitStages.creatingCampaign')
  createdCampaign.value = null

  try {
    const options = actionRequestOptions(controller)

    if (isEditMode.value) {
      const updatedCampaign = await updateCampaign(campaignId.value, buildCampaignPayload(), options)
      submitStage.value = t('pages.campaignWizard.submitStages.savingSchedules')
      const scheduleFailures = await saveChangedEditSchedules(options)
      if (scheduleFailures.length) {
        submitError.value = t('pages.campaignWizard.errors.partialScheduleSave', { count: scheduleFailures.length })
        targetMutationError.value = scheduleFailures.join(' ')
        notifyWarning(t('pages.campaignWizard.notifications.scheduleNeedsAttentionTitle'), submitError.value)
        await loadEditTargets()
        return
      }
      editCampaign.value = updatedCampaign || editCampaign.value
      await loadEditTargets()
      markWizardSnapshotClean()
      markWizardClean()
      notifySuccess(t('pages.campaignWizard.notifications.campaignUpdatedTitle'), t('pages.campaignWizard.notifications.campaignUpdatedMessage'))
      await router.push(campaignDetailPath())
      return
    }

    const campaign = await createCampaign(buildCampaignPayload(), options)
    createdCampaign.value = campaign

    if (mode === 'draft') {
      markWizardClean()
      notifySuccess(t('pages.campaignWizard.notifications.draftSavedTitle'), t('pages.campaignWizard.notifications.draftSavedMessage'))
      await router.push(`/campaigns/${campaign.id}`)
      return
    }

    submitStage.value = t('pages.campaignWizard.submitStages.creatingTargets')
    const failures = []
    const scheduleFailures = []
    const createdTargets = []

    for (const employee of selectedEmployees.value) {
      try {
        const target = await createTarget({
          employee_id: itemId(employee),
          campaign_id: itemId(campaign),
        }, options)
        createdTargets.push(target)

        const scheduledAt = scheduleDraftToIso(itemId(employee))
        if (scheduledAt) {
          try {
            await scheduleTarget(itemId(target), scheduledAt, options)
          } catch (error) {
            if (isAuthError(error)) {
              throw error
            }
            scheduleFailures.push({
              employee,
              message: error?.message || t('pages.campaignWizard.errors.scheduleSaveFailed'),
            })
          }
        }
      } catch (error) {
        if (isAuthError(error)) {
          throw error
        }

        failures.push({
          employee,
          message: error?.message || t('pages.campaignWizard.errors.targetCreateFailed'),
        })
      }
    }

    if (failures.length || scheduleFailures.length) {
      partialFailure.value = {
        campaign,
        createdTargetCount: createdTargets.length,
        failures,
        distributionError: scheduleFailures.length
          ? t('pages.campaignWizard.errors.scheduleFailuresCount', { count: scheduleFailures.length })
          : '',
      }
      submitError.value = scheduleFailures.length
        ? t('pages.campaignWizard.errors.createdSchedulePartial')
        : t('pages.campaignWizard.errors.createdTargetsPartial')
      notifyWarning(t('pages.campaignWizard.notifications.campaignNeedsAttentionTitle'), submitError.value)
      activeStepIndex.value = steps.value.length - 1
      return
    }

    markWizardClean()
    if (mode === 'distribute') {
      notifySuccess(t('pages.campaignWizard.notifications.targetsDistributedTitle'), t('pages.campaignWizard.notifications.campaignTargetsDistributedMessage'))
    } else {
      notifySuccess(t('pages.campaignWizard.notifications.campaignCreatedTitle'), t('pages.campaignWizard.notifications.campaignCreatedMessage'))
    }
    await router.push(`/campaigns/${campaign.id}`)
  } catch (error) {
    if (error?.name === 'AbortError') {
      return
    }

    if (isAuthError(error)) {
      await handleAuthFailure()
      submitError.value = 'Session expired. Sign in again to continue.'
      notifyError('Session expired', 'Sign in again to continue.')
      return
    }

    submitError.value = error?.message || `${submitLabel(submitMode)} failed.`
    notifyError(`${submitLabel(submitMode)} failed`, submitError.value)
  } finally {
    if (submitController === controller) {
      submitController = null
    }

    pendingSubmit.value = ''
    submitStage.value = ''
  }
}

function handleBeforeUnload(event) {
  if (!isWizardDirty.value) {
    return
  }

  event.preventDefault()
  event.returnValue = ''
}

watch([employeeSearch, departmentFilter], () => {
  targetPage.value = 1
})

watch(() => form.value.type, () => {
  const selectedType = MESSAGE_TYPES.includes(selectedMessage.value?.type) ? selectedMessage.value.type : 'EMAIL'
  if (selectedMessage.value && selectedType !== form.value.type) {
    selectedMessageId.value = ''
  }

  if (isEmailCampaign.value) {
    form.value.includeSite = true
    selectedMaxEducationTextId.value = ''
  } else if (!selectedLandingId.value && !form.value.domain.trim()) {
    form.value.includeSite = false
  }

  if (!campaignUsesSite.value) {
    form.value.domain = ''
    selectedLandingId.value = ''
    selectedEducationId.value = ''
  }

  clearSubmitState()
})

watch(() => form.value.includeSite, (enabled) => {
  if (isEmailCampaign.value && !enabled) {
    form.value.includeSite = true
    return
  }

  if (!enabled) {
    form.value.domain = ''
    selectedLandingId.value = ''
    selectedEducationId.value = ''
    delete formErrors.value.domain
    delete formErrors.value.landing
    delete formErrors.value.education
  }

  clearSubmitState()
})

watch(() => steps.value.length, () => {
  if (activeStepIndex.value > steps.value.length - 1) {
    activeStepIndex.value = Math.max(0, steps.value.length - 1)
  }
})

watch(() => filteredEmployees.value.length, () => {
  if (targetPage.value > targetTotalPages.value) {
    targetPage.value = targetTotalPages.value
  }
})

onBeforeRouteLeave(() => {
  if (!isWizardDirty.value) {
    return true
  }

  const title = isEditMode.value ? t('pages.campaignWizard.discard.editTitle') : t('pages.campaignWizard.discard.createTitle')
  return window.confirm(`${title}\n\n${t('pages.campaignWizard.discard.message')}`)
})

onMounted(() => {
  loadInitialData()
  window.addEventListener('beforeunload', handleBeforeUnload)
})

onBeforeUnmount(() => {
  abortEditCampaignRequest()
  abortEditTargetsRequest()
  abortTargetMutationRequest()
  abortReferenceRequest()
  abortSubmitRequest()
  window.removeEventListener('beforeunload', handleBeforeUnload)
})
</script>

<template>
  <div class="campaign-wizard-page">
    <PageHeader
      :eyebrow="t('sections.campaigns')"
      :title="pageTitle"
      :description="pageDescription"
    >
      <template #actions>
        <UiButton
          v-if="isEditMode"
          :loading="pendingSubmit === 'edit'"
          :disabled="isSubmitting || Boolean(pendingTargetAction) || !canShowWizard"
          :title="isSubmitting ? t('common.states.submitting') : ''"
          @click="submitWizard('edit')"
        >
          {{ pendingSubmit === 'edit' ? submitLoadingLabel('edit') : t('common.actions.saveChanges') }}
        </UiButton>
        <UiButton
          v-if="!isEditMode"
          variant="secondary"
          :loading="pendingSubmit === 'draft'"
          :disabled="isSubmitting"
          :title="isSubmitting ? t('common.states.submitting') : ''"
          @click="submitWizard('draft')"
        >
          {{ pendingSubmit === 'draft' ? submitLoadingLabel('draft') : t('common.actions.saveDraft') }}
        </UiButton>
        <UiButton
          v-if="!isEditMode"
          :loading="pendingSubmit === 'create'"
          :disabled="isSubmitting"
          :title="isSubmitting ? t('common.states.submitting') : ''"
          @click="submitWizard('create')"
        >
          {{ pendingSubmit === 'create' ? submitLoadingLabel('create') : t('common.actions.createCampaign') }}
        </UiButton>
      </template>
    </PageHeader>

    <section
      v-if="isEditMode && editLoading"
      class="campaign-wizard-state"
      :aria-label="t('pages.campaignWizard.editLoadingState')"
    >
      <SkeletonBlock :rows="5" />
    </section>

    <EmptyState
      v-else-if="isEditMode && editNotFound"
      :title="t('pages.campaignDetail.notFoundTitle')"
      :description="t('pages.campaignDetail.notFoundDescription')"
    >
      <RouterLink class="ui-button ui-button--secondary" to="/campaigns">
        {{ t('pages.campaignDetail.backToCampaigns') }}
      </RouterLink>
    </EmptyState>

    <section
      v-else-if="isEditMode && editAuthSuppressed"
      class="campaign-wizard-state"
      :aria-label="t('pages.campaignWizard.editSessionState')"
    >
      <SkeletonBlock :rows="4" />
    </section>

    <UiAlert
      v-else-if="isEditMode && editLoadError"
      variant="error"
      :title="t('pages.campaignDetail.loadTitle')"
      :message="editLoadError"
    >
      <UiButton variant="secondary" @click="loadInitialData">
        {{ t('pages.campaignDetail.retry') }}
      </UiButton>
    </UiAlert>

    <UiAlert
      v-if="canShowWizard && submitError && !partialFailure"
      variant="error"
      :title="isEditMode ? t('pages.campaignWizard.errors.updateFailedTitle') : t('pages.campaignWizard.errors.createFailedTitle')"
      :message="submitError"
      dismissible
      @dismiss="clearSubmitState"
    />

    <UiAlert
      v-if="canShowWizard && isSubmitting"
      variant="info"
      :title="submitLoadingLabel(pendingSubmit)"
      :message="submitStage || t('common.states.submitting')"
    />

    <WizardStepper
      v-if="canShowWizard"
      :active-index="activeStepIndex"
      :steps="steps"
      :aria-label="t('pages.campaignWizard.stepsAria')"
      @step-click="goToStep"
    />

    <section v-if="canShowWizard" class="campaign-wizard-layout" :aria-label="isEditMode ? t('pages.campaignWizard.editTitle') : t('pages.campaignWizard.createTitle')">
      <main class="campaign-wizard-main" :aria-labelledby="`campaign-wizard-${activeStepIndex}`">
        <header v-if="showPanelHeader" class="campaign-wizard-panel-header">
          <div>
            <h2 :id="`campaign-wizard-${activeStepIndex}`">{{ activeStep }}</h2>
            <p>
              {{ isEditMode
                ? t('pages.campaignWizard.panelDescriptionEdit')
                : t('pages.campaignWizard.panelDescriptionCreate') }}
            </p>
          </div>
        </header>

        <template v-if="activeStepKey === 'basic'">
          <section class="campaign-wizard-form-grid" :aria-label="t('pages.campaignWizard.steps.basicInfo')">
            <label class="ui-field">
              <span class="ui-field-label">{{ t('common.labels.type') }}</span>
              <select v-model="form.type" class="ui-select">
                <option
                  v-for="option in campaignTypeOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
            </label>
            <label
              class="campaign-wizard-site-toggle"
              :class="{ 'is-disabled': isEmailCampaign }"
            >
              <input
                v-model="form.includeSite"
                type="checkbox"
                :disabled="isEmailCampaign"
              />
              <span class="campaign-wizard-site-toggle-copy">
                <strong>{{ t('pages.campaignWizard.siteOptionLabel') }}</strong>
                <small>
                  {{ isEmailCampaign
                    ? t('pages.campaignWizard.siteOptionEmailRequired')
                    : t('pages.campaignWizard.siteOptionHelper') }}
                </small>
              </span>
            </label>
            <UiInput
              id="campaign-label"
              v-model="form.label"
              :label="t('campaign.labels.label')"
              placeholder="Quarterly awareness simulation"
              :error="formErrors.label"
            />
            <UiInput
              v-if="campaignUsesSite"
              id="campaign-domain"
              v-model="form.domain"
              :label="t('common.labels.domain')"
              placeholder="https://training.example.com"
              :helper="t('pages.campaignWizard.domainHelper')"
              :error="formErrors.domain"
            />
            <UiDatePicker
              id="campaign-date-from"
              v-model="form.dateFrom"
              :label="t('common.labels.dateFrom')"
              :error="formErrors.dateFrom"
            />
            <UiDatePicker
              id="campaign-date-to"
              v-model="form.dateTo"
              :label="t('common.labels.dateTo')"
              :error="formErrors.dateTo"
            />

            <section class="campaign-wizard-attributes campaign-wizard-field--wide" :aria-label="t('common.labels.attributes')">
              <div class="campaign-wizard-summary-title">
                <h3>{{ t('common.labels.attributes') }}</h3>
                <UiButton variant="secondary" @click="addAttribute">
                  {{ t('common.actions.addAttribute') }}
                </UiButton>
              </div>
              <div
                v-for="attribute in form.attributes"
                :key="attribute.id"
                class="campaign-wizard-attribute-row"
              >
                <label class="ui-field">
                  <span class="ui-field-label">{{ t('common.labels.key') }}</span>
                  <input v-model="attribute.key" class="ui-input" placeholder="region" />
                </label>
                <label class="ui-field">
                  <span class="ui-field-label">{{ t('common.labels.value') }}</span>
                  <input v-model="attribute.value" class="ui-input" placeholder="internal" />
                </label>
                <UiButton variant="ghost" @click="removeAttribute(attribute.id)">
                  {{ t('common.actions.delete') }}
                </UiButton>
              </div>
            </section>
          </section>
        </template>

        <template v-else-if="activeStepKey === 'message'">
          <section class="campaign-wizard-picker-step" :aria-label="t('pages.campaignWizard.messageSelectionAria')">
            <CampaignResourcePickerPanel
              v-model="selectedMessageId"
              resource-type="message"
              :message-type="form.type"
              :title="isMaxCampaign ? t('pages.emailTemplates.tabs.max') : t('nav.emailTemplates')"
              :selected-title="t('pages.campaignWizard.selectedMessage')"
              :selected-empty-text="t('pages.campaignWizard.selectMessage')"
              :empty-title="t('pages.campaignWizard.noMessageTemplates')"
              :empty-description="t('pages.campaignWizard.noMessageTemplatesDescription')"
              :search-placeholder="t('pages.campaignWizard.messageSearchPlaceholder')"
              :items="filteredMessages"
              :loading="referenceLoading && !referenceLoaded"
              :error="referenceErrors.messages"
              :create-to="{ name: 'email-template-new', query: { type: form.type } }"
              :create-label="t('common.actions.createMessage')"
              @selection-change="handleMessageSelection"
              @preview="previewMessageResource"
              @refresh="loadReferenceData"
            />
            <UiAlert
              v-if="maxMessageRequiresSite"
              variant="warning"
              :title="t('pages.campaignWizard.siteRequiredByMessageTitle')"
              :message="t('pages.campaignWizard.siteRequiredByMessageMessage')"
            />
            <div v-if="formErrors.message" class="campaign-wizard-inline-error">
              {{ formErrors.message }}
            </div>
          </section>
        </template>

        <template v-else-if="activeStepKey === 'landing'">
          <section class="campaign-wizard-picker-step" :aria-label="t('pages.campaignWizard.landingSelectionAria')">
            <CampaignResourcePickerPanel
              v-model="selectedLandingId"
              resource-type="landing"
              :title="t('nav.landingPages')"
              :selected-title="t('pages.campaignWizard.selectedLanding')"
              :selected-empty-text="t('pages.campaignWizard.selectLanding')"
              :empty-title="t('pages.campaignWizard.noLandingPages')"
              :empty-description="t('pages.campaignWizard.noLandingPagesDescription')"
              :search-placeholder="t('pages.campaignWizard.landingSearchPlaceholder')"
              :items="landings"
              :loading="referenceLoading && !referenceLoaded"
              :error="referenceErrors.landings"
              :create-to="{ name: 'landing-page-new' }"
              :create-label="t('common.actions.createLanding')"
              @selection-change="handleLandingSelection"
              @preview="previewLandingResource"
              @refresh="loadReferenceData"
            />
            <div v-if="formErrors.landing" class="campaign-wizard-inline-error">
              {{ formErrors.landing }}
            </div>
          </section>
        </template>

        <template v-else-if="activeStepKey === 'education'">
          <section class="campaign-wizard-picker-step" :aria-label="t('pages.campaignWizard.educationSelectionAria')">
            <CampaignResourcePickerPanel
              v-if="isEmailCampaign || campaignUsesSite"
              v-model="selectedEducationId"
              resource-type="education"
              :title="isMaxCampaign ? t('pages.campaignWizard.siteEducationTemplates') : t('pages.campaignWizard.educationTemplates')"
              :selected-title="isMaxCampaign ? t('pages.campaignWizard.selectedSiteEducation') : t('pages.campaignWizard.selectedEducation')"
              :selected-empty-text="isMaxCampaign ? t('pages.campaignWizard.selectSiteEducation') : t('pages.campaignWizard.selectEducation')"
              :empty-title="t('pages.campaignWizard.noEducationAssets')"
              :search-placeholder="t('pages.campaignWizard.educationSearchPlaceholder')"
              :items="landings"
              :loading="referenceLoading && !referenceLoaded"
              :error="referenceErrors.landings"
              :create-to="{ name: 'landing-page-new' }"
              :create-label="t('common.actions.createLanding')"
              @selection-change="handleEducationSelection"
              @preview="previewLandingResource($event, t('common.labels.educationAsset'))"
              @refresh="loadReferenceData"
            />
            <div v-if="formErrors.education" class="campaign-wizard-inline-error">
              {{ formErrors.education }}
            </div>
            <AttachmentPickerPanel
              v-if="isMaxCampaign"
              v-model="selectedMaxEducationTextId"
              selection-mode="single"
              :title="t('pages.campaignWizard.maxEducationText')"
              :empty-text="t('pages.campaignWizard.noMaxEducationTexts')"
              :allowed-types="['.txt']"
              default-type=".txt"
              show-preview-action
              show-download-action
              show-upload-action
              @preview="openResourcePreview({
                title: t('pages.campaignWizard.maxEducationText'),
                kind: t('common.labels.educationAsset'),
                attachmentId: $event?.id,
                attachmentLabel: $event?.label,
                attachmentType: $event?.type || '.txt',
                meta: [[t('common.labels.name'), $event?.label || t('common.placeholder')]],
              })"
              @selection-change="handleMaxEducationSelection"
            />
            <div v-if="isMaxCampaign && formErrors.maxEducationText" class="campaign-wizard-inline-error">
              {{ formErrors.maxEducationText }}
            </div>
          </section>
        </template>

        <template v-else-if="activeStepKey === 'targets'">
          <section class="campaign-wizard-targets" :aria-label="t('pages.campaignWizard.targetSelectionAria')">
            <div class="campaign-wizard-target-controls">
              <UiInput
                id="employee-search"
                v-model="employeeSearch"
                :label="t('pages.campaignWizard.employeeSearch')"
                :placeholder="t('pages.campaignWizard.employeeSearchPlaceholder')"
              />
              <label class="ui-field">
                <span class="ui-field-label">{{ t('common.labels.department') }}</span>
                <select v-model="departmentFilter" class="ui-input">
                  <option value="">{{ t('common.filters.allDepartments') }}</option>
                  <option
                    v-for="department in departments"
                    :key="itemId(department)"
                    :value="itemId(department)"
                  >
                    {{ department.label }}
                  </option>
                </select>
              </label>
              <div class="campaign-wizard-target-actions">
                <UiButton
                  variant="secondary"
                  :disabled="!filteredEmployees.length"
                  @click="selectVisibleEmployees"
                >
                  {{ t('common.actions.selectVisible') }}
                </UiButton>
                <UiButton
                  variant="ghost"
                  :disabled="!selectedTargetCount"
                  @click="clearSelectedEmployees"
                >
                  {{ t('common.actions.clearSelection') }}
                </UiButton>
                <UiBadge :variant="selectedTargetCount ? 'primary' : 'neutral'">
                  {{ t('pages.campaignWizard.selectedCount', { count: selectedTargetCount }) }}
                </UiBadge>
                <UiButton
                  v-if="isEditMode"
                  :loading="pendingTargetAction === 'add-targets'"
                  :disabled="!selectedTargetCount || Boolean(pendingTargetAction)"
                  :title="targetMutationDisabledReason"
                  @click="addSelectedTargetsInEdit"
                >
                  <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
                  {{ t('common.actions.addSelected') }}
                </UiButton>
              </div>
            </div>

            <UiAlert
              v-if="targetMutationError"
              variant="error"
              :title="t('pages.campaignWizard.targetOperationFailed')"
              :message="targetMutationError"
              dismissible
              @dismiss="targetMutationError = ''"
            />

            <UiAlert
              v-if="isEditMode && targetMutationDisabledReason"
              variant="warning"
              :title="t('pages.campaignWizard.targetChangesDisabled')"
              :message="targetMutationDisabledReason"
            />

            <section
              v-if="isEditMode"
              class="campaign-wizard-current-targets"
              aria-labelledby="campaign-wizard-current-targets-title"
            >
              <div class="campaign-wizard-summary-title">
                <div>
                  <h3 id="campaign-wizard-current-targets-title">{{ t('pages.campaignWizard.currentTargetsTitle') }}</h3>
                  <p class="campaign-wizard-muted">{{ t('pages.campaignWizard.currentTargetsDescription') }}</p>
                </div>
                <UiBadge :variant="editTargetRows.length ? 'primary' : 'neutral'">
                  {{ t('pages.campaignWizard.targetCount', { count: editTargetRows.length }) }}
                </UiBadge>
              </div>

              <SkeletonBlock v-if="editTargetsLoading && !editTargetsLoaded" :rows="4" />
              <div v-else-if="editTargetsError" class="campaign-wizard-inline-error">
                {{ editTargetsError }}
              </div>
              <EmptyState
                v-else-if="editTargetsLoaded && !editTargetRows.length"
                :title="t('pages.campaignWizard.noCampaignTargets')"
                :description="t('pages.campaignWizard.noCampaignTargetsDescription')"
              />
              <div v-else class="campaign-wizard-target-table campaign-wizard-target-table--edit" role="table" :aria-label="t('pages.campaignWizard.currentTargetsTitle')">
                <div class="campaign-wizard-target-row campaign-wizard-target-row--head" role="row">
                  <span role="columnheader">{{ t('common.labels.employee') }}</span>
                  <span role="columnheader">{{ t('common.labels.status') }}</span>
                  <span role="columnheader">{{ t('common.labels.scheduledAt') }}</span>
                  <span role="columnheader">{{ t('common.labels.createdAt') }}</span>
                  <span role="columnheader">{{ t('common.labels.token') }}</span>
                  <span role="columnheader">{{ t('pages.campaigns.columns.actions') }}</span>
                </div>
                <div
                  v-for="target in editTargetRows"
                  :key="target.id"
                  class="campaign-wizard-target-row"
                  role="row"
                >
                  <span role="cell">{{ target.employeeName }}</span>
                  <span role="cell">
                    <UiBadge :variant="targetStatusVariant(target.status)">
                      {{ displayTargetStatus(target.status) }}
                    </UiBadge>
                  </span>
                  <span role="cell">{{ formatFullDateTime(target.scheduled_at) }}</span>
                  <span role="cell">{{ formatFullDateTime(target.created_at) }}</span>
                  <span role="cell"><code>{{ formatTechnicalId(target.token) }}</code></span>
                  <span role="cell">
                    <UiButton
                      variant="danger"
                      :disabled="Boolean(pendingTargetAction) || !canMutateTargetsInWizard"
                      :title="targetMutationDisabledReason"
                      @click="openDeleteTargetDialog(target)"
                    >
                      <Trash2 :size="16" stroke-width="1.8" aria-hidden="true" />
                      {{ t('common.actions.delete') }}
                    </UiButton>
                  </span>
                </div>
              </div>
            </section>

            <SkeletonBlock v-if="referenceLoading && !referenceLoaded" :rows="5" />
            <div v-else-if="referenceErrors.employees" class="campaign-wizard-inline-error">
              {{ referenceErrors.employees }}
            </div>
            <EmptyState
              v-else-if="!employees.length"
                :title="t('pages.campaignWizard.noEmployees')"
              :description="t('pages.campaignWizard.noEmployeesDescription')"
            />
            <EmptyState
              v-else-if="!filteredEmployees.length"
                :title="t('pages.campaignWizard.noMatchingEmployees')"
              :description="t('pages.campaignWizard.noMatchingEmployeesDescription')"
            />
            <div v-else class="campaign-wizard-table" role="table" :aria-label="t('pages.campaignWizard.employeeTableAria')">
              <div class="campaign-wizard-table-row campaign-wizard-table-head" role="row">
                <span role="columnheader">{{ t('common.actions.select') }}</span>
                <span role="columnheader">{{ t('common.labels.fullName') }}</span>
                <span role="columnheader">{{ t('common.labels.email') }}</span>
                <span role="columnheader">{{ t('common.labels.phone') }}</span>
                <span role="columnheader">{{ t('common.labels.department') }}</span>
              </div>
              <div
                v-for="employee in pagedEmployees"
                :key="itemId(employee)"
                class="campaign-wizard-table-row"
                role="row"
              >
                <span role="cell">
                  <label class="campaign-wizard-checkbox">
                    <input
                      type="checkbox"
                      :checked="isEmployeeSelected(employee)"
                      :aria-label="`Select ${employeeName(employee)}`"
                      @change="toggleEmployee(employee)"
                    />
                  </label>
                </span>
                <span role="cell">{{ employeeName(employee) }}</span>
                <span role="cell">{{ employee.email }}</span>
                <span role="cell">{{ formatNullable(employee.phone) }}</span>
                <span role="cell">{{ departmentLabel(employee.department_id) }}</span>
              </div>
            </div>
            <ResourcePagination
              v-if="referenceLoaded && employees.length && filteredEmployees.length"
              :page="targetPage"
              :rows="targetRows"
              :total="filteredEmployees.length"
              :loaded-count="pagedEmployees.length"
              :loading="referenceLoading && !referenceLoaded"
              @update:page="setTargetPage"
              @update:rows="setTargetRows"
            />

            <div class="campaign-wizard-target-summary">
              <div class="campaign-wizard-note">
                {{ isEditMode
                  ? t('pages.campaignWizard.targetsEditNote')
                  : t('pages.campaignWizard.targetsCreateNote') }}
              </div>
            </div>
            <div v-if="formErrors.targets" class="campaign-wizard-inline-error">
              {{ formErrors.targets }}
            </div>
          </section>
        </template>

        <template v-else>
          <section class="campaign-wizard-review" :aria-label="t('pages.campaignWizard.steps.review')">
            <section class="campaign-wizard-schedule-panel" aria-labelledby="campaign-wizard-schedule-title">
              <header class="campaign-wizard-schedule-header">
                <div>
                  <h3 id="campaign-wizard-schedule-title">{{ t('pages.campaignWizard.scheduleTitle') }}</h3>
                  <p class="campaign-wizard-muted">
                    {{ isEditMode
                      ? t('pages.campaignWizard.scheduleEditDescription')
                      : t('pages.campaignWizard.scheduleCreateDescription') }}
                  </p>
                </div>
                <div class="campaign-wizard-schedule-actions">
                  <UiBadge :variant="scheduledScheduleCount ? 'primary' : 'neutral'">
                    {{ t('pages.campaignWizard.scheduledCount', { count: scheduledScheduleCount }) }}
                  </UiBadge>
                  <UiBadge :variant="unscheduledScheduleCount ? 'warning' : 'neutral'">
                    {{ t('pages.campaignWizard.unscheduledCount', { count: unscheduledScheduleCount }) }}
                  </UiBadge>
                  <UiButton
                    :loading="pendingTargetAction === 'distribute'"
                    :disabled="!canAutoDistributeSchedule || Boolean(pendingTargetAction)"
                    :title="autoDistributeDisabledReason"
                    @click="runWizardAutoDistribute"
                  >
                    <Shuffle :size="16" stroke-width="1.8" aria-hidden="true" />
                    {{ pendingTargetAction === 'distribute' ? t('common.states.distributing') : t('common.actions.autoDistribute') }}
                  </UiButton>
                </div>
              </header>

              <UiAlert
                v-if="targetMutationError"
                variant="error"
                :title="t('pages.campaignWizard.scheduleOperationFailed')"
                :message="targetMutationError"
                dismissible
                @dismiss="targetMutationError = ''"
              />

              <div v-if="isEditMode && editTargetsLoading && !editTargetsLoaded" class="campaign-wizard-inline-state">
                <SkeletonBlock :rows="4" />
              </div>
              <EmptyState
                v-else-if="!allScheduleRows.length"
                :title="t('pages.campaignWizard.noTargetsToSchedule')"
                :description="isEditMode ? t('pages.campaignWizard.noTargetsToScheduleEdit') : t('pages.campaignWizard.noTargetsToScheduleCreate')"
              />
              <template v-else>
                <CampaignScheduleTimelineChart
                  :targets="scheduleTimelineRows"
                  :empty-text="t('charts.schedule.empty')"
                  @select-target="focusScheduleTarget"
                />

                <div class="campaign-wizard-schedule-table" role="table" :aria-label="t('pages.campaignWizard.scheduleEditorAria')">
                  <div class="campaign-wizard-schedule-row campaign-wizard-schedule-row--head" role="row">
                    <span role="columnheader">{{ t('common.labels.employee') }}</span>
                    <span role="columnheader">{{ t('common.labels.status') }}</span>
                    <span role="columnheader">{{ t('common.labels.currentSchedule') }}</span>
                    <span role="columnheader">{{ t('pages.campaignWizard.newScheduledAt') }}</span>
                    <span role="columnheader">{{ t('pages.campaigns.columns.actions') }}</span>
                  </div>
                  <div
                    v-for="target in allScheduleRows"
                    :key="scheduleRowKey(target)"
                    class="campaign-wizard-schedule-row"
                    :class="{ 'is-focused': focusedScheduleTargetId === scheduleRowKey(target) }"
                    role="row"
                  >
                    <span role="cell">
                      <strong>{{ target.employeeName }}</strong>
                    </span>
                    <span role="cell">
                      <UiBadge :variant="targetStatusVariant(target.status)">
                        {{ displayTargetStatus(target.status) }}
                      </UiBadge>
                    </span>
                    <span role="cell">{{ formatFullDateTime(target.scheduled_at) }}</span>
                    <span role="cell">
                      <div class="campaign-wizard-schedule-input">
                        <UiDatePicker
                          :id="`schedule-date-${scheduleRowKey(target)}`"
                          :label="t('pages.campaignWizard.scheduleDate')"
                          :model-value="getScheduleDraftDate(scheduleRowKey(target))"
                          :placeholder="t('common.date.placeholder')"
                          :min="form.dateFrom"
                          :max="form.dateTo"
                          :disabled="!canScheduleRow(target) || Boolean(pendingTargetAction)"
                          @update:model-value="updateScheduleDraftDate(scheduleRowKey(target), $event)"
                        />
                        <UiTimePicker
                          :id="`schedule-time-${scheduleRowKey(target)}`"
                          :label="t('pages.campaignWizard.scheduleTime')"
                          :model-value="getScheduleDraftTime(scheduleRowKey(target))"
                          :placeholder="t('common.time.placeholder')"
                          :disabled="!canScheduleRow(target) || Boolean(pendingTargetAction)"
                          @update:model-value="updateScheduleDraftTime(scheduleRowKey(target), $event)"
                        />
                        <span
                          v-if="formErrors[`schedule-${scheduleRowKey(target)}`]"
                          class="ui-field-error campaign-wizard-schedule-error"
                        >
                          {{ formErrors[`schedule-${scheduleRowKey(target)}`] }}
                        </span>
                      </div>
                    </span>
                    <span role="cell">
                      <span class="campaign-wizard-row-actions">
                        <UiButton
                          v-if="isEditMode"
                          variant="secondary"
                          :loading="pendingTargetAction === `schedule-${scheduleRowKey(target)}`"
                          :disabled="!canScheduleRow(target) || Boolean(pendingTargetAction)"
                          :title="scheduleDisabledReason(target)"
                          @click="saveEditTargetSchedule(target)"
                        >
                          <CalendarClock :size="16" stroke-width="1.8" aria-hidden="true" />
                          {{ t('common.actions.save') }}
                        </UiButton>
                        <UiButton
                          v-else
                          variant="ghost"
                          :disabled="!scheduleDrafts[scheduleRowKey(target)] || Boolean(pendingTargetAction)"
                          @click="clearScheduleDraft(target)"
                        >
                          {{ t('common.actions.clear') }}
                        </UiButton>
                      </span>
                    </span>
                  </div>
                </div>
              </template>
            </section>

            <div class="campaign-wizard-review-grid">
              <article class="campaign-wizard-review-card">
                <h3>{{ t('pages.campaignWizard.steps.basicInfo') }}</h3>
                <dl class="campaign-wizard-description-list">
                  <div><dt>{{ t('common.labels.type') }}</dt><dd>{{ isMaxCampaign ? t('pages.emailTemplates.tabs.max') : t('pages.emailTemplates.tabs.email') }}</dd></div>
                  <div><dt>{{ t('campaign.labels.label') }}</dt><dd>{{ form.label || t('common.placeholder') }}</dd></div>
                  <div><dt>{{ t('pages.campaignWizard.siteOptionLabel') }}</dt><dd>{{ campaignUsesSite ? t('pages.campaignWizard.siteEnabled') : t('pages.campaignWizard.siteDisabled') }}</dd></div>
                  <div v-if="campaignUsesSite"><dt>{{ t('common.labels.domain') }}</dt><dd>{{ form.domain || t('common.placeholder') }}</dd></div>
                </dl>
              </article>

              <article class="campaign-wizard-review-card">
                <h3>{{ t('common.labels.message') }}</h3>
                <dl class="campaign-wizard-description-list">
                  <div><dt>{{ t('common.labels.name') }}</dt><dd>{{ selectedMessage?.label || t('common.placeholder') }}</dd></div>
                  <template v-if="isMaxCampaign">
                    <div><dt>{{ t('pages.emailTemplates.maxAccount') }}</dt><dd>{{ formatTechnicalId(selectedMessage?.max_account_id) }}</dd></div>
                    <div><dt>{{ t('common.labels.textBody') }}</dt><dd>{{ selectedMessageTextBodyAttachment?.label || formatTechnicalId(selectedMessage?.text_body_id) }}</dd></div>
                  </template>
                  <template v-else>
                    <div><dt>{{ t('common.labels.subject') }}</dt><dd>{{ formatNullable(selectedMessage?.subject) }}</dd></div>
                    <div><dt>{{ t('common.labels.from') }}</dt><dd>{{ selectedMessage?.from_email || t('common.placeholder') }}</dd></div>
                  </template>
                </dl>
              </article>

              <article v-if="campaignUsesSite" class="campaign-wizard-review-card">
                <h3>{{ t('common.labels.landingPage') }}</h3>
                <dl class="campaign-wizard-description-list">
                  <div><dt>{{ t('common.labels.name') }}</dt><dd>{{ selectedLanding?.label || t('common.placeholder') }}</dd></div>
                  <div><dt>{{ t('common.labels.htmlBody') }}</dt><dd>{{ selectedLanding?.html_body_id ? t('common.resource.ready') : t('common.resource.missing') }}</dd></div>
                </dl>
              </article>

              <article class="campaign-wizard-review-card">
                <h3>{{ t('common.labels.educationAsset') }}</h3>
                <dl class="campaign-wizard-description-list">
                  <template v-if="isMaxCampaign">
                    <div><dt>{{ t('common.labels.textBody') }}</dt><dd>{{ selectedMaxEducationAttachment?.label || t('common.placeholder') }}</dd></div>
                    <div><dt>{{ t('common.labels.id') }}</dt><dd>{{ formatTechnicalId(selectedMaxEducationTextId) }}</dd></div>
                    <template v-if="campaignUsesSite">
                      <div><dt>{{ t('pages.campaignWizard.siteEducationTemplates') }}</dt><dd>{{ selectedEducation?.label || t('common.placeholder') }}</dd></div>
                      <div><dt>{{ t('common.labels.htmlBody') }}</dt><dd>{{ selectedEducation?.html_body_id ? t('common.resource.ready') : t('common.resource.missing') }}</dd></div>
                    </template>
                  </template>
                  <template v-else>
                    <div><dt>{{ t('common.labels.name') }}</dt><dd>{{ selectedEducation?.label || t('common.placeholder') }}</dd></div>
                    <div><dt>{{ t('common.labels.htmlBody') }}</dt><dd>{{ selectedEducation?.html_body_id ? t('common.resource.ready') : t('common.resource.missing') }}</dd></div>
                    <div><dt>{{ t('common.labels.attachments') }}</dt><dd>{{ selectedEducationAttachment?.label || t('common.placeholder') }}</dd></div>
                  </template>
                </dl>
              </article>

              <article class="campaign-wizard-review-card">
                <h3>{{ t('common.labels.targets') }}</h3>
                <dl class="campaign-wizard-description-list">
                  <div><dt>{{ isEditMode ? t('common.labels.currentCount') : t('common.labels.selectedCount') }}</dt><dd>{{ isEditMode ? editTargetRows.length : selectedTargetCount }}</dd></div>
                  <div><dt>{{ t('common.labels.scheduled') }}</dt><dd>{{ scheduledScheduleCount }}</dd></div>
                </dl>
              </article>

              <article class="campaign-wizard-review-card">
                <h3>{{ t('common.labels.schedule') }}</h3>
                <dl class="campaign-wizard-description-list">
                  <div><dt>{{ t('common.labels.dateRange') }}</dt><dd>{{ formatReviewDateRange() }}</dd></div>
                </dl>
                <p class="campaign-wizard-muted">
                  {{ isEditMode
                    ? t('pages.campaignWizard.autoDistributeEditNote')
                    : t('pages.campaignWizard.autoDistributeCreateNote') }}
                </p>
              </article>

              <article class="campaign-wizard-review-card campaign-wizard-review-card--wide">
                <h3>{{ t('common.labels.attributes') }}</h3>
                <div v-if="Object.keys(normalizedAttributes).length" class="campaign-wizard-key-values">
                  <div
                    v-for="(value, key) in normalizedAttributes"
                    :key="key"
                    class="campaign-wizard-field"
                  >
                    <span>{{ key }}</span>
                    <strong>{{ value || '—' }}</strong>
                  </div>
                </div>
                <p v-else class="campaign-wizard-muted">{{ t('common.attributes.noneAdded') }}</p>
              </article>

              <UiAlert
                v-if="reviewWarnings.length"
                class="campaign-wizard-review-card--wide"
                variant="warning"
                :title="t('pages.campaignWizard.validationWarnings')"
                :items="reviewWarnings"
              />
              <UiAlert
                v-else
                class="campaign-wizard-review-card--wide"
                variant="success"
                :title="t('pages.campaignWizard.noValidationWarnings')"
                :message="isEditMode ? t('pages.campaignWizard.readyToSave') : t('pages.campaignWizard.readyForFinalAction')"
              />
            </div>

            <div class="campaign-wizard-final-action">
              <div>
                <span>{{ isEditMode ? t('pages.campaignWizard.finalAction') : t('pages.campaignWizard.finalActions') }}</span>
                <p class="campaign-wizard-muted">
                  {{ isEditMode
                    ? t('pages.campaignWizard.finalActionEditDescription')
                    : t('pages.campaignWizard.finalActionCreateDescription') }}
                </p>
              </div>
            </div>

            <UiAlert
              v-if="Object.keys(formErrors).length"
              variant="error"
              :title="t('errors.resolveValidation')"
              :message="t('pages.campaignWizard.completeRequiredBeforeSubmit')"
            />

            <UiAlert
              v-if="!isEditMode && partialFailure"
              variant="warning"
              :title="t('pages.campaignWizard.notifications.campaignNeedsAttentionTitle')"
              :message="t('pages.campaignWizard.partialFailureMessage', { created: partialFailure.createdTargetCount, failed: partialFailure.failures.length, distributionError: partialFailure.distributionError })"
            >
              <RouterLink
                class="campaign-wizard-inline-link"
                :to="`/campaigns/${partialFailure.campaign.id}`"
                @click="markWizardClean"
              >
                {{ t('common.actions.openCreatedCampaign') }}
              </RouterLink>
            </UiAlert>
          </section>
        </template>
      </main>
    </section>

    <footer v-if="canShowWizard" class="campaign-wizard-footer" :aria-label="t('pages.campaignWizard.navigationAria')">
      <UiButton
        variant="ghost"
        :disabled="isSubmitting"
        :title="isSubmitting ? t('common.states.submitting') : ''"
        @click="requestCancelWizard"
      >
          {{ t('common.actions.cancel') }}
      </UiButton>
      <UiButton
        v-if="!isEditMode"
        variant="secondary"
        :loading="pendingSubmit === 'draft'"
        :disabled="isSubmitting"
        :title="isSubmitting ? t('common.states.submitting') : ''"
        @click="submitWizard('draft')"
      >
        {{ pendingSubmit === 'draft' ? submitLoadingLabel('draft') : t('common.actions.saveDraft') }}
      </UiButton>
      <UiButton variant="secondary" :disabled="isFirstStep || isSubmitting" @click="goBack">
        {{ t('common.actions.back') }}
      </UiButton>
      <UiButton v-if="!isFinalStep" :disabled="isSubmitting" @click="goNext">
        {{ t('common.actions.next') }}
      </UiButton>
      <template v-else>
        <UiButton
          v-if="!isEditMode"
          variant="secondary"
          :loading="pendingSubmit === 'create'"
          :disabled="isSubmitting"
          :title="isSubmitting ? t('common.states.submitting') : ''"
          @click="submitWizard('create')"
        >
          {{ pendingSubmit === 'create' ? submitLoadingLabel('create') : t('common.actions.createCampaign') }}
        </UiButton>
        <UiButton
          v-if="!isEditMode"
          :loading="pendingSubmit === 'distribute'"
          :disabled="isSubmitting"
          :title="isSubmitting ? t('common.states.submitting') : ''"
          @click="submitWizard('distribute')"
        >
          {{ pendingSubmit === 'distribute' ? submitLoadingLabel('distribute') : t('common.actions.createAndDistribute') }}
        </UiButton>
        <UiButton
          v-else
          :loading="pendingSubmit === 'edit'"
          :disabled="isSubmitting || Boolean(pendingTargetAction)"
          :title="isSubmitting ? t('common.states.submitting') : ''"
          @click="submitWizard('edit')"
        >
          {{ pendingSubmit === 'edit' ? submitLoadingLabel('edit') : t('common.actions.saveChanges') }}
        </UiButton>
      </template>
    </footer>

    <PreviewDrawer
      :open="resourcePreviewOpen"
      :title="resourcePreviewTitle"
      :subtitle="t('common.preview.rawTemplate')"
      @close="closeResourcePreview"
    >
      <section class="campaign-resource-preview-meta">
        <div>
          <span>{{ t('common.labels.resourceType') }}</span>
          <strong>{{ resourcePreviewKind || t('common.placeholder') }}</strong>
        </div>
        <div v-for="[label, value] in resourcePreviewMeta" :key="label">
          <span>{{ label }}</span>
          <strong>{{ value }}</strong>
        </div>
      </section>
      <HtmlAttachmentPreview
        :attachment="resourcePreviewAttachment"
        :attachment-id="resourcePreviewAttachment?.id || ''"
        :title="t('common.preview.campaignResource')"
        :empty-text="t('common.preview.selectResource')"
      />
    </PreviewDrawer>

    <ConfirmDialog
      :open="discardDialogOpen"
      :title="isEditMode ? t('pages.campaignWizard.discard.editTitle') : t('pages.campaignWizard.discard.createTitle')"
      :message="t('pages.campaignWizard.discard.message')"
      :confirm-label="t('common.actions.discard')"
      :cancel-label="t('common.actions.continueEditing')"
      variant="danger"
      @cancel="closeDiscardDialog"
      @confirm="discardWizardDraft"
    />

    <ConfirmDialog
      :open="Boolean(targetDeleteCandidate)"
      :title="t('pages.campaignWizard.deleteTarget.title')"
      :message="t('pages.campaignWizard.deleteTarget.message')"
      :confirm-label="t('pages.campaignWizard.deleteTarget.confirm')"
      :loading-label="t('common.states.deleting')"
      :cancel-label="t('common.actions.keepTarget')"
      variant="danger"
      :loading="pendingTargetAction === 'delete-target'"
      @cancel="closeDeleteTargetDialog"
      @confirm="confirmDeleteTarget"
    />
  </div>
</template>
