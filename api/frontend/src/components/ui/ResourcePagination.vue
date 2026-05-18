<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import IconButton from './IconButton.vue'

const props = defineProps({
  page: {
    type: Number,
    required: true,
  },
  rows: {
    type: Number,
    required: true,
  },
  total: {
    type: Number,
    required: true,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  rowsOptions: {
    type: Array,
    default: () => [10, 25, 50, 100],
  },
  loadedCount: {
    type: Number,
    default: null,
  },
})

const emit = defineEmits(['update:page', 'update:rows'])

const { t } = useI18n()

const safeRows = computed(() => {
  const rows = Number(props.rows)
  return Number.isFinite(rows) && rows > 0 ? rows : 10
})
const safeTotal = computed(() => {
  const total = Number(props.total)
  return Number.isFinite(total) && total > 0 ? total : 0
})
const totalPages = computed(() => Math.max(1, Math.ceil(safeTotal.value / safeRows.value)))
const currentPage = computed(() => {
  const page = Number(props.page)
  const normalized = Number.isFinite(page) && page > 0 ? page : 1
  return Math.min(normalized, totalPages.value)
})
const currentLoadedCount = computed(() => {
  const count = Number(props.loadedCount)
  return Number.isFinite(count) && count >= 0 ? count : null
})
const showingFrom = computed(() => {
  if (!safeTotal.value) {
    return 0
  }

  if (currentLoadedCount.value === 0) {
    return 0
  }

  return ((currentPage.value - 1) * safeRows.value) + 1
})
const showingTo = computed(() => {
  if (!safeTotal.value || currentLoadedCount.value === 0) {
    return 0
  }

  const loadedLimit = currentLoadedCount.value === null
    ? safeRows.value
    : currentLoadedCount.value

  return Math.min(safeTotal.value, showingFrom.value + loadedLimit - 1)
})
const summary = computed(() => {
  if (!safeTotal.value) {
    return t('common.pagination.empty')
  }

  return t('common.pagination.showing', {
    from: showingFrom.value,
    to: showingTo.value,
    total: safeTotal.value,
  })
})

function emitPage(nextPage) {
  if (props.loading) {
    return
  }

  const normalized = Math.min(Math.max(1, nextPage), totalPages.value)
  if (normalized !== props.page) {
    emit('update:page', normalized)
  }
}

function emitRows(event) {
  if (props.loading) {
    return
  }

  const nextRows = Number(event.target.value)
  if (Number.isFinite(nextRows) && nextRows > 0 && nextRows !== props.rows) {
    emit('update:rows', nextRows)
  }
}
</script>

<template>
  <footer class="resource-pagination" aria-live="polite">
    <span class="resource-pagination-summary">{{ summary }}</span>
    <div class="resource-pagination-controls">
      <IconButton
        :label="t('common.pagination.previous')"
        :disabled="currentPage <= 1 || loading"
        @click="emitPage(currentPage - 1)"
      >
        <ChevronLeft :size="16" stroke-width="1.8" aria-hidden="true" />
      </IconButton>
      <strong>{{ t('common.pagination.page') }} {{ currentPage }} {{ t('common.pagination.of') }} {{ totalPages }}</strong>
      <IconButton
        :label="t('common.pagination.next')"
        :disabled="currentPage >= totalPages || loading"
        @click="emitPage(currentPage + 1)"
      >
        <ChevronRight :size="16" stroke-width="1.8" aria-hidden="true" />
      </IconButton>
      <label class="resource-pagination-rows">
        <span>{{ t('common.pagination.rows') }}</span>
        <select :value="safeRows" class="ui-select" :disabled="loading" @change="emitRows">
          <option v-for="option in rowsOptions" :key="option" :value="option">
            {{ option }}
          </option>
        </select>
      </label>
    </div>
  </footer>
</template>
