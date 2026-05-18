<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import SkeletonBlock from './SkeletonBlock.vue'
import { getAttachmentContent, getAttachmentUrl } from '../../resources/attachments'
import { nullableId } from '../../utils/resourceFormat'

const props = defineProps({
  attachment: {
    type: Object,
    default: null,
  },
  attachmentId: {
    type: [String, Object],
    default: '',
  },
  title: {
    type: String,
    default: '',
  },
  emptyText: {
    type: String,
    default: '',
  },
})

const { t } = useI18n()
const status = ref('idle')
const content = ref('')
const error = ref('')
let controller = null

const resolvedId = computed(() => nullableId(props.attachmentId) || nullableId(props.attachment?.id) || '')
const resolvedType = computed(() => String(props.attachment?.type || '').toLowerCase())
const isHtmlLike = computed(() => ['.html', 'html', 'text/html'].includes(resolvedType.value))
const isTextLike = computed(() => isHtmlLike.value || ['.txt', 'txt', 'text/plain'].includes(resolvedType.value))
const canLoad = computed(() => Boolean(resolvedId.value) && (!props.attachment || isTextLike.value))
const downloadUrl = computed(() => (resolvedId.value ? getAttachmentUrl(resolvedId.value) : ''))
const resolvedTitle = computed(() => props.title || t('common.preview.template'))
const resolvedEmptyText = computed(() => props.emptyText || t('common.preview.selectHtmlAsset'))

function abortRequest() {
  if (controller) {
    controller.abort()
    controller = null
  }
}

async function loadPreview() {
  abortRequest()
  content.value = ''
  error.value = ''

  if (!resolvedId.value) {
    status.value = 'empty'
    return
  }

  if (!canLoad.value) {
    status.value = 'unsupported'
    return
  }

  const requestController = new AbortController()
  controller = requestController
  status.value = 'loading'

  try {
    const result = await getAttachmentContent(resolvedId.value, {
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

watch(resolvedId, () => {
  loadPreview()
}, { immediate: true })

watch(resolvedType, () => {
  loadPreview()
})

onBeforeUnmount(() => {
  abortRequest()
})
</script>

<template>
  <section class="html-preview-panel">
    <div class="html-preview-frame">
      <SkeletonBlock v-if="status === 'loading'" :rows="5" />
      <iframe
        v-else-if="status === 'loaded' && isHtmlLike"
        :srcdoc="content"
        :title="resolvedTitle"
        sandbox=""
      />
      <pre v-else-if="status === 'loaded'">{{ content }}</pre>
      <div v-else-if="status === 'unsupported'" class="html-preview-empty">
        <h4>{{ t('common.preview.unavailable') }}</h4>
        <p>{{ t('common.preview.unsupported') }}</p>
        <a v-if="downloadUrl" class="action-link" :href="downloadUrl" target="_blank" rel="noreferrer">
          {{ t('common.actions.download') }}
        </a>
      </div>
      <div v-else-if="status === 'error'" class="html-preview-empty html-preview-empty--error">
        <h4>{{ t('common.preview.failed') }}</h4>
        <p>{{ error }}</p>
      </div>
      <div v-else class="html-preview-empty">
        <h4>{{ t('common.preview.noneSelected') }}</h4>
        <p>{{ resolvedEmptyText }}</p>
      </div>
    </div>
  </section>
</template>
