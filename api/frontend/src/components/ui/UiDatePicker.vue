<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { CalendarDays, ChevronLeft, ChevronRight, X } from 'lucide-vue-next'

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
  min: {
    type: String,
    default: '',
  },
  max: {
    type: String,
    default: '',
  },
  error: {
    type: String,
    default: '',
  },
  id: {
    type: String,
    default: '',
  },
  name: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['update:modelValue', 'blur', 'change'])

const { t, tm, locale } = useI18n()
const rootRef = ref(null)
const inputRef = ref(null)
const isOpen = ref(false)
const focusedIso = ref('')
const visibleMonth = ref(startOfMonth(parseIsoDate(props.modelValue) || new Date()))

const inputId = computed(() => props.id || undefined)
const selectedDate = computed(() => parseIsoDate(props.modelValue))
const minDate = computed(() => parseIsoDate(props.min))
const maxDate = computed(() => parseIsoDate(props.max))
const isRussianLocale = computed(() => locale.value === 'ru')
const weekStart = computed(() => (isRussianLocale.value ? 1 : 0))
const localizedPlaceholder = computed(() => props.placeholder || t('common.date.placeholder'))
const monthNames = computed(() => {
  locale.value
  const value = tm('common.date.months')
  return Array.isArray(value) ? value : []
})
const weekdayNames = computed(() => {
  locale.value
  const value = tm('common.date.weekdaysShort')
  return Array.isArray(value) ? value : []
})
const visibleMonthLabel = computed(() => monthNames.value[visibleMonth.value.getMonth()] || '')
const visibleYear = computed(() => visibleMonth.value.getFullYear())
const displayValue = computed(() => formatDisplayDate(selectedDate.value, isRussianLocale.value))
const orderedWeekdays = computed(() => weekdayNames.value)
const calendarDays = computed(() => buildCalendarDays(visibleMonth.value, weekStart.value))

watch(() => props.modelValue, (value) => {
  const nextDate = parseIsoDate(value)
  if (nextDate) {
    visibleMonth.value = startOfMonth(nextDate)
    focusedIso.value = toIsoDate(nextDate)
  }
})

watch(locale, () => {
  if (isOpen.value) {
    visibleMonth.value = startOfMonth(visibleMonth.value)
  }
})

function parseIsoDate(value) {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(String(value || ''))
  if (!match) {
    return null
  }

  const year = Number(match[1])
  const month = Number(match[2])
  const day = Number(match[3])
  const date = new Date(year, month - 1, day)

  if (
    date.getFullYear() !== year
    || date.getMonth() !== month - 1
    || date.getDate() !== day
  ) {
    return null
  }

  return date
}

function toIsoDate(date) {
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, '0'),
    String(date.getDate()).padStart(2, '0'),
  ].join('-')
}

function startOfMonth(date) {
  return new Date(date.getFullYear(), date.getMonth(), 1)
}

function addMonths(date, amount) {
  return new Date(date.getFullYear(), date.getMonth() + amount, 1)
}

function formatDisplayDate(date, russian) {
  if (!date) {
    return ''
  }

  const day = String(date.getDate()).padStart(2, '0')
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const year = date.getFullYear()

  return russian ? `${day}.${month}.${year}` : `${month}/${day}/${year}`
}

function buildCalendarDays(monthDate, firstDayOfWeek) {
  const first = startOfMonth(monthDate)
  const leadingDays = (first.getDay() - firstDayOfWeek + 7) % 7
  const start = new Date(first)
  start.setDate(first.getDate() - leadingDays)

  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(start)
    date.setDate(start.getDate() + index)
    const iso = toIsoDate(date)

    return {
      date,
      iso,
      day: date.getDate(),
      isCurrentMonth: date.getMonth() === monthDate.getMonth(),
      isSelected: iso === props.modelValue,
      isToday: iso === toIsoDate(new Date()),
      isFocused: iso === focusedIso.value,
      isDisabled: isDateDisabled(date),
    }
  })
}

function isDateDisabled(date) {
  const iso = toIsoDate(date)
  const minIso = minDate.value ? toIsoDate(minDate.value) : ''
  const maxIso = maxDate.value ? toIsoDate(maxDate.value) : ''

  return Boolean((minIso && iso < minIso) || (maxIso && iso > maxIso))
}

function openCalendar() {
  if (props.disabled) {
    return
  }

  const baseDate = selectedDate.value || new Date()
  visibleMonth.value = startOfMonth(baseDate)
  focusedIso.value = toIsoDate(baseDate)
  isOpen.value = true
}

function toggleCalendar() {
  if (isOpen.value) {
    closeCalendar()
  } else {
    openCalendar()
  }
}

function closeCalendar() {
  if (!isOpen.value) {
    return
  }

  isOpen.value = false
  emit('blur')
}

function emitDate(iso) {
  emit('update:modelValue', iso)
  emit('change', iso)
}

function selectDate(day) {
  if (day?.isDisabled || props.disabled) {
    return
  }

  focusedIso.value = day.iso
  visibleMonth.value = startOfMonth(day.date)
  emitDate(day.iso)
  isOpen.value = false
  inputRef.value?.focus()
}

function selectIsoDate(iso) {
  const date = parseIsoDate(iso)
  if (!date || isDateDisabled(date)) {
    return
  }

  selectDate({
    date,
    iso,
    isDisabled: false,
  })
}

function clearValue() {
  if (props.disabled) {
    return
  }

  emitDate('')
  focusedIso.value = ''
  isOpen.value = false
  inputRef.value?.focus()
}

function selectToday() {
  selectIsoDate(toIsoDate(new Date()))
}

function moveMonth(amount) {
  visibleMonth.value = addMonths(visibleMonth.value, amount)
  const focusedDate = parseIsoDate(focusedIso.value)
  if (!focusedDate || focusedDate.getMonth() !== visibleMonth.value.getMonth()) {
    focusedIso.value = toIsoDate(visibleMonth.value)
  }
}

function handleInputKeydown(event) {
  if (event.key === 'Escape') {
    closeCalendar()
    return
  }

  if (event.key === 'Enter') {
    event.preventDefault()
    if (!isOpen.value) {
      openCalendar()
      return
    }

    selectIsoDate(focusedIso.value || toIsoDate(new Date()))
  }
}

function handleDayKeydown(event, day) {
  if (event.key === 'Escape') {
    closeCalendar()
    inputRef.value?.focus()
    return
  }

  if (event.key === 'Enter') {
    event.preventDefault()
    selectDate(day)
  }
}

function handleDocumentPointerDown(event) {
  if (!isOpen.value || rootRef.value?.contains(event.target)) {
    return
  }

  closeCalendar()
}

onMounted(() => {
  document.addEventListener('pointerdown', handleDocumentPointerDown)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', handleDocumentPointerDown)
})
</script>

<template>
  <div ref="rootRef" class="ui-field ui-date-picker" :class="{ 'is-open': isOpen, 'is-disabled': disabled }">
    <label v-if="label" class="ui-field-label" :for="inputId">{{ label }}</label>

    <div class="ui-date-picker-control">
      <input
        :id="inputId"
        ref="inputRef"
        class="ui-input ui-date-picker-input"
        :class="{ 'has-error': error }"
        type="text"
        inputmode="none"
        readonly
        :name="name || undefined"
        :value="displayValue"
        :placeholder="localizedPlaceholder"
        :disabled="disabled"
        :aria-expanded="isOpen ? 'true' : 'false'"
        :aria-invalid="error ? 'true' : undefined"
        autocomplete="off"
        role="combobox"
        @focus="openCalendar"
        @click="openCalendar"
        @keydown="handleInputKeydown"
      />

      <button
        v-if="modelValue && !disabled"
        type="button"
        class="ui-date-picker-action ui-date-picker-clear"
        :aria-label="t('common.date.clear')"
        :title="t('common.date.clear')"
        @click="clearValue"
      >
        <X :size="15" stroke-width="1.9" aria-hidden="true" />
      </button>

      <button
        type="button"
        class="ui-date-picker-action ui-date-picker-trigger"
        :aria-label="label || t('common.date.open')"
        :disabled="disabled"
        @click="toggleCalendar"
      >
        <CalendarDays :size="16" stroke-width="1.8" aria-hidden="true" />
      </button>
    </div>

    <div v-if="isOpen" class="ui-date-picker-popover" role="dialog">
      <header class="ui-date-picker-header">
        <button
          type="button"
          class="ui-date-picker-nav"
          :aria-label="t('common.date.previousMonth')"
          :title="t('common.date.previousMonth')"
          @click="moveMonth(-1)"
        >
          <ChevronLeft :size="16" stroke-width="1.9" aria-hidden="true" />
        </button>
        <strong>{{ visibleMonthLabel }} {{ visibleYear }}</strong>
        <button
          type="button"
          class="ui-date-picker-nav"
          :aria-label="t('common.date.nextMonth')"
          :title="t('common.date.nextMonth')"
          @click="moveMonth(1)"
        >
          <ChevronRight :size="16" stroke-width="1.9" aria-hidden="true" />
        </button>
      </header>

      <div class="ui-date-picker-weekdays" aria-hidden="true">
        <span v-for="dayName in orderedWeekdays" :key="dayName">{{ dayName }}</span>
      </div>

      <div class="ui-date-picker-grid">
        <button
          v-for="day in calendarDays"
          :key="day.iso"
          type="button"
          class="ui-date-picker-day"
          :class="{
            'is-outside': !day.isCurrentMonth,
            'is-selected': day.isSelected,
            'is-today': day.isToday,
          }"
          :disabled="day.isDisabled"
          :tabindex="day.isFocused ? 0 : -1"
          @click="selectDate(day)"
          @focus="focusedIso = day.iso"
          @keydown="handleDayKeydown($event, day)"
        >
          {{ day.day }}
        </button>
      </div>

      <footer class="ui-date-picker-footer">
        <button type="button" class="ui-date-picker-today" @click="selectToday">
          {{ t('common.date.today') }}
        </button>
      </footer>
    </div>

    <span v-if="error" class="ui-field-error">{{ error }}</span>
  </div>
</template>
