<script setup>
defineProps({
  id: {
    type: String,
    default: '',
  },
  label: {
    type: String,
    default: '',
  },
  modelValue: {
    type: [String, Number],
    default: '',
  },
  type: {
    type: String,
    default: 'text',
  },
  placeholder: {
    type: String,
    default: '',
  },
  name: {
    type: String,
    default: '',
  },
  autocomplete: {
    type: String,
    default: '',
  },
  autofocus: {
    type: Boolean,
    default: false,
  },
  helper: {
    type: String,
    default: '',
  },
  error: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

defineEmits(['update:modelValue'])
</script>

<template>
  <label class="ui-field" :for="id || undefined">
    <span v-if="label" class="ui-field-label">{{ label }}</span>
    <input
      :id="id || undefined"
      class="ui-input"
      :class="{ 'has-error': error }"
      :name="name || undefined"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :autocomplete="autocomplete || undefined"
      :autofocus="autofocus"
      :disabled="disabled"
      :aria-invalid="error ? 'true' : undefined"
      @input="$emit('update:modelValue', $event.target.value)"
    />
    <span v-if="error" class="ui-field-error">{{ error }}</span>
    <span v-else-if="helper" class="ui-field-helper">{{ helper }}</span>
  </label>
</template>
