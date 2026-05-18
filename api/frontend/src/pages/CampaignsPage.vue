<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Eye } from 'lucide-vue-next'
import EmptyState from '../components/ui/EmptyState.vue'
import HtmlAttachmentPreview from '../components/ui/HtmlAttachmentPreview.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import PreviewDrawer from '../components/ui/PreviewDrawer.vue'
import ResourcePagination from '../components/ui/ResourcePagination.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiDatePicker from '../components/ui/UiDatePicker.vue'
import UiInput from '../components/ui/UiInput.vue'
import { useNotifications } from '../composables/useNotifications'
import { useSession } from '../composables/useSession'
import { translateCampaignStatus } from '../i18n'
import { CAMPAIGN_STATUSES, isCampaignAuthError, listCampaigns } from '../resources/campaigns'
import { getLanding } from '../resources/landings'
import { getMessage } from '../resources/messages'
import { EVENT_TYPES, listVTargets } from '../resources/vtargets'
import { loadAllPages } from '../utils/resourcePagination'
import { errorMessage, formatDateRange as formatLocalizedDateRange, formatPercent as formatLocalizedPercent, nullableId } from '../utils/resourceFormat'

const { loadCurrentSession } = useSession()
const { notifyError } = useNotifications()
const { t } = useI18n()
const supportedStatuses = CAMPAIGN_STATUSES
const SENT_OR_LATER_EVENTS = ['MESSAGE_SENT', 'EMAIL_OPENED', 'LINK_OPENED', 'DATA_SENT']

const filters = reactive({
  label: '',
  status: '',
  dateFrom: '',
  dateTo: '',
})

const columns = computed(() => [
  t('pages.campaigns.columns.campaign'),
  t('pages.campaigns.columns.status'),
  t('pages.campaigns.columns.domain'),
  t('pages.campaigns.columns.dateRange'),
  t('pages.campaigns.columns.resources'),
  t('pages.campaigns.columns.targets'),
  t('pages.campaigns.columns.opens'),
  t('pages.campaigns.columns.clicks'),
  t('pages.campaigns.columns.submissions'),
  t('pages.campaigns.columns.actions'),
])

const campaigns = ref([])
const campaignMetrics = ref({})
const page = ref(1)
const rows = ref(10)
const total = ref(0)
const isLoading = ref(false)
const isMetricsLoading = ref(false)
const hasLoaded = ref(false)
const error = ref('')
const metricsError = ref('')
const authSuppressed = ref(false)
const assetPreviewOpen = ref(false)
const assetPreviewTitle = ref('')
const assetPreviewKind = ref('')
const assetPreviewAttachment = ref(null)
const assetPreviewMeta = ref([])
const assetPreviewLoadingKey = ref('')

let loadTimer = 0
let requestSequence = 0
let activeController = null

const hasCampaigns = computed(() => campaigns.value.length > 0)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / rows.value)))

function scheduleCampaignLoad(delay = 250) {
  window.clearTimeout(loadTimer)
  loadTimer = window.setTimeout(() => {
    loadCampaigns()
  }, delay)
}

function resetFilters() {
  const alreadyReset = !filters.label && !filters.status && !filters.dateFrom && !filters.dateTo

  page.value = 1
  filters.label = ''
  filters.status = ''
  filters.dateFrom = ''
  filters.dateTo = ''

  if (alreadyReset) {
    scheduleCampaignLoad(0)
  }
}

function abortActiveRequest() {
  if (activeController) {
    activeController.abort()
    activeController = null
  }
}

function supportedEventsForTarget(target) {
  return Array.isArray(target?.events)
    ? target.events.filter((event) => EVENT_TYPES.includes(event?.type))
    : []
}

function hasAnyEvent(target, eventTypes) {
  return supportedEventsForTarget(target).some((event) => eventTypes.includes(event.type))
}

function percent(numerator, denominator) {
  return denominator ? Math.round((numerator / denominator) * 100) : null
}

function formatPercent(value) {
  return formatLocalizedPercent(value)
}

function buildCampaignMetrics(vtargets) {
  return (vtargets || []).reduce((acc, target) => {
    const campaignId = String(target?.campaign_id || '')
    if (!campaignId) {
      return acc
    }

    if (!acc[campaignId]) {
      acc[campaignId] = {
        targets: 0,
        sent: 0,
        opened: 0,
        clicked: 0,
        submitted: 0,
      }
    }

    acc[campaignId].targets += 1
    acc[campaignId].sent += hasAnyEvent(target, SENT_OR_LATER_EVENTS) ? 1 : 0
    acc[campaignId].opened += hasAnyEvent(target, ['EMAIL_OPENED']) ? 1 : 0
    acc[campaignId].clicked += hasAnyEvent(target, ['LINK_OPENED']) ? 1 : 0
    acc[campaignId].submitted += hasAnyEvent(target, ['DATA_SENT']) ? 1 : 0

    return acc
  }, {})
}

function metricForCampaign(campaign) {
  return campaignMetrics.value[String(campaign?.id || '')] || null
}

function displayMetricCount(campaign) {
  if (isMetricsLoading.value || metricsError.value) {
    return '—'
  }

  const metrics = metricForCampaign(campaign)
  return metrics ? String(metrics.targets) : '0'
}

function displayMetricRate(campaign, key) {
  if (isMetricsLoading.value || metricsError.value) {
    return '—'
  }

  const metrics = metricForCampaign(campaign)
  if (!metrics) {
    return '—'
  }

  return formatPercent(percent(metrics[key], metrics.sent))
}

function isAssetPreviewLoading(campaign, type) {
  return assetPreviewLoadingKey.value === `${type}-${campaign?.id || ''}`
}

function assetButtonTitle(campaign, type) {
  if (type === 'message') {
    return campaign?.message_id ? t('pages.campaigns.preview.message') : t('pages.campaigns.preview.messageMissing')
  }

  if (type === 'landing') {
    return campaign?.landing_id ? t('pages.campaigns.preview.landing') : t('pages.campaigns.preview.landingMissing')
  }

  return campaign?.education_id ? t('pages.campaigns.preview.education') : t('pages.campaigns.preview.educationMissing')
}

function closeAssetPreview() {
  assetPreviewOpen.value = false
  assetPreviewAttachment.value = null
  assetPreviewMeta.value = []
  assetPreviewKind.value = ''
}

function openAssetPreview({ title, kind, attachmentId, attachmentLabel, meta }) {
  assetPreviewTitle.value = title
  assetPreviewKind.value = kind
  assetPreviewAttachment.value = {
    id: attachmentId,
    label: attachmentLabel,
    type: '.html',
  }
  assetPreviewMeta.value = meta
  assetPreviewOpen.value = true
}

async function previewMessageAsset(campaign) {
  const id = nullableId(campaign?.message_id)
  if (!id) {
    return
  }

  const key = `message-${campaign.id || id}`
  assetPreviewLoadingKey.value = key

  try {
    const message = await getMessage(id)
    const htmlBodyId = nullableId(message?.html_body_id)
    if (!htmlBodyId) {
      notifyError(t('common.preview.unavailable'), t('pages.campaigns.preview.messageNoHtml'))
      return
    }

    openAssetPreview({
      title: t('pages.campaigns.preview.message'),
      kind: t('common.labels.message'),
      attachmentId: htmlBodyId,
      attachmentLabel: message.label || t('common.labels.htmlBody'),
      meta: [
        [t('common.labels.name'), message.label || t('common.placeholder')],
        [t('common.labels.fromEmail'), message.from_email || t('common.placeholder')],
        [t('common.labels.fromName'), message.from_name || t('common.placeholder')],
        [t('common.labels.subject'), message.subject || t('common.placeholder')],
      ],
    })
  } catch (err) {
    notifyError(t('common.preview.failed'), err?.message || t('pages.campaigns.preview.loadFailed'))
  } finally {
    if (assetPreviewLoadingKey.value === key) {
      assetPreviewLoadingKey.value = ''
    }
  }
}

async function previewLandingAsset(campaign, type = 'landing') {
  const id = type === 'education'
    ? nullableId(campaign?.education_id)
    : nullableId(campaign?.landing_id)
  if (!id) {
    return
  }

  const key = `${type}-${campaign.id || id}`
  assetPreviewLoadingKey.value = key

  try {
    const landing = await getLanding(id)
    const htmlBodyId = nullableId(landing?.html_body_id)
    if (!htmlBodyId) {
      notifyError(t('common.preview.unavailable'), t('pages.campaigns.preview.landingNoHtml'))
      return
    }

    openAssetPreview({
      title: type === 'education' ? t('pages.campaigns.preview.education') : t('pages.campaigns.preview.landing'),
      kind: type === 'education' ? t('common.labels.educationAsset') : t('common.labels.landingPage'),
      attachmentId: htmlBodyId,
      attachmentLabel: landing.label || t('common.labels.htmlBody'),
      meta: [
        [t('common.labels.name'), landing.label || t('common.placeholder')],
      ],
    })
  } catch (err) {
    notifyError(t('common.preview.failed'), err?.message || t('pages.campaigns.preview.loadFailed'))
  } finally {
    if (assetPreviewLoadingKey.value === key) {
      assetPreviewLoadingKey.value = ''
    }
  }
}

async function loadCampaigns() {
  const requestId = requestSequence + 1
  requestSequence = requestId
  abortActiveRequest()

  const controller = new AbortController()
  activeController = controller
  isLoading.value = true
  isMetricsLoading.value = false
  error.value = ''
  metricsError.value = ''
  authSuppressed.value = false
  campaignMetrics.value = {}

  try {
    const response = await listCampaigns({
      label: filters.label,
      status: filters.status,
      dateFrom: filters.dateFrom,
      dateTo: filters.dateTo,
      page: page.value,
      rows: rows.value,
    }, {
      signal: controller.signal,
    })

    if (requestId !== requestSequence) {
      return
    }

    campaigns.value = Array.isArray(response?.items) ? response.items : []
    total.value = Number.isFinite(response?.total) ? response.total : campaigns.value.length
    page.value = Number.isFinite(response?.page) ? response.page : page.value
    rows.value = Number.isFinite(response?.rowsPerPage) ? response.rowsPerPage : rows.value
    hasLoaded.value = true

    if (requestId === requestSequence) {
      isLoading.value = false
    }

    isMetricsLoading.value = true

    try {
      const vtargets = await loadAllPages(listVTargets, {}, {
        signal: controller.signal,
      })

      if (requestId !== requestSequence) {
        return
      }

      campaignMetrics.value = buildCampaignMetrics(vtargets)
    } catch (metricsLoadError) {
      if (metricsLoadError?.name === 'AbortError' || requestId !== requestSequence) {
        return
      }

       metricsError.value = errorMessage(metricsLoadError, 'pages.campaigns.errors.metricsLoad')
    } finally {
      if (requestId === requestSequence) {
        isMetricsLoading.value = false
      }
    }
  } catch (err) {
    if (err?.name === 'AbortError' || requestId !== requestSequence) {
      return
    }

    campaigns.value = []
    campaignMetrics.value = {}
    hasLoaded.value = true

    if (isCampaignAuthError(err)) {
      authSuppressed.value = true
      error.value = ''
      await loadCurrentSession()
      return
    }

    error.value = errorMessage(err, 'pages.campaigns.errors.load')
  } finally {
    if (requestId === requestSequence) {
      isLoading.value = false
      isMetricsLoading.value = false
      activeController = null
    }
  }
}

function retryLoad() {
  scheduleCampaignLoad(0)
}

function setPage(nextPage) {
  page.value = Math.min(Math.max(1, nextPage), totalPages.value)
  scheduleCampaignLoad(0)
}

function setRows(nextRows) {
  rows.value = nextRows
  page.value = 1
  scheduleCampaignLoad(0)
}

function statusVariant(status) {
  switch (status) {
    case 'ACTIVE':
      return 'success'
    case 'PAUSED':
      return 'warning'
    case 'COMPLETED':
      return 'info'
    case 'CANCELED':
      return 'danger'
    default:
      return 'neutral'
  }
}

function displayStatus(status) {
  return CAMPAIGN_STATUSES.includes(status) ? translateCampaignStatus(status) : t('common.placeholder')
}

function formatDateRange(campaign) {
  return formatLocalizedDateRange(campaign.date_from, campaign.date_to)
}

watch(
  () => [filters.label, filters.status, filters.dateFrom, filters.dateTo],
  () => {
    page.value = 1
    scheduleCampaignLoad()
  },
)

onMounted(() => {
  loadCampaigns()
})

onBeforeUnmount(() => {
  window.clearTimeout(loadTimer)
  requestSequence += 1
  abortActiveRequest()
})
</script>

<template>
  <div class="campaigns-page">
    <PageHeader
      :eyebrow="t('sections.operations')"
      :title="t('pages.campaigns.title')"
      :description="t('pages.campaigns.description')"
    >
      <template #actions>
        <RouterLink class="ui-button ui-button--primary" to="/campaigns/new">
          {{ t('common.actions.createCampaign') }}
        </RouterLink>
      </template>
    </PageHeader>

    <section class="campaign-filter-panel" :aria-label="t('pages.campaigns.filtersAria')">
      <UiInput
        id="campaign-label-search"
        v-model="filters.label"
        :label="t('pages.campaigns.searchLabel')"
        :placeholder="t('pages.campaigns.searchPlaceholder')"
        autocomplete="off"
      />

      <label class="ui-field" for="campaign-status-filter">
        <span class="ui-field-label">{{ t('common.labels.status') }}</span>
        <select id="campaign-status-filter" v-model="filters.status" class="ui-select">
          <option value="">{{ t('common.filters.anyStatus') }}</option>
          <option v-for="status in supportedStatuses" :key="status" :value="status">
            {{ translateCampaignStatus(status) }}
          </option>
        </select>
      </label>

      <UiDatePicker
        id="campaign-date-from"
        v-model="filters.dateFrom"
        :label="t('common.labels.dateFrom')"
      />

      <UiDatePicker
        id="campaign-date-to"
        v-model="filters.dateTo"
        :label="t('common.labels.dateTo')"
      />

      <UiButton class="campaign-filter-reset" variant="secondary" @click="resetFilters">
        {{ t('common.actions.reset') }}
      </UiButton>
    </section>

    <section class="campaign-table-panel" aria-labelledby="campaign-table-title">
      <header class="campaign-table-header">
        <div>
          <h2 id="campaign-table-title">{{ t('pages.campaigns.tableTitle') }}</h2>
          <p>
            {{ t('pages.campaigns.tableDescription') }}
          </p>
        </div>
      </header>

      <div v-if="isLoading || hasCampaigns" class="campaigns-table" role="table" :aria-label="t('pages.campaigns.tableAria')">
        <div class="campaigns-table-row campaigns-table-head" role="row">
            <span v-for="column in columns" :key="column" role="columnheader">
              {{ column }}
            </span>
        </div>

        <template v-if="isLoading">
          <div
            v-for="row in 5"
            :key="`loading-${row}`"
            class="campaigns-table-row"
            role="row"
          >
            <span v-for="column in columns" :key="column" role="cell">
              <SkeletonBlock :rows="1" />
            </span>
          </div>
        </template>

        <template v-else>
          <div
            v-for="campaign in campaigns"
            :key="campaign.id"
            class="campaigns-table-row"
            role="row"
          >
            <span role="cell">
              <span class="campaigns-table-value campaigns-table-value--strong">
                {{ campaign.label || t('common.placeholder') }}
              </span>
            </span>
            <span role="cell">
              <UiBadge :variant="statusVariant(campaign.status)">
                {{ displayStatus(campaign.status) }}
              </UiBadge>
            </span>
            <span role="cell">
              <span class="campaigns-table-value" :title="campaign.domain || undefined">
                {{ campaign.domain || t('common.placeholder') }}
              </span>
            </span>
            <span role="cell">
              <span class="campaigns-table-value">
                {{ formatDateRange(campaign) }}
              </span>
            </span>
            <span role="cell" class="campaign-assets-cell">
              <button
                type="button"
                class="campaign-asset-button"
                :disabled="!campaign.message_id || isAssetPreviewLoading(campaign, 'message')"
                :title="assetButtonTitle(campaign, 'message')"
                :aria-label="t('pages.campaigns.preview.message')"
                @click="previewMessageAsset(campaign)"
              >
                <Eye :size="14" stroke-width="1.8" aria-hidden="true" />
                {{ t('common.labels.message') }}
              </button>
              <button
                type="button"
                class="campaign-asset-button"
                :disabled="!campaign.landing_id || isAssetPreviewLoading(campaign, 'landing')"
                :title="assetButtonTitle(campaign, 'landing')"
                :aria-label="t('pages.campaigns.preview.landing')"
                @click="previewLandingAsset(campaign)"
              >
                <Eye :size="14" stroke-width="1.8" aria-hidden="true" />
                {{ t('common.labels.landing') }}
              </button>
              <button
                type="button"
                class="campaign-asset-button"
                :disabled="!campaign.education_id || isAssetPreviewLoading(campaign, 'education')"
                :title="assetButtonTitle(campaign, 'education')"
                :aria-label="t('pages.campaigns.preview.education')"
                @click="previewLandingAsset(campaign, 'education')"
              >
                <Eye :size="14" stroke-width="1.8" aria-hidden="true" />
                {{ t('common.labels.education') }}
              </button>
            </span>
            <span role="cell">
              <span class="campaigns-table-derived">{{ displayMetricCount(campaign) }}</span>
            </span>
            <span role="cell">
              <span class="campaigns-table-derived">{{ displayMetricRate(campaign, 'opened') }}</span>
            </span>
            <span role="cell">
              <span class="campaigns-table-derived">{{ displayMetricRate(campaign, 'clicked') }}</span>
            </span>
            <span role="cell">
              <span class="campaigns-table-derived">{{ displayMetricRate(campaign, 'submitted') }}</span>
            </span>
            <span role="cell" class="campaign-actions-cell">
              <RouterLink
                v-if="campaign.id"
                class="icon-button icon-button--secondary"
                :to="`/campaigns/${campaign.id}`"
                :aria-label="t('common.actions.view')"
                :title="t('common.actions.view')"
              >
                <Eye :size="16" stroke-width="1.8" aria-hidden="true" />
              </RouterLink>
            </span>
          </div>
        </template>
      </div>

      <div v-else-if="authSuppressed" class="campaign-inline-state">
        <SkeletonBlock :rows="3" />
      </div>

      <div v-else-if="error" class="campaign-table-error" role="status">
        <div>
          <h2>{{ t('pages.campaigns.errors.load') }}</h2>
          <p>{{ error }}</p>
        </div>
        <UiButton variant="secondary" @click="retryLoad">
          {{ t('common.actions.retry') }}
        </UiButton>
      </div>

      <EmptyState
        v-else-if="hasLoaded"
        :title="t('pages.campaigns.emptyTitle')"
        :description="t('pages.campaigns.emptyDescription')"
      />

      <ResourcePagination
        v-if="(hasLoaded || isLoading) && !authSuppressed && !error"
        :page="page"
        :rows="rows"
        :total="total"
        :loaded-count="campaigns.length"
        :loading="isLoading"
        @update:page="setPage"
        @update:rows="setRows"
      />
    </section>

    <PreviewDrawer
      :open="assetPreviewOpen"
      :title="assetPreviewTitle"
      :subtitle="t('common.preview.rawTemplate')"
      @close="closeAssetPreview"
    >
      <section class="campaign-asset-preview-meta">
        <div>
          <span>{{ t('common.labels.resourceType') }}</span>
          <strong>{{ assetPreviewKind || t('common.placeholder') }}</strong>
        </div>
        <div v-for="[label, value] in assetPreviewMeta" :key="label">
          <span>{{ label }}</span>
          <strong>{{ value }}</strong>
        </div>
      </section>
      <HtmlAttachmentPreview
        :attachment="assetPreviewAttachment"
        :attachment-id="assetPreviewAttachment?.id || ''"
        :title="t('common.preview.campaignResource')"
        :empty-text="t('pages.campaigns.preview.selectResource')"
      />
    </PreviewDrawer>
  </div>
</template>
