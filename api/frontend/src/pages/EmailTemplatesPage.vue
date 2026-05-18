<script setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  Copy,
  Download,
  Edit3,
  Eye,
  Plus,
  Trash2,
} from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import AttachmentPickerPanel from '../components/ui/AttachmentPickerPanel.vue'
import ConfirmDialog from '../components/ui/ConfirmDialog.vue'
import HtmlAttachmentPreview from '../components/ui/HtmlAttachmentPreview.vue'
import HtmlLiveEditor from '../components/ui/HtmlLiveEditor.vue'
import IconButton from '../components/ui/IconButton.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import PreviewDrawer from '../components/ui/PreviewDrawer.vue'
import ResourcePagination from '../components/ui/ResourcePagination.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiInput from '../components/ui/UiInput.vue'
import WizardStepper from '../components/ui/WizardStepper.vue'
import { useNotifications } from '../composables/useNotifications'
import { useResourceActions } from '../composables/useResourceActions'
import {
  getAttachmentContent,
  getAttachmentUrl,
  listAttachments,
  uploadAttachment,
} from '../resources/attachments'
import {
  createMessage,
  deleteMessage,
  getMessage,
  listMessages,
  updateMessage,
} from '../resources/messages'
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
  t('common.labels.htmlBody'),
  t('common.labels.attachments'),
  t('pages.emailTemplates.steps.review'),
])

const messages = ref([])
const attachments = ref([])
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
const previewTitle = ref('')
const htmlEditorOpen = ref(false)
const htmlEditorInitial = ref('')
const htmlDraftFilename = ref('')
const selectedHtmlBodyAttachment = ref(null)
const emailHtmlBodyPicker = ref(null)

const filters = reactive({
  label: '',
  fromEmail: '',
  fromName: '',
  subject: '',
})

const form = reactive({
  label: '',
  from_email: '',
  from_name: '',
  subject: '',
  html_body_id: '',
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
const previewOpen = computed(() => Boolean(previewAttachment.value))
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / rows.value)))

function clearForm() {
  form.label = ''
  form.from_email = ''
  form.from_name = ''
  form.subject = ''
  form.html_body_id = ''
  form.attachment_ids = []
  formError.value = ''
  activeStep.value = 0
  htmlEditorOpen.value = false
  htmlEditorInitial.value = ''
  htmlDraftFilename.value = ''
  selectedHtmlBodyAttachment.value = null
}

function fillForm(message) {
  form.label = message?.label || ''
  form.from_email = message?.from_email || ''
  form.from_name = message?.from_name || ''
  form.subject = message?.subject || ''
  form.html_body_id = nullableId(message?.html_body_id)
  form.attachment_ids = Array.isArray(message?.attachment_ids)
    ? message.attachment_ids.map((id) => nullableId(id) || String(id)).filter(Boolean)
    : []
  formError.value = ''
  activeStep.value = 0
  htmlEditorOpen.value = false
  selectedHtmlBodyAttachment.value = null
}

function fillDuplicateForm(message) {
  fillForm(message)
  form.label = t('pages.emailTemplates.copyLabel', { label: message?.label || t('common.labels.message') })
}

function buildPayload() {
  return {
    label: form.label.trim(),
    from_email: form.from_email.trim(),
    from_name: form.from_name.trim(),
    subject: form.subject.trim(),
    html_body_id: nullableId(form.html_body_id),
    attachment_ids: form.attachment_ids,
  }
}

function validateForm() {
  if (!form.label.trim()) {
    return t('pages.emailTemplates.validation.labelRequired')
  }

  if (!form.from_email.trim()) {
    return t('pages.emailTemplates.validation.fromEmailRequired')
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

function openPreview(attachment, title = t('common.preview.template')) {
  if (!attachment) {
    notifyError(t('common.preview.unavailable'), t('common.preview.selectHtmlOrText'))
    return
  }

  previewAttachment.value = attachment
  previewTitle.value = title
}

function closePreview() {
  previewAttachment.value = null
}

function attachmentDownloadUrl(attachment) {
  return attachment?.id ? getAttachmentUrl(attachment.id) : ''
}

async function loadVisibleHtmlBodyAttachments(items) {
  const ids = [...new Set(items.map((message) => nullableId(message?.html_body_id)).filter(Boolean))]

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
  htmlEditorInitial.value = ''
  htmlDraftFilename.value = form.label || 'message-body'
  htmlEditorOpen.value = true
}

function closeHtmlEditor() {
  htmlEditorOpen.value = false
  htmlEditorInitial.value = ''
}

function trackHtmlBodySelection(attachment) {
  selectedHtmlBodyAttachment.value = attachment
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
    htmlEditorInitial.value = typeof content === 'string' ? content : ''
    htmlDraftFilename.value = htmlCopyFileName(selectedHtmlBodyAttachment.value)
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
    htmlEditorOpen.value = false
    htmlEditorInitial.value = ''
    await nextTick()
    await emailHtmlBodyPicker.value?.refresh?.()
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

async function submitForm() {
  const validation = validateForm()
  if (validation) {
    formError.value = validation
    activeStep.value = 0
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
    },
  })
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
          <RouterLink class="ui-button ui-button--primary" :to="{ name: 'email-template-new' }">
            <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('common.actions.createMessage') }}
          </RouterLink>
        </template>
      </PageHeader>

      <section class="resource-filter-panel">
        <UiInput v-model="filters.label" :label="t('common.labels.name')" :placeholder="t('common.resource.searchByLabel')" />
        <UiInput v-model="filters.fromEmail" :label="t('common.labels.fromEmail')" placeholder="sender@example.com" />
        <UiInput v-model="filters.fromName" :label="t('common.labels.fromName')" :placeholder="t('common.labels.fromName')" />
        <UiInput v-model="filters.subject" :label="t('common.labels.subject')" :placeholder="t('common.labels.subject')" />
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

        <div class="resource-table resource-email-table">
          <div class="resource-table-row resource-table-head">
            <span>{{ t('common.labels.name') }}</span>
            <span>{{ t('common.labels.fromEmail') }}</span>
            <span>{{ t('common.labels.fromName') }}</span>
            <span>{{ t('common.labels.subject') }}</span>
            <span>{{ t('common.labels.htmlBody') }}</span>
            <span>{{ t('common.labels.attachments') }}</span>
            <span>{{ t('pages.campaigns.columns.actions') }}</span>
          </div>
          <template v-if="isLoading">
            <div v-for="index in 4" :key="index" class="resource-table-row">
              <span v-for="cell in 7" :key="cell"><SkeletonBlock :rows="1" /></span>
            </div>
          </template>
          <template v-else-if="messages.length">
            <div
              v-for="message in messages"
              :key="message.id"
              class="resource-table-row"
            >
              <span><strong>{{ message.label }}</strong></span>
              <span>{{ message.from_email || t('common.placeholder') }}</span>
              <span>{{ message.from_name || t('common.placeholder') }}</span>
              <span>{{ message.subject || t('common.placeholder') }}</span>
              <span><code>{{ formatTechnicalId(message.html_body_id) }}</code></span>
              <span>{{ Array.isArray(message.attachment_ids) ? message.attachment_ids.length : 0 }}</span>
              <span class="resource-row-actions">
                <IconButton
                  :label="t('pages.emailTemplates.previewHtmlBody')"
                  :disabled="!attachmentById[nullableId(message.html_body_id)]"
                  @click="openPreview(attachmentById[nullableId(message.html_body_id)], t('pages.emailTemplates.emailHtmlPreview'))"
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
        v-if="!htmlEditorOpen"
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
          :initial-html="htmlEditorInitial"
          :title="t('pages.emailTemplates.htmlBodyTitle')"
          filename-placeholder="message-body.html"
          :save-label="t('pages.emailTemplates.saveHtmlAttachment')"
          :saving="isUploadingHtml"
          @save="saveHtmlSource"
          @cancel="closeHtmlEditor"
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
            <UiInput v-model="form.label" :label="t('common.labels.name')" :placeholder="t('pages.emailTemplates.placeholders.label')" />
            <UiInput v-model="form.from_email" :label="t('common.labels.fromEmail')" placeholder="training@example.com" />
            <UiInput v-model="form.from_name" :label="t('common.labels.fromName')" :placeholder="t('pages.emailTemplates.placeholders.fromName')" />
            <UiInput v-model="form.subject" :label="t('common.labels.subject')" :placeholder="t('pages.emailTemplates.placeholders.subject')" />
          </div>
        </section>

        <section v-else-if="activeStep === 1" class="resource-editor-panel">
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
            show-download-action
            show-create-action
            @preview="openPreview($event, t('pages.emailTemplates.emailHtmlPreview'))"
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
        </section>

        <section v-else-if="activeStep === 2" class="resource-editor-panel">
          <AttachmentPickerPanel
            v-model="form.attachment_ids"
            selection-mode="multiple"
            :title="t('pages.emailTemplates.templateAttachments')"
            :empty-text="t('common.empty.noMatchingRecords')"
            show-preview-action
            show-download-action
            show-upload-action
            @preview="openPreview($event, t('common.preview.attachment'))"
          />
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
            <div><dt>{{ t('common.labels.fromEmail') }}</dt><dd>{{ form.from_email || t('common.placeholder') }}</dd></div>
            <div><dt>{{ t('common.labels.fromName') }}</dt><dd>{{ form.from_name || t('common.placeholder') }}</dd></div>
            <div><dt>{{ t('common.labels.subject') }}</dt><dd>{{ form.subject || t('common.placeholder') }}</dd></div>
            <div><dt>{{ t('common.labels.htmlBody') }}</dt><dd><code>{{ formatTechnicalId(form.html_body_id) }}</code></dd></div>
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

    <PreviewDrawer
      :open="previewOpen"
      :title="previewTitle"
      :subtitle="t('common.preview.rawTemplate')"
      @close="closePreview"
    >
      <template #actions>
        <a
          v-if="previewAttachment"
          class="ui-button ui-button--secondary"
          :href="attachmentDownloadUrl(previewAttachment)"
          target="_blank"
          rel="noreferrer"
        >
          <Download :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ t('common.actions.download') }}
        </a>
      </template>
      <HtmlAttachmentPreview
        :attachment="previewAttachment"
        :attachment-id="previewAttachment?.id || ''"
        :title="t('common.preview.template')"
        :empty-text="t('common.preview.selectHtmlOrText')"
      />
    </PreviewDrawer>

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
