<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Download } from 'lucide-vue-next'
import PreviewDrawer from './PreviewDrawer.vue'
import SkeletonBlock from './SkeletonBlock.vue'
import WaveformPlayer from './WaveformPlayer.vue'
import { getAttachmentContent, getAttachmentUrl } from '../../resources/attachments'
import {
  isAudioAttachmentType,
  isHtmlAttachmentType,
  isImageAttachmentType,
  isTextAttachmentType,
  isTextPreviewAttachmentType,
  normalizeAttachmentType,
} from '../../utils/attachmentTypes'
import { nullableId } from '../../utils/resourceFormat'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  attachment: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['close'])
const { t } = useI18n()

const status = ref('idle')
const content = ref('')
const error = ref('')
let controller = null

const attachmentId = computed(() => nullableId(props.attachment?.id) || '')
const attachmentType = computed(() => normalizeAttachmentType(props.attachment?.type))
const title = computed(() => props.attachment?.label || t('common.preview.attachment'))
const subtitle = computed(() => props.attachment?.type || '')
const downloadUrl = computed(() => (attachmentId.value ? getAttachmentUrl(attachmentId.value) : ''))
const isAudio = computed(() => isAudioAttachmentType(attachmentType.value))
const isImage = computed(() => isImageAttachmentType(attachmentType.value))
const isHtml = computed(() => isHtmlAttachmentType(attachmentType.value))
const isText = computed(() => isTextAttachmentType(attachmentType.value))
const canLoadText = computed(() => isTextPreviewAttachmentType(attachmentType.value))
const previewLabel = computed(() => {
  if (isImage.value) {
    return t('common.preview.image')
  }

  if (isAudio.value) {
    return t('common.preview.audio')
  }

  if (isText.value) {
    return t('common.preview.text')
  }

  if (isHtml.value) {
    return t('common.preview.rawTemplate')
  }

  return t('common.preview.unavailable')
})

function abortRequest() {
  if (controller) {
    controller.abort()
    controller = null
  }
}

async function loadContent() {
  abortRequest()
  content.value = ''
  error.value = ''

  if (!props.open || !attachmentId.value) {
    status.value = 'empty'
    return
  }

  if (isImage.value || isAudio.value) {
    status.value = 'loaded'
    return
  }

  if (!canLoadText.value) {
    status.value = 'unsupported'
    return
  }

  const requestController = new AbortController()
  controller = requestController
  status.value = 'loading'

  try {
    const result = await getAttachmentContent(attachmentId.value, {
      signal: requestController.signal,
    })

    content.value = typeof result === 'string' ? result : ''
    status.value = content.value ? 'loaded' : 'empty'
  } catch (err) {
    if (err?.name === 'AbortError') {
      return
    }

    error.value = err?.message || t('common.preview.failed')
    status.value = 'error'
  } finally {
    if (controller === requestController) {
      controller = null
    }
  }
}

watch([
  () => props.open,
  attachmentId,
  attachmentType,
], () => {
  loadContent()
}, { immediate: true })

onBeforeUnmount(() => {
  abortRequest()
})
</script>

<template>
  <PreviewDrawer
    :open="open"
    :title="title"
    :subtitle="subtitle"
    @close="emit('close')"
  >
    <template #actions>
      <a
        v-if="downloadUrl"
        class="ui-button ui-button--secondary"
        :href="downloadUrl"
        target="_blank"
        rel="noreferrer"
      >
        <Download :size="16" stroke-width="1.8" aria-hidden="true" />
        {{ t('common.actions.downloadFile') }}
      </a>
    </template>

    <section class="attachment-preview" :aria-label="previewLabel">
      <SkeletonBlock v-if="status === 'loading'" :rows="6" />

      <img
        v-else-if="status === 'loaded' && isImage"
        class="attachment-preview-image"
        :src="downloadUrl"
        :alt="t('common.preview.image')"
      />

      <WaveformPlayer
        v-else-if="status === 'loaded' && isAudio"
        :src="downloadUrl"
      />

      <template v-else-if="status === 'loaded' && isHtml">
        <p class="attachment-preview-note">{{ t('common.preview.rawTemplate') }}</p>
        <iframe
          class="attachment-preview-frame"
          :srcdoc="content"
          :title="title"
          sandbox=""
        />
      </template>

      <pre v-else-if="status === 'loaded' && isText" class="attachment-preview-text">{{ content }}</pre>

      <div v-else-if="status === 'error'" class="attachment-preview-empty attachment-preview-empty--error">
        <h4>{{ t('common.preview.failed') }}</h4>
        <p>{{ error }}</p>
      </div>

      <div v-else-if="status === 'unsupported'" class="attachment-preview-empty">
        <h4>{{ t('common.preview.unavailable') }}</h4>
        <p>{{ t('common.preview.unsupported') }}</p>
        <a
          v-if="downloadUrl"
          class="ui-button ui-button--secondary"
          :href="downloadUrl"
          target="_blank"
          rel="noreferrer"
        >
          <Download :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ t('common.actions.downloadFile') }}
        </a>
      </div>

      <div v-else class="attachment-preview-empty">
        <h4>{{ t('common.preview.noneSelected') }}</h4>
        <p>{{ t('common.preview.selectResource') }}</p>
      </div>
    </section>
  </PreviewDrawer>
</template>
