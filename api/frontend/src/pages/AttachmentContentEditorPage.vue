<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import HtmlLiveEditor from '../components/ui/HtmlLiveEditor.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import { useNotifications } from '../composables/useNotifications'
import { useResourceActions } from '../composables/useResourceActions'
import {
  getAttachmentContent,
  listAttachments,
  updateAttachmentContent,
} from '../resources/attachments'
import {
  isHtmlAttachmentType,
  isTextAttachmentType,
  normalizeAttachmentType,
} from '../utils/attachmentTypes'
import {
  errorMessage,
  formatTechnicalId,
  nullableId,
} from '../utils/resourceFormat'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()

const attachment = ref(null)
const content = ref('')
const loadError = ref('')
const formError = ref('')
const isLoading = ref(false)
const isSaving = ref(false)

const attachmentId = computed(() => nullableId(route.params.id))
const attachmentType = computed(() => normalizeAttachmentType(attachment.value?.type))
const isHtml = computed(() => isHtmlAttachmentType(attachmentType.value))
const isText = computed(() => isTextAttachmentType(attachmentType.value))
const isEditable = computed(() => isHtml.value || isText.value)
const editorType = computed(() => (isHtml.value ? 'html' : 'txt'))
const returnTo = computed(() => {
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
  return redirect.startsWith('/') && !redirect.startsWith('//') ? redirect : '/templates/attachments'
})
const filename = computed(() => {
  const label = attachment.value?.label || formatTechnicalId(attachmentId.value)

  if (isHtml.value && !label.toLowerCase().endsWith('.html')) {
    return `${label}.html`
  }

  if (isText.value && !label.toLowerCase().endsWith('.txt')) {
    return `${label}.txt`
  }

  return label
})
const editorTitle = computed(() => (
  isHtml.value ? t('common.labels.htmlBody') : t('common.labels.textBody')
))
const editorDescription = computed(() => (
  isHtml.value ? t('common.preview.rawTemplate') : t('common.preview.text')
))
const filenamePlaceholder = computed(() => (
  isHtml.value ? 'message-body.html' : 'message-body.txt'
))

async function goBack() {
  await router.push(returnTo.value)
}

async function loadAttachmentContent() {
  if (!attachmentId.value) {
    loadError.value = t('common.preview.selectResource')
    return
  }

  isLoading.value = true
  loadError.value = ''
  formError.value = ''
  attachment.value = null
  content.value = ''

  try {
    const response = await listAttachments({ id: attachmentId.value, rows: 1 })
    attachment.value = Array.isArray(response?.items) ? response.items[0] : null

    if (!attachment.value?.id) {
      loadError.value = t('common.preview.selectResource')
      return
    }

    if (!isEditable.value) {
      loadError.value = t('common.preview.unsupported')
      return
    }

    const result = await getAttachmentContent(attachmentId.value)
    content.value = typeof result === 'string' ? result : ''
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    loadError.value = errorMessage(error, 'errors.attachmentsLoad')
  } finally {
    isLoading.value = false
  }
}

async function saveContent(nextContent) {
  if (!attachmentId.value || !isEditable.value) {
    return
  }

  if (!String(nextContent || '').trim()) {
    formError.value = t('common.resource.contentEmpty')
    return
  }

  if (!(await ensureCsrfToken())) {
    return
  }

  isSaving.value = true
  formError.value = ''

  try {
    const saved = await updateAttachmentContent(attachmentId.value, nextContent, mutationOptions())
    attachment.value = saved
    content.value = nextContent
    notifySuccess(t('pages.attachments.notifications.contentUpdated'))
    await goBack()
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.attachmentsLoad')
    notifyError(t('pages.attachments.notifications.contentUpdateFailedTitle'), formError.value)
  } finally {
    isSaving.value = false
  }
}

watch(attachmentId, () => {
  loadAttachmentContent()
})

onMounted(() => {
  loadAttachmentContent()
})
</script>

<template>
  <section class="resource-page">
    <SkeletonBlock v-if="isLoading" :rows="8" />

    <UiAlert
      v-else-if="loadError"
      variant="error"
      :title="t('common.preview.failed')"
      :message="loadError"
    >
      <UiButton variant="secondary" @click="goBack">{{ t('common.actions.back') }}</UiButton>
      <UiButton variant="secondary" @click="loadAttachmentContent">{{ t('common.actions.retry') }}</UiButton>
    </UiAlert>

    <section v-else class="resource-html-workspace">
      <UiAlert
        v-if="formError"
        variant="error"
        :title="t('pages.attachments.updateErrorTitle')"
        :message="formError"
        dismissible
        @dismiss="formError = ''"
      />

      <HtmlLiveEditor
        :filename="filename"
        workspace
        filename-readonly
        :initial-content="content"
        :content-type="editorType"
        :title="editorTitle"
        :description="editorDescription"
        :filename-placeholder="filenamePlaceholder"
        :save-label="t('common.actions.saveContent')"
        :saving="isSaving"
        :show-preview="isHtml"
        :show-image-action="isHtml"
        @save="saveContent"
        @cancel="goBack"
      />
    </section>
  </section>
</template>
