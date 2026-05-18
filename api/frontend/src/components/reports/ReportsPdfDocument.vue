<script setup>
import { computed, nextTick, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseEChart from '../charts/BaseEChart.vue'
import { EVENT_LABELS } from '../../resources/vtargets'

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    default: '',
  },
  dateRangeLabel: {
    type: String,
    required: true,
  },
  campaignLabel: {
    type: String,
    required: true,
  },
  departmentLabel: {
    type: String,
    required: true,
  },
  generatedAtLabel: {
    type: String,
    required: true,
  },
  kpis: {
    type: Array,
    default: () => [],
  },
  charts: {
    type: Array,
    default: () => [],
  },
  topEmployees: {
    type: Array,
    default: () => [],
  },
})

const { t } = useI18n()

const rootEl = ref(null)
const chartImages = ref({})
const chartRefs = new Map()
const renderableCharts = computed(() => props.charts.filter((chart) => !chart.empty))

function waitForFrame() {
  return new Promise((resolve) => {
    requestAnimationFrame(resolve)
  })
}

function setChartRef(key, element) {
  if (element) {
    chartRefs.set(key, element)
    return
  }

  chartRefs.delete(key)
}

function chartHeight(chart) {
  return chart.height || '300px'
}

function chartRendererStyle(chart) {
  return {
    width: '730px',
    height: chartHeight(chart),
  }
}

function chartFallbackText(chart) {
  return chart.empty ? chart.emptyText : t('pages.reports.chartExportUnavailable')
}

async function waitForImages() {
  const images = Array.from(rootEl.value?.querySelectorAll('img') || [])

  await Promise.all(images.map((image) => {
    if (image.complete) {
      return Promise.resolve()
    }

    return new Promise((resolve) => {
      image.addEventListener('load', resolve, { once: true })
      image.addEventListener('error', resolve, { once: true })
    })
  }))
}

async function prepareCharts() {
  const nextImages = {}

  await nextTick()
  await waitForFrame()
  await waitForFrame()

  renderableCharts.value.forEach((chart) => {
    chartRefs.get(chart.key)?.resizeNow?.()
  })

  await waitForFrame()

  props.charts.forEach((chart) => {
    if (chart.empty) {
      nextImages[chart.key] = ''
      return
    }

    const dataURL = chartRefs.get(chart.key)?.getDataURL?.({
      type: 'png',
      pixelRatio: 2,
      backgroundColor: '#ffffff',
    })

    nextImages[chart.key] = typeof dataURL === 'string' && dataURL.startsWith('data:image/')
      ? dataURL
      : ''
  })

  chartImages.value = nextImages
  await nextTick()
  await waitForImages()
}

function getRootElement() {
  return rootEl.value
}

defineExpose({
  getRootElement,
  prepareCharts,
  rootEl,
  waitForImages,
})
</script>

<template>
  <div class="reports-pdf-export-stage" aria-hidden="true">
    <article ref="rootEl" class="reports-pdf-document">
      <header class="reports-pdf-header">
        <span>{{ t('sections.analytics') }}</span>
        <h1>{{ title }}</h1>
        <p v-if="description">{{ description }}</p>
      </header>

      <section class="reports-pdf-section reports-pdf-summary" :aria-label="t('pages.reports.currentScope')">
        <div>
          <span>{{ t('common.labels.timeRange') }}</span>
          <strong>{{ dateRangeLabel }}</strong>
        </div>
        <div>
          <span>{{ t('common.labels.campaign') }}</span>
          <strong>{{ campaignLabel }}</strong>
        </div>
        <div>
          <span>{{ t('common.labels.department') }}</span>
          <strong>{{ departmentLabel }}</strong>
        </div>
        <div>
          <span>{{ t('common.labels.generatedAt') }}</span>
          <strong>{{ generatedAtLabel }}</strong>
        </div>
      </section>

      <section class="reports-pdf-section">
        <div class="reports-pdf-kpi-grid">
          <article v-for="metric in kpis" :key="metric.label" class="reports-pdf-kpi-card">
            <span>{{ metric.label }}</span>
            <strong>{{ metric.value }}</strong>
            <p v-if="metric.description">{{ metric.description }}</p>
          </article>
        </div>
      </section>

      <section class="reports-pdf-section reports-pdf-chart-list">
        <article v-for="chart in charts" :key="chart.key" class="reports-pdf-chart-card">
          <header>
            <h2>{{ chart.title }}</h2>
            <p v-if="chart.description">{{ chart.description }}</p>
          </header>
          <img
            v-if="chartImages[chart.key]"
            class="reports-pdf-chart-image"
            :src="chartImages[chart.key]"
            :alt="chart.title"
          >
          <div v-else class="reports-pdf-chart-fallback">
            {{ chartFallbackText(chart) }}
          </div>
        </article>
      </section>

      <section class="reports-pdf-section reports-pdf-table-section">
        <header class="reports-pdf-section-header">
          <h2>{{ t('pages.reports.charts.topEmployeesTitle') }}</h2>
          <p>{{ t('pages.reports.charts.topEmployeesDescription') }}</p>
        </header>

        <table class="reports-pdf-table">
          <thead>
            <tr>
              <th>{{ t('common.labels.employee') }}</th>
              <th>{{ t('common.labels.department') }}</th>
              <th>{{ t('common.labels.riskIndex') }}</th>
              <th>{{ EVENT_LABELS.LINK_OPENED }}</th>
              <th>{{ EVENT_LABELS.DATA_SENT }}</th>
              <th>{{ t('common.labels.activity') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="employee in topEmployees" :key="employee.id">
              <td><strong>{{ employee.label }}</strong></td>
              <td>{{ employee.departmentLabel }}</td>
              <td>{{ employee.risk }}</td>
              <td>{{ employee.linkOpenedCount }}</td>
              <td>{{ employee.dataSubmittedCount }}</td>
              <td>{{ employee.lastActivityLabel }}</td>
            </tr>
            <tr v-if="!topEmployees.length">
              <td colspan="6">{{ t('pages.reports.charts.topEmployeesEmpty') }}</td>
            </tr>
          </tbody>
        </table>
      </section>
    </article>

    <div class="reports-pdf-chart-renderers">
      <div
        v-for="chart in renderableCharts"
        :key="chart.key"
        class="reports-pdf-chart-renderer"
        :style="chartRendererStyle(chart)"
      >
        <BaseEChart
          :ref="(element) => setChartRef(chart.key, element)"
          :option="chart.option"
          :empty="false"
          :empty-text="chart.emptyText"
          :height="chartHeight(chart)"
          :aria-label="chart.title"
        />
      </div>
    </div>
  </div>
</template>
