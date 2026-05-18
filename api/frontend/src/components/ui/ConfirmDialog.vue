<script setup>
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import UiButton from './UiButton.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    required: true,
  },
  message: {
    type: String,
    required: true,
  },
  confirmLabel: {
    type: String,
    default: '',
  },
  loadingLabel: {
    type: String,
    default: '',
  },
  cancelLabel: {
    type: String,
    default: '',
  },
  variant: {
    type: String,
    default: 'primary',
    validator: (value) => ['primary', 'danger'].includes(value),
  },
  loading: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['cancel', 'confirm'])
const dialogRef = ref(null)
const { t } = useI18n()
const resolvedConfirmLabel = computed(() => props.confirmLabel || t('common.actions.confirm'))
const resolvedCancelLabel = computed(() => props.cancelLabel || t('common.actions.cancel'))

watch(() => props.open, (open) => {
  if (open) {
    nextTick(() => dialogRef.value?.focus())
  }
})

function cancel() {
  if (!props.loading) {
    emit('cancel')
  }
}

function handleKeydown(event) {
  if (event.key === 'Escape') {
    cancel()
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="ui-confirm-backdrop"
      role="presentation"
      @click.self="cancel"
      @keydown="handleKeydown"
    >
      <section
        ref="dialogRef"
        class="ui-confirm-dialog"
        :class="`ui-confirm-dialog--${variant}`"
        role="dialog"
        aria-modal="true"
        aria-labelledby="ui-confirm-title"
        aria-describedby="ui-confirm-description"
        tabindex="-1"
      >
        <header>
          <h2 id="ui-confirm-title">{{ title }}</h2>
          <p id="ui-confirm-description">{{ message }}</p>
        </header>
        <div class="ui-confirm-actions">
          <UiButton
            variant="ghost"
            :disabled="loading"
            @click="cancel"
          >
            {{ resolvedCancelLabel }}
          </UiButton>
          <UiButton
            :variant="variant === 'danger' ? 'danger' : 'primary'"
            :loading="loading"
            @click="$emit('confirm')"
          >
            {{ loading && loadingLabel ? loadingLabel : resolvedConfirmLabel }}
          </UiButton>
        </div>
      </section>
    </div>
  </Teleport>
</template>
