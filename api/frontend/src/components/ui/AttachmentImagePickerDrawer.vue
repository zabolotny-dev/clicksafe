<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Copy, Eye, RefreshCw } from 'lucide-vue-next'
import AttachmentPreviewDrawer from './AttachmentPreviewDrawer.vue'
import EmptyState from './EmptyState.vue'
import IconButton from './IconButton.vue'
import PreviewDrawer from './PreviewDrawer.vue'
import SkeletonBlock from './SkeletonBlock.vue'
import UiAlert from './UiAlert.vue'
import UiButton from './UiButton.vue'
import UiInput from './UiInput.vue'
import { useNotifications } from '../../composables/useNotifications'
import { translateAttachmentVisibility } from '../../i18n'
import { getAttachmentUrl, listAttachments } from '../../resources/attachments'
import { isImageAttachmentType, normalizeAttachmentType } from '../../utils/attachmentTypes'
import { errorMessage, formatTechnicalId, nullableId } from '../../utils/resourceFormat'

const IMAGE_TYPES = ['.png', '.jpg', '.jpeg', '.gif', '.webp']
const PAGE_SIZE = 20
const CAMPAIGN_DOMAIN_TEMPLATE = '{{ .Campaign.Domain }}'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close'])
const { t } = useI18n()
const { notifySuccess, notifyWarning, notifyError } = useNotifications()

const searchQuery = ref('')
const typeFilter = ref('')
const attachments = ref([])
const total = ref(0)
const nextPage = ref(1)
const isInitialLoading = ref(false)
const isLoadingMore = ref(false)
const isComplete = ref(false)
const loadError = ref('')
const previewAttachment = ref(null)

let searchTimer = null
let requestSeq = 0
let controller = null

const loadedCount = computed(() => attachments.value.length)
const hasResults = computed(() => attachments.value.length > 0)
const canLoadMore = computed(() => !isComplete.value && !isInitialLoading.value && !isLoadingMore.value)

function normalizeId(value) {
  return nullableId(value) || (value ? String(value) : '')
}

function isSupportedImage(attachment) {
  return isImageAttachmentType(attachment?.type) && IMAGE_TYPES.includes(normalizeAttachmentType(attachment?.type))
}

function sortByUploadDate(items) {
  return [...items].sort((left, right) => (
    new Date(right.uploaded_at || 0).getTime() - new Date(left.uploaded_at || 0).getTime()
  ))
}

function dedupeAttachments(items) {
  const byId = new Map()
  items.forEach((attachment) => {
    const id = normalizeId(attachment?.id)
    if (id) {
      byId.set(id, attachment)
    }
  })

  return [...byId.values()]
}

function mergeAttachments(nextItems, reset) {
  const merged = reset
    ? nextItems
    : [...attachments.value, ...nextItems]

  attachments.value = sortByUploadDate(dedupeAttachments(merged))
}

function abortRequest() {
  if (controller) {
    controller.abort()
    controller = null
  }
}

function responseItems(response) {
  return Array.isArray(response?.items)
    ? response.items.filter(isSupportedImage)
    : []
}

function responseTotal(response, fallback) {
  return Number.isFinite(response?.total) ? response.total : fallback
}

async function loadImages(reset = false) {
  if (!props.open) {
    return
  }

  if (!reset && !canLoadMore.value) {
    return
  }

  if (reset) {
    abortRequest()
    attachments.value = []
    total.value = 0
    nextPage.value = 1
    isComplete.value = false
  }

  const page = reset ? 1 : nextPage.value
  const seq = ++requestSeq
  const requestController = new AbortController()
  controller = requestController
  loadError.value = ''

  if (reset) {
    isInitialLoading.value = true
  } else {
    isLoadingMore.value = true
  }

  try {
    const query = searchQuery.value.trim()
    let nextItems = []
    let nextTotal = 0

    if (typeFilter.value) {
      const response = await listAttachments({
        label: query,
        type: typeFilter.value,
        page,
        rows: PAGE_SIZE,
      }, { signal: requestController.signal })

      nextItems = responseItems(response)
      nextTotal = responseTotal(response, nextItems.length)
    } else {
      const responses = await Promise.all(IMAGE_TYPES.map((type) => listAttachments({
        label: query,
        type,
        page,
        rows: PAGE_SIZE,
      }, { signal: requestController.signal })))

      nextItems = responses.flatMap(responseItems)
      nextTotal = responses.reduce((sum, response) => (
        sum + responseTotal(response, responseItems(response).length)
      ), 0)
    }

    if (seq !== requestSeq) {
      return
    }

    mergeAttachments(nextItems, reset)
    total.value = nextTotal
    nextPage.value = page + 1
    isComplete.value = attachments.value.length >= nextTotal || nextItems.length === 0
  } catch (error) {
    if (error?.name === 'AbortError') {
      return
    }

    loadError.value = errorMessage(error, 'errors.attachmentsLoad')
    isComplete.value = true
  } finally {
    if (controller === requestController) {
      controller = null
    }

    if (seq === requestSeq) {
      isInitialLoading.value = false
      isLoadingMore.value = false
    }
  }
}

function scheduleSearch() {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }

  searchTimer = window.setTimeout(() => {
    loadImages(true)
  }, 300)
}

function refresh() {
  loadImages(true)
}

function handleScroll(event) {
  const target = event.currentTarget
  const distanceToBottom = target.scrollHeight - target.scrollTop - target.clientHeight

  if (distanceToBottom < 160) {
    loadImages(false)
  }
}

function closeDrawer() {
  emit('close')
}

function handleEscape(event) {
  if (event.key !== 'Escape' || !props.open) {
    return
  }

  if (previewAttachment.value) {
    previewAttachment.value = null
    return
  }

  closeDrawer()
}

function attachmentLabel(attachment) {
  return attachment?.label || formatTechnicalId(attachment?.id)
}

function thumbnailUrl(attachment) {
  return attachment?.id ? getAttachmentUrl(attachment.id) : ''
}

function imageTemplateUrl(attachment) {
  if (!attachment?.public || !attachment?.public_path) {
    return ''
  }

  const path = String(attachment.public_path).startsWith('/')
    ? attachment.public_path
    : `/${attachment.public_path}`

  return `${CAMPAIGN_DOMAIN_TEMPLATE}${path}`
}

async function copyImageUrl(attachment) {
  const imageUrl = imageTemplateUrl(attachment)
  if (!imageUrl) {
    notifyWarning(t('components.imageAssetDrawer.copyPrivateWarning'))
    return
  }

  if (typeof navigator === 'undefined' || !navigator.clipboard?.writeText) {
    notifyError(t('components.imageAssetDrawer.copyError'))
    return
  }

  try {
    await navigator.clipboard.writeText(imageUrl)
    notifySuccess(t('components.imageAssetDrawer.copySuccess'))
  } catch {
    notifyError(t('components.imageAssetDrawer.copyError'))
  }
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    document.addEventListener('keydown', handleEscape)
    loadImages(true)
    return
  }

  document.removeEventListener('keydown', handleEscape)
  abortRequest()
  previewAttachment.value = null
}, { immediate: true })

watch(typeFilter, () => {
  if (props.open) {
    loadImages(true)
  }
})

watch(searchQuery, () => {
  if (props.open) {
    scheduleSearch()
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleEscape)
  abortRequest()

  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }
})
</script>

<template>
  <PreviewDrawer
    :open="open"
    :title="t('components.imageAssetDrawer.title')"
    :subtitle="t('components.imageAssetDrawer.subtitle')"
    :resizable="false"
    :initial-width="500"
    :min-width="420"
    :max-width="560"
    @close="closeDrawer"
  >
    <section class="image-asset-drawer" :aria-label="t('components.imageAssetDrawer.title')">
      <div class="image-asset-drawer-toolbar">
        <UiInput
          v-model="searchQuery"
          :label="t('common.labels.search')"
          :placeholder="t('components.imageAssetDrawer.searchPlaceholder')"
        />
        <label class="ui-field">
          <span class="ui-field-label">{{ t('common.labels.type') }}</span>
          <select v-model="typeFilter" class="ui-select">
            <option value="">{{ t('components.imageAssetDrawer.typeAllImages') }}</option>
            <option v-for="type in IMAGE_TYPES" :key="type" :value="type">
              {{ type }}
            </option>
          </select>
        </label>
        <IconButton :label="t('common.actions.refresh')" :disabled="isInitialLoading || isLoadingMore" @click="refresh">
          <RefreshCw :size="16" stroke-width="1.8" aria-hidden="true" />
        </IconButton>
      </div>

      <UiAlert
        v-if="loadError"
        variant="error"
        :title="t('common.resource.attachmentsUnavailable')"
        :message="loadError"
      >
        <UiButton variant="secondary" @click="refresh">{{ t('common.actions.retry') }}</UiButton>
      </UiAlert>

      <div class="image-asset-drawer-summary">
        <span>{{ loadedCount }} / {{ total }}</span>
      </div>

      <div class="image-asset-drawer-list" @scroll="handleScroll">
        <template v-if="isInitialLoading">
          <article v-for="index in 5" :key="index" class="image-asset-row">
            <SkeletonBlock :rows="2" />
          </article>
        </template>

        <template v-else-if="hasResults">
          <article
            v-for="attachment in attachments"
            :key="attachment.id"
            class="image-asset-row"
          >
            <img
              class="image-asset-thumb"
              :src="thumbnailUrl(attachment)"
              :alt="attachmentLabel(attachment)"
              loading="lazy"
            />
            <div class="image-asset-meta">
              <strong :title="attachmentLabel(attachment)">{{ attachmentLabel(attachment) }}</strong>
              <span>{{ attachment.type || t('common.placeholder') }} / {{ translateAttachmentVisibility(attachment.public) }}</span>
              <code v-if="attachment.public_path">{{ attachment.public_path }}</code>
            </div>
            <div class="image-asset-actions">
              <IconButton :label="t('components.imageAssetDrawer.preview')" @click="previewAttachment = attachment">
                <Eye :size="16" stroke-width="1.8" aria-hidden="true" />
              </IconButton>
              <IconButton :label="t('components.imageAssetDrawer.copyUrl')" @click="copyImageUrl(attachment)">
                <Copy :size="16" stroke-width="1.8" aria-hidden="true" />
              </IconButton>
            </div>
          </article>
        </template>

        <EmptyState
          v-else
          :title="t('components.imageAssetDrawer.empty')"
          :description="t('common.resource.adjustAttachmentFilters')"
        />

        <div v-if="isLoadingMore" class="image-asset-loading-more">
          {{ t('components.imageAssetDrawer.loadingMore') }}
        </div>
      </div>
    </section>
  </PreviewDrawer>

  <AttachmentPreviewDrawer
    :open="Boolean(previewAttachment)"
    :attachment="previewAttachment"
    @close="previewAttachment = null"
  />
</template>
