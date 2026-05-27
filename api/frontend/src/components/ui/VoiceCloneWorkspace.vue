<script setup>
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import {
  ArrowLeft,
  RefreshCw,
  Save,
  Settings,
  Upload,
  Volume2,
} from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import UiAlert from './UiAlert.vue'
import UiButton from './UiButton.vue'
import UiInput from './UiInput.vue'
import { useNotifications } from '../../composables/useNotifications'
import { useResourceActions } from '../../composables/useResourceActions'
import { uploadAttachment } from '../../resources/attachments'
import { cloneVoice } from '../../resources/voice'
import {
  errorMessage,
  formatKilobytes,
} from '../../utils/resourceFormat'

const ACCEPTED_AUDIO_TYPES = '.wav,.mp3,.ogg,.opus,.flac,.m4a,.aac,.weba,.webm,audio/*'

const props = defineProps({
  status: {
    type: Object,
    default: () => ({}),
  },
  initialLabel: {
    type: String,
    default: '',
  },
  statusLoading: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['cancel', 'saved', 'refresh-status'])

const { t } = useI18n()
const { notifySuccess, notifyError } = useNotifications()
const { mutationOptions, ensureCsrfToken, handleAuthError } = useResourceActions()

const form = reactive({
  label: '',
  text: '',
  language: 'auto',
  instruct: '',
  speed: 1,
  durationSeconds: '',
  inferenceSteps: 32,
  guidanceScale: 2,
  denoise: true,
  preprocessPrompt: true,
  postprocessOutput: true,
})

const referenceFile = ref(null)
const referenceUrl = ref('')
const result = ref(null)
const resultUrl = ref('')
const formError = ref('')
const isDragging = ref(false)
const isGenerating = ref(false)
const isSaving = ref(false)
const fileInput = ref(null)
const savedAttachment = ref(null)

const languageOptions = computed(() => [
  { value: 'auto', label: t('pages.emailTemplates.voiceClone.languages.auto') },
  { value: 'ru', label: t('pages.emailTemplates.voiceClone.languages.ru') },
  { value: 'en', label: t('pages.emailTemplates.voiceClone.languages.en') },
])

const statusText = computed(() => {
  if (props.statusLoading) {
    return t('pages.emailTemplates.voiceClone.status.checking')
  }

  if (!props.status?.available) {
    return props.status?.message || t('pages.emailTemplates.voiceClone.status.unavailable')
  }

  return [
    props.status?.model_provider,
    props.status?.model_id,
    props.status?.device,
  ].filter(Boolean).join(' · ') || t('pages.emailTemplates.voiceClone.status.available')
})

const canGenerate = computed(() => (
  Boolean(referenceFile.value)
    && Boolean(form.text.trim())
    && props.status?.available
    && !isGenerating.value
    && !isSaving.value
))
const canSave = computed(() => Boolean(result.value?.blob) && !isSaving.value && !isGenerating.value)
const resultDurationText = computed(() => {
  if (!result.value?.durationMs) {
    return t('common.placeholder')
  }

  return `${(result.value.durationMs / 1000).toFixed(1)} ${t('pages.emailTemplates.voiceClone.secondsShort')}`
})
const resultSizeText = computed(() => (
  result.value?.totalBytes ? formatKilobytes(result.value.totalBytes) : t('common.placeholder')
))

function defaultLabel() {
  const base = props.initialLabel?.trim()
  return base
    ? t('pages.emailTemplates.voiceClone.defaultLabelFromTemplate', { label: base })
    : t('pages.emailTemplates.voiceClone.defaultLabel')
}

function ensureDefaultLabel() {
  if (!form.label.trim()) {
    form.label = defaultLabel()
  }
}

function revokeUrl(url) {
  if (url) {
    URL.revokeObjectURL(url)
  }
}

function resetResult() {
  revokeUrl(resultUrl.value)
  resultUrl.value = ''
  result.value = null
  savedAttachment.value = null
}

function updateReferenceFile(file) {
  if (!file) {
    referenceFile.value = null
    revokeUrl(referenceUrl.value)
    referenceUrl.value = ''
    return
  }

  referenceFile.value = file
  revokeUrl(referenceUrl.value)
  referenceUrl.value = URL.createObjectURL(file)
  resetResult()
}

function handleFileChange(event) {
  updateReferenceFile(event.target.files?.[0] || null)
}

function handleDrop(event) {
  event.preventDefault()
  isDragging.value = false
  updateReferenceFile(event.dataTransfer?.files?.[0] || null)
}

function openFileDialog() {
  fileInput.value?.click()
}

function cleanExtension(extension) {
  const raw = String(extension || '.ogg').trim().toLowerCase()
  return raw.startsWith('.') ? raw : `.${raw}`
}

function outputFilename() {
  const extension = cleanExtension(result.value?.extension)
  const label = (form.label || defaultLabel())
    .trim()
    .replace(/[<>:"/\\|?*\u0000-\u001F]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .slice(0, 96)
    || 'Voice Clone'

  return label.toLowerCase().endsWith(extension) ? label : `${label}${extension}`
}

function appendIfPresent(formData, field, value) {
  const text = String(value ?? '').trim()
  if (text) {
    formData.append(field, text)
  }
}

function buildRequestFormData() {
  const formData = new FormData()
  formData.append('reference_audio', referenceFile.value)
  formData.append('text', form.text.trim())
  formData.append('language', form.language || 'auto')
  formData.append('speed', String(form.speed))
  formData.append('inference_steps', String(form.inferenceSteps))
  formData.append('guidance_scale', String(form.guidanceScale))
  formData.append('denoise', String(Boolean(form.denoise)))
  formData.append('preprocess_prompt', String(Boolean(form.preprocessPrompt)))
  formData.append('postprocess_output', String(Boolean(form.postprocessOutput)))
  appendIfPresent(formData, 'duration_seconds', form.durationSeconds)
  appendIfPresent(formData, 'instruct', form.instruct)
  return formData
}

async function generateVoice() {
  ensureDefaultLabel()

  if (!referenceFile.value) {
    formError.value = t('pages.emailTemplates.voiceClone.validation.referenceAudioRequired')
    return
  }

  if (!form.text.trim()) {
    formError.value = t('pages.emailTemplates.voiceClone.validation.textRequired')
    return
  }

  if (!(await ensureCsrfToken())) {
    return
  }

  isGenerating.value = true
  formError.value = ''
  resetResult()

  try {
    const generated = await cloneVoice(buildRequestFormData(), mutationOptions())
    result.value = generated
    resultUrl.value = URL.createObjectURL(generated.blob)
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'pages.emailTemplates.voiceClone.errors.generateFailed')
    notifyError(t('pages.emailTemplates.voiceClone.generateFailedTitle'), formError.value)
  } finally {
    isGenerating.value = false
  }
}

async function saveResult() {
  if (!result.value?.blob) {
    return
  }

  ensureDefaultLabel()

  if (!(await ensureCsrfToken())) {
    return
  }

  isSaving.value = true
  formError.value = ''

  try {
    const file = new File([result.value.blob], outputFilename(), { type: result.value.mimeType || 'audio/ogg' })
    const uploaded = await uploadAttachment(file, { public: false }, mutationOptions())
    savedAttachment.value = uploaded
    notifySuccess(t('pages.emailTemplates.voiceClone.savedTitle'), t('pages.emailTemplates.voiceClone.savedMessage'))
    emit('saved', uploaded)
  } catch (error) {
    if (await handleAuthError(error)) {
      return
    }

    formError.value = errorMessage(error, 'errors.attachmentsLoad')
    notifyError(t('pages.emailTemplates.voiceClone.saveFailedTitle'), formError.value)
  } finally {
    isSaving.value = false
  }
}

watch(() => props.initialLabel, ensureDefaultLabel, { immediate: true })

onBeforeUnmount(() => {
  revokeUrl(referenceUrl.value)
  revokeUrl(resultUrl.value)
})
</script>

<template>
  <section class="voice-clone-workspace" :aria-label="t('pages.emailTemplates.voiceClone.title')">
    <header class="voice-clone-header">
      <div>
        <p>{{ t('pages.emailTemplates.voiceClone.eyebrow') }}</p>
        <h2>{{ t('pages.emailTemplates.voiceClone.title') }}</h2>
        <span>{{ t('pages.emailTemplates.voiceClone.description') }}</span>
      </div>
      <div class="voice-clone-header-actions">
        <UiButton variant="ghost" :disabled="isGenerating || isSaving" @click="emit('refresh-status')">
          <RefreshCw :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ t('pages.emailTemplates.voiceClone.refreshStatus') }}
        </UiButton>
        <UiButton variant="secondary" :disabled="isGenerating || isSaving" @click="emit('cancel')">
          <ArrowLeft :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ t('common.actions.back') }}
        </UiButton>
      </div>
    </header>

    <UiAlert
      v-if="formError"
      variant="error"
      :title="t('pages.emailTemplates.formAttention')"
      :message="formError"
      dismissible
      @dismiss="formError = ''"
    />

    <section class="voice-clone-grid">
      <div class="voice-clone-main">
        <section class="voice-clone-panel voice-clone-panel--text">
          <header class="voice-clone-panel-header">
            <div>
              <Volume2 :size="17" stroke-width="1.8" aria-hidden="true" />
              <h3>{{ t('pages.emailTemplates.voiceClone.textTitle') }}</h3>
            </div>
            <span class="voice-clone-status-pill" :class="{ 'is-unavailable': !props.status?.available }">
              {{ statusText }}
            </span>
          </header>

          <label class="ui-field">
            <span class="ui-field-label">{{ t('pages.emailTemplates.voiceClone.outputLabel') }}</span>
            <textarea
              v-model="form.text"
              class="voice-clone-textarea"
              rows="7"
              :placeholder="t('pages.emailTemplates.voiceClone.outputPlaceholder')"
            />
          </label>
        </section>

        <section class="voice-clone-panel">
          <header class="voice-clone-panel-header">
            <div>
              <Upload :size="17" stroke-width="1.8" aria-hidden="true" />
              <h3>{{ t('pages.emailTemplates.voiceClone.referenceTitle') }}</h3>
            </div>
            <span>{{ t('pages.emailTemplates.voiceClone.recommendation') }}</span>
          </header>

          <div
            class="voice-clone-dropzone"
            :class="{ 'is-dragging': isDragging }"
            @dragover.prevent="isDragging = true"
            @dragleave.prevent="isDragging = false"
            @drop="handleDrop"
          >
            <Upload :size="30" stroke-width="1.6" aria-hidden="true" />
            <strong>{{ referenceFile ? referenceFile.name : t('pages.emailTemplates.voiceClone.dropTitle') }}</strong>
            <span>{{ t('pages.emailTemplates.voiceClone.dropDescription') }}</span>
            <UiButton variant="secondary" type="button" :disabled="isGenerating || isSaving" @click="openFileDialog">
              {{ t('common.actions.select') }}
            </UiButton>
            <input
              ref="fileInput"
              class="ui-file-input ui-file-input--hidden"
              type="file"
              :accept="ACCEPTED_AUDIO_TYPES"
              @change="handleFileChange"
            />
          </div>

          <audio v-if="referenceUrl" class="voice-clone-audio" :src="referenceUrl" controls />
        </section>

        <section class="voice-clone-panel voice-clone-panel--details">
          <header class="voice-clone-panel-header">
            <div>
              <Settings :size="17" stroke-width="1.8" aria-hidden="true" />
              <h3>{{ t('pages.emailTemplates.voiceClone.detailsTitle') }}</h3>
            </div>
          </header>

          <div class="voice-clone-two-col">
            <label class="ui-field">
              <span class="ui-field-label">{{ t('pages.emailTemplates.voiceClone.languageLabel') }}</span>
              <select v-model="form.language" class="ui-select">
                <option v-for="option in languageOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
              <span class="ui-field-helper">{{ t('pages.emailTemplates.voiceClone.languageHint') }}</span>
            </label>
          </div>

          <label class="ui-field">
            <span class="ui-field-label">{{ t('pages.emailTemplates.voiceClone.instructLabel') }}</span>
            <textarea
              v-model="form.instruct"
              class="voice-clone-textarea voice-clone-textarea--compact"
              rows="3"
              :placeholder="t('pages.emailTemplates.voiceClone.instructPlaceholder')"
            />
            <span class="ui-field-helper">{{ t('pages.emailTemplates.voiceClone.instructHint') }}</span>
          </label>
        </section>
      </div>

      <aside class="voice-clone-side">
        <section class="voice-clone-panel">
          <header class="voice-clone-panel-header">
            <div>
              <Settings :size="17" stroke-width="1.8" aria-hidden="true" />
              <h3>{{ t('pages.emailTemplates.voiceClone.settingsTitle') }}</h3>
            </div>
          </header>

          <div class="voice-clone-setting">
            <label>
              <span>{{ t('pages.emailTemplates.voiceClone.speed') }}</span>
              <input v-model.number="form.speed" type="number" min="0.5" max="1.5" step="0.05" class="ui-input" />
            </label>
            <input v-model.number="form.speed" type="range" min="0.5" max="1.5" step="0.05" />
            <small>{{ t('pages.emailTemplates.voiceClone.speedHint') }}</small>
          </div>

          <UiInput
            v-model="form.durationSeconds"
            type="number"
            :label="t('pages.emailTemplates.voiceClone.duration')"
            :helper="t('pages.emailTemplates.voiceClone.durationHint')"
            placeholder="0"
          />

          <div class="voice-clone-setting">
            <label>
              <span>{{ t('pages.emailTemplates.voiceClone.inferenceSteps') }}</span>
              <input v-model.number="form.inferenceSteps" type="number" min="4" max="64" step="1" class="ui-input" />
            </label>
            <input v-model.number="form.inferenceSteps" type="range" min="4" max="64" step="1" />
            <small>{{ t('pages.emailTemplates.voiceClone.inferenceHint') }}</small>
          </div>

          <div class="voice-clone-setting">
            <label>
              <span>{{ t('pages.emailTemplates.voiceClone.guidanceScale') }}</span>
              <input v-model.number="form.guidanceScale" type="number" min="0" max="4" step="0.1" class="ui-input" />
            </label>
            <input v-model.number="form.guidanceScale" type="range" min="0" max="4" step="0.1" />
            <small>{{ t('pages.emailTemplates.voiceClone.guidanceHint') }}</small>
          </div>

          <div class="voice-clone-checks">
            <label>
              <input v-model="form.denoise" type="checkbox" />
              <span>{{ t('pages.emailTemplates.voiceClone.denoise') }}</span>
            </label>
            <label>
              <input v-model="form.preprocessPrompt" type="checkbox" />
              <span>{{ t('pages.emailTemplates.voiceClone.preprocessPrompt') }}</span>
            </label>
            <label>
              <input v-model="form.postprocessOutput" type="checkbox" />
              <span>{{ t('pages.emailTemplates.voiceClone.postprocessOutput') }}</span>
            </label>
          </div>
        </section>

        <section class="voice-clone-panel voice-clone-output">
          <header class="voice-clone-panel-header">
            <div>
              <Volume2 :size="17" stroke-width="1.8" aria-hidden="true" />
              <h3>{{ t('pages.emailTemplates.voiceClone.outputTitle') }}</h3>
            </div>
          </header>

          <div v-if="resultUrl" class="voice-clone-output-preview">
            <audio class="voice-clone-audio" :src="resultUrl" controls />
            <dl>
              <div><dt>{{ t('pages.emailTemplates.voiceClone.duration') }}</dt><dd>{{ resultDurationText }}</dd></div>
              <div><dt>{{ t('common.labels.size') }}</dt><dd>{{ resultSizeText }}</dd></div>
            </dl>
          </div>
          <div v-else class="voice-clone-output-empty">
            <Volume2 :size="30" stroke-width="1.5" aria-hidden="true" />
            <span>{{ isGenerating ? t('pages.emailTemplates.voiceClone.generating') : t('pages.emailTemplates.voiceClone.outputEmpty') }}</span>
          </div>

          <UiInput
            v-model="form.label"
            :label="t('pages.emailTemplates.voiceClone.saveName')"
            :placeholder="defaultLabel()"
          />

          <div class="voice-clone-actions">
            <UiButton :loading="isGenerating" :disabled="!canGenerate" @click="generateVoice">
              <Volume2 :size="16" stroke-width="1.8" aria-hidden="true" />
              {{ isGenerating ? t('pages.emailTemplates.voiceClone.generating') : t('pages.emailTemplates.voiceClone.generate') }}
            </UiButton>
            <UiButton variant="secondary" :loading="isSaving" :disabled="!canSave" @click="saveResult">
              <Save :size="16" stroke-width="1.8" aria-hidden="true" />
              {{ isSaving ? t('common.states.saving') : t('pages.emailTemplates.voiceClone.saveResult') }}
            </UiButton>
          </div>
        </section>
      </aside>
    </section>
  </section>
</template>
