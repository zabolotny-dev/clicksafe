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
  createLanding,
  deleteLanding,
  getLanding,
  listLandings,
  updateLanding,
} from '../resources/landings'
import { errorMessage, formatTechnicalId, nullableId } from '../utils/resourceFormat'

const route = useRoute()
const router = useRouter()
const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()
const { t } = useI18n()

const steps = computed(() => [
  t('pages.landingPages.steps.basicInfo'),
  t('common.labels.htmlBody'),
  t('pages.landingPages.steps.review'),
])

const landings = ref([])
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
const htmlEditorOpen = ref(false)
const htmlEditorInitial = ref('')
const htmlDraftFilename = ref('')
const selectedHtmlBodyAttachment = ref(null)
const landingHtmlBodyPicker = ref(null)

const filters = reactive({
  label: '',
})

const form = reactive({
  label: '',
  html_body_id: '',
})

const isEditorRoute = computed(() => ['landing-page-new', 'landing-page-edit'].includes(route.name))
const isEditMode = computed(() => route.name === 'landing-page-edit')
const editingId = computed(() => String(route.params.id || ''))
const pageTitle = computed(() => (isEditMode.value ? t('pages.landingPages.editTitle') : t('pages.landingPages.createTitle')))
const pageDescription = computed(() => (
  isEditMode.value
    ? t('pages.landingPages.editDescription')
    : t('pages.landingPages.createDescription')
))
const attachmentById = computed(() => attachments.value.reduce((acc, attachment) => {
  acc[attachment.id] = attachment
  return acc
}, {}))
const previewOpen = computed(() => Boolean(previewAttachment.value))
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / rows.value)))

function clearForm() {
  form.label = ''
  form.html_body_id = ''
  formError.value = ''
  activeStep.value = 0
  htmlEditorOpen.value = false
  htmlEditorInitial.value = ''
  htmlDraftFilename.value = ''
  selectedHtmlBodyAttachment.value = null
}

function fillForm(landing) {
  form.label = landing?.label || ''
  form.html_body_id = nullableId(landing?.html_body_id)
  formError.value = ''
  activeStep.value = 0
  htmlEditorOpen.value = false
  selectedHtmlBodyAttachment.value = null
}

function fillDuplicateForm(landing) {
  fillForm(landing)
  form.label = t('pages.landingPages.copyLabel', { label: landing?.label || t('common.labels.landingPage') })
}

function validateForm() {
  if (!form.label.trim()) {
    return t('pages.landingPages.validation.labelRequired')
  }

  return ''
}

function buildPayload() {
  return {
    label: form.label.trim(),
    html_body_id: nullableId(form.html_body_id),
  }
}

function htmlFileName() {
  const base = (htmlDraftFilename.value || form.label || 'landing-body')
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9-]+/g, '-')
    .replace(/^-|-$/g, '') || 'landing-body'

  return base.endsWith('.html') ? base : `${base}.html`
}

function openPreview(attachment) {
  if (!attachment) {
    notifyError(t('common.preview.unavailable'), t('common.preview.selectHtmlAsset'))
    return
  }

  previewAttachment.value = attachment
}

function closePreview() {
  previewAttachment.value = null
}

function attachmentDownloadUrl(attachment) {
  return attachment?.id ? getAttachmentUrl(attachment.id) : ''
}

async function loadVisibleHtmlBodyAttachments(items) {
  const ids = [...new Set(items.map((landing) => nullableId(landing?.html_body_id)).filter(Boolean))]

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
    const landingResponse = await listLandings({
      ...filters,
      page: page.value,
      rows: rows.value,
    })

    landings.value = Array.isArray(landingResponse?.items) ? landingResponse.items : []
    total.value = Number.isFinite(landingResponse?.total) ? landingResponse.total : landings.value.length
    page.value = Number.isFinite(landingResponse?.page) ? landingResponse.page : page.value
    rows.value = Number.isFinite(landingResponse?.rowsPerPage) ? landingResponse.rowsPerPage : rows.value

    if (options.backfillEmptyPage && total.value > 0 && !landings.value.length && page.value > 1) {
      page.value -= 1
      await loadData()
      return
    }

    if (isEditorRoute.value) {
      attachments.value = []
    } else {
      await loadVisibleHtmlBodyAttachments(landings.value)
    }

    await prepareEditorFromRoute()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    loadError.value = errorMessage(error, 'errors.landingsLoad')
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
    const landing = landings.value.find((item) => item.id === editingId.value) || await getLanding(editingId.value)
    fillForm(landing)
    return
  }

  const duplicateId = String(route.query.duplicate || '')
  if (duplicateId) {
    const source = landings.value.find((item) => item.id === duplicateId) || await getLanding(duplicateId)
    fillDuplicateForm(source)
  }
}

function openBlankHtmlEditor() {
  formError.value = ''
  htmlEditorInitial.value = ''
  htmlDraftFilename.value = form.label || 'landing-body'
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
  const label = attachment?.label || form.label || 'landing-body'
  const cleanLabel = label.replace(/\.html$/i, '')
  return `${cleanLabel}-copy.html`
}

async function openSelectedHtmlEditor() {
  if (!selectedHtmlBodyAttachment.value?.id) {
    notifyError(t('pages.landingPages.notifications.editorUnavailableTitle'), t('pages.landingPages.notifications.editorUnavailableMessage'))
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

    notifyError(t('pages.landingPages.notifications.htmlSourceUnavailableTitle'), errorMessage(error, 'errors.attachmentsLoad'))
  }
}

async function saveHtmlSource(html) {
  if (!html.trim()) {
    formError.value = t('pages.landingPages.validation.htmlEmpty')
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
    await landingHtmlBodyPicker.value?.refresh?.()
    notifySuccess(t('pages.landingPages.notifications.htmlSavedTitle'), t('pages.landingPages.notifications.htmlSavedMessage'))
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.attachmentsLoad')
    notifyError(t('pages.landingPages.notifications.htmlUploadFailedTitle'), formError.value)
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
      await updateLanding(editingId.value, buildPayload(), mutationOptions())
      notifySuccess(t('pages.landingPages.notifications.updated'))
    } else {
      await createLanding(buildPayload(), mutationOptions())
      notifySuccess(t('pages.landingPages.notifications.created'))
    }

    await router.push({ name: 'landing-pages' })
    await loadData()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.landingsLoad')
    notifyError(t('pages.landingPages.notifications.saveFailedTitle'), formError.value)
  } finally {
    isSaving.value = false
  }
}

function openDeleteDialog(landing) {
  deleteTarget.value = landing
  deleteDialogOpen.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value || !(await ensureCsrfToken())) {
    return
  }

  isDeleting.value = true

  try {
    await deleteLanding(deleteTarget.value.id, mutationOptions())
    notifySuccess(t('pages.landingPages.notifications.deleted'))
    deleteDialogOpen.value = false
    deleteTarget.value = null
    await loadData({ backfillEmptyPage: true })
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    notifyError(t('pages.landingPages.notifications.deleteFailedTitle'), errorMessage(error, 'errors.landingsLoad'))
  } finally {
    isDeleting.value = false
  }
}

function goToDuplicate(landing) {
  router.push({
    name: 'landing-page-new',
    query: {
      duplicate: landing.id,
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
        :title="t('pages.landingPages.title')"
        :description="t('pages.landingPages.description')"
      >
        <template #actions>
          <RouterLink class="ui-button ui-button--primary" :to="{ name: 'landing-page-new' }">
            <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('common.actions.createLanding') }}
          </RouterLink>
        </template>
      </PageHeader>

      <section class="resource-filter-panel resource-filter-panel--compact">
        <UiInput v-model="filters.label" :label="t('common.labels.name')" :placeholder="t('common.resource.searchByLabel')" />
        <UiButton variant="secondary" :loading="isLoading" @click="applyFilters">
          {{ t('common.filters.apply') }}
        </UiButton>
      </section>

      <UiAlert
        v-if="loadError"
        variant="error"
        :title="t('pages.landingPages.unavailable')"
        :message="loadError"
      >
        <UiButton variant="secondary" @click="loadData">{{ t('common.actions.retry') }}</UiButton>
      </UiAlert>

      <section class="resource-panel resource-panel--table">
        <header class="resource-panel-header">
          <div>
            <h2>{{ t('pages.landingPages.tableTitle') }}</h2>
            <p>{{ t('common.loadedCount', { count: landings.length }) }}</p>
          </div>
        </header>

        <div class="resource-table resource-landing-table">
          <div class="resource-table-row resource-table-head">
            <span>{{ t('common.labels.name') }}</span>
            <span>{{ t('common.labels.htmlBody') }}</span>
            <span>{{ t('pages.campaigns.columns.actions') }}</span>
          </div>
          <template v-if="isLoading">
            <div v-for="index in 4" :key="index" class="resource-table-row">
              <span v-for="cell in 3" :key="cell"><SkeletonBlock :rows="1" /></span>
            </div>
          </template>
          <template v-else-if="landings.length">
            <div
              v-for="landing in landings"
              :key="landing.id"
              class="resource-table-row"
            >
              <span><strong>{{ landing.label }}</strong></span>
              <span><code>{{ formatTechnicalId(landing.html_body_id) }}</code></span>
              <span class="resource-row-actions">
                <IconButton
                  :label="t('pages.landingPages.previewHtml')"
                  :disabled="!attachmentById[nullableId(landing.html_body_id)]"
                  @click="openPreview(attachmentById[nullableId(landing.html_body_id)])"
                >
                  <Eye :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
                <RouterLink
                  class="icon-button icon-button--ghost"
                  :to="{ name: 'landing-page-edit', params: { id: landing.id } }"
                  :aria-label="t('common.actions.edit')"
                  :title="t('common.actions.edit')"
                >
                  <Edit3 :size="16" stroke-width="1.8" aria-hidden="true" />
                </RouterLink>
                <IconButton :label="t('common.actions.duplicate')" @click="goToDuplicate(landing)">
                  <Copy :size="16" stroke-width="1.8" aria-hidden="true" />
                </IconButton>
                <IconButton :label="t('common.actions.delete')" variant="danger" @click="openDeleteDialog(landing)">
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
          :loaded-count="landings.length"
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
          <RouterLink class="ui-button ui-button--secondary" :to="{ name: 'landing-pages' }">
            {{ t('pages.landingPages.backToList') }}
          </RouterLink>
        </template>
      </PageHeader>

      <UiAlert
        v-if="formError"
        variant="error"
        :title="t('pages.landingPages.formAttention')"
        :message="formError"
        dismissible
        @dismiss="formError = ''"
      />

      <section v-if="htmlEditorOpen" class="resource-html-workspace">
        <HtmlLiveEditor
          v-model:filename="htmlDraftFilename"
          workspace
          :initial-html="htmlEditorInitial"
          :title="t('pages.landingPages.htmlBodyTitle')"
          filename-placeholder="landing-body.html"
          :save-label="t('pages.landingPages.saveHtmlAttachment')"
          :saving="isUploadingHtml"
          @save="saveHtmlSource"
          @cancel="closeHtmlEditor"
        />
      </section>

      <section v-else class="resource-editor-shell">
        <WizardStepper
          v-model:active-index="activeStep"
          :steps="steps"
          :aria-label="t('pages.landingPages.stepsAria')"
        />

        <section v-if="activeStep === 0" class="resource-editor-panel">
          <header class="resource-panel-header resource-panel-header--plain">
            <div>
              <h2>{{ t('pages.landingPages.steps.basicInfo') }}</h2>
              <p>{{ t('pages.landingPages.basicInfoDescription') }}</p>
            </div>
          </header>
          <div class="resource-form-grid resource-form-grid--single">
            <UiInput v-model="form.label" :label="t('common.labels.name')" :placeholder="t('pages.landingPages.labelPlaceholder')" />
          </div>
        </section>

        <section v-else-if="activeStep === 1" class="resource-editor-panel">
          <AttachmentPickerPanel
            ref="landingHtmlBodyPicker"
            v-model="form.html_body_id"
            selection-mode="single"
            :title="t('pages.landingPages.htmlBodyAttachment')"
            :empty-text="t('pages.landingPages.noHtmlAttachments')"
            :allowed-types="['.html']"
            default-type=".html"
            show-preview-action
            show-duplicate-action
            show-download-action
            show-create-action
            @preview="openPreview"
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
              {{ t('pages.landingPages.editAsNewVersion') }}
            </UiButton>
          </div>
        </section>

        <section v-else class="resource-editor-panel">
          <header class="resource-panel-header resource-panel-header--plain">
            <div>
              <h2>{{ t('pages.landingPages.steps.review') }}</h2>
              <p>{{ t('pages.landingPages.reviewDescription') }}</p>
            </div>
          </header>
          <dl class="resource-review-list">
            <div><dt>{{ t('common.labels.name') }}</dt><dd>{{ form.label || t('common.placeholder') }}</dd></div>
            <div><dt>{{ t('common.labels.htmlBody') }}</dt><dd><code>{{ formatTechnicalId(form.html_body_id) }}</code></dd></div>
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
            {{ isSaving ? t('common.states.saving') : isEditMode ? t('common.actions.saveChanges') : t('common.actions.createLanding') }}
          </UiButton>
        </footer>
      </section>
    </template>

    <PreviewDrawer
      :open="previewOpen"
      :title="t('pages.landingPages.htmlPreviewTitle')"
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
        :empty-text="t('common.preview.selectHtmlAsset')"
      />
    </PreviewDrawer>

    <ConfirmDialog
      :open="deleteDialogOpen"
      :title="t('pages.landingPages.delete.title')"
      :message="t('pages.landingPages.delete.message')"
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
