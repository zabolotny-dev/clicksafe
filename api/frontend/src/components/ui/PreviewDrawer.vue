<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { X } from 'lucide-vue-next'
import IconButton from './IconButton.vue'

const props = defineProps({
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
  resizable: {
    type: Boolean,
    default: true,
  },
  initialWidth: {
    type: Number,
    default: 760,
  },
  minWidth: {
    type: Number,
    default: 460,
  },
  maxWidth: {
    type: Number,
    default: 1120,
  },
})

const emit = defineEmits(['close'])
const { t } = useI18n()

const drawerWidth = ref(props.initialWidth)
const isResizing = ref(false)
let dragStartX = 0
let dragStartWidth = 0
let resizeHandle = null
let resizePointerId = null

const drawerStyle = computed(() => ({
  '--drawer-width': `${drawerWidth.value}px`,
}))

function clampWidth(value) {
  const viewportMax = typeof window === 'undefined'
    ? props.maxWidth
    : Math.max(props.minWidth, Math.min(props.maxWidth, window.innerWidth - 32))

  return Math.max(props.minWidth, Math.min(viewportMax, value))
}

function stopResize() {
  isResizing.value = false
  document.removeEventListener('pointermove', resizeDrawer)
  document.removeEventListener('pointerup', stopResize)
  document.removeEventListener('pointercancel', stopResize)

  if (resizeHandle && resizePointerId !== null) {
    if (resizeHandle.hasPointerCapture?.(resizePointerId)) {
      resizeHandle.releasePointerCapture?.(resizePointerId)
    }
    resizeHandle = null
    resizePointerId = null
  }
}

function resizeDrawer(event) {
  drawerWidth.value = clampWidth(dragStartWidth + (dragStartX - event.clientX))
}

function startResize(event) {
  if (!props.resizable) {
    return
  }

  resizeHandle = event.currentTarget
  resizePointerId = event.pointerId
  resizeHandle.setPointerCapture?.(event.pointerId)
  isResizing.value = true
  dragStartX = event.clientX
  dragStartWidth = drawerWidth.value
  document.addEventListener('pointermove', resizeDrawer)
  document.addEventListener('pointerup', stopResize)
  document.addEventListener('pointercancel', stopResize)
}

watch(() => props.initialWidth, (value) => {
  drawerWidth.value = clampWidth(value)
})

onBeforeUnmount(() => {
  stopResize()
})
</script>

<template>
  <Teleport to="body">
    <Transition name="preview-drawer">
      <div
        v-if="open"
        class="preview-drawer-layer"
        :class="{ 'is-resizing': isResizing }"
        role="presentation"
      >
        <button
          class="preview-drawer-backdrop"
          type="button"
          :aria-label="t('common.drawers.closePreview')"
          @click="emit('close')"
        />
        <aside
          class="preview-drawer"
          role="dialog"
          aria-modal="true"
          :aria-label="title"
          :style="drawerStyle"
        >
          <button
            v-if="resizable"
            type="button"
            class="preview-drawer-resize"
            :aria-label="t('common.drawers.resize')"
            @pointerdown.prevent="startResize"
          />
          <header class="preview-drawer-header">
            <div class="preview-drawer-copy">
              <h2>{{ title }}</h2>
              <p v-if="subtitle">{{ subtitle }}</p>
            </div>
            <IconButton :label="t('common.drawers.closePreview')" @click="emit('close')">
              <X :size="18" stroke-width="1.8" aria-hidden="true" />
            </IconButton>
          </header>
          <div v-if="$slots.actions" class="preview-drawer-actions">
            <slot name="actions" />
          </div>
          <div class="preview-drawer-body">
            <slot />
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>
