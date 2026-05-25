<script setup>
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Eye,
  Plus,
  RefreshCw,
} from 'lucide-vue-next'
import EmptyState from './EmptyState.vue'
import IconButton from './IconButton.vue'
import ResourcePagination from './ResourcePagination.vue'
import SkeletonBlock from './SkeletonBlock.vue'
import UiAlert from './UiAlert.vue'
import UiButton from './UiButton.vue'
import UiInput from './UiInput.vue'
import { formatTechnicalId, nullableId } from '../../utils/resourceFormat'

const ROW_OPTIONS = [5, 10, 25]

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  items: {
    type: Array,
    default: () => [],
  },
  resourceType: {
    type: String,
    default: 'message',
    validator: (value) => ['message', 'landing', 'education'].includes(value),
  },
  messageType: {
    type: String,
    default: 'EMAIL',
    validator: (value) => ['EMAIL', 'MAX'].includes(value),
  },
  title: {
    type: String,
    required: true,
  },
  emptyTitle: {
    type: String,
    default: 'No resources found',
  },
  emptyDescription: {
    type: String,
    default: '',
  },
  searchPlaceholder: {
    type: String,
    default: 'Search by label',
  },
  selectedTitle: {
    type: String,
    default: 'Selected resource',
  },
  selectedEmptyText: {
    type: String,
    default: 'Nothing selected',
  },
  createTo: {
    type: Object,
    default: null,
  },
  createLabel: {
    type: String,
    default: '',
  },
  loading: {
    type: Boolean,
    default: false,
  },
  error: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['update:modelValue', 'selection-change', 'preview', 'refresh'])
const { t } = useI18n()

const search = ref('')
const page = ref(1)
const rows = ref(5)
const pickerName = `campaign-resource-picker-${Math.random().toString(36).slice(2)}`

const isMessage = computed(() => props.resourceType === 'message')
const isMaxMessage = computed(() => isMessage.value && props.messageType === 'MAX')
const resolvedEmptyTitle = computed(() => props.emptyTitle || t('common.resource.noResourcesFound'))
const resolvedEmptyDescription = computed(() => props.emptyDescription || t('common.resource.createBeforeAssigning'))
const resolvedSearchPlaceholder = computed(() => props.searchPlaceholder || t('common.resource.searchByLabel'))
const resolvedSelectedTitle = computed(() => props.selectedTitle || t('common.resource.selectedResource'))
const resolvedSelectedEmptyText = computed(() => props.selectedEmptyText || t('common.selection.none'))
const resolvedCreateLabel = computed(() => props.createLabel || t('common.actions.create'))
const previewLabel = computed(() => {
  if (props.resourceType === 'message') {
    return t('common.actions.preview')
  }

  if (props.resourceType === 'education') {
    return t('common.actions.preview')
  }

  return t('common.actions.preview')
})
const selectedItem = computed(() => props.items.find((item) => itemId(item) === props.modelValue) || null)
const filteredItems = computed(() => {
  const query = search.value.trim().toLowerCase()

  if (!query) {
    return props.items
  }

  return props.items.filter((item) => {
    const values = isMessage.value
      ? [item.label, item.from_email, item.from_name, item.subject]
      : [item.label]

    return values
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(query))
  })
})
const totalPages = computed(() => Math.max(1, Math.ceil(filteredItems.value.length / rows.value)))
const pagedItems = computed(() => {
  const start = (page.value - 1) * rows.value
  return filteredItems.value.slice(start, start + rows.value)
})
const selectedSummary = computed(() => selectedItem.value?.label || resolvedSelectedEmptyText.value)
const skeletonCellCount = computed(() => (isMessage.value ? 6 : 3))

function itemId(item) {
  return nullableId(item?.id) || (item?.id ? String(item.id) : '')
}

function attachmentCount(item) {
  return Array.isArray(item?.attachment_ids) ? item.attachment_ids.length : 0
}

function fromLabel(item) {
  if (isMaxMessage.value) {
    return formatTechnicalId(item?.max_account_id)
  }

  return [item?.from_name, item?.from_email].filter(Boolean).join(' / ') || '—'
}

function messageBodyId(item) {
  return nullableId(isMaxMessage.value ? item?.text_body_id : item?.html_body_id)
}

function messageBodyLabel() {
  return isMaxMessage.value ? t('common.labels.textBody') : t('common.labels.subject')
}

function messageSenderLabel() {
  return isMaxMessage.value ? t('pages.emailTemplates.maxAccount') : t('common.labels.from')
}

function selectItem(item) {
  const id = itemId(item)
  emit('update:modelValue', id)
  emit('selection-change', item)
}

function setPage(nextPage) {
  page.value = Math.min(totalPages.value, Math.max(1, nextPage))
}

function setRows(nextRows) {
  rows.value = nextRows
  page.value = 1
}

function resetPage() {
  page.value = 1
}

function previewItem(item) {
  emit('preview', item)
}

watch(search, resetPage)
watch(() => props.items.length, () => {
  if (page.value > totalPages.value) {
    page.value = totalPages.value
  }
})
</script>

<template>
  <section class="attachment-picker-panel campaign-resource-picker" :aria-label="title">
    <header class="attachment-picker-header campaign-resource-picker-header">
      <div>
        <h3>{{ title }}</h3>
        <p>{{ t('common.availableCount', { filtered: filteredItems.length, total: items.length }) }}</p>
      </div>
      <div class="attachment-picker-header-actions campaign-resource-picker-actions">
        <RouterLink
          v-if="createTo"
          class="ui-button ui-button--secondary"
          :to="createTo"
        >
          <Plus :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ resolvedCreateLabel }}
        </RouterLink>
        <IconButton :label="t('common.resource.refreshResources')" :disabled="loading" @click="emit('refresh')">
          <RefreshCw :size="16" stroke-width="1.8" aria-hidden="true" />
        </IconButton>
      </div>
    </header>

    <section class="attachment-picker-toolbar campaign-resource-picker-toolbar">
      <UiInput
        v-model="search"
        :label="t('common.labels.search')"
        :placeholder="resolvedSearchPlaceholder"
        autocomplete="off"
      />
    </section>

    <UiAlert
      v-if="error"
      variant="error"
      :title="t('common.resource.resourcesUnavailable')"
      :message="error"
    >
      <UiButton variant="secondary" @click="emit('refresh')">{{ t('common.actions.retry') }}</UiButton>
    </UiAlert>

    <section class="attachment-picker-selection-bar campaign-resource-selected">
      <div>
        <span>{{ resolvedSelectedTitle }}</span>
        <strong>{{ selectedSummary }}</strong>
      </div>
    </section>

    <EmptyState
      v-if="!loading && !items.length"
      :title="resolvedEmptyTitle"
      :description="resolvedEmptyDescription"
    />
    <EmptyState
      v-else-if="!loading && !filteredItems.length"
      :title="t('common.resource.noMatchingResources')"
      :description="t('common.resource.adjustSearchQuery')"
    />

    <div
      v-else
      class="attachment-picker-table campaign-resource-table"
      :class="isMessage ? 'campaign-resource-table--message' : 'campaign-resource-table--landing'"
      role="table"
      :aria-label="title"
    >
      <div class="attachment-picker-row attachment-picker-head campaign-resource-row campaign-resource-head" role="row">
        <span>{{ t('common.actions.select') }}</span>
        <span>{{ t('common.labels.name') }}</span>
        <template v-if="isMessage">
          <span>{{ messageSenderLabel() }}</span>
          <span>{{ messageBodyLabel() }}</span>
          <span>{{ t('common.labels.attachments') }}</span>
        </template>
        <span>{{ t('pages.campaigns.columns.actions') }}</span>
      </div>

      <template v-if="loading">
        <div
          v-for="index in 5"
          :key="index"
          class="attachment-picker-row campaign-resource-row"
          role="row"
        >
          <span v-for="cell in skeletonCellCount" :key="cell"><SkeletonBlock :rows="1" /></span>
        </div>
      </template>

      <template v-else>
        <div
          v-for="item in pagedItems"
          :key="itemId(item)"
          class="attachment-picker-row campaign-resource-row"
          :class="{ 'is-selected': modelValue === itemId(item) }"
          role="row"
          @click="selectItem(item)"
        >
          <span>
            <input
              :name="pickerName"
              type="radio"
              :checked="modelValue === itemId(item)"
              :aria-label="`${t('common.actions.select')} ${item.label || t('common.placeholder')}`"
              @click.stop
              @change="selectItem(item)"
            />
          </span>
          <span>
            <span class="campaign-resource-label">{{ item.label || '—' }}</span>
          </span>
          <template v-if="isMessage">
            <span>{{ fromLabel(item) }}</span>
            <span>{{ isMaxMessage ? formatTechnicalId(item.text_body_id) : (item.subject || '—') }}</span>
            <span>{{ attachmentCount(item) }}</span>
          </template>
          <span class="attachment-picker-actions campaign-resource-row-actions" @click.stop>
            <IconButton
              :label="messageBodyId(item) ? previewLabel : t('common.resource.noHtmlBody')"
              :disabled="!messageBodyId(item)"
              @click="previewItem(item)"
            >
              <Eye :size="16" stroke-width="1.8" aria-hidden="true" />
            </IconButton>
          </span>
        </div>
      </template>
    </div>

    <ResourcePagination
      :page="page"
      :rows="rows"
      :total="filteredItems.length"
      :loaded-count="pagedItems.length"
      :loading="loading"
      :rows-options="ROW_OPTIONS"
      @update:page="setPage"
      @update:rows="setRows"
    />
  </section>
</template>
