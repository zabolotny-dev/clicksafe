<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Code,
  Download,
  Edit3,
  Eye,
  FileText,
  FileUp,
  Image,
  Trash2,
  UploadCloud,
  Volume2,
} from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import AttachmentPreviewDrawer from '../components/ui/AttachmentPreviewDrawer.vue'
import ConfirmDialog from '../components/ui/ConfirmDialog.vue'
import IconButton from '../components/ui/IconButton.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import PreviewDrawer from '../components/ui/PreviewDrawer.vue'
import ResourcePagination from '../components/ui/ResourcePagination.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiInput from '../components/ui/UiInput.vue'
import { useNotifications } from '../composables/useNotifications'
import { useResourceActions } from '../composables/useResourceActions'
import {
  deleteAttachment,
  getAttachmentUrl,
  listAttachments,
  updateAttachment,
  uploadAttachment,
} from '../resources/attachments'
import {
  errorMessage,
  formatDateTime,
  formatKilobytes,
  formatTechnicalId,
} from '../utils/resourceFormat'
import {
  isAudioAttachmentType,
  isHtmlAttachmentType,
  isImageAttachmentType,
  isTextPreviewAttachmentType,
  SUPPORTED_ATTACHMENT_TYPES,
} from '../utils/attachmentTypes'

const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()
const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const ATTACHMENT_TYPE_OPTIONS = SUPPORTED_ATTACHMENT_TYPES

const attachments = ref([])
const page = ref(1)
const rows = ref(10)
const total = ref(0)
const selectedFile = ref(null)
const fileInput = ref(null)
const isDragging = ref(false)
const isLoading = ref(false)
const isUploading = ref(false)
const isSaving = ref(false)
const isDeleting = ref(false)
const loadError = ref('')
const formError = ref('')
const uploadDrawerOpen = ref(false)
const editDrawerOpen = ref(false)
const deleteDialogOpen = ref(false)
const previewAttachment = ref(null)
const editTarget = ref(null)
const deleteTarget = ref(null)

const filters = reactive({
  id: '',
  label: '',
  type: '',
})

const uploadForm = reactive({
  label: '',
  public: false,
})

const editForm = reactive({
  label: '',
  public: false,
})

const previewOpen = computed(() => Boolean(previewAttachment.value))
const editDownloadUrl = computed(() => editTarget.value ? getAttachmentUrl(editTarget.value.id) : '')
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / rows.value)))
const selectedFileType = computed(() => {
  if (!selectedFile.value?.name) {
    return t('common.placeholder')
  }

  const parts = selectedFile.value.name.split('.')
  return parts.length > 1 ? `.${parts.pop().toLowerCase()}` : t('common.labels.unknown')
})

function resetUploadForm() {
  uploadForm.label = ''
  uploadForm.public = false
  selectedFile.value = null
  formError.value = ''

  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

function openUploadDrawer() {
  resetUploadForm()
  uploadDrawerOpen.value = true
}

function closeUploadDrawer() {
  if (isUploading.value) {
    return
  }

  uploadDrawerOpen.value = false
  resetUploadForm()
}

function handleFileChange(event) {
  selectedFile.value = event.target.files?.[0] || null
}

function handleDrop(event) {
  event.preventDefault()
  isDragging.value = false
  selectedFile.value = event.dataTransfer?.files?.[0] || null
}

function openPreview(attachment) {
  previewAttachment.value = attachment
}

function closePreview() {
  previewAttachment.value = null
}

function canEditContent(attachment) {
  return isTextPreviewAttachmentType(attachment?.type)
}

function isMediaPreviewable(attachment) {
  return isImageAttachmentType(attachment?.type) || isAudioAttachmentType(attachment?.type)
}

function isHtmlLike(attachment) {
  return isHtmlAttachmentType(attachment?.type)
}

function attachmentIcon(attachment) {
  if (isMediaPreviewable(attachment)) {
    return isAudioAttachmentType(attachment?.type) ? Volume2 : Image
  }

  return isHtmlLike(attachment) ? Code : FileText
}

function openContentEditor(attachment) {
  if (!canEditContent(attachment)) {
    return
  }

  router.push({
    name: 'attachment-content-edit',
    params: { id: attachment.id },
    query: { redirect: route.fullPath || '/templates/attachments' },
  })
}

function openEditDrawer(attachment) {
  editTarget.value = attachment
  editForm.label = attachment.label || ''
  editForm.public = Boolean(attachment.public)
  formError.value = ''
  editDrawerOpen.value = true
}

function closeEditDrawer() {
  if (isSaving.value) {
    return
  }

  editDrawerOpen.value = false
  editTarget.value = null
  editForm.label = ''
  editForm.public = false
}

function openDeleteDialog(attachment) {
  deleteTarget.value = attachment
  deleteDialogOpen.value = true
}

function attachmentDownloadUrl(attachment) {
  return attachment?.id ? getAttachmentUrl(attachment.id) : ''
}

async function copyPublicPath(path) {
  if (!path || typeof navigator === 'undefined' || !navigator.clipboard?.writeText) {
    notifyError(t('pages.attachments.notifications.copyUnavailableTitle'), t('pages.attachments.notifications.copyUnavailableMessage'))
    return
  }

  try {
    await navigator.clipboard.writeText(path)
    notifySuccess(t('pages.attachments.notifications.publicPathCopied'))
  } catch (error) {
    notifyError(t('pages.attachments.notifications.copyFailedTitle'), t('pages.attachments.notifications.copyFailedMessage'))
  }
}

async function loadData(options = {}) {
  isLoading.value = true
  loadError.value = ''

  try {
    const response = await listAttachments({
      ...filters,
      page: page.value,
      rows: rows.value,
    })
    attachments.value = Array.isArray(response?.items) ? response.items : []
    total.value = Number.isFinite(response?.total) ? response.total : attachments.value.length
    page.value = Number.isFinite(response?.page) ? response.page : page.value
    rows.value = Number.isFinite(response?.rowsPerPage) ? response.rowsPerPage : rows.value

    if (options.backfillEmptyPage && total.value > 0 && !attachments.value.length && page.value > 1) {
      page.value -= 1
      await loadData()
      return
    }

    if (editTarget.value) {
      const current = attachments.value.find((attachment) => attachment.id === editTarget.value.id)
      if (current) {
        openEditDrawer(current)
      }
    }
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    loadError.value = errorMessage(error, 'errors.attachmentsLoad')
  } finally {
    isLoading.value = false
  }
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

async function uploadSelectedFile() {
  if (!selectedFile.value) {
    formError.value = t('pages.attachments.validation.selectFile')
    return
  }

  if (!(await ensureCsrfToken())) {
    return
  }

  isUploading.value = true
  formError.value = ''

  try {
    let uploaded = await uploadAttachment(selectedFile.value, {
      public: uploadForm.public,
    }, mutationOptions())

    if (uploadForm.label.trim()) {
      uploaded = await updateAttachment(uploaded.id, {
        label: uploadForm.label.trim(),
        public: uploadForm.public,
      }, mutationOptions())
    }

    notifySuccess(t('pages.attachments.notifications.uploaded'))
    uploadDrawerOpen.value = false
    resetUploadForm()
    await loadData()
    openPreview(uploaded)
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.attachmentsLoad')
    notifyError(t('pages.attachments.notifications.uploadFailedTitle'), formError.value)
  } finally {
    isUploading.value = false
  }
}

async function saveAttachment() {
  if (!editTarget.value) {
    formError.value = t('pages.attachments.validation.selectFileFirst')
    return
  }

  if (!editForm.label.trim()) {
    formError.value = t('pages.attachments.validation.labelRequired')
    return
  }

  if (!(await ensureCsrfToken())) {
    return
  }

  isSaving.value = true
  formError.value = ''

  try {
    const saved = await updateAttachment(editTarget.value.id, {
      label: editForm.label.trim(),
      public: editForm.public,
    }, mutationOptions())

    notifySuccess(t('pages.attachments.notifications.updated'))
    await loadData()
    openEditDrawer(saved)
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.attachmentsLoad')
    notifyError(t('pages.attachments.notifications.updateFailedTitle'), formError.value)
  } finally {
    isSaving.value = false
  }
}

async function confirmDelete() {
  if (!deleteTarget.value || !(await ensureCsrfToken())) {
    return
  }

  isDeleting.value = true

  try {
    const deletedId = deleteTarget.value.id
    await deleteAttachment(deletedId, mutationOptions())
    notifySuccess(t('pages.attachments.notifications.deleted'))
    deleteDialogOpen.value = false
    deleteTarget.value = null

    if (editTarget.value?.id === deletedId) {
      closeEditDrawer()
    }

    await loadData({ backfillEmptyPage: true })
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    notifyError(t('pages.attachments.notifications.deleteFailedTitle'), errorMessage(error, 'errors.attachmentsLoad'))
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
      :eyebrow="t('sections.templates')"
      :title="t('pages.attachments.title')"
      :description="t('pages.attachments.description')"
    >
      <template #actions>
        <UiButton @click="openUploadDrawer">
          <UploadCloud :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ t('common.actions.upload') }}
        </UiButton>
      </template>
    </PageHeader>

    <section class="resource-filter-panel resource-filter-panel--attachments">
      <UiInput v-model="filters.id" label="ID" placeholder="Attachment ID" />
      <UiInput v-model="filters.label" :label="t('common.labels.name')" :placeholder="t('common.resource.searchByLabel')" />
      <label class="ui-field">
        <span class="ui-field-label">{{ t('common.labels.type') }}</span>
        <select v-model="filters.type" class="ui-select">
          <option value="">{{ t('common.resource.allTypes') }}</option>
          <option v-for="type in ATTACHMENT_TYPE_OPTIONS" :key="type" :value="type">
            {{ type }}
          </option>
        </select>
      </label>
      <UiButton variant="secondary" :loading="isLoading" @click="applyFilters">
        {{ t('common.filters.apply') }}
      </UiButton>
    </section>

    <UiAlert
      v-if="loadError"
      variant="error"
      :title="t('common.resource.attachmentsUnavailable')"
      :message="loadError"
    >
      <UiButton variant="secondary" @click="loadData">{{ t('common.actions.retry') }}</UiButton>
    </UiAlert>

    <section class="resource-panel resource-panel--table">
      <header class="resource-panel-header">
        <div>
          <h2>{{ t('pages.attachments.libraryTitle') }}</h2>
          <p>{{ t('common.loadedCount', { count: attachments.length }) }}</p>
        </div>
      </header>

      <div class="resource-table resource-attachment-table">
        <div class="resource-table-row resource-table-head">
          <span>{{ t('common.labels.name') }}</span>
          <span>{{ t('common.labels.type') }}</span>
          <span>{{ t('common.labels.requiredVars') }}</span>
          <span>{{ t('common.labels.publicPath') }}</span>
          <span>{{ t('common.labels.uploadedAt') }}</span>
          <span>{{ t('pages.campaigns.columns.actions') }}</span>
        </div>
        <template v-if="isLoading">
          <div v-for="index in 4" :key="index" class="resource-table-row">
            <span v-for="cell in 6" :key="cell"><SkeletonBlock :rows="1" /></span>
          </div>
        </template>
        <template v-else-if="attachments.length">
          <div
            v-for="attachment in attachments"
            :key="attachment.id"
            class="resource-table-row"
          >
            <span class="attachment-picker-label-cell">
              <span class="attachment-picker-file-icon" aria-hidden="true">
                <component :is="attachmentIcon(attachment)" :size="16" stroke-width="1.8" />
              </span>
              <strong :title="attachment.label">{{ attachment.label || formatTechnicalId(attachment.id) }}</strong>
            </span>
            <span>{{ attachment.type }}</span>
            <span>{{ attachment.required_vars?.length ? attachment.required_vars.join(', ') : '-' }}</span>
            <span>
              <button
                v-if="attachment.public_path"
                type="button"
                class="resource-path-chip"
                :title="t('common.actions.copyPublicPath')"
                @click="copyPublicPath(attachment.public_path)"
              >
                <code>{{ attachment.public_path }}</code>
              </button>
              <span v-else class="resource-muted-value">
                {{ attachment.public ? t('common.placeholder') : t('common.labels.private') }}
              </span>
            </span>
            <span>{{ formatDateTime(attachment.uploaded_at) }}</span>
            <span class="resource-row-actions">
              <IconButton :label="t('common.preview.attachment')" @click="openPreview(attachment)">
                <Eye :size="16" stroke-width="1.8" aria-hidden="true" />
              </IconButton>
              <a
                class="icon-button icon-button--ghost"
                :href="attachmentDownloadUrl(attachment)"
                target="_blank"
                rel="noreferrer"
                :aria-label="t('common.actions.download')"
                :title="t('common.actions.download')"
              >
                <Download :size="16" stroke-width="1.8" aria-hidden="true" />
              </a>
              <IconButton :label="t('common.actions.edit')" @click="openEditDrawer(attachment)">
                <Edit3 :size="16" stroke-width="1.8" aria-hidden="true" />
              </IconButton>
              <IconButton :label="t('common.actions.delete')" variant="danger" @click="openDeleteDialog(attachment)">
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
        :loaded-count="attachments.length"
        :loading="isLoading"
        @update:page="setPage"
        @update:rows="setRows"
      />
    </section>

    <PreviewDrawer
      :open="uploadDrawerOpen"
      :title="t('pages.attachments.uploadTitle')"
      :subtitle="t('common.resource.singleFileUpload')"
      @close="closeUploadDrawer"
    >
      <UiAlert
        v-if="formError"
        variant="error"
        :title="t('pages.attachments.uploadErrorTitle')"
        :message="formError"
        dismissible
        @dismiss="formError = ''"
      />

      <section
        class="attachment-dropzone"
        :class="{ 'is-dragging': isDragging }"
        @dragover.prevent="isDragging = true"
        @dragleave.prevent="isDragging = false"
        @drop="handleDrop"
      >
        <UploadCloud :size="28" stroke-width="1.6" aria-hidden="true" />
        <h3>{{ selectedFile ? selectedFile.name : t('common.resource.dropOneFile') }}</h3>
        <p>{{ t('common.resource.singleFileUpload') }}</p>
        <input ref="fileInput" class="ui-file-input" type="file" :accept="ATTACHMENT_TYPE_OPTIONS.join(',')" @change="handleFileChange" />
      </section>

      <div class="resource-form-grid resource-form-grid--single">
        <UiInput v-model="uploadForm.label" :label="t('common.labels.name')" :placeholder="t('common.resource.optionalUploadLabel')" />
        <label class="resource-check-row resource-check-row--inline">
          <input v-model="uploadForm.public" type="checkbox" />
          <span>
            <strong>{{ t('common.resource.publicAttachment') }}</strong>
            <small>{{ t('common.resource.publicAttachmentHint') }}</small>
          </span>
        </label>
      </div>

      <dl class="resource-review-list">
        <div><dt>{{ t('common.labels.fileName') }}</dt><dd>{{ selectedFile?.name || t('common.placeholder') }}</dd></div>
        <div><dt>{{ t('common.resource.detectedType') }}</dt><dd>{{ selectedFileType }}</dd></div>
        <div><dt>{{ t('common.labels.size') }}</dt><dd>{{ selectedFile ? formatKilobytes(selectedFile.size) : t('common.placeholder') }}</dd></div>
      </dl>

      <footer class="resource-drawer-footer">
        <UiButton variant="secondary" :disabled="isUploading" @click="closeUploadDrawer">
          {{ t('common.actions.cancel') }}
        </UiButton>
        <UiButton :loading="isUploading" @click="uploadSelectedFile">
          <FileUp :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ isUploading ? t('common.states.uploading') : t('common.actions.upload') }}
        </UiButton>
      </footer>
    </PreviewDrawer>

    <PreviewDrawer
      :open="editDrawerOpen"
      :title="t('pages.attachments.editTitle')"
      :subtitle="t('pages.attachments.editDescription')"
      @close="closeEditDrawer"
    >
      <UiAlert
        v-if="formError"
        variant="error"
        :title="t('pages.attachments.updateErrorTitle')"
        :message="formError"
        dismissible
        @dismiss="formError = ''"
      />

      <template v-if="editTarget">
        <div class="resource-form-grid resource-form-grid--single">
          <UiInput v-model="editForm.label" :label="t('common.labels.name')" />
          <label class="resource-check-row resource-check-row--inline">
            <input v-model="editForm.public" type="checkbox" />
            <span>
              <strong>{{ t('common.labels.public') }}</strong>
              <small>{{ editTarget.public_path || t('pages.attachments.privateFileNoPublicPath') }}</small>
            </span>
          </label>
        </div>

        <dl class="resource-review-list">
          <div><dt>ID</dt><dd><code>{{ formatTechnicalId(editTarget.id) }}</code></dd></div>
          <div><dt>{{ t('common.labels.type') }}</dt><dd>{{ editTarget.type }}</dd></div>
          <div><dt>{{ t('common.labels.requiredVars') }}</dt><dd>{{ editTarget.required_vars?.length ? editTarget.required_vars.join(', ') : t('common.placeholder') }}</dd></div>
          <div><dt>{{ t('common.labels.uploadedAt') }}</dt><dd>{{ formatDateTime(editTarget.uploaded_at) }}</dd></div>
        </dl>

        <footer class="resource-drawer-footer">
          <UiButton
            v-if="canEditContent(editTarget)"
            variant="secondary"
            :disabled="isSaving"
            @click="openContentEditor(editTarget)"
          >
            <Edit3 :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('common.actions.editContent') }}
          </UiButton>
          <a class="ui-button ui-button--secondary" :href="editDownloadUrl" target="_blank" rel="noreferrer">
            <Download :size="16" stroke-width="1.8" aria-hidden="true" />
            {{ t('common.actions.download') }}
          </a>
          <UiButton variant="danger" :disabled="isSaving" @click="openDeleteDialog(editTarget)">
            {{ t('common.actions.delete') }}
          </UiButton>
          <UiButton :loading="isSaving" @click="saveAttachment">
            {{ isSaving ? t('common.states.saving') : t('common.actions.saveChanges') }}
          </UiButton>
        </footer>
      </template>
    </PreviewDrawer>

    <AttachmentPreviewDrawer
      :open="previewOpen"
      :attachment="previewAttachment"
      @close="closePreview"
    />

    <ConfirmDialog
      :open="deleteDialogOpen"
      :title="t('pages.attachments.delete.title')"
      :message="t('pages.attachments.delete.message')"
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
