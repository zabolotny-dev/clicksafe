<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Clock, X } from 'lucide-vue-next'

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  label: {
    type: String,
    default: '',
  },
  placeholder: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  id: {
    type: String,
    default: '',
  },
  name: {
    type: String,
    default: '',
  },
  error: {
    type: String,
    default: '',
  },
  minuteStep: {
    type: Number,
    default: 1,
  },
})

const emit = defineEmits(['update:modelValue', 'blur', 'change'])

const { t } = useI18n()
const rootRef = ref(null)
const inputRef = ref(null)
const isOpen = ref(false)

const inputId = computed(() => props.id || undefined)
const localizedPlaceholder = computed(() => props.placeholder || t('common.time.placeholder'))
const normalizedValue = computed(() => (isValidTime(props.modelValue) ? props.modelValue : ''))
const selectedHour = computed(() => normalizedValue.value.split(':')[0] || '')
const selectedMinute = computed(() => normalizedValue.value.split(':')[1] || '')
const resolvedMinuteStep = computed(() => {
  const step = Number.parseInt(props.minuteStep, 10)
  return Number.isFinite(step) ? Math.min(Math.max(step, 1), 59) : 1
})
const hourOptions = computed(() => Array.from({ length: 24 }, (_, index) => formatPart(index)))
const minuteOptions = computed(() => {
  const values = []
  for (let minute = 0; minute < 60; minute += resolvedMinuteStep.value) {
    values.push(formatPart(minute))
  }
  return values
})

function formatPart(value) {
  return String(value).padStart(2, '0')
}

function isValidTime(value) {
  return /^([01]\d|2[0-3]):[0-5]\d$/.test(String(value || ''))
}

function openPicker() {
  if (props.disabled) {
    return
  }

  isOpen.value = true
}

function togglePicker() {
  if (isOpen.value) {
    closePicker()
  } else {
    openPicker()
  }
}

function closePicker() {
  if (!isOpen.value) {
    return
  }

  isOpen.value = false
  emit('blur')
}

function emitTime(value) {
  const nextValue = isValidTime(value) ? value : ''
  emit('update:modelValue', nextValue)
  emit('change', nextValue)
}

function selectHour(hour) {
  const minute = selectedMinute.value || '00'
  emitTime(`${hour}:${minute}`)
  inputRef.value?.focus()
}

function selectMinute(minute) {
  const hour = selectedHour.value || '00'
  emitTime(`${hour}:${minute}`)
  inputRef.value?.focus()
}

function clearValue() {
  if (props.disabled) {
    return
  }

  emitTime('')
  closePicker()
  inputRef.value?.focus()
}

function handleInputKeydown(event) {
  if (event.key === 'Escape') {
    closePicker()
    return
  }

  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    togglePicker()
  }
}

function handlePopoverKeydown(event) {
  if (event.key === 'Escape') {
    closePicker()
    inputRef.value?.focus()
  }
}

function handleFocusOut(event) {
  if (!isOpen.value || rootRef.value?.contains(event.relatedTarget)) {
    return
  }

  closePicker()
}

function handleDocumentPointerDown(event) {
  if (!isOpen.value || rootRef.value?.contains(event.target)) {
    return
  }

  closePicker()
}

onMounted(() => {
  document.addEventListener('pointerdown', handleDocumentPointerDown)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', handleDocumentPointerDown)
})
</script>

<template>
  <div
    ref="rootRef"
    class="ui-field ui-time-picker"
    :class="{ 'is-open': isOpen, 'is-disabled': disabled }"
    @focusout="handleFocusOut"
  >
    <label v-if="label" class="ui-field-label" :for="inputId">{{ label }}</label>

    <div class="ui-time-picker-control">
      <input
        :id="inputId"
        ref="inputRef"
        class="ui-input ui-time-picker-input"
        :class="{ 'has-error': error }"
        type="text"
        inputmode="none"
        readonly
        :name="name || undefined"
        :value="normalizedValue"
        :placeholder="localizedPlaceholder"
        :disabled="disabled"
        :aria-expanded="isOpen ? 'true' : 'false'"
        :aria-invalid="error ? 'true' : undefined"
        autocomplete="off"
        role="combobox"
        @focus="openPicker"
        @click="openPicker"
        @keydown="handleInputKeydown"
      />

      <button
        v-if="normalizedValue && !disabled"
        type="button"
        class="ui-time-picker-action ui-time-picker-clear"
        :aria-label="t('common.time.clear')"
        :title="t('common.time.clear')"
        @click="clearValue"
      >
        <X :size="15" stroke-width="1.9" aria-hidden="true" />
      </button>

      <button
        type="button"
        class="ui-time-picker-action ui-time-picker-trigger"
        :aria-label="label || t('common.time.placeholder')"
        :disabled="disabled"
        @click="togglePicker"
      >
        <Clock :size="16" stroke-width="1.8" aria-hidden="true" />
      </button>
    </div>

    <div
      v-if="isOpen"
      class="ui-time-picker-popover"
      role="dialog"
      @keydown="handlePopoverKeydown"
    >
      <div class="ui-time-picker-columns">
        <section class="ui-time-picker-column" :aria-label="t('common.time.hours')">
          <strong>{{ t('common.time.hours') }}</strong>
          <div class="ui-time-picker-list">
            <button
              v-for="hour in hourOptions"
              :key="hour"
              type="button"
              class="ui-time-picker-option"
              :class="{ 'is-selected': hour === selectedHour }"
              @click="selectHour(hour)"
            >
              {{ hour }}
            </button>
          </div>
        </section>

        <section class="ui-time-picker-column" :aria-label="t('common.time.minutes')">
          <strong>{{ t('common.time.minutes') }}</strong>
          <div class="ui-time-picker-list">
            <button
              v-for="minute in minuteOptions"
              :key="minute"
              type="button"
              class="ui-time-picker-option"
              :class="{ 'is-selected': minute === selectedMinute }"
              @click="selectMinute(minute)"
            >
              {{ minute }}
            </button>
          </div>
        </section>
      </div>
    </div>

    <span v-if="error" class="ui-field-error">{{ error }}</span>
  </div>
</template>
