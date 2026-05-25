<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowLeft, Image as ImageIcon, Save } from 'lucide-vue-next'
import AttachmentImagePickerDrawer from './AttachmentImagePickerDrawer.vue'
import UiAlert from './UiAlert.vue'
import UiButton from './UiButton.vue'

const props = defineProps({
  initialHtml: {
    type: String,
    default: '',
  },
  initialContent: {
    type: String,
    default: '',
  },
  contentType: {
    type: String,
    default: 'html',
  },
  language: {
    type: String,
    default: '',
  },
  title: {
    type: String,
    default: 'HTML editor',
  },
  description: {
    type: String,
    default: '',
  },
  saveLabel: {
    type: String,
    default: 'Save HTML',
  },
  filename: {
    type: String,
    default: '',
  },
  filenamePlaceholder: {
    type: String,
    default: 'template.html',
  },
  showFilename: {
    type: Boolean,
    default: true,
  },
  filenameReadonly: {
    type: Boolean,
    default: false,
  },
  workspace: {
    type: Boolean,
    default: false,
  },
  saving: {
    type: Boolean,
    default: false,
  },
  showPreview: {
    type: Boolean,
    default: true,
  },
  showImageAction: {
    type: Boolean,
    default: true,
  },
  placeholder: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['save', 'cancel', 'update', 'update:filename'])
const { t } = useI18n()

const editorHost = ref(null)
const resizeHost = ref(null)
const localHtml = ref(props.initialContent || props.initialHtml)
const isEditorLoading = ref(false)
const isMonacoReady = ref(false)
const editorError = ref('')
const isResizing = ref(false)
const splitPercent = ref(50)
const imageDrawerOpen = ref(false)

let editor = null
let monacoModule = null
let changeSubscription = null
let resizePointerId = null

const isHtmlMode = computed(() => String(props.contentType || '').toLowerCase() === 'html')
const showPreviewPane = computed(() => props.showPreview && isHtmlMode.value)
const showImageButton = computed(() => props.showImageAction && isHtmlMode.value)
const editorLanguage = computed(() => props.language || (isHtmlMode.value ? 'html' : 'plaintext'))
const editorDescription = computed(() => props.description || t('common.preview.rawTemplate'))
const codePaneLabel = computed(() => (isHtmlMode.value ? t('common.labels.htmlBody') : t('common.preview.text')))
const editorPlaceholder = computed(() => props.placeholder || (isHtmlMode.value ? '<html>...</html>' : ''))
const previewTitle = computed(() => `${props.title} ${t('common.actions.preview').toLowerCase()}`)
const editorGridStyle = computed(() => ({
  '--editor-split': `${splitPercent.value}%`,
}))

function clampSplit(value) {
  return Math.min(65, Math.max(35, value))
}

function resizePane(event) {
  if (!resizeHost.value || !isResizing.value) {
    return
  }

  const rect = resizeHost.value.getBoundingClientRect()
  if (!rect.width) {
    return
  }

  const nextPercent = ((event.clientX - rect.left) / rect.width) * 100
  splitPercent.value = clampSplit(nextPercent)
}

function stopResize() {
  isResizing.value = false
  resizePointerId = null
  document.removeEventListener('pointermove', resizePane)
  document.removeEventListener('pointerup', stopResize)
  document.removeEventListener('pointercancel', stopResize)
}

function startResize(event) {
  if (!resizeHost.value) {
    return
  }

  isResizing.value = true
  resizePointerId = event.pointerId
  event.currentTarget?.setPointerCapture?.(resizePointerId)
  resizePane(event)
  document.addEventListener('pointermove', resizePane)
  document.addEventListener('pointerup', stopResize)
  document.addEventListener('pointercancel', stopResize)
}

function disposeMonaco() {
  if (changeSubscription) {
    changeSubscription.dispose()
    changeSubscription = null
  }

  if (editor) {
    editor.dispose()
    editor = null
  }
}

async function mountMonaco() {
  if (!editorHost.value || editor || isEditorLoading.value) {
    return
  }

  isEditorLoading.value = true
  editorError.value = ''

  try {
    monacoModule = await import('monaco-editor')
    await nextTick()

    editor = monacoModule.editor.create(editorHost.value, {
      value: localHtml.value,
      language: editorLanguage.value,
      theme: 'vs-dark',
      automaticLayout: true,
      minimap: { enabled: false },
      wordWrap: 'on',
      scrollBeyondLastLine: false,
      fontSize: 13,
      lineNumbersMinChars: 4,
      tabSize: 2,
      quickSuggestions: true,
      suggestOnTriggerCharacters: true,
      autoClosingBrackets: 'always',
      autoClosingQuotes: 'always',
      autoIndent: 'full',
      formatOnPaste: true,
    })

    changeSubscription = editor.onDidChangeModelContent(() => {
      localHtml.value = editor.getValue()
      emit('update', localHtml.value)
    })
    isMonacoReady.value = true
  } catch (error) {
    disposeMonaco()
    isMonacoReady.value = false
    editorError.value = error?.message || t('errors.requestFailed')
  } finally {
    isEditorLoading.value = false
  }
}

function handleTextareaInput() {
  emit('update', localHtml.value)
}

function save() {
  emit('save', localHtml.value)
}

watch(() => [props.initialContent, props.initialHtml], ([initialContent, initialHtml]) => {
  const value = initialContent || initialHtml
  localHtml.value = value

  if (editor && editor.getValue() !== value) {
    editor.setValue(value)
  }
})

watch(editorLanguage, (language) => {
  const model = editor?.getModel?.()
  if (monacoModule && model) {
    monacoModule.editor.setModelLanguage(model, language)
  }
})

watch(localHtml, (value) => {
  if (editor && editor.getValue() !== value) {
    editor.setValue(value)
  }
})

onMounted(() => {
  mountMonaco()
})

onBeforeUnmount(() => {
  stopResize()
  disposeMonaco()
})
</script>

<template>
  <section
    class="html-live-editor"
    :class="{
      'html-live-editor--workspace': workspace,
      'html-live-editor--code-only': !showPreviewPane,
      'is-resizing': isResizing,
    }"
    :aria-label="title"
  >
    <header class="html-live-editor-toolbar">
      <div>
        <h2>{{ title }}</h2>
        <p>{{ editorDescription }}</p>
      </div>
      <label v-if="workspace && showFilename" class="html-live-editor-filename ui-field">
        <span class="ui-field-label">{{ t('common.labels.fileName') }}</span>
        <input
          class="ui-input"
          :value="filename"
          :placeholder="filenamePlaceholder"
          :readonly="filenameReadonly"
          @input="emit('update:filename', $event.target.value)"
        />
      </label>
      <div class="html-live-editor-actions">
        <UiButton variant="secondary" :disabled="saving" @click="emit('cancel')">
          <ArrowLeft :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ t('common.actions.back') }}
        </UiButton>
        <UiButton v-if="showImageButton" variant="secondary" @click="imageDrawerOpen = true">
          <ImageIcon :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ t('components.htmlEditor.addImage') }}
        </UiButton>
        <UiButton :loading="saving" @click="save">
          <Save :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ saving ? t('common.states.saving') : saveLabel }}
        </UiButton>
      </div>
    </header>

    <UiAlert
      v-if="editorError"
      variant="error"
      :title="t('common.preview.unavailable')"
      :message="editorError"
    />

    <div ref="resizeHost" class="html-live-editor-grid" :style="editorGridStyle">
      <section class="html-code-pane" :aria-label="codePaneLabel">
        <div v-show="isMonacoReady" ref="editorHost" class="html-monaco-host" />
        <textarea
          v-if="!isMonacoReady"
          v-model="localHtml"
          class="ui-textarea ui-textarea--code html-fallback-textarea"
          spellcheck="false"
          :placeholder="editorPlaceholder"
          @input="handleTextareaInput"
        />
        <div v-if="isEditorLoading" class="html-editor-loading">{{ t('common.states.loadingEllipsis') }}</div>
      </section>

      <button
        v-if="showPreviewPane"
        type="button"
        class="html-live-editor-resize"
        :aria-label="t('common.drawers.resize')"
        @pointerdown.prevent="startResize"
      />

      <section v-if="showPreviewPane" class="html-preview-pane" :aria-label="previewTitle">
        <iframe :srcdoc="localHtml" :title="previewTitle" sandbox="" />
      </section>
    </div>

    <AttachmentImagePickerDrawer
      :open="imageDrawerOpen"
      @close="imageDrawerOpen = false"
    />
  </section>
</template>
