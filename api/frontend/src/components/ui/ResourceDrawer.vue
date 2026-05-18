<script setup>
import { useI18n } from 'vue-i18n'
import { X } from 'lucide-vue-next'
import IconButton from './IconButton.vue'

defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    required: true,
  },
  subtitle: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['close'])
const { t } = useI18n()
</script>

<template>
  <Teleport to="body">
    <Transition name="preview-drawer">
      <div v-if="open" class="preview-drawer-layer" role="presentation">
        <button
          class="preview-drawer-backdrop"
          type="button"
          :aria-label="t('common.drawers.close')"
          @click="emit('close')"
        />
        <aside
          class="preview-drawer resource-drawer"
          role="dialog"
          aria-modal="true"
          :aria-label="title"
        >
          <header class="preview-drawer-header">
            <div class="preview-drawer-copy">
              <h2>{{ title }}</h2>
              <p v-if="subtitle">{{ subtitle }}</p>
            </div>
            <IconButton :label="t('common.drawers.close')" @click="emit('close')">
              <X :size="18" stroke-width="1.8" aria-hidden="true" />
            </IconButton>
          </header>
          <div class="preview-drawer-body resource-drawer-body">
            <slot />
          </div>
          <footer v-if="$slots.footer" class="resource-drawer-footer">
            <slot name="footer" />
          </footer>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>
