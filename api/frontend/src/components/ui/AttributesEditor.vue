<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import UiButton from './UiButton.vue'
import UiInput from './UiInput.vue'
import { attributesToRows, normalizeAttributes } from '../../utils/resourceFormat'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({}),
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue'])
const { t } = useI18n()

let nextRowId = 1
const rows = ref(makeRows(props.modelValue))

function makeRows(attributes) {
  return attributesToRows(attributes).map((row) => ({
    id: nextRowId += 1,
    key: row.key,
    value: row.value,
  }))
}

function currentValue() {
  return normalizeAttributes(rows.value)
}

watch(() => props.modelValue, (value) => {
  if (JSON.stringify(value || {}) !== JSON.stringify(currentValue())) {
    rows.value = makeRows(value)
  }
}, { deep: true })

function emitValue() {
  emit('update:modelValue', currentValue())
}

function updateRow(index, field, value) {
  rows.value[index] = {
    ...rows.value[index],
    [field]: value,
  }
  emitValue()
}

function addRow() {
  rows.value.push({
    id: nextRowId += 1,
    key: '',
    value: '',
  })
}

function removeRow(index) {
  rows.value.splice(index, 1)

  if (!rows.value.length) {
    addRow()
  }

  emitValue()
}
</script>

<template>
  <div class="attributes-editor">
    <div
      v-for="(row, index) in rows"
      :key="row.id"
      class="attributes-editor-row"
    >
      <UiInput
        :model-value="row.key"
        :label="t('common.labels.key')"
        :placeholder="t('common.attributes.keyPlaceholder')"
        :disabled="disabled"
        @update:model-value="updateRow(index, 'key', $event)"
      />
      <UiInput
        :model-value="row.value"
        :label="t('common.labels.value')"
        :placeholder="t('common.attributes.valuePlaceholder')"
        :disabled="disabled"
        @update:model-value="updateRow(index, 'value', $event)"
      />
      <UiButton
        variant="ghost"
        :disabled="disabled"
        @click="removeRow(index)"
      >
        {{ t('common.actions.delete') }}
      </UiButton>
    </div>
    <UiButton
      variant="secondary"
      :disabled="disabled"
      @click="addRow"
    >
      {{ t('common.actions.addAttribute') }}
    </UiButton>
  </div>
</template>
