<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ChevronLeft,
  ChevronRight,
  Code,
  Copy,
  Download,
  Edit3,
  Eye,
  FileText,
  FileUp,
  Image,
  Plus,
  RefreshCw,
  Upload,
  Volume2,
} from 'lucide-vue-next'
import EmptyState from './EmptyState.vue'
import IconButton from './IconButton.vue'
import PreviewDrawer from './PreviewDrawer.vue'
import SkeletonBlock from './SkeletonBlock.vue'
import UiAlert from './UiAlert.vue'
import UiButton from './UiButton.vue'
import UiInput from './UiInput.vue'
import { useNotifications } from '../../composables/useNotifications'
import { useResourceActions } from '../../composables/useResourceActions'
import { translateAttachmentVisibility } from '../../i18n'
import {
  getAttachmentContent,
  getAttachmentUrl,
  listAttachments,
  updateAttachment,
  uploadAttachment,
} from '../../resources/attachments'
import {
  isAudioAttachmentType,
  isHtmlAttachmentType,
  isImageAttachmentType,
  isTextPreviewAttachmentType,
  SUPPORTED_ATTACHMENT_TYPES,
} from '../../utils/attachmentTypes'
import {
  errorMessage,
  formatDateTime,
  formatKilobytes,
  formatTechnicalId,
  nullableId,
} from '../../utils/resourceFormat'

const ROW_OPTIONS = [5, 10, 25, 50]

const props = defineProps({
  modelValue: {
    type: [String, Array],
    default: '',
  },
  selectionMode: {
    type: String,
    default: 'single',
    validator: (value) => ['single', 'multiple'].includes(value),
  },
  title: {
    type: String,
    default: '',
  },
  emptyText: {
    type: String,
    default: '',
  },
  allowedTypes: {
    type: Array,
    default: () => [],
  },
  defaultType: {
    type: String,
    default: '',
  },
  showPreviewAction: {
    type: Boolean,
    default: false,
  },
  showDuplicateAction: {
    type: Boolean,
    default: false,
  },
  showDownloadAction: {
    type: Boolean,
    default: false,
  },
  showCreateAction: {
    type: Boolean,
    default: false,
  },
  showUploadAction: {
    type: Boolean,
    default: false,
  },
  showMediaPreviewAction: {
    type: Boolean,
    default: false,
  },
  showEditContentAction: {
    type: Boolean,
    default: false,
  },
  showSelectedSummary: {
    type: Boolean,
    default: true,
  },
  showRequiredVarsColumn: {
    type: Boolean,
    default: true,
  },
  showHeaderClearAction: {
    type: Boolean,
    default: false,
  },
  restrictToAllowedTypes: {
    type: Boolean,
    default: false,
  },
  unsupportedTypeMessage: {
    type: String,
    default: '',
  },
  allTypesLabel: {
    type: String,
    default: '',
  },
  uploadSuccessTitle: {
    type: String,
    default: '',
  },
  uploadSuccessMessage: {
    type: String,
    default: '',
  },
})

const emit = defineEmits([
  'update:modelValue',
  'preview',
  'create',
  'selection-change',
  'edit-content',
  'uploaded',
  'duplicated',
])

const { t } = useI18n()
const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()

const pickerName = `attachment-picker-${Math.random().toString(36).slice(2)}`
const attachments = ref([])
const selectedAttachments = ref([])
const isLoading = ref(false)
const isUploading = ref(false)
const duplicatingId = ref('')
const loadError = ref('')
const formError = ref('')
const labelSearch = ref('')
const typeFilter = ref(props.defaultType || '')
const page = ref(1)
const rows = ref(10)
const total = ref(0)
const uploadDrawerOpen = ref(false)
const selectedFile = ref(null)
const fileInput = ref(null)
const isDragging = ref(false)

const uploadForm = reactive({
  label: '',
  public: false,
})

let searchTimer = null
let requestSeq = 0

const typeOptions = computed(() => {
  const values = props.allowedTypes.length ? props.allowedTypes : SUPPORTED_ATTACHMENT_TYPES
  return values.map((type) => String(type).toLowerCase())
})
const resolvedTitle = computed(() => props.title || t('nav.attachments'))
const resolvedEmptyText = computed(() => props.emptyText || t('common.empty.noMatchingRecords'))
const typeFilterAllowsAll = computed(() => !props.allowedTypes.length || props.allowedTypes.length > 1)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / rows.value)))
const skeletonCellCount = computed(() => (props.showRequiredVarsColumn ? 7 : 6))
const selectedIdList = computed(() => normalizeModelValue(props.modelValue))
const selectedIdSet = computed(() => new Set(selectedIdList.value))
const selectedCount = computed(() => selectedIdList.value.length)
const selectedSummary = computed(() => {
  if (!selectedIdList.value.length) {
    return t('common.selection.none')
  }

  const labels = selectedIdList.value.map((id) => {
    const attachment = selectedAttachments.value.find((item) => normalizeId(item?.id) === id)
    return attachment?.label || formatTechnicalId(id)
  })

  if (props.selectionMode === 'single') {
    return labels[0]
  }

  return labels.length > 3
    ? `${labels.slice(0, 3).join(', ')} +${labels.length - 3}`
    : labels.join(', ')
})
const createActionLabel = computed(() => (
  typeFilter.value === '.html' ? t('common.actions.create') : t('common.actions.create')
))
const selectedTitleLabel = computed(() => (
  props.selectionMode === 'multiple'
    ? t('common.resource.selectedAttachments')
    : t('common.resource.selectedAttachment')
))
const selectedFileType = computed(() => {
  if (!selectedFile.value) {
    return t('common.placeholder')
  }

  return selectedFileExtension(selectedFile.value) || t('common.labels.unknown')
})
const acceptedTypes = computed(() => typeOptions.value.join(','))
const allowedTypeText = computed(() => typeOptions.value.join(', '))
const allTypesLabel = computed(() => {
  if (props.allTypesLabel) {
    return props.allTypesLabel
  }

  return props.restrictToAllowedTypes && props.allowedTypes.length
    ? t('common.resource.allImageTypes')
    : t('common.resource.allSupportedTypes')
})

function normalizeId(value) {
  return nullableId(value) || (value ? String(value) : '')
}

function normalizeModelValue(value) {
  if (props.selectionMode === 'multiple') {
    return Array.isArray(value)
      ? [...new Set(value.map((id) => normalizeId(id)).filter(Boolean))]
      : []
  }

  const id = normalizeId(value)
  return id ? [id] : []
}

function isHtmlLike(attachment) {
  return isHtmlAttachmentType(attachment?.type)
}

function isPreviewable(attachment) {
  return isTextPreviewAttachmentType(attachment?.type)
}

function isMediaPreviewable(attachment) {
  return isImageAttachmentType(attachment?.type) || isAudioAttachmentType(attachment?.type)
}

function canPreview(attachment) {
  return isPreviewable(attachment) || (props.showMediaPreviewAction && isMediaPreviewable(attachment))
}

function canDuplicate(attachment) {
  return isPreviewable(attachment)
}

function canEditContent(attachment) {
  return isPreviewable(attachment)
}

function attachmentDownloadUrl(attachment) {
  return attachment?.id ? getAttachmentUrl(attachment.id) : ''
}

function requiredVarsText(vars) {
  return Array.isArray(vars) && vars.length ? vars.join(', ') : t('common.placeholder')
}

function attachmentIcon(attachment) {
  if (isMediaPreviewable(attachment)) {
    return isAudioAttachmentType(attachment?.type) ? Volume2 : Image
  }

  return isHtmlLike(attachment) ? Code : FileText
}

function emitSelectionChange(ids) {
  const byId = new Map([
    ...attachments.value,
    ...selectedAttachments.value,
  ].map((attachment) => [normalizeId(attachment?.id), attachment]))
  const selected = ids.map((id) => byId.get(id)).filter(Boolean)

  emit('selection-change', props.selectionMode === 'multiple' ? selected : selected[0] || null)
}

function updateSelection(ids) {
  const cleaned = [...new Set(ids.map((id) => normalizeId(id)).filter(Boolean))]

  if (props.selectionMode === 'multiple') {
    emit('update:modelValue', cleaned)
  } else {
    emit('update:modelValue', cleaned[0] || '')
  }

  emitSelectionChange(cleaned)
}

function toggleSelection(attachment) {
  const id = normalizeId(attachment?.id)
  if (!id) {
    return
  }

  if (props.selectionMode === 'single') {
    updateSelection([id])
    return
  }

  if (selectedIdSet.value.has(id)) {
    updateSelection(selectedIdList.value.filter((item) => item !== id))
  } else {
    updateSelection([...selectedIdList.value, id])
  }
}

function selectAttachment(attachment) {
  const id = normalizeId(attachment?.id)
  if (!id) {
    return
  }

  const byId = new Map(selectedAttachments.value.map((item) => [normalizeId(item?.id), item]))
  byId.set(id, attachment)
  selectedAttachments.value = selectedIdList.value
    .filter((item) => item !== id)
    .map((item) => byId.get(item))
    .filter(Boolean)
    .concat(attachment)

  if (props.selectionMode === 'multiple') {
    updateSelection([...selectedIdList.value, id])
  } else {
    updateSelection([id])
  }
}

function clearSelection() {
  updateSelection([])
}

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

function openFileDialog() {
  fileInput.value?.click()
}

function selectedFileExtension(file) {
  if (!file?.name) {
    return ''
  }

  const name = String(file.name)
  const extensionIndex = name.lastIndexOf('.')
  return extensionIndex === -1 ? '' : name.slice(extensionIndex).toLowerCase()
}

function uploadTypeError() {
  if (!props.allowedTypes.length) {
    return ''
  }

  const extension = selectedFileExtension(selectedFile.value)
  const allowed = new Set(typeOptions.value)

  if (extension && allowed.has(extension)) {
    return ''
  }

  return props.unsupportedTypeMessage || t('common.resource.unsupportedType', { types: allowedTypeText.value })
}

function handleDrop(event) {
  event.preventDefault()
  isDragging.value = false
  selectedFile.value = event.dataTransfer?.files?.[0] || null
}

function setPage(nextPage) {
  page.value = Math.max(1, Math.min(totalPages.value, nextPage))
  loadAttachments()
}

function changeRows() {
  page.value = 1
  loadAttachments()
}

function changeTypeFilter() {
  page.value = 1
  loadAttachments()
}

function filterAllowedAttachments(items) {
  if (!props.restrictToAllowedTypes || !props.allowedTypes.length) {
    return items
  }

  const allowed = new Set(typeOptions.value)
  return items.filter((attachment) => allowed.has(String(attachment?.type || '').toLowerCase()))
}

function syncSelectedFromPage() {
  if (!selectedIdList.value.length) {
    selectedAttachments.value = []
    return
  }

  const known = new Map(selectedAttachments.value.map((attachment) => [normalizeId(attachment?.id), attachment]))
  attachments.value.forEach((attachment) => known.set(normalizeId(attachment?.id), attachment))
  selectedAttachments.value = selectedIdList.value.map((id) => known.get(id)).filter(Boolean)
}

async function loadSelectedMetadata() {
  if (!selectedIdList.value.length) {
    selectedAttachments.value = []
    return
  }

  const known = new Map([
    ...selectedAttachments.value,
    ...attachments.value,
  ].map((attachment) => [normalizeId(attachment?.id), attachment]))
  const missing = selectedIdList.value.filter((id) => !known.has(id))

  if (missing.length) {
    try {
      const responses = await Promise.all(missing.map((id) => listAttachments({ id, rows: 1 })))
      responses.forEach((response) => {
        const attachment = Array.isArray(response?.items) ? response.items[0] : null
        if (attachment?.id) {
          known.set(normalizeId(attachment.id), attachment)
        }
      })
    } catch (error) {
      if (await handleAuthError(error)) {
        return
      }

      loadError.value = errorMessage(error, 'errors.attachmentsLoad')
    }
  }

  selectedAttachments.value = selectedIdList.value.map((id) => known.get(id)).filter(Boolean)
  emitSelectionChange(selectedIdList.value)
}

async function loadAttachments() {
  const seq = ++requestSeq
  isLoading.value = true
  loadError.value = ''

  try {
    if (props.restrictToAllowedTypes && props.allowedTypes.length && !typeFilter.value) {
      await loadRestrictedAttachments(seq)
      return
    }

    const response = await listAttachments({
      label: labelSearch.value,
      type: typeFilter.value,
      page: page.value,
      rows: rows.value,
    })

    if (seq !== requestSeq) {
      return
    }

    const items = filterAllowedAttachments(Array.isArray(response?.items) ? response.items : [])
    const nextTotal = Number.isFinite(response?.total) ? response.total : items.length
    const nextPage = Number.isFinite(response?.page) ? response.page : page.value
    const nextRows = Number.isFinite(response?.rowsPerPage) ? response.rowsPerPage : rows.value

    attachments.value = items
    total.value = nextTotal
    page.value = nextPage
    rows.value = nextRows

    if (nextTotal > 0 && !items.length && page.value > 1) {
      page.value = Math.max(1, Math.ceil(nextTotal / rows.value))
      await loadAttachments()
      return
    }

    syncSelectedFromPage()
    await loadSelectedMetadata()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    loadError.value = errorMessage(error, 'errors.attachmentsLoad')
  } finally {
    if (seq === requestSeq) {
      isLoading.value = false
    }
  }
}

async function loadRestrictedAttachments(seq) {
  async function fetchAllForType(type) {
    const items = []
    let currentPage = 1
    let fetchedForType = 0
    let totalForType = 0

    do {
      const response = await listAttachments({
        label: labelSearch.value,
        type,
        page: currentPage,
        rows: 100,
      })
      const pageItems = Array.isArray(response?.items) ? response.items : []
      totalForType = Number.isFinite(response?.total) ? response.total : fetchedForType + pageItems.length
      fetchedForType += pageItems.length
      items.push(...pageItems)
      currentPage += 1
    } while (fetchedForType < totalForType && currentPage < 100)

    return items
  }

  const results = await Promise.all(typeOptions.value.map(fetchAllForType))
  const fetched = results.flat()

  if (seq !== requestSeq) {
    return
  }

  const byId = new Map()
  filterAllowedAttachments(fetched).forEach((attachment) => {
    if (attachment?.id) {
      byId.set(normalizeId(attachment.id), attachment)
    }
  })

  const allItems = [...byId.values()].sort((a, b) => (
    new Date(b.uploaded_at || 0).getTime() - new Date(a.uploaded_at || 0).getTime()
  ))
  const nextRows = rows.value
  const nextTotal = allItems.length
  const maxPage = Math.max(1, Math.ceil(nextTotal / nextRows))
  const nextPage = Math.min(page.value, maxPage)
  const start = (nextPage - 1) * nextRows

  attachments.value = allItems.slice(start, start + nextRows)
  total.value = nextTotal
  page.value = nextPage

  syncSelectedFromPage()
  await loadSelectedMetadata()
}

function refresh() {
  loadAttachments()
}

function duplicateFileName(attachment) {
  const extension = String(attachment?.type || '.html').startsWith('.')
    ? String(attachment.type)
    : '.html'
  const base = (attachment?.label || 'attachment')
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9-]+/g, '-')
    .replace(/^-|-$/g, '') || 'attachment'

  return `${base}-copy-${Date.now()}${extension}`
}

async function duplicateAttachment(attachment) {
  if (!canDuplicate(attachment)) {
    notifyError(t('common.actions.duplicate'), t('common.preview.unavailable'))
    return
  }

  if (!(await ensureCsrfToken())) {
    return
  }

  duplicatingId.value = normalizeId(attachment?.id)

  try {
    const content = await getAttachmentContent(attachment.id)
    const type = isHtmlLike(attachment) ? 'text/html' : 'text/plain'
    const file = new File([typeof content === 'string' ? content : ''], duplicateFileName(attachment), { type })
    const uploaded = await uploadAttachment(file, { public: false }, mutationOptions())

    selectAttachment(uploaded)
    await loadAttachments()
    emit('duplicated', uploaded)
    notifySuccess(t('common.actions.duplicate'), t('common.resource.selectedAttachment'))
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    notifyError(t('common.actions.duplicate'), errorMessage(error, 'errors.attachmentsLoad'))
  } finally {
    duplicatingId.value = ''
  }
}

async function uploadSelectedFile() {
  if (!selectedFile.value) {
    formError.value = t('common.resource.dropOneFile')
    return
  }

  const typeError = uploadTypeError()
  if (typeError) {
    formError.value = typeError
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

    selectAttachment(uploaded)
    uploadDrawerOpen.value = false
    resetUploadForm()
    await loadAttachments()
    emit('uploaded', uploaded)
    notifySuccess(
      props.uploadSuccessTitle || t('common.actions.upload'),
      props.uploadSuccessMessage || t('common.resource.selectedAttachment'),
    )
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.attachmentsLoad')
    notifyError(t('common.actions.upload'), formError.value)
  } finally {
    isUploading.value = false
  }
}

watch(labelSearch, () => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }

  searchTimer = window.setTimeout(() => {
    page.value = 1
    loadAttachments()
  }, 350)
})

watch(() => props.defaultType, (value) => {
  typeFilter.value = value || ''
  page.value = 1
  loadAttachments()
})

watch(() => selectedIdList.value.join('|'), () => {
  loadSelectedMetadata()
})

onMounted(() => {
  typeFilter.value = props.defaultType || (props.allowedTypes.length === 1 ? String(props.allowedTypes[0]).toLowerCase() : '')

  loadAttachments()
})

onBeforeUnmount(() => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }
})

defineExpose({
  refresh,
})
</script>

<template>
  <section class="attachment-picker-panel" :aria-label="resolvedTitle">
    <header class="attachment-picker-header">
      <div>
        <h3>{{ resolvedTitle }}</h3>
        <p>{{ t('common.resource.searchAttachments') }}</p>
      </div>
      <div class="attachment-picker-header-actions">
        <slot name="actions" />
        <UiButton v-if="showCreateAction" variant="secondary" @click="emit('create')">
          <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ createActionLabel || t('common.actions.create') }}
        </UiButton>
        <UiButton v-if="showUploadAction" variant="secondary" @click="openUploadDrawer">
          <Upload :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ t('common.actions.upload') }}
        </UiButton>
        <UiButton
          v-if="showHeaderClearAction && selectedCount"
          variant="ghost"
          :disabled="isLoading"
          @click="clearSelection"
        >
          {{ t('common.actions.clear') }}
        </UiButton>
      </div>
    </header>

    <section class="attachment-picker-toolbar">
      <UiInput v-model="labelSearch" :label="t('common.labels.search')" :placeholder="t('common.resource.searchByLabel')" />
      <label class="ui-field">
        <span class="ui-field-label">{{ t('common.labels.type') }}</span>
        <select v-model="typeFilter" class="ui-select" @change="changeTypeFilter">
          <option v-if="typeFilterAllowsAll" value="">{{ allTypesLabel }}</option>
          <option v-for="type in typeOptions" :key="type" :value="type">
            {{ type }}
          </option>
        </select>
      </label>
      <IconButton :label="t('common.resource.refreshAttachments')" :disabled="isLoading" @click="refresh">
        <RefreshCw :size="16" stroke-width="1.8" aria-hidden="true" />
      </IconButton>
    </section>

    <section v-if="showSelectedSummary" class="attachment-picker-selection-bar">
      <div>
        <span>{{ selectedTitleLabel }}</span>
        <strong>{{ selectedSummary }}</strong>
      </div>
      <UiButton
        v-if="selectedCount"
        variant="ghost"
        :disabled="isLoading"
        @click="clearSelection"
      >
        {{ t('common.actions.clear') }}
      </UiButton>
    </section>

    <UiAlert
      v-if="loadError"
      variant="error"
      :title="t('common.resource.attachmentsUnavailable')"
      :message="loadError"
    >
      <UiButton variant="secondary" @click="refresh">{{ t('common.actions.retry') }}</UiButton>
    </UiAlert>

    <div class="attachment-picker-table" role="table" :aria-label="resolvedTitle">
      <div
        class="attachment-picker-row attachment-picker-head"
        :class="{ 'attachment-picker-row--no-vars': !showRequiredVarsColumn }"
        role="row"
      >
        <span>{{ t('common.actions.select') }}</span>
        <span>{{ t('common.labels.name') }}</span>
        <span>{{ t('common.labels.type') }}</span>
        <span v-if="showRequiredVarsColumn">{{ t('common.labels.requiredVars') }}</span>
        <span>{{ t('common.labels.public') }}</span>
        <span>{{ t('common.labels.uploadedAt') }}</span>
        <span>{{ t('pages.campaigns.columns.actions') }}</span>
      </div>

      <template v-if="isLoading">
        <div
          v-for="index in 5"
          :key="index"
          class="attachment-picker-row"
          :class="{ 'attachment-picker-row--no-vars': !showRequiredVarsColumn }"
          role="row"
        >
          <span v-for="cell in skeletonCellCount" :key="cell"><SkeletonBlock :rows="1" /></span>
        </div>
      </template>

      <template v-else-if="attachments.length">
        <div
          v-for="attachment in attachments"
          :key="attachment.id"
          class="attachment-picker-row"
          :class="{
            'is-selected': selectedIdSet.has(normalizeId(attachment.id)),
            'attachment-picker-row--no-vars': !showRequiredVarsColumn,
          }"
          role="row"
          @click="toggleSelection(attachment)"
        >
          <span>
            <input
              :name="pickerName"
              :type="selectionMode === 'multiple' ? 'checkbox' : 'radio'"
              :checked="selectedIdSet.has(normalizeId(attachment.id))"
              :aria-label="`${t('common.actions.select')} ${attachment.label}`"
              @click.stop
              @change="toggleSelection(attachment)"
            />
          </span>
          <span class="attachment-picker-label-cell">
            <span class="attachment-picker-file-icon" aria-hidden="true">
              <component :is="attachmentIcon(attachment)" :size="16" stroke-width="1.8" />
            </span>
            <strong :title="attachment.label">{{ attachment.label || formatTechnicalId(attachment.id) }}</strong>
          </span>
          <span>{{ attachment.type || t('common.placeholder') }}</span>
          <span v-if="showRequiredVarsColumn" :title="requiredVarsText(attachment.required_vars)">
            {{ requiredVarsText(attachment.required_vars) }}
          </span>
          <span>{{ translateAttachmentVisibility(attachment.public) }}</span>
          <span>{{ formatDateTime(attachment.uploaded_at) }}</span>
          <span class="attachment-picker-actions" @click.stop>
            <IconButton
              v-if="showPreviewAction && canPreview(attachment)"
              :label="t('common.actions.preview')"
              @click="emit('preview', attachment)"
            >
              <Eye :size="16" stroke-width="1.8" aria-hidden="true" />
            </IconButton>
            <IconButton
              v-if="showDuplicateAction"
              :label="t('common.actions.duplicate')"
              :disabled="!canDuplicate(attachment) || duplicatingId === normalizeId(attachment.id)"
              @click="duplicateAttachment(attachment)"
            >
              <Copy :size="16" stroke-width="1.8" aria-hidden="true" />
            </IconButton>
            <IconButton
              v-if="showEditContentAction && canEditContent(attachment)"
              :label="t('common.actions.editContent')"
              @click="emit('edit-content', attachment)"
            >
              <Edit3 :size="16" stroke-width="1.8" aria-hidden="true" />
            </IconButton>
            <a
              v-if="showDownloadAction"
              class="icon-button icon-button--ghost"
              :href="attachmentDownloadUrl(attachment)"
              target="_blank"
              rel="noreferrer"
              :aria-label="t('common.actions.download')"
              :title="t('common.actions.download')"
            >
              <Download :size="16" stroke-width="1.8" aria-hidden="true" />
            </a>
          </span>
        </div>
      </template>

      <EmptyState
        v-else
        :title="resolvedEmptyText"
        :description="t('common.resource.adjustAttachmentFilters')"
      />
    </div>

    <footer class="attachment-picker-footer">
      <span>{{ selectedCount }} / {{ total }}</span>
      <div class="attachment-picker-pagination">
        <IconButton :label="t('common.pagination.previousPage')" :disabled="page <= 1 || isLoading" @click="setPage(page - 1)">
          <ChevronLeft :size="16" stroke-width="1.8" aria-hidden="true" />
        </IconButton>
        <strong>{{ t('common.pagination.pageOf', { page, total: totalPages }) }}</strong>
        <IconButton :label="t('common.pagination.nextPage')" :disabled="page >= totalPages || isLoading" @click="setPage(page + 1)">
          <ChevronRight :size="16" stroke-width="1.8" aria-hidden="true" />
        </IconButton>
        <label class="attachment-picker-rows">
          <span>{{ t('common.labels.rows') }}</span>
          <select v-model.number="rows" class="ui-select" @change="changeRows">
            <option v-for="option in ROW_OPTIONS" :key="option" :value="option">
              {{ option }}
            </option>
          </select>
        </label>
      </div>
    </footer>

    <PreviewDrawer
      :open="uploadDrawerOpen"
      :title="t('common.actions.upload')"
      :subtitle="t('common.resource.singleFileUpload')"
      @close="closeUploadDrawer"
    >
      <UiAlert
        v-if="formError"
        variant="error"
        :title="t('common.resource.attachmentsUnavailable')"
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
        <Upload :size="28" stroke-width="1.6" aria-hidden="true" />
        <h3>{{ selectedFile ? selectedFile.name : t('common.resource.dropOneFile') }}</h3>
        <p>{{ t('common.resource.singleFileUpload') }}</p>
        <UiButton variant="secondary" type="button" @click="openFileDialog">
          <FileUp :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ t('common.actions.select') }}
        </UiButton>
        <input
          ref="fileInput"
          class="ui-file-input ui-file-input--hidden"
          type="file"
          :accept="acceptedTypes"
          @change="handleFileChange"
        />
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
  </section>
</template>
