<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Activity,
  AlertTriangle,
  Building2,
  CalendarDays,
  Database,
  ExternalLink,
  MailOpen,
  MousePointerClick,
  PieChart,
  Send,
  ShieldAlert,
} from 'lucide-vue-next'
import DashboardActivityCalendarChart from '../components/charts/DashboardActivityCalendarChart.vue'
import DashboardEventDistributionChart from '../components/charts/DashboardEventDistributionChart.vue'
import DashboardTopDepartmentsChart from '../components/charts/DashboardTopDepartmentsChart.vue'
import MetricCard from '../components/dashboard/MetricCard.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import { useResourceActions } from '../composables/useResourceActions'
import { translateCampaignStatus } from '../i18n'
import { listCampaigns } from '../resources/campaigns'
import { listDepartments } from '../resources/departments'
import { listEmployees } from '../resources/employees'
import { listVTargets } from '../resources/vtargets'
import { buildDashboardAnalytics } from '../utils/dashboardAnalytics'
import { errorMessage } from '../utils/resourceFormat'
import { loadAllPages } from '../utils/resourcePagination'

const selectedYear = ref(new Date().getFullYear())
const campaigns = ref([])
const vtargets = ref([])
const employees = ref([])
const departments = ref([])
const isLoading = ref(false)
const loadError = ref('')
const partialErrors = ref([])

const { handleAuthError } = useResourceActions()
const { t, locale } = useI18n()

const hasEmployeeJoin = computed(() => !partialErrors.value.some((item) => item.source === 'employees'))
const hasDepartmentJoin = computed(() => (
  hasEmployeeJoin.value
  && !partialErrors.value.some((item) => item.source === 'departments')
))

const dashboard = computed(() => {
  locale.value

  return buildDashboardAnalytics({
    campaigns: campaigns.value,
    vtargets: vtargets.value,
    employees: hasEmployeeJoin.value ? employees.value : [],
    departments: hasDepartmentJoin.value ? departments.value : [],
    selectedYear: selectedYear.value,
  })
})

const availableYears = computed(() => dashboard.value.availableYears)
const noActivity = computed(() => !isLoading.value && !dashboard.value.hasActivity)

const metricCards = computed(() => [
  {
    label: t('pages.dashboard.metrics.activeCampaigns.label'),
    value: dashboard.value.kpis.activeCampaigns,
    description: t('pages.dashboard.metrics.activeCampaigns.description'),
    icon: Activity,
  },
  {
    label: t('pages.dashboard.metrics.riskIndex.label'),
    value: dashboard.value.kpis.riskIndex,
    description: t('pages.dashboard.metrics.riskIndex.description', { year: selectedYear.value }),
    icon: ShieldAlert,
    help: true,
  },
  {
    label: t('pages.dashboard.metrics.openRate.label'),
    value: dashboard.value.kpis.openRate,
    description: t('pages.dashboard.metrics.openRate.description'),
    icon: MailOpen,
  },
  {
    label: t('pages.dashboard.metrics.clickRate.label'),
    value: dashboard.value.kpis.clickRate,
    description: t('pages.dashboard.metrics.clickRate.description'),
    icon: MousePointerClick,
  },
  {
    label: t('pages.dashboard.metrics.submitRate.label'),
    value: dashboard.value.kpis.submitRate,
    description: t('pages.dashboard.metrics.submitRate.description'),
    icon: Send,
  },
])

const riskTooltip = computed(() => t('pages.dashboard.riskTooltip'))

function eventIcon(eventType) {
  switch (eventType) {
    case 'MESSAGE_SENT':
      return Send
    case 'DELIVERY_FAILED':
      return AlertTriangle
    case 'EMAIL_OPENED':
      return MailOpen
    case 'LINK_OPENED':
      return MousePointerClick
    case 'DATA_SENT':
      return Database
    default:
      return Activity
  }
}

function eventClass(eventType) {
  return `dashboard-event-icon--${String(eventType || '').toLowerCase()}`
}

function statusVariant(status) {
  switch (status) {
    case 'ACTIVE':
      return 'success'
    case 'PAUSED':
      return 'warning'
    case 'CANCELED':
      return 'danger'
    case 'COMPLETED':
      return 'info'
    case 'DRAFT':
      return 'neutral'
    default:
      return 'neutral'
  }
}

async function loadDashboard() {
  isLoading.value = true
  loadError.value = ''
  partialErrors.value = []

  try {
    const [campaignResult, targetResult, employeeResult, departmentResult] = await Promise.allSettled([
      loadAllPages(listCampaigns),
      loadAllPages(listVTargets),
      loadAllPages(listEmployees),
      loadAllPages(listDepartments),
    ])

    if (campaignResult.status === 'rejected') {
      if (await handleAuthError(campaignResult.reason)) {
        return
      }
      throw campaignResult.reason
    }

    if (targetResult.status === 'rejected') {
      if (await handleAuthError(targetResult.reason)) {
        return
      }
      throw targetResult.reason
    }

    campaigns.value = campaignResult.value
    vtargets.value = targetResult.value

    if (employeeResult.status === 'fulfilled') {
      employees.value = employeeResult.value
    } else {
        partialErrors.value.push({
          source: 'employees',
          message: errorMessage(employeeResult.reason, 'pages.dashboard.employeeLookupUnavailable'),
        })
      employees.value = []
    }

    if (departmentResult.status === 'fulfilled') {
      departments.value = departmentResult.value
    } else {
        partialErrors.value.push({
          source: 'departments',
          message: errorMessage(departmentResult.reason, 'pages.dashboard.departmentLookupUnavailable'),
        })
      departments.value = []
    }
  } catch (error) {
    loadError.value = errorMessage(error, 'errors.dashboardLoad')
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadDashboard()
})
</script>

<template>
  <div class="dashboard-page">
    <PageHeader
      :eyebrow="t('sections.operations')"
      :title="t('pages.dashboard.title')"
      :description="t('pages.dashboard.description')"
    >
      <template #actions>
        <RouterLink class="ui-button ui-button--secondary" to="/reports">
          {{ t('common.actions.viewReports') }}
        </RouterLink>
        <RouterLink class="ui-button ui-button--primary" to="/campaigns/new">
          {{ t('common.actions.createCampaign') }}
        </RouterLink>
      </template>
    </PageHeader>

    <UiAlert
      v-if="loadError"
      variant="error"
      :title="t('pages.dashboard.unavailable')"
      :message="loadError"
    >
      <UiButton variant="secondary" @click="loadDashboard">{{ t('common.actions.retry') }}</UiButton>
    </UiAlert>

    <UiAlert
      v-for="partialError in partialErrors"
      :key="partialError.source"
      variant="warning"
      :title="t('pages.dashboard.partialData')"
      :message="partialError.message"
    />

    <template v-if="isLoading">
      <section class="metrics-grid" :aria-label="t('pages.dashboard.metricsLoading')">
        <article v-for="index in 5" :key="index" class="metric-card">
          <SkeletonBlock :rows="3" />
        </article>
      </section>
      <section class="dashboard-panel">
        <SkeletonBlock :rows="8" />
      </section>
      <section class="dashboard-chart-row">
        <article class="dashboard-panel"><SkeletonBlock :rows="6" /></article>
        <article class="dashboard-panel"><SkeletonBlock :rows="6" /></article>
      </section>
      <section class="dashboard-operational-grid">
        <article class="dashboard-panel"><SkeletonBlock :rows="8" /></article>
        <article class="dashboard-panel"><SkeletonBlock :rows="8" /></article>
      </section>
    </template>

    <template v-else-if="!loadError">
      <section class="metrics-grid" :aria-label="t('pages.dashboard.metrics')">
        <MetricCard
          v-for="metric in metricCards"
          :key="metric.label"
          :label="metric.label"
          :value="metric.value"
          :description="metric.description"
          :icon="metric.icon"
        >
          <template v-if="metric.help" #label-actions>
            <span class="dashboard-help">
              <ShieldAlert :size="14" stroke-width="1.8" aria-hidden="true" />
              <span class="dashboard-help-tooltip" role="tooltip">{{ riskTooltip }}</span>
            </span>
          </template>
        </MetricCard>
      </section>

      <UiAlert
        v-if="noActivity"
        variant="info"
        :title="t('pages.dashboard.noActivityTitle')"
        :message="t('pages.dashboard.noActivityMessage', { year: selectedYear })"
      />

      <section class="dashboard-panel dashboard-calendar-panel" aria-labelledby="dashboard-activity-calendar-title">
        <header class="dashboard-panel-header dashboard-calendar-header">
          <div>
            <h2 id="dashboard-activity-calendar-title">{{ t('pages.dashboard.activityCalendarTitle') }}</h2>
            <p>{{ t('pages.dashboard.activityCalendarDescription') }}</p>
          </div>
          <label class="dashboard-year-select">
            <span>{{ t('pages.dashboard.year') }}</span>
            <select v-model.number="selectedYear" class="ui-select">
              <option v-for="year in availableYears" :key="year" :value="year">
                {{ year }}
              </option>
            </select>
          </label>
        </header>
        <div class="dashboard-panel-body">
          <DashboardActivityCalendarChart
            :year="selectedYear"
            :days="dashboard.calendarDays"
            :empty-text="t('pages.dashboard.noActivityMessage', { year: selectedYear })"
          />
        </div>
      </section>

      <section class="dashboard-chart-row" :aria-label="t('pages.dashboard.compactAnalytics')">
        <article class="dashboard-panel">
          <header class="dashboard-panel-header">
            <div>
              <h2>{{ t('pages.dashboard.eventDistributionTitle') }}</h2>
              <p>{{ t('pages.dashboard.eventDistributionDescription') }}</p>
            </div>
            <PieChart :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>
          <div class="dashboard-panel-body">
            <DashboardEventDistributionChart
              :items="dashboard.eventDistribution"
              :empty-text="t('charts.eventDistribution.empty')"
            />
          </div>
        </article>

        <article class="dashboard-panel">
          <header class="dashboard-panel-header">
            <div>
              <h2>{{ t('pages.dashboard.topDepartmentsTitle') }}</h2>
              <p>{{ t('pages.dashboard.topDepartmentsDescription') }}</p>
            </div>
            <Building2 :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>
          <div class="dashboard-panel-body">
            <DashboardTopDepartmentsChart
              :departments="hasDepartmentJoin ? dashboard.topDepartments : []"
              :empty-text="t('charts.departments.empty')"
            />
          </div>
        </article>
      </section>

      <section class="dashboard-operational-grid" :aria-label="t('pages.dashboard.operationalData')">
        <section class="dashboard-panel active-campaigns-preview" aria-labelledby="active-campaigns-title">
          <header class="dashboard-panel-header">
            <div>
              <h2 id="active-campaigns-title">{{ t('pages.dashboard.activeCampaignsTitle') }}</h2>
              <p>{{ t('pages.dashboard.activeCampaignsDescription') }}</p>
            </div>
            <CalendarDays :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>

          <div v-if="dashboard.activeCampaignRows.length" class="campaign-preview-table" role="table" :aria-label="t('pages.dashboard.activeCampaignsAria')">
            <div class="campaign-preview-row campaign-preview-head" role="row">
              <span role="columnheader">{{ t('common.labels.campaign') }}</span>
              <span role="columnheader">{{ t('common.labels.status') }}</span>
              <span role="columnheader">{{ t('common.labels.dateRange') }}</span>
              <span role="columnheader">{{ t('common.labels.targets') }}</span>
              <span role="columnheader">{{ t('pages.dashboard.columns.opens') }}</span>
              <span role="columnheader">{{ t('pages.dashboard.columns.clicks') }}</span>
              <span role="columnheader">{{ t('pages.dashboard.columns.submissions') }}</span>
              <span role="columnheader">{{ t('pages.dashboard.columns.actions') }}</span>
            </div>

            <div
              v-for="campaign in dashboard.activeCampaignRows"
              :key="campaign.id"
              class="campaign-preview-row"
              role="row"
            >
              <span role="cell">
                <strong class="dashboard-table-strong" :title="campaign.label">{{ campaign.label }}</strong>
              </span>
              <span role="cell">
                <UiBadge :variant="statusVariant(campaign.status)">{{ translateCampaignStatus(campaign.status) }}</UiBadge>
              </span>
              <span role="cell">{{ campaign.dateRange }}</span>
              <span role="cell">{{ campaign.targetCount }}</span>
              <span role="cell">{{ campaign.openRate }}</span>
              <span role="cell">{{ campaign.clickRate }}</span>
              <span role="cell">{{ campaign.submitRate }}</span>
              <span role="cell">
                <RouterLink class="ui-button ui-button--secondary dashboard-table-action" :to="`/campaigns/${campaign.id}`">
                  <ExternalLink :size="14" stroke-width="1.8" aria-hidden="true" />
                  {{ t('common.actions.view') }}
                </RouterLink>
              </span>
            </div>
          </div>

          <EmptyState
            v-else
            :title="t('pages.dashboard.noActiveCampaignsTitle')"
            :description="t('pages.dashboard.noActiveCampaignsDescription')"
          >
            <RouterLink class="ui-button ui-button--primary" to="/campaigns/new">
              {{ t('common.actions.createCampaign') }}
            </RouterLink>
            <RouterLink class="ui-button ui-button--secondary" to="/campaigns">
              {{ t('common.actions.viewCampaigns') }}
            </RouterLink>
          </EmptyState>
        </section>

        <section class="dashboard-panel event-timeline" aria-labelledby="recent-events-title">
          <header class="dashboard-panel-header">
            <div>
              <h2 id="recent-events-title">{{ t('pages.dashboard.recentEventsTitle') }}</h2>
              <p>{{ t('pages.dashboard.recentEventsDescription') }}</p>
            </div>
            <Activity :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>

          <ol v-if="dashboard.recentEvents.length" class="event-list">
            <li v-for="event in dashboard.recentEvents" :key="event.key" class="event-item">
              <span class="dashboard-event-icon" :class="eventClass(event.eventType)" aria-hidden="true">
                <component :is="eventIcon(event.eventType)" :size="15" stroke-width="1.9" />
              </span>
              <span class="dashboard-event-copy">
                <strong>{{ event.label }}</strong>
                <span>{{ event.employeeLabel }} · {{ event.campaignLabel }}</span>
              </span>
              <time class="event-placeholder" :datetime="event.occurredAt">
                {{ event.occurredAtLabel }}
              </time>
            </li>
          </ol>

          <EmptyState
            v-else
            :title="t('pages.dashboard.noRecentEventsTitle')"
            :description="t('pages.dashboard.noRecentEventsDescription', { year: selectedYear })"
          />
        </section>
      </section>
    </template>
  </div>
</template>
