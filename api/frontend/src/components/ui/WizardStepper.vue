<script setup>
import { computed } from 'vue'

const props = defineProps({
  steps: {
    type: Array,
    required: true,
  },
  activeIndex: {
    type: Number,
    default: 0,
  },
  activeKey: {
    type: String,
    default: '',
  },
  completedKeys: {
    type: Array,
    default: () => [],
  },
  clickable: {
    type: Boolean,
    default: true,
  },
  ariaLabel: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['update:activeIndex', 'step-click'])

const normalizedSteps = computed(() => props.steps.map((step, index) => (
  typeof step === 'string'
    ? { key: step, label: step, index }
    : {
        key: String(step.key ?? step.label ?? index),
        label: step.label ?? String(step.key ?? index + 1),
        completed: Boolean(step.completed),
        index,
      }
)))

const resolvedActiveIndex = computed(() => {
  if (props.activeKey) {
    const keyIndex = normalizedSteps.value.findIndex((step) => step.key === props.activeKey)
    if (keyIndex !== -1) {
      return keyIndex
    }
  }

  return props.activeIndex
})

const stepperStyle = computed(() => ({
  '--wizard-step-count': String(Math.max(1, normalizedSteps.value.length)),
}))

function isComplete(step, index) {
  return step.completed || props.completedKeys.includes(step.key) || index < resolvedActiveIndex.value
}

function selectStep(index) {
  if (!props.clickable) {
    return
  }

  emit('update:activeIndex', index)
  emit('step-click', index)
}
</script>

<template>
  <nav
    class="wizard-stepper"
    :style="stepperStyle"
    :aria-label="ariaLabel"
  >
    <button
      v-for="(step, index) in normalizedSteps"
      :key="step.key"
      type="button"
      class="wizard-step"
      :class="{
        'is-active': resolvedActiveIndex === index,
        'is-complete': isComplete(step, index),
      }"
      :disabled="!clickable"
      :aria-current="resolvedActiveIndex === index ? 'step' : undefined"
      @click="selectStep(index)"
    >
      <span class="wizard-step-marker" aria-hidden="true" />
      <span>{{ step.label }}</span>
    </button>
  </nav>
</template>
