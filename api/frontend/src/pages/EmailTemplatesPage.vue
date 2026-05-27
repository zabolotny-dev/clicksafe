<script setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  Copy,
  Edit3,
  Eye,
  Mail,
  MessageCircle,
  Plus,
  Trash2,
  Volume2,
} from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import AttachmentPickerPanel from '../components/ui/AttachmentPickerPanel.vue'
import AttachmentPreviewDrawer from '../components/ui/AttachmentPreviewDrawer.vue'
import ConfirmDialog from '../components/ui/ConfirmDialog.vue'
import HtmlLiveEditor from '../components/ui/HtmlLiveEditor.vue'
import IconButton from '../components/ui/IconButton.vue'
import MaxAccountPickerPanel from '../components/ui/MaxAccountPickerPanel.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import ResourcePagination from '../components/ui/ResourcePagination.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiInput from '../components/ui/UiInput.vue'
import VoiceCloneWorkspace from '../components/ui/VoiceCloneWorkspace.vue'
import WizardStepper from '../components/ui/WizardStepper.vue'
import { useNotifications } from '../composables/useNotifications'
import { useResourceActions } from '../composables/useResourceActions'
import {
  getAttachmentContent,
  listAttachments,
  updateAttachmentContent,
  uploadAttachment,
} from '../resources/attachments'
import {
  createMessage,
  deleteMessage,
  getMessage,
  listMessages,
  MESSAGE_TYPES,
  updateMessage,
} from '../resources/messages'
import { listMaxAccounts } from '../resources/maxAccounts'
import { getVoiceStatus } from '../resources/voice'
import {
  isHtmlAttachmentType,
  isTextAttachmentType,
  normalizeAttachmentType,
  SUPPORTED_ATTACHMENT_TYPES,
} from '../utils/attachmentTypes'
import {
  errorMessage,
  formatTechnicalId,
  nullableId,
} from '../utils/resourceFormat'

const route = useRoute()
const router = useRouter()
const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()
const { t } = useI18n()

const steps = computed(() => [
  t('pages.emailTemplates.steps.basicInfo'),
  form.type === 'MAX' ? t('common.labels.textBody') : t('common.labels.htmlBody'),
  t('common.labels.attachments'),
  t('pages.emailTemplates.steps.review'),
])

const messages = ref([])
const attachments = ref([])
const maxAccounts = ref([])
const activeMessageType = ref(MESSAGE_TYPES.includes(route.query.type) ? route.query.type : 'EMAIL')
const page = ref(1)
const rows = ref(10)
const total = ref(0)
const isLoading = ref(false)
const isSaving = ref(false)
const isUploadingHtml = ref(false)
const isDeleting = ref(false)
const loadError = ref('')
const formError = ref('')
const activeStep = ref(0)
const deleteDialogOpen = ref(false)
const deleteTarget = ref(null)
const previewAttachment = ref(null)
const htmlEditorOpen = ref(false)
const voiceCloneOpen = ref(false)
const htmlEditorMode = ref('create')
const htmlEditorContentType = ref('html')
const htmlEditorInitial = ref('')
const htmlDraftFilename = ref('')
const selectedHtmlBodyAttachment = ref(null)
const selectedMaxAccount = ref(null)
const selectedTemplateAttachments = ref([])
const emailHtmlBodyPicker = ref(null)
const maxTextBodyPicker = ref(null)
const templateAttachmentsPicker = ref(null)
const contentEditAttachment = ref(null)
const maxAttachmentTypes = SUPPORTED_ATTACHMENT_TYPES.filter((type) => type !== '.html')
const voiceStatus = reactive({
  checked: false,
  available: false,
  message: '',
  status: '',
  model_provider: '',
  model_id: '',
  device: '',
})
const isVoiceStatusLoading = ref(false)

const filters = reactive({
  label: '',
  fromEmail: '',
  fromName: '',
  subject: '',
})

const form = reactive({
  type: 'EMAIL',
  label: '',
  from_email: '',
  from_name: '',
  subject: '',
  html_body_id: '',
  text_body_id: '',
  max_account_id: '',
  attachment_ids: [],
})

const isEditorRoute = computed(() => ['email-template-new', 'email-template-edit'].includes(route.name))
const isEditMode = computed(() => route.name === 'email-template-edit')
const editingId = computed(() => String(route.params.id || ''))
const pageTitle = computed(() => (isEditMode.value ? t('pages.emailTemplates.editTitle') : t('pages.emailTemplates.createTitle')))
const pageDescription = computed(() => (
  isEditMode.value
    ? t('pages.emailTemplates.editDescription')
    : t('pages.emailTemplates.createDescription')
))
const attachmentById = computed(() => attachments.value.reduce((acc, attachment) => {
  acc[attachment.id] = attachment
  return acc
}, {}))
const maxAccountById = computed(() => maxAccounts.value.reduce((acc, account) => {
  acc[account.id] = account
  return acc
}, {}))
const previewOpen = computed(() => Boolean(previewAttachment.value))
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / rows.value)))
const isMaxForm = computed(() => form.type === 'MAX')
const isUpdatingAttachmentContent = computed(() => htmlEditorMode.value === 'update')
const htmlWorkspaceIsHtml = computed(() => htmlEditorContentType.value === 'html')
const htmlWorkspaceTitle = computed(() => {
  if (isUpdatingAttachmentContent.value) {
    return htmlWorkspaceIsHtml.value ? t('common.labels.htmlBody') : t('common.labels.textBody')
  }

  return t('pages.emailTemplates.htmlBodyTitle')
})
const htmlWorkspaceDescription = computed(() => (
  htmlWorkspaceIsHtml.value
    ? t('common.preview.rawTemplate')
    : t('common.preview.text')
))
const htmlWorkspaceSaveLabel = computed(() => (
  isUpdatingAttachmentContent.value
    ? t('common.actions.saveContent')
    : t('pages.emailTemplates.saveHtmlAttachment')
))
const htmlWorkspaceFilenamePlaceholder = computed(() => (
  htmlWorkspaceIsHtml.value ? 'message-body.html' : 'message-body.txt'
))
const selectedMaxAccountDisplay = computed(() => {
  const account = selectedMaxAccount.value || maxAccountById.value[form.max_account_id]
  if (account) {
    return [account.label, account.phone].filter(Boolean).join(' · ') || formatTechnicalId(account.id)
  }

  return form.max_account_id ? formatTechnicalId(form.max_account_id) : t('common.placeholder')
})
const maxUnsupportedAttachments = computed(() => {
  if (!isMaxForm.value) {
    return []
  }

  return selectedTemplateAttachments.value.filter((attachment) => (
    String(attachment?.type || '').toLowerCase() === '.html'
  ))
})
const hasMaxUnsupportedAttachments = computed(() => maxUnsupportedAttachments.value.length > 0)
const voiceUnavailableMessage = computed(() => (
  isVoiceStatusLoading.value
    ? t('pages.emailTemplates.voiceClone.status.checking')
    : voiceStatus.message || t('pages.emailTemplates.voiceClone.status.unavailable')
))
const canOpenVoiceClone = computed(() => isMaxForm.value && voiceStatus.available && !isVoiceStatusLoading.value)
const messageTypeTabs = computed(() => [
  { value: 'EMAIL', label: t('pages.emailTemplates.tabs.email'), icon: Mail },
  { value: 'MAX', label: t('pages.emailTemplates.tabs.max'), icon: MessageCircle },
])
const messageTableColumnCount = computed(() => (activeMessageType.value === 'MAX' ? 4 : 6))

function clearForm() {
  form.type = MESSAGE_TYPES.includes(route.query.type) ? route.query.type : activeMessageType.value
  form.label = ''
  form.from_email = ''
  form.from_name = ''
  form.subject = ''
  form.html_body_id = ''
  form.text_body_id = ''
  form.max_account_id = ''
  form.attachment_ids = []
  formError.value = ''
  activeStep.value = 0
  htmlEditorOpen.value = false
  voiceCloneOpen.value = false
  htmlEditorMode.value = 'create'
  htmlEditorContentType.value = 'html'
  htmlEditorInitial.value = ''
  htmlDraftFilename.value = ''
  contentEditAttachment.value = null
  selectedHtmlBodyAttachment.value = null
  selectedMaxAccount.value = null
  selectedTemplateAttachments.value = []
}

function fillForm(message) {
  form.type = MESSAGE_TYPES.includes(message?.type) ? message.type : 'EMAIL'
  form.label = message?.label || ''
  form.from_email = message?.from_email || ''
  form.from_name = message?.from_name || ''
  form.subject = message?.subject || ''
  form.html_body_id = nullableId(message?.html_body_id)
  form.text_body_id = nullableId(message?.text_body_id)
  form.max_account_id = nullableId(message?.max_account_id)
  form.attachment_ids = Array.isArray(message?.attachment_ids)
    ? message.attachment_ids.map((id) => nullableId(id) || String(id)).filter(Boolean)
    : []
  formError.value = ''
  activeStep.value = 0
  htmlEditorOpen.value = false
  voiceCloneOpen.value = false
  htmlEditorMode.value = 'create'
  htmlEditorContentType.value = 'html'
  contentEditAttachment.value = null
  selectedHtmlBodyAttachment.value = null
  selectedTemplateAttachments.value = []
  selectedMaxAccount.value = maxAccountById.value[form.max_account_id] || null
}

function fillDuplicateForm(message) {
  fillForm(message)
  form.label = t('pages.emailTemplates.copyLabel', { label: message?.label || t('common.labels.message') })
}

function buildPayload() {
  return {
    type: form.type,
    label: form.label.trim(),
    from_email: form.type === 'EMAIL' ? form.from_email.trim() : '',
    from_name: form.type === 'EMAIL' ? form.from_name.trim() : '',
    subject: form.type === 'EMAIL' ? form.subject.trim() : '',
    html_body_id: form.type === 'EMAIL' ? nullableId(form.html_body_id) : '',
    text_body_id: form.type === 'MAX' ? nullableId(form.text_body_id) : '',
    max_account_id: form.type === 'MAX' ? nullableId(form.max_account_id) : '',
    attachment_ids: form.attachment_ids,
  }
}

function validateForm() {
  if (!form.label.trim()) {
    return t('pages.emailTemplates.validation.labelRequired')
  }

  if (form.type === 'EMAIL' && !form.from_email.trim()) {
    return t('pages.emailTemplates.validation.fromEmailRequired')
  }

  if (form.type === 'MAX' && !form.max_account_id) {
    return t('pages.emailTemplates.validation.maxAccountRequired')
  }

  if (form.type === 'MAX' && !form.text_body_id) {
    return t('pages.emailTemplates.validation.textBodyRequired')
  }

  if (form.type === 'MAX' && hasMaxUnsupportedAttachments.value) {
    return t('pages.emailTemplates.validation.maxHtmlAttachmentUnsupported')
  }

  return ''
}

function htmlFileName() {
  const base = (htmlDraftFilename.value || form.label || 'message-body')
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9-]+/g, '-')
    .replace(/^-|-$/g, '') || 'message-body'

  return base.endsWith('.html') ? base : `${base}.html`
}

function openPreview(attachment) {
  if (!attachment) {
    notifyError(t('common.preview.unavailable'), t('common.preview.selectResource'))
    return
  }

  previewAttachment.value = attachment
}

function closePreview() {
  previewAttachment.value = null
}

function editableAttachmentContentType(attachment) {
  const type = normalizeAttachmentType(attachment?.type)

  if (isHtmlAttachmentType(type)) {
    return 'html'
  }

  if (isTextAttachmentType(type)) {
    return 'txt'
  }

  return ''
}

function attachmentEditorFilename(attachment) {
  const type = editableAttachmentContentType(attachment)
  const label = attachment?.label || 'message-body'

  if (type === 'html' && !label.toLowerCase().endsWith('.html')) {
    return `${label}.html`
  }

  if (type === 'txt' && !label.toLowerCase().endsWith('.txt')) {
    return `${label}.txt`
  }

  return label
}

async function openContentEditor(attachment) {
  const contentType = editableAttachmentContentType(attachment)

  if (!attachment?.id || !contentType) {
    notifyError(t('common.preview.unavailable'), t('common.preview.unsupported'))
    return
  }

  formError.value = ''

  try {
    const content = await getAttachmentContent(attachment.id)
    htmlEditorMode.value = 'update'
    htmlEditorContentType.value = contentType
    htmlEditorInitial.value = typeof content === 'string' ? content : ''
    htmlDraftFilename.value = attachmentEditorFilename(attachment)
    contentEditAttachment.value = attachment
    htmlEditorOpen.value = true
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    notifyError(t('common.preview.failed'), errorMessage(error, 'errors.attachmentsLoad'))
  }
}

async function loadVisibleHtmlBodyAttachments(items) {
  const ids = [...new Set(items.flatMap((message) => [
    nullableId(message?.html_body_id),
    nullableId(message?.text_body_id),
  ]).filter(Boolean))]

  if (!ids.length) {
    attachments.value = []
    return
  }

  const responses = await Promise.all(ids.map((id) => listAttachments({ id, rows: 1 })))
  attachments.value = responses
    .map((response) => (Array.isArray(response?.items) ? response.items[0] : null))
    .filter((attachment) => attachment?.id)
}

async function loadData(options = {}) {
  isLoading.value = true
  loadError.value = ''

  try {
    const messageResponse = await listMessages({
      ...filters,
      type: activeMessageType.value,
      page: page.value,
      rows: rows.value,
    })

    messages.value = Array.isArray(messageResponse?.items) ? messageResponse.items : []
    total.value = Number.isFinite(messageResponse?.total) ? messageResponse.total : messages.value.length
    page.value = Number.isFinite(messageResponse?.page) ? messageResponse.page : page.value
    rows.value = Number.isFinite(messageResponse?.rowsPerPage) ? messageResponse.rowsPerPage : rows.value

    if (options.backfillEmptyPage && total.value > 0 && !messages.value.length && page.value > 1) {
      page.value -= 1
      await loadData()
      return
    }

    if (isEditorRoute.value) {
      attachments.value = []
    } else {
      await loadVisibleHtmlBodyAttachments(messages.value)
    }

    if (activeMessageType.value === 'MAX' || isEditorRoute.value) {
      await loadMaxAccounts()
    }

    await prepareEditorFromRoute()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    loadError.value = errorMessage(error, 'errors.messagesLoad')
  } finally {
    isLoading.value = false
  }
}

async function loadMaxAccounts() {
  const response = await listMaxAccounts({ rows: 100 })
  maxAccounts.value = Array.isArray(response?.items) ? response.items : []
  selectedMaxAccount.value = maxAccountById.value[form.max_account_id] || selectedMaxAccount.value
}

function voiceStatusMessage(message) {
  const text = String(message || '').trim()
  if (!text || text === 'voice service is unavailable' || text === 'voice service is not configured') {
    return t('pages.emailTemplates.voiceClone.status.unavailable')
  }

  return text
}

async function loadVoiceStatus() {
  if (isVoiceStatusLoading.value) {
    return
  }

  isVoiceStatusLoading.value = true

  try {
    const response = await getVoiceStatus()
    voiceStatus.checked = true
    voiceStatus.available = Boolean(response?.available)
    voiceStatus.message = response?.available ? '' : voiceStatusMessage(response?.message)
    voiceStatus.status = response?.status || ''
    voiceStatus.model_provider = response?.model_provider || ''
    voiceStatus.model_id = response?.model_id || ''
    voiceStatus.device = response?.device || ''
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    voiceStatus.checked = true
    voiceStatus.available = false
    voiceStatus.message = voiceStatusMessage(errorMessage(error, 'pages.emailTemplates.voiceClone.status.unavailable'))
  } finally {
    isVoiceStatusLoading.value = false
  }
}

async function prepareEditorFromRoute() {
  if (!isEditorRoute.value) {
    return
  }

  clearForm()

  if (isEditMode.value && editingId.value) {
    const message = messages.value.find((item) => item.id === editingId.value) || await getMessage(editingId.value)
    fillForm(message)
    return
  }

  const duplicateId = String(route.query.duplicate || '')
  if (duplicateId) {
    const source = messages.value.find((item) => item.id === duplicateId) || await getMessage(duplicateId)
    fillDuplicateForm(source)
  }
}

function openBlankHtmlEditor() {
  formError.value = ''
  htmlEditorMode.value = 'create'
  htmlEditorContentType.value = 'html'
  htmlEditorInitial.value = ''
  htmlDraftFilename.value = form.label || 'message-body'
  contentEditAttachment.value = null
  htmlEditorOpen.value = true
}

function closeHtmlEditor() {
  htmlEditorOpen.value = false
  htmlEditorMode.value = 'create'
  htmlEditorContentType.value = 'html'
  htmlEditorInitial.value = ''
  htmlDraftFilename.value = ''
  contentEditAttachment.value = null
}

function openVoiceClone() {
  if (!canOpenVoiceClone.value) {
    return
  }

  formError.value = ''
  voiceCloneOpen.value = true
}

function closeVoiceClone() {
  voiceCloneOpen.value = false
}

function trackHtmlBodySelection(attachment) {
  selectedHtmlBodyAttachment.value = attachment
}

function trackMaxAccountSelection(account) {
  selectedMaxAccount.value = account
}

function trackTemplateAttachmentsSelection(selected) {
  selectedTemplateAttachments.value = Array.isArray(selected) ? selected : []
}

async function handleVoiceCloneSaved(uploaded) {
  const id = nullableId(uploaded?.id)
  if (!id) {
    return
  }

  if (!form.attachment_ids.includes(id)) {
    form.attachment_ids = [...form.attachment_ids, id]
  }

  const byId = new Map(selectedTemplateAttachments.value.map((attachment) => [nullableId(attachment?.id), attachment]))
  byId.set(id, uploaded)
  selectedTemplateAttachments.value = form.attachment_ids
    .map((attachmentID) => byId.get(nullableId(attachmentID)))
    .filter(Boolean)

  voiceCloneOpen.value = false
  await refreshAttachmentPickers()
}

async function refreshAttachmentPickers() {
  await nextTick()
  await Promise.all([
    emailHtmlBodyPicker.value?.refresh?.(),
    maxTextBodyPicker.value?.refresh?.(),
    templateAttachmentsPicker.value?.refresh?.(),
  ].filter(Boolean))
}

function syncAttachmentContentSaved(saved) {
  if (saved?.id) {
    attachments.value = attachments.value.map((attachment) => (
      nullableId(attachment.id) === nullableId(saved.id) ? saved : attachment
    ))

    if (nullableId(selectedHtmlBodyAttachment.value?.id) === nullableId(saved.id)) {
      selectedHtmlBodyAttachment.value = saved
    }

    selectedTemplateAttachments.value = selectedTemplateAttachments.value.map((attachment) => (
      nullableId(attachment?.id) === nullableId(saved.id) ? saved : attachment
    ))

    contentEditAttachment.value = saved
  }
}

function htmlCopyFileName(attachment) {
  const label = attachment?.label || form.label || 'message-body'
  const cleanLabel = label.replace(/\.html$/i, '')
  return `${cleanLabel}-copy.html`
}

async function openSelectedHtmlEditor() {
  if (!selectedHtmlBodyAttachment.value?.id) {
    notifyError(t('pages.emailTemplates.notifications.editorUnavailableTitle'), t('pages.emailTemplates.notifications.editorUnavailableMessage'))
    return
  }

  formError.value = ''

  try {
    const content = await getAttachmentContent(selectedHtmlBodyAttachment.value.id)
    htmlEditorMode.value = 'create'
    htmlEditorContentType.value = 'html'
    htmlEditorInitial.value = typeof content === 'string' ? content : ''
    htmlDraftFilename.value = htmlCopyFileName(selectedHtmlBodyAttachment.value)
    contentEditAttachment.value = null
    htmlEditorOpen.value = true
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    notifyError(t('pages.emailTemplates.notifications.htmlSourceUnavailableTitle'), errorMessage(error, 'errors.attachmentsLoad'))
  }
}

async function saveHtmlSource(html) {
  if (!html.trim()) {
    formError.value = t('pages.emailTemplates.validation.htmlEmpty')
    return
  }

  if (!(await ensureCsrfToken())) {
    return
  }

  isUploadingHtml.value = true
  formError.value = ''

  try {
    const file = new File([html], htmlFileName(), { type: 'text/html' })
    const uploaded = await uploadAttachment(file, { public: false }, mutationOptions())
    form.html_body_id = uploaded.id
    selectedHtmlBodyAttachment.value = uploaded
    closeHtmlEditor()
    await refreshAttachmentPickers()
    notifySuccess(t('pages.emailTemplates.notifications.htmlSavedTitle'), t('pages.emailTemplates.notifications.htmlSavedMessage'))
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.attachmentsLoad')
    notifyError(t('pages.emailTemplates.notifications.htmlUploadFailedTitle'), formError.value)
  } finally {
    isUploadingHtml.value = false
  }
}

async function saveExistingAttachmentContent(content) {
  if (!String(content || '').trim()) {
    formError.value = t('common.resource.contentEmpty')
    return
  }

  if (!contentEditAttachment.value?.id || !(await ensureCsrfToken())) {
    return
  }

  isUploadingHtml.value = true
  formError.value = ''

  try {
    const saved = await updateAttachmentContent(contentEditAttachment.value.id, content, mutationOptions())
    syncAttachmentContentSaved(saved)
    notifySuccess(t('pages.attachments.notifications.contentUpdated'))
    closeHtmlEditor()
    await refreshAttachmentPickers()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.attachmentsLoad')
    notifyError(t('pages.attachments.notifications.contentUpdateFailedTitle'), formError.value)
  } finally {
    isUploadingHtml.value = false
  }
}

function saveWorkspaceSource(content) {
  if (isUpdatingAttachmentContent.value) {
    return saveExistingAttachmentContent(content)
  }

  return saveHtmlSource(content)
}

async function submitForm() {
  const validation = validateForm()
  if (validation) {
    formError.value = validation
    if (form.type === 'MAX' && !form.max_account_id) {
      activeStep.value = 0
    } else if (form.type === 'MAX' && !form.text_body_id) {
      activeStep.value = 1
    } else if (form.type === 'MAX' && hasMaxUnsupportedAttachments.value) {
      activeStep.value = 2
    } else {
      activeStep.value = 0
    }
    return
  }

  if (!(await ensureCsrfToken())) {
    return
  }

  isSaving.value = true
  formError.value = ''

  try {
    if (isEditMode.value && editingId.value) {
      await updateMessage(editingId.value, buildPayload(), mutationOptions())
      notifySuccess(t('pages.emailTemplates.notifications.updated'))
    } else {
      await createMessage(buildPayload(), mutationOptions())
      notifySuccess(t('pages.emailTemplates.notifications.created'))
    }

    await router.push({ name: 'email-templates' })
    await loadData()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.messagesLoad')
    notifyError(t('pages.emailTemplates.notifications.saveFailedTitle'), formError.value)
  } finally {
    isSaving.value = false
  }
}

function openDeleteDialog(message) {
  deleteTarget.value = message
  deleteDialogOpen.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value || !(await ensureCsrfToken())) {
    return
  }

  isDeleting.value = true

  try {
    await deleteMessage(deleteTarget.value.id, mutationOptions())
    notifySuccess(t('pages.emailTemplates.notifications.deleted'))
    deleteDialogOpen.value = false
    deleteTarget.value = null
    await loadData({ backfillEmptyPage: true })
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    notifyError(t('pages.emailTemplates.notifications.deleteFailedTitle'), errorMessage(error, 'errors.messagesLoad'))
  } finally {
    isDeleting.value = false
  }
}

function goToDuplicate(message) {
  router.push({
    name: 'email-template-new',
    query: {
      duplicate: message.id,
      type: message.type || activeMessageType.value,
    },
  })
}

function setActiveMessageType(type) {
  if (!MESSAGE_TYPES.includes(type) || activeMessageType.value === type) {
    return
  }
  activeMessageType.value = type
  page.value = 1
  loadData()
}

function applyFilters() {
  page.value = 1
  loadData()
}

function setPage(nextPage) {
  page.value = Math.min(Math.max(1, nextPage), totalPages.value)
  loadData()
}

function setRows(nextRows) {
  rows.value = nextRows
  page.value = 1
  loadData()
}

watch(() => route.fullPath, () => {
  loadData()
})

watch(() => [isEditorRoute.value, isMaxForm.value], ([editor, max]) => {
  if (editor && max && !voiceStatus.checked) {
    loadVoiceStatus()
  }
})

watch(activeStep, (step) => {
  if (step === 2 && isEditorRoute.value && isMaxForm.value) {
    loadVoiceStatus()
  }
})

onMounted(() => {
  loadData()
})
</script>

<template>
  <section class="resource-page">
    <template v-if="!isEditorRoute">
      <PageHeader
        :eyebrow="t('sections.templates')"
        :title="t('pages.emailTemplates.title')"
        :description="t('pages.emailTemplates.description')"
      >
        <template #actions>
          <RouterLink class="ui-button ui-button--primary" :to="{ name: 'email-template-new', query: { type: activeMessageType } }">
            <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('common.actions.createMessage') }}
          </RouterLink>
        </template>
      </PageHeader>

      <div class="resource-type-tabs" role="tablist" :aria-label="t('pages.emailTemplates.typeTabsAria')">
        <button
          v-for="tab in messageTypeTabs"
          :key="tab.value"
          class="resource-type-tab"
          :class="{ 'is-active': activeMessageType === tab.value }"
          type="button"
          role="tab"
          :aria-selected="activeMessageType === tab.value"
          @click="setActiveMessageType(tab.value)"
        >
          <component :is="tab.icon" :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ tab.label }}
        </button>
      </div>

      <section class="resource-filter-panel" :class="{ 'resource-filter-panel--compact': activeMessageType === 'MAX' }">
        <UiInput v-model="filters.label" :label="t('common.labels.name')" :placeholder="t('common.resource.searchByLabel')" />
        <UiInput v-if="activeMessageType === 'EMAIL'" v-model="filters.fromEmail" :label="t('common.labels.fromEmail')" placeholder="sender@example.com" />
        <UiInput v-if="activeMessageType === 'EMAIL'" v-model="filters.fromName" :label="t('common.labels.fromName')" :placeholder="t('common.labels.fromName')" />
        <UiInput v-if="activeMessageType === 'EMAIL'" v-model="filters.subject" :label="t('common.labels.subject')" :placeholder="t('common.labels.subject')" />
        <UiButton variant="secondary" :loading="isLoading" @click="applyFilters">
          {{ t('common.filters.apply') }}
        </UiButton>
      </section>

      <UiAlert
        v-if="loadError"
        variant="error"
        :title="t('pages.emailTemplates.unavailable')"
        :message="loadError"
      >
        <UiButton variant="secondary" @click="loadData">{{ t('common.actions.retry') }}</UiButton>
      </UiAlert>

      <section class="resource-panel resource-panel--table">
        <header class="resource-panel-header">
          <div>
            <h2>{{ t('pages.emailTemplates.tableTitle') }}</h2>
            <p>{{ t('common.loadedCount', { count: messages.length }) }}</p>
          </div>
        </header>

        <div
          class="resource-table resource-message-table"
          :class="activeMessageType === 'MAX' ? 'resource-message-table--max' : 'resource-message-table--email'"
        >
          <div class="resource-table-row resource-table-head">
            <span>{{ t('common.labels.name') }}</span>
            <span>{{ activeMessageType === 'MAX' ? t('pages.emailTemplates.maxAccount') : t('common.labels.fromEmail') }}</span>
            <span v-if="activeMessageType === 'EMAIL'">{{ t('common.labels.subject') }}</span>
            <span v-if="activeMessageType === 'EMAIL'">{{ t('common.labels.htmlBody') }}</span>
            <span>{{ t('common.labels.attachments') }}</span>
            <span>{{ t('pages.campaigns.columns.actions') }}</span>
          </div>
          <template v-if="isLoading">
            <div v-for="index in 4" :key="index" class="resource-table-row">
              <span v-for="cell in messageTableColumnCount" :key="cell"><SkeletonBlock :rows="1" /></span>
            </div>
          </template>
          <template v-else-if="messages.length">
            <div
              v-for="message in messages"
              :key="message.id"
              class="resource-table-row"
            >
              <span><strong>{{ message.label }}</strong></span>
              <span>{{ activeMessageType === 'MAX' ? (maxAccountById[nullableId(message.max_account_id)]?.label || formatTechnicalId(message.max_account_id)) : (message.from_email || t('common.placeholder')) }}</span>
              <span v-if="activeMessageType === 'EMAIL'">{{ message.subject || t('common.placeholder') }}</span>
              <span v-if="activeMessageType === 'EMAIL'"><code>{{ formatTechnicalId(message.html_body_id) }}</code></span>
              <span>{{ Array.isArray(message.attachment_ids) ? message.attachment_ids.length : 0 }}</span>
              <span class="resource-row-actions">
                <IconButton
                  :label="activeMessageType === 'MAX' ? t('pages.emailTemplates.previewTextBody') : t('pages.emailTemplates.previewHtmlBody')"
                  :disabled="!attachmentById[nullableId(activeMessageType === 'MAX' ? message.text_body_id : message.html_body_id)]"
                  @click="openPreview(attachmentById[nullableId(activeMessageType === 'MAX' ? message.text_body_id : message.html_body_id)])"
                >
                  <Eye :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
                <RouterLink
                  class="icon-button icon-button--ghost"
                  :to="{ name: 'email-template-edit', params: { id: message.id } }"
                  :aria-label="t('common.actions.edit')"
                  :title="t('common.actions.edit')"
                >
                  <Edit3 :size="16" stroke-width="1.8" aria-hidden="true" />
                </RouterLink>
                <IconButton :label="t('common.actions.duplicate')" @click="goToDuplicate(message)">
                  <Copy :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
                <IconButton :label="t('common.actions.delete')" variant="danger" @click="openDeleteDialog(message)">
                  <Trash2 :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
              </span>
            </div>
          </template>
          <div v-else class="resource-empty-row">
            {{ t('common.empty.noMatchingRecords') }}
          </div>
        </div>
        <ResourcePagination
          :page="page"
          :rows="rows"
          :total="total"
          :loaded-count="messages.length"
          :loading="isLoading"
          @update:page="setPage"
          @update:rows="setRows"
        />
      </section>
    </template>

    <template v-else>
      <PageHeader
        v-if="!htmlEditorOpen && !voiceCloneOpen"
        :eyebrow="t('sections.templates')"
        :title="pageTitle"
        :description="pageDescription"
      >
        <template #actions>
          <RouterLink class="ui-button ui-button--secondary" :to="{ name: 'email-templates' }">
            {{ t('pages.emailTemplates.backToList') }}
          </RouterLink>
        </template>
      </PageHeader>

      <UiAlert
        v-if="formError"
        variant="error"
        :title="t('pages.emailTemplates.formAttention')"
        :message="formError"
        dismissible
        @dismiss="formError = ''"
      />

      <section v-if="htmlEditorOpen" class="resource-html-workspace">
        <HtmlLiveEditor
          v-model:filename="htmlDraftFilename"
          workspace
          :initial-content="htmlEditorInitial"
          :content-type="htmlEditorContentType"
          :title="htmlWorkspaceTitle"
          :description="htmlWorkspaceDescription"
          :filename-placeholder="htmlWorkspaceFilenamePlaceholder"
          :filename-readonly="isUpdatingAttachmentContent"
          :save-label="htmlWorkspaceSaveLabel"
          :saving="isUploadingHtml"
          :show-preview="htmlWorkspaceIsHtml"
          :show-image-action="htmlWorkspaceIsHtml"
          @save="saveWorkspaceSource"
          @cancel="closeHtmlEditor"
        />
      </section>

      <section v-else-if="voiceCloneOpen" class="resource-html-workspace">
        <VoiceCloneWorkspace
          :status="voiceStatus"
          :status-loading="isVoiceStatusLoading"
          :initial-label="form.label"
          @cancel="closeVoiceClone"
          @refresh-status="loadVoiceStatus"
          @saved="handleVoiceCloneSaved"
        />
      </section>

      <section v-else class="resource-editor-shell">
        <WizardStepper
          v-model:active-index="activeStep"
          :steps="steps"
          :aria-label="t('pages.emailTemplates.stepsAria')"
        />

        <section v-if="activeStep === 0" class="resource-editor-panel">
          <header class="resource-panel-header resource-panel-header--plain">
            <div>
              <h2>{{ t('pages.emailTemplates.steps.basicInfo') }}</h2>
              <p>{{ t('pages.emailTemplates.senderIdentityDescription') }}</p>
            </div>
          </header>
          <div class="resource-form-grid">
            <label class="ui-field">
              <span>{{ t('common.labels.type') }}</span>
              <select v-model="form.type" class="ui-select">
                <option value="EMAIL">{{ t('pages.emailTemplates.tabs.email') }}</option>
                <option value="MAX">{{ t('pages.emailTemplates.tabs.max') }}</option>
              </select>
            </label>
            <UiInput v-model="form.label" :label="t('common.labels.name')" :placeholder="t('pages.emailTemplates.placeholders.label')" />
            <template v-if="!isMaxForm">
              <UiInput v-model="form.from_email" :label="t('common.labels.fromEmail')" placeholder="training@example.com" />
              <UiInput v-model="form.from_name" :label="t('common.labels.fromName')" :placeholder="t('pages.emailTemplates.placeholders.fromName')" />
              <UiInput v-model="form.subject" :label="t('common.labels.subject')" :placeholder="t('pages.emailTemplates.placeholders.subject')" />
            </template>
          </div>
          <MaxAccountPickerPanel
            v-if="isMaxForm"
            v-model="form.max_account_id"
            :title="t('pages.emailTemplates.maxAccount')"
            :selected-title="t('pages.emailTemplates.selectedMaxAccount')"
            :selected-empty-text="t('pages.emailTemplates.selectMaxAccount')"
            :empty-title="t('pages.maxAccounts.emptyTitle')"
            :empty-description="t('pages.maxAccounts.emptyDescription')"
            :create-to="{ name: 'max-accounts-new' }"
            :create-label="t('pages.maxAccounts.addTitle')"
            @selection-change="trackMaxAccountSelection"
          />
        </section>

        <section v-else-if="activeStep === 1" class="resource-editor-panel">
          <template v-if="!isMaxForm">
            <AttachmentPickerPanel
              ref="emailHtmlBodyPicker"
              v-model="form.html_body_id"
              selection-mode="single"
              :title="t('pages.emailTemplates.htmlBodyAttachment')"
              :empty-text="t('pages.emailTemplates.noHtmlAttachments')"
              :allowed-types="['.html']"
              default-type=".html"
              show-preview-action
              show-duplicate-action
              show-edit-content-action
              show-download-action
              show-create-action
              @preview="openPreview"
              @edit-content="openContentEditor"
              @create="openBlankHtmlEditor"
              @selection-change="trackHtmlBodySelection"
            />

            <div class="resource-inline-actions resource-picker-actions">
              <UiButton
                variant="secondary"
                :disabled="!selectedHtmlBodyAttachment || isUploadingHtml"
                @click="openSelectedHtmlEditor"
              >
                <Edit3 :size="16" stroke-width="1.8" aria-hidden="true" />
                {{ t('pages.emailTemplates.editAsNewVersion') }}
              </UiButton>
            </div>
          </template>
          <AttachmentPickerPanel
            v-else
            ref="maxTextBodyPicker"
            v-model="form.text_body_id"
            selection-mode="single"
            :title="t('pages.emailTemplates.textBodyAttachment')"
            :empty-text="t('pages.emailTemplates.noTextAttachments')"
            :allowed-types="['.txt']"
            default-type=".txt"
            show-preview-action
            show-edit-content-action
            show-download-action
            show-upload-action
            @preview="openPreview"
            @edit-content="openContentEditor"
          />
        </section>

        <section v-else-if="activeStep === 2" class="resource-editor-panel">
          <UiAlert
            v-if="isMaxForm"
            variant="warning"
            :title="t('pages.emailTemplates.maxHtmlAttachmentWarningTitle')"
            :message="t('pages.emailTemplates.maxHtmlAttachmentWarningMessage')"
          />
          <AttachmentPickerPanel
            ref="templateAttachmentsPicker"
            v-model="form.attachment_ids"
            selection-mode="multiple"
            :title="t('pages.emailTemplates.templateAttachments')"
            :empty-text="t('common.empty.noMatchingRecords')"
            :allowed-types="isMaxForm ? maxAttachmentTypes : []"
            :restrict-to-allowed-types="isMaxForm"
            :unsupported-type-message="isMaxForm ? t('pages.emailTemplates.validation.maxHtmlAttachmentUnsupported') : ''"
            :all-types-label="isMaxForm ? t('pages.emailTemplates.maxAllowedAttachmentTypes') : ''"
            show-preview-action
            show-media-preview-action
            show-edit-content-action
            show-download-action
            show-upload-action
            @preview="openPreview"
            @edit-content="openContentEditor"
            @selection-change="trackTemplateAttachmentsSelection"
          >
            <template v-if="isMaxForm" #actions>
              <span class="voice-clone-button-wrap" :title="canOpenVoiceClone ? '' : voiceUnavailableMessage">
                <UiButton
                  variant="secondary"
                  :disabled="!canOpenVoiceClone"
                  @click="openVoiceClone"
                >
                  <Volume2 :size="16" stroke-width="1.8" aria-hidden="true" />
                  {{ t('pages.emailTemplates.voiceClone.button') }}
                </UiButton>
              </span>
            </template>
          </AttachmentPickerPanel>
        </section>

        <section v-else class="resource-editor-panel">
          <header class="resource-panel-header resource-panel-header--plain">
            <div>
              <h2>{{ t('pages.emailTemplates.steps.review') }}</h2>
              <p>{{ t('pages.emailTemplates.reviewDescription') }}</p>
            </div>
          </header>
          <dl class="resource-review-list">
            <div><dt>{{ t('common.labels.name') }}</dt><dd>{{ form.label || t('common.placeholder') }}</dd></div>
            <div><dt>{{ t('common.labels.type') }}</dt><dd>{{ form.type }}</dd></div>
            <template v-if="!isMaxForm">
              <div><dt>{{ t('common.labels.fromEmail') }}</dt><dd>{{ form.from_email || t('common.placeholder') }}</dd></div>
              <div><dt>{{ t('common.labels.fromName') }}</dt><dd>{{ form.from_name || t('common.placeholder') }}</dd></div>
              <div><dt>{{ t('common.labels.subject') }}</dt><dd>{{ form.subject || t('common.placeholder') }}</dd></div>
              <div><dt>{{ t('common.labels.htmlBody') }}</dt><dd><code>{{ formatTechnicalId(form.html_body_id) }}</code></dd></div>
            </template>
            <template v-else>
              <div><dt>{{ t('pages.emailTemplates.maxAccount') }}</dt><dd>{{ selectedMaxAccountDisplay }}</dd></div>
              <div><dt>{{ t('common.labels.textBody') }}</dt><dd><code>{{ formatTechnicalId(form.text_body_id) }}</code></dd></div>
            </template>
            <div><dt>{{ t('common.labels.attachments') }}</dt><dd>{{ form.attachment_ids.length }}</dd></div>
          </dl>
        </section>

        <footer class="resource-editor-actions">
          <UiButton variant="secondary" :disabled="activeStep === 0 || isSaving" @click="activeStep -= 1">
            {{ t('common.actions.back') }}
          </UiButton>
          <UiButton
            v-if="activeStep < steps.length - 1"
            variant="secondary"
            :disabled="isSaving"
            @click="activeStep += 1"
          >
            {{ t('common.actions.continue') }}
          </UiButton>
          <UiButton v-else :loading="isSaving" @click="submitForm">
            {{ isSaving ? t('common.states.saving') : isEditMode ? t('common.actions.saveChanges') : t('common.actions.createMessage') }}
          </UiButton>
        </footer>
      </section>
    </template>

    <AttachmentPreviewDrawer
      :open="previewOpen"
      :attachment="previewAttachment"
      @close="closePreview"
    />

    <ConfirmDialog
      :open="deleteDialogOpen"
      :title="t('pages.emailTemplates.delete.title')"
      :message="t('pages.emailTemplates.delete.message')"
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
