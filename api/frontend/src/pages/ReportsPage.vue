<script setup>
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Activity,
  BarChart3,
  Building2,
  Gauge,
  Info,
  MailOpen,
  MousePointerClick,
  Printer,
  Send,
  ShieldAlert,
  TrendingUp,
  Users,
} from 'lucide-vue-next'
import BaseEChart from '../components/charts/BaseEChart.vue'
import { chartTokens, eventChartColor, riskValueColor } from '../components/charts/chartPalette'
import MetricCard from '../components/dashboard/MetricCard.vue'
import ReportsPdfDocument from '../components/reports/ReportsPdfDocument.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiDatePicker from '../components/ui/UiDatePicker.vue'
import { useNotifications } from '../composables/useNotifications'
import { useResourceActions } from '../composables/useResourceActions'
import { listCampaigns } from '../resources/campaigns'
import { listDepartments } from '../resources/departments'
import { listEmployees } from '../resources/employees'
import { EVENT_LABELS, EVENT_TYPES, listVTargets } from '../resources/vtargets'
import { errorMessage, formatDateTime } from '../utils/resourceFormat'
import { loadAllPages } from '../utils/resourcePagination'
import {
  buildReportAnalytics,
} from '../utils/reportAnalytics'

const filters = reactive({
  dateFrom: '',
  dateTo: '',
  campaignId: '',
  departmentId: '',
})
const appliedFilters = ref({ ...filters })
const campaigns = ref([])
const vtargets = ref([])
const employees = ref([])
const departments = ref([])
const isLoading = ref(false)
const loadError = ref('')
const partialErrors = ref([])
const generatedAt = ref(new Date())
const isPdfExporting = ref(false)
const pdfDocumentRef = ref(null)

const { handleAuthError } = useResourceActions()
const { notifyError } = useNotifications()
const { t, locale } = useI18n()
const riskTooltip = computed(() => t('pages.reports.riskTooltip'))

function hexToRgba(color, alpha, fallback = 'rgba(31, 22, 51, 0.08)') {
  const value = String(color || '').trim()
  const hex = value.startsWith('#') ? value.slice(1) : ''
  const normalized = hex.length === 3
    ? hex.split('').map((char) => `${char}${char}`).join('')
    : hex

  if (normalized.length !== 6) {
    return fallback
  }

  const channels = [
    Number.parseInt(normalized.slice(0, 2), 16),
    Number.parseInt(normalized.slice(2, 4), 16),
    Number.parseInt(normalized.slice(4, 6), 16),
  ]

  if (channels.some((channel) => Number.isNaN(channel))) {
    return fallback
  }

  return `rgba(${channels.join(', ')}, ${alpha})`
}

const hasEmployeeJoin = computed(() => !partialErrors.value.some((item) => item.source === 'employees'))
const hasDepartmentJoin = computed(() => (
  hasEmployeeJoin.value
  && !partialErrors.value.some((item) => item.source === 'departments')
))
const effectiveFilters = computed(() => ({
  ...appliedFilters.value,
  departmentId: hasEmployeeJoin.value ? appliedFilters.value.departmentId : '',
}))
const report = computed(() => {
  locale.value

  return buildReportAnalytics({
    campaigns: campaigns.value,
    vtargets: vtargets.value,
    employees: employees.value,
    departments: departments.value,
    filters: effectiveFilters.value,
  })
})
const reportEmpty = computed(() => !isLoading.value && !report.value.targets.length)
const noSupportedEvents = computed(() => !isLoading.value && report.value.targets.length > 0 && !report.value.hasSupportedEvents)
const selectedCampaignLabel = computed(() => {
  if (!appliedFilters.value.campaignId) {
    return t('common.filters.allCampaigns')
  }

  return campaigns.value.find((campaign) => String(campaign.id) === appliedFilters.value.campaignId)?.label || t('pages.reports.selectedCampaign')
})
const selectedDepartmentLabel = computed(() => {
  if (!appliedFilters.value.departmentId) {
    return t('common.filters.allDepartments')
  }

  return departments.value.find((department) => String(department.id) === appliedFilters.value.departmentId)?.label || t('pages.reports.selectedDepartment')
})
const reportDateRangeLabel = computed(() => {
  const from = appliedFilters.value.dateFrom || t('pages.reports.rangeStart')
  const to = appliedFilters.value.dateTo || t('pages.reports.rangeNow')

  if (!appliedFilters.value.dateFrom && !appliedFilters.value.dateTo) {
    return t('pages.reports.allRecordedDates')
  }

  return `${from} — ${to}`
})
const generatedAtLabel = computed(() => formatDateTime(generatedAt.value))

const kpiCards = computed(() => [
  {
    label: t('pages.reports.kpis.companyRisk.label'),
    value: report.value.kpis.companyRisk,
    description: t('pages.reports.kpis.companyRisk.description'),
    icon: ShieldAlert,
    help: true,
  },
  {
    label: t('pages.reports.kpis.campaignsAnalyzed.label'),
    value: report.value.kpis.campaignsAnalyzed,
    description: t('pages.reports.kpis.campaignsAnalyzed.description'),
    icon: BarChart3,
  },
  {
    label: t('pages.reports.kpis.employeesTested.label'),
    value: report.value.kpis.employeesTested,
    description: t('pages.reports.kpis.employeesTested.description'),
    icon: Users,
  },
  {
    label: t('pages.reports.kpis.emailOpenRate.label'),
    value: report.value.kpis.emailOpenRate,
    description: t('pages.reports.kpis.emailOpenRate.description'),
    icon: MailOpen,
  },
  {
    label: t('pages.reports.kpis.linkClickRate.label'),
    value: report.value.kpis.linkClickRate,
    description: t('pages.reports.kpis.linkClickRate.description'),
    icon: MousePointerClick,
  },
  {
    label: t('pages.reports.kpis.dataSubmittedRate.label'),
    value: report.value.kpis.dataSubmittedRate,
    description: t('pages.reports.kpis.dataSubmittedRate.description'),
    icon: Send,
  },
  {
    label: t('pages.reports.kpis.totalEvents.label'),
    value: report.value.kpis.totalEvents,
    description: t('pages.reports.kpis.totalEvents.description'),
    icon: Activity,
  },
])

const chartTheme = computed(() => {
  const tokens = chartTokens()
  return {
    tokens,
    eventColors: Object.fromEntries(EVENT_TYPES.map((eventType) => [
      eventType,
      eventChartColor(eventType, tokens),
    ])),
    riskColor: riskValueColor(report.value.companyRiskValue ?? 0, tokens),
  }
})

const commonTooltip = computed(() => ({
  borderColor: chartTheme.value.tokens.hairlineCool,
  backgroundColor: chartTheme.value.tokens.surface,
  textStyle: {
    color: chartTheme.value.tokens.inkSecondary,
    fontSize: 12,
  },
  appendToBody: true,
  borderWidth: 1,
  extraCssText: `box-shadow: 0 10px 28px ${hexToRgba(chartTheme.value.tokens.ink, 0.08)}; border-radius: 8px;`,
}))

const departmentPalette = computed(() => {
  const tokens = chartTheme.value.tokens
  return [
    tokens.primary,
    tokens.accentViolet,
    tokens.accentPink,
    tokens.warning,
    tokens.danger,
    tokens.info,
    tokens.accentIndigo,
    tokens.accentLime,
  ]
})

const departmentColorMap = computed(() => {
  const colors = departmentPalette.value
  const map = new Map()
  report.value.departmentRows.forEach((department, index) => {
    map.set(department.label, colors[index % colors.length])
  })
  return map
})

function reportEventColor(eventType) {
  return chartTheme.value.eventColors[eventType] || chartTheme.value.tokens.inkMuted
}

const gaugeOption = computed(() => {
  const tokens = chartTheme.value.tokens
  const currentRiskColor = chartTheme.value.riskColor

  return {
    series: [
      {
        type: 'gauge',
        startAngle: 180,
        endAngle: 0,
        center: ['50%', '72%'],
        radius: '116%',
        min: 0,
        max: 100,
        splitNumber: 4,
        axisLine: {
          lineStyle: {
            width: 10,
            color: [
              [0.45, tokens.primary],
              [0.75, tokens.warning],
              [1, tokens.danger],
            ],
          },
        },
        pointer: {
          show: true,
          length: '58%',
          width: 6,
          offsetCenter: [0, 0],
          itemStyle: { color: currentRiskColor },
        },
        anchor: {
          show: true,
          showAbove: true,
          size: 10,
          itemStyle: {
            color: tokens.surface,
            borderColor: currentRiskColor,
            borderWidth: 3,
          },
        },
        axisTick: {
          length: 7,
          lineStyle: { color: 'auto', width: 1 },
        },
        splitLine: {
          length: 13,
          lineStyle: { color: 'auto', width: 2 },
        },
        axisLabel: {
          color: tokens.inkMuted,
          fontSize: 11,
          distance: 14,
          formatter(value) {
            if (value === 25) return t('charts.risk.low')
            if (value === 50) return t('charts.risk.moderate')
            if (value === 75) return t('charts.risk.high')
            return ''
          },
        },
        detail: {
          fontSize: 30,
          fontWeight: 700,
          offsetCenter: [0, '-18%'],
          valueAnimation: true,
          formatter(value) {
            return `${Math.round(value)}`
          },
          color: tokens.ink,
        },
        title: {
          offsetCenter: [0, '10%'],
          fontSize: 13,
          color: tokens.inkMuted,
        },
        data: [
          {
            value: report.value.companyRiskValue ?? 0,
            name: t('common.labels.riskIndex'),
          },
        ],
      },
    ],
  }
})

const campaignComparisonOption = computed(() => ({
  color: EVENT_TYPES.map((eventType) => reportEventColor(eventType)),
  tooltip: {
    ...commonTooltip.value,
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    formatter(params) {
      const first = params[0]
      const row = report.value.campaignRows.find((campaign) => campaign.label === first?.axisValue)
      const lines = [`<strong>${first?.axisValue || t('common.labels.campaign')}</strong>`]
      params.forEach((item) => {
        lines.push(`${item.marker}${item.seriesName}: ${item.value}`)
      })
      lines.push(`${t('common.labels.riskIndex')}: ${row?.risk ?? t('common.placeholder')}`)
      return lines.join('<br/>')
    },
  },
  legend: {
    top: 0,
    textStyle: { color: chartTheme.value.tokens.inkMuted, fontSize: 11 },
  },
  grid: {
    left: 12,
    right: 24,
    top: 42,
    bottom: 24,
    containLabel: true,
  },
  xAxis: {
    type: 'value',
    minInterval: 1,
    splitLine: { lineStyle: { color: chartTheme.value.tokens.hairline } },
    axisLabel: { color: chartTheme.value.tokens.inkSubtle, fontSize: 11 },
  },
  yAxis: {
    type: 'category',
    data: report.value.campaignRows.map((campaign) => campaign.label),
    axisLabel: { color: chartTheme.value.tokens.inkSecondary, fontSize: 12 },
    axisTick: { show: false },
  },
  series: EVENT_TYPES.map((eventType) => ({
    name: EVENT_LABELS[eventType],
    type: 'bar',
    stack: 'total',
    emphasis: { focus: 'series' },
    itemStyle: { color: reportEventColor(eventType) },
    data: report.value.campaignRows.map((campaign) => campaign.eventCounts[eventType] || 0),
  })),
}))

const departmentRiskOption = computed(() => ({
  dataset: {
    source: [
      ['risk', 'employees', 'department'],
      ...report.value.departmentRows.map((department) => [
        department.risk,
        department.employeeCount,
        department.label,
      ]),
    ],
  },
  tooltip: {
    ...commonTooltip.value,
    trigger: 'item',
    formatter(params) {
      const row = params.data || []
      return [
        `<strong>${row[2]}</strong>`,
        `${t('common.labels.riskIndex')}: ${row[0]}`,
        `${t('common.labels.employees')}: ${row[1]}`,
        t('pages.reports.departmentRiskTooltip'),
      ].join('<br/>')
    },
  },
  grid: { left: 12, right: 24, top: 24, bottom: 48, containLabel: true },
  xAxis: {
    name: t('common.labels.riskIndex'),
    max: 100,
    axisLabel: { color: chartTheme.value.tokens.inkSubtle, fontSize: 11 },
    splitLine: { lineStyle: { color: chartTheme.value.tokens.hairline } },
  },
  yAxis: {
    type: 'category',
    axisLabel: { color: chartTheme.value.tokens.inkSecondary, fontSize: 12 },
    axisTick: { show: false },
  },
  visualMap: {
    orient: 'horizontal',
    left: 'center',
    bottom: 4,
    min: 0,
    max: 100,
    text: [t('pages.reports.highRisk'), t('pages.reports.lowRisk')],
    dimension: 0,
    inRange: {
      color: [
        chartTheme.value.tokens.primary,
        chartTheme.value.tokens.warning,
        chartTheme.value.tokens.danger,
      ],
    },
    textStyle: { color: chartTheme.value.tokens.inkMuted, fontSize: 11 },
  },
  series: [
    {
      type: 'bar',
      encode: {
        x: 'risk',
        y: 'department',
      },
      barMaxWidth: 24,
    },
  ],
}))

const employeeScatterOption = computed(() => {
  const departmentsInChart = Array.from(new Set(report.value.employeeRows.map((employee) => employee.departmentLabel)))

  return {
    tooltip: {
      ...commonTooltip.value,
      trigger: 'item',
      formatter(params) {
        return [
          `<strong>${params.data.employeeLabel}</strong>`,
          `${t('pages.reports.risk')}: ${params.data.risk}`,
          `${t('common.labels.events')}: ${params.data.eventCount}`,
          `${t('common.labels.department')}: ${params.data.departmentLabel}`,
        ].join('<br/>')
      },
    },
    legend: {
      top: 0,
      textStyle: { color: chartTheme.value.tokens.inkMuted, fontSize: 11 },
    },
    grid: {
      left: 48,
      right: 24,
      top: 48,
      bottom: 40,
    },
    xAxis: {
      type: 'value',
      min: 0,
      name: t('pages.reports.riskSignals'),
      nameTextStyle: { color: chartTheme.value.tokens.inkMuted, fontSize: 11 },
      axisLabel: { color: chartTheme.value.tokens.inkSubtle, fontSize: 11 },
      splitLine: { lineStyle: { color: chartTheme.value.tokens.hairline } },
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: 100,
      name: t('common.labels.riskIndex'),
      nameTextStyle: { color: chartTheme.value.tokens.inkMuted, fontSize: 11 },
      axisLabel: { color: chartTheme.value.tokens.inkSubtle, fontSize: 11 },
      splitLine: { lineStyle: { color: chartTheme.value.tokens.hairline } },
    },
    series: departmentsInChart.map((departmentLabel) => ({
      name: departmentLabel,
      type: 'scatter',
      symbolSize(data) {
        return Math.max(8, Math.min(28, 8 + Number(data?.[2] || 0) * 2))
      },
      itemStyle: {
        color: departmentColorMap.value.get(departmentLabel) || chartTheme.value.tokens.inkMuted,
      },
      data: report.value.employeeRows
        .filter((employee) => employee.departmentLabel === departmentLabel)
        .map((employee) => ({
          value: [employee.riskSignalCount, employee.risk, employee.eventCount],
          employeeLabel: employee.label,
          risk: employee.risk,
          eventCount: employee.eventCount,
          departmentLabel: employee.departmentLabel,
        })),
    })),
  }
})

const funnelOption = computed(() => ({
  tooltip: {
    ...commonTooltip.value,
    trigger: 'item',
    formatter: '{b}: {c}',
  },
  legend: {
    top: 0,
    data: report.value.funnelStages.map((stage) => stage.label),
    textStyle: { color: chartTheme.value.tokens.inkMuted, fontSize: 11 },
  },
  series: [
    {
      name: t('pages.reports.eventFunnel'),
      type: 'funnel',
      left: '8%',
      top: 40,
      bottom: 24,
      width: '84%',
      min: 0,
      max: Math.max(...report.value.funnelStages.map((stage) => stage.count), 1),
      minSize: '0%',
      maxSize: '100%',
      sort: 'descending',
      gap: 2,
      label: {
        show: true,
        position: 'inside',
        color: chartTheme.value.tokens.surface,
        fontWeight: 700,
      },
      itemStyle: {
        borderColor: chartTheme.value.tokens.surface,
        borderWidth: 1,
      },
      data: report.value.funnelStages.map((stage) => ({
        name: stage.label,
        value: stage.count,
        itemStyle: { color: reportEventColor(stage.eventType) },
      })),
    },
  ],
}))

const trendOption = computed(() => ({
  tooltip: {
    ...commonTooltip.value,
    trigger: 'axis',
    axisPointer: { type: 'cross' },
    formatter(params) {
      const item = params[0]
      return [
        `<strong>${item?.axisValue || t('pages.reports.day')}</strong>`,
        `${t('pages.reports.riskSignalVolume')}: ${item?.value ?? 0}`,
        t('pages.reports.weightedRiskTooltip'),
      ].join('<br/>')
    },
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: report.value.trendBuckets.map((bucket) => bucket.label),
    axisLabel: { color: chartTheme.value.tokens.inkSubtle, fontSize: 11 },
    axisLine: { lineStyle: { color: chartTheme.value.tokens.hairline } },
  },
  yAxis: {
    type: 'value',
    axisLabel: {
      color: chartTheme.value.tokens.inkSubtle,
      fontSize: 11,
      formatter: '{value}',
    },
    splitLine: { lineStyle: { color: chartTheme.value.tokens.hairline } },
    axisPointer: { snap: true },
  },
  visualMap: {
    show: false,
    dimension: 0,
    pieces: report.value.trendBuckets.map((bucket, index) => ({
      gt: index - 1,
      lte: index,
      color: index > 0 && bucket.value > report.value.trendBuckets[index - 1].value
        ? chartTheme.value.tokens.warning
        : chartTheme.value.tokens.primary,
    })),
  },
  series: [
    {
      name: t('pages.reports.riskSignals'),
      type: 'line',
      smooth: true,
      data: report.value.trendBuckets.map((bucket) => bucket.value),
      symbolSize: 7,
      lineStyle: { width: 3 },
      markArea: {
        itemStyle: {
          color: hexToRgba(chartTheme.value.tokens.danger, 0.08, 'rgba(201, 42, 58, 0.08)'),
        },
        data: report.value.trendMarkAreas,
      },
    },
  ],
}))

const eventDistributionOption = computed(() => ({
  color: EVENT_TYPES.map((eventType) => reportEventColor(eventType)),
  tooltip: {
    ...commonTooltip.value,
    trigger: 'item',
    formatter: ({ name, value, percent }) => `${name}: ${value} (${percent}%)`,
  },
  legend: {
    bottom: 0,
    left: 'center',
    itemWidth: 8,
    itemHeight: 8,
    icon: 'circle',
    itemGap: 12,
    textStyle: { color: chartTheme.value.tokens.inkMuted, fontSize: 11 },
  },
  series: [
    {
      name: t('pages.reports.eventActivityVolume'),
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['50%', '43%'],
      itemStyle: {
        borderRadius: 6,
        borderColor: chartTheme.value.tokens.surface,
        borderWidth: 2,
      },
      label: { show: false },
      labelLine: { show: false },
      data: report.value.eventDistribution.map((item) => ({
        name: item.label,
        value: item.count,
        itemStyle: { color: reportEventColor(item.eventType) },
      })),
    },
  ],
}))

const pdfCharts = computed(() => [
  {
    key: 'company-risk-index',
    title: t('pages.reports.charts.companyRiskIndexTitle'),
    description: t('pages.reports.charts.companyRiskIndexDescription'),
    option: gaugeOption.value,
    empty: report.value.companyRiskValue === null,
    emptyText: t('charts.risk.empty'),
    height: '300px',
  },
  {
    key: 'event-distribution',
    title: t('pages.reports.charts.eventDistributionTitle'),
    description: t('pages.reports.charts.eventDistributionDescription'),
    option: eventDistributionOption.value,
    empty: !report.value.hasSupportedEvents,
    emptyText: t('charts.eventDistribution.empty'),
    height: '300px',
  },
  {
    key: 'campaign-comparison',
    title: t('pages.reports.charts.campaignComparisonTitle'),
    description: t('pages.reports.charts.campaignComparisonDescription'),
    option: campaignComparisonOption.value,
    empty: !report.value.campaignRows.length || !report.value.hasSupportedEvents,
    emptyText: t('common.empty.noMatchingRecords'),
    height: '360px',
  },
  {
    key: 'funnel',
    title: t('pages.reports.charts.funnelTitle'),
    description: t('pages.reports.charts.funnelDescription'),
    option: funnelOption.value,
    empty: !report.value.funnelStages.some((stage) => stage.count > 0),
    emptyText: t('charts.funnel.empty'),
    height: '320px',
  },
  {
    key: 'trend',
    title: t('pages.reports.charts.trendTitle'),
    description: t('pages.reports.charts.trendDescription'),
    option: trendOption.value,
    empty: !report.value.trendBuckets.length,
    emptyText: t('common.empty.noMatchingRecords'),
    height: '320px',
  },
  {
    key: 'department-risk',
    title: t('pages.reports.charts.departmentRiskTitle'),
    description: t('pages.reports.charts.departmentRiskDescription'),
    option: departmentRiskOption.value,
    empty: !hasDepartmentJoin.value || !report.value.departmentRows.length,
    emptyText: t('charts.departments.empty'),
    height: '330px',
  },
  {
    key: 'employee-risk',
    title: t('pages.reports.charts.employeeRiskTitle'),
    description: t('pages.reports.charts.employeeRiskDescription'),
    option: employeeScatterOption.value,
    empty: !hasEmployeeJoin.value || !report.value.employeeRows.length,
    emptyText: t('common.empty.noMatchingRecords'),
    height: '330px',
  },
])

function applyFilters() {
  appliedFilters.value = { ...filters }
  loadReport()
}

function clearFilters() {
  filters.dateFrom = ''
  filters.dateTo = ''
  filters.campaignId = ''
  filters.departmentId = ''
  appliedFilters.value = { ...filters }
  loadReport()
}

function waitForFrame() {
  return new Promise((resolve) => {
    requestAnimationFrame(resolve)
  })
}

async function loadReport() {
  isLoading.value = true
  loadError.value = ''
  partialErrors.value = []

  try {
    const [campaignResult, targetResult, employeeResult, departmentResult] = await Promise.allSettled([
      loadAllPages(listCampaigns),
      loadAllPages(listVTargets, { campaignId: appliedFilters.value.campaignId }),
      loadAllPages(listEmployees),
      loadAllPages(listDepartments),
    ])

    if (targetResult.status === 'rejected') {
      if (await handleAuthError(targetResult.reason)) {
        return
      }
      throw targetResult.reason
    }

    if (campaignResult.status === 'fulfilled') {
      campaigns.value = campaignResult.value
    } else {
      partialErrors.value.push({
        source: 'campaigns',
        message: errorMessage(campaignResult.reason, 'errors.campaignsLoad'),
      })
      campaigns.value = []
    }

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
    loadError.value = errorMessage(error, 'errors.reportsLoad')
  } finally {
    isLoading.value = false
  }
}

function pdfFileDate(value) {
  const date = value instanceof Date ? value : new Date(value)

  if (Number.isNaN(date.getTime())) {
    return 'report'
  }

  return date.toISOString().slice(0, 10)
}

function addCanvasToPdf(pdf, canvas) {
  const pageWidth = pdf.internal.pageSize.getWidth()
  const pageHeight = pdf.internal.pageSize.getHeight()
  const pixelsPerMillimeter = canvas.width / pageWidth
  const pageHeightPixels = Math.max(1, Math.floor(pageHeight * pixelsPerMillimeter))
  let offsetY = 0
  let pageIndex = 0

  while (offsetY < canvas.height) {
    const sliceHeight = Math.min(pageHeightPixels, canvas.height - offsetY)
    const pageCanvas = document.createElement('canvas')
    pageCanvas.width = canvas.width
    pageCanvas.height = pageHeightPixels

    const context = pageCanvas.getContext('2d')
    if (!context) {
      throw new Error('Canvas context unavailable')
    }

    context.fillStyle = '#ffffff'
    context.fillRect(0, 0, pageCanvas.width, pageCanvas.height)
    context.drawImage(
      canvas,
      0,
      offsetY,
      canvas.width,
      sliceHeight,
      0,
      0,
      canvas.width,
      sliceHeight,
    )

    if (pageIndex > 0) {
      pdf.addPage()
    }

    pdf.addImage(pageCanvas.toDataURL('image/png'), 'PNG', 0, 0, pageWidth, pageHeight)
    offsetY += sliceHeight
    pageIndex += 1
  }
}

async function exportReportPdf() {
  if (isPdfExporting.value) {
    return
  }

  generatedAt.value = new Date()
  isPdfExporting.value = true

  try {
    await nextTick()
    await waitForFrame()
    await pdfDocumentRef.value?.prepareCharts?.()
    await pdfDocumentRef.value?.waitForImages?.()

    const pdfRoot = pdfDocumentRef.value?.getRootElement?.()
    if (!pdfRoot) {
      throw new Error(t('pages.reports.exportFailed'))
    }

    const html2canvas = (await import('html2canvas')).default
    const { jsPDF } = await import('jspdf')
    const canvas = await html2canvas(pdfRoot, {
      scale: 2,
      backgroundColor: '#ffffff',
      useCORS: true,
    })
    if (!canvas.width || !canvas.height) {
      throw new Error(t('pages.reports.exportFailed'))
    }

    const pdf = new jsPDF('p', 'mm', 'a4')
    addCanvasToPdf(pdf, canvas)
    pdf.save(`clicksafe-report-${pdfFileDate(generatedAt.value)}.pdf`)
  } catch (error) {
    notifyError(t('pages.reports.exportFailed'), error?.message || t('errors.requestFailed'))
  } finally {
    isPdfExporting.value = false
  }
}

onMounted(() => {
  loadReport()
})
</script>

<template>
  <section class="reports-page">
    <PageHeader
      :eyebrow="t('sections.analytics')"
      :title="t('pages.reports.title')"
      :description="t('pages.reports.description')"
    >
      <template #actions>
        <UiButton
          variant="secondary"
          :disabled="isLoading || isPdfExporting || (!report.targets.length && !report.hasSupportedEvents)"
          :loading="isPdfExporting"
          @click="exportReportPdf"
        >
          <Printer :size="16" stroke-width="1.8" aria-hidden="true" />
          {{ isPdfExporting ? t('pages.reports.exportPreparing') : t('pages.reports.exportPdf') }}
        </UiButton>
      </template>
    </PageHeader>

    <section class="reports-filter-panel" :aria-label="t('pages.reports.filtersAria')">
      <UiDatePicker v-model="filters.dateFrom" :label="t('common.labels.dateFrom')" />
      <UiDatePicker v-model="filters.dateTo" :label="t('common.labels.dateTo')" />
      <label class="ui-field">
        <span class="ui-field-label">{{ t('common.labels.campaign') }}</span>
        <select v-model="filters.campaignId" class="ui-select">
          <option value="">{{ t('common.filters.allCampaigns') }}</option>
          <option v-for="campaign in campaigns" :key="campaign.id" :value="campaign.id">
            {{ campaign.label }}
          </option>
        </select>
      </label>
      <label class="ui-field">
        <span class="ui-field-label">{{ t('common.labels.department') }}</span>
        <select v-model="filters.departmentId" class="ui-select" :disabled="!hasEmployeeJoin">
          <option value="">{{ t('common.filters.allDepartments') }}</option>
          <option v-for="department in departments" :key="department.id" :value="department.id">
            {{ department.label }}
          </option>
        </select>
        <span v-if="!hasEmployeeJoin" class="ui-field-helper">
          {{ t('pages.reports.departmentFilterRequiresEmployees') }}
        </span>
      </label>
      <div class="reports-filter-actions">
        <UiButton :loading="isLoading" @click="applyFilters">{{ t('common.filters.apply') }}</UiButton>
        <UiButton variant="ghost" :disabled="isLoading" @click="clearFilters">{{ t('common.filters.clear') }}</UiButton>
      </div>
    </section>

    <section class="reports-print-summary" :aria-label="t('pages.reports.currentScope')">
      <div>
        <span>{{ t('common.labels.timeRange') }}</span>
        <strong>{{ reportDateRangeLabel }}</strong>
      </div>
      <div>
        <span>{{ t('common.labels.campaign') }}</span>
        <strong>{{ selectedCampaignLabel }}</strong>
      </div>
      <div>
        <span>{{ t('common.labels.department') }}</span>
        <strong>{{ selectedDepartmentLabel }}</strong>
      </div>
      <div>
        <span>{{ t('common.labels.generatedAt') }}</span>
        <strong>{{ generatedAtLabel }}</strong>
      </div>
    </section>

    <UiAlert
      v-if="loadError"
      variant="error"
      :title="t('pages.reports.unavailable')"
      :message="loadError"
    >
      <UiButton variant="secondary" @click="loadReport">{{ t('common.actions.retry') }}</UiButton>
    </UiAlert>

    <UiAlert
      v-for="partialError in partialErrors"
      :key="partialError.source"
      variant="warning"
      :title="t('pages.reports.partialData')"
      :message="partialError.message"
    />

    <template v-if="isLoading">
      <section class="reports-metrics-grid">
        <article v-for="index in 7" :key="index" class="metric-card">
          <SkeletonBlock :rows="3" />
        </article>
      </section>
      <section class="reports-chart-grid">
        <article v-for="index in 6" :key="index" class="reports-panel">
          <SkeletonBlock :rows="6" />
        </article>
      </section>
    </template>

    <EmptyState
      v-else-if="reportEmpty"
      :title="t('pages.reports.emptyTitle')"
      :description="t('pages.reports.emptyDescription')"
    >
      <UiButton variant="secondary" @click="clearFilters">{{ t('common.filters.clear') }}</UiButton>
    </EmptyState>

    <template v-else>
      <UiAlert
        v-if="noSupportedEvents"
        variant="info"
        :title="t('pages.reports.noSupportedEventsTitle')"
        :message="t('pages.reports.noSupportedEventsDescription')"
      />

      <section class="reports-metrics-grid" :aria-label="t('pages.reports.aria.kpiCards')">
        <MetricCard
          v-for="metric in kpiCards"
          :key="metric.label"
          :label="metric.label"
          :value="metric.value"
          :description="metric.description"
          :icon="metric.icon"
        >
          <template v-if="metric.help" #label-actions>
            <span class="reports-help">
              <Info :size="14" stroke-width="1.8" aria-hidden="true" />
              <span class="reports-help-tooltip" role="tooltip">{{ riskTooltip }}</span>
            </span>
          </template>
        </MetricCard>
      </section>

      <section class="reports-chart-grid" :aria-label="t('pages.reports.aria.charts')">
        <article class="reports-panel reports-panel--compact">
          <header class="reports-panel-header">
            <div>
              <h2>{{ t('pages.reports.charts.companyRiskIndexTitle') }}</h2>
              <p>{{ t('pages.reports.charts.companyRiskIndexDescription') }}</p>
            </div>
            <Gauge :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>
          <BaseEChart
            :option="gaugeOption"
            :empty="report.companyRiskValue === null"
            :empty-text="t('charts.risk.empty')"
            height="300px"
            :aria-label="t('charts.aria.campaignRiskGauge')"
          />
        </article>

        <article class="reports-panel reports-panel--compact">
          <header class="reports-panel-header">
            <div>
              <h2>{{ t('pages.reports.charts.eventDistributionTitle') }}</h2>
              <p>{{ t('pages.reports.charts.eventDistributionDescription') }}</p>
            </div>
            <Activity :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>
          <BaseEChart
            :option="eventDistributionOption"
            :empty="!report.hasSupportedEvents"
            :empty-text="t('charts.eventDistribution.empty')"
            height="300px"
            :aria-label="t('charts.aria.campaignEventDistribution')"
          />
        </article>

        <article class="reports-panel reports-panel--wide">
          <header class="reports-panel-header">
            <div>
              <h2>{{ t('pages.reports.charts.campaignComparisonTitle') }}</h2>
              <p>{{ t('pages.reports.charts.campaignComparisonDescription') }}</p>
            </div>
            <BarChart3 :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>
          <BaseEChart
            :option="campaignComparisonOption"
            :empty="!report.campaignRows.length || !report.hasSupportedEvents"
            :empty-text="t('common.empty.noMatchingRecords')"
            height="360px"
            :aria-label="t('pages.reports.charts.campaignComparisonTitle')"
          />
        </article>

        <article class="reports-panel">
          <header class="reports-panel-header">
            <div>
              <h2>{{ t('pages.reports.charts.funnelTitle') }}</h2>
              <p>{{ t('pages.reports.charts.funnelDescription') }}</p>
            </div>
            <TrendingUp :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>
          <BaseEChart
            :option="funnelOption"
            :empty="!report.funnelStages.some((stage) => stage.count > 0)"
            :empty-text="t('charts.funnel.empty')"
            height="320px"
            :aria-label="t('charts.aria.campaignFunnel')"
          />
        </article>

        <article class="reports-panel">
          <header class="reports-panel-header">
            <div>
              <h2>{{ t('pages.reports.charts.trendTitle') }}</h2>
              <p>{{ t('pages.reports.charts.trendDescription') }}</p>
            </div>
            <Activity :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>
          <BaseEChart
            :option="trendOption"
            :empty="!report.trendBuckets.length"
            :empty-text="t('common.empty.noMatchingRecords')"
            height="320px"
            :aria-label="t('pages.reports.charts.trendTitle')"
          />
        </article>

        <article class="reports-panel">
          <header class="reports-panel-header">
            <div>
              <h2>{{ t('pages.reports.charts.departmentRiskTitle') }}</h2>
              <p>{{ t('pages.reports.charts.departmentRiskDescription') }}</p>
            </div>
            <Building2 :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>
          <BaseEChart
            :option="departmentRiskOption"
            :empty="!hasDepartmentJoin || !report.departmentRows.length"
            :empty-text="t('charts.departments.empty')"
            height="330px"
            :aria-label="t('pages.reports.charts.departmentRiskTitle')"
          />
        </article>

        <article class="reports-panel">
          <header class="reports-panel-header">
            <div>
              <h2>{{ t('pages.reports.charts.employeeRiskTitle') }}</h2>
              <p>{{ t('pages.reports.charts.employeeRiskDescription') }}</p>
            </div>
            <Users :size="18" stroke-width="1.8" aria-hidden="true" />
          </header>
          <BaseEChart
            :option="employeeScatterOption"
            :empty="!hasEmployeeJoin || !report.employeeRows.length"
            :empty-text="t('common.empty.noMatchingRecords')"
            height="330px"
            :aria-label="t('pages.reports.charts.employeeRiskTitle')"
          />
        </article>
      </section>

      <section class="reports-panel reports-table-panel" aria-labelledby="top-risky-employees-title">
        <header class="reports-panel-header">
          <div>
            <h2 id="top-risky-employees-title">{{ t('pages.reports.charts.topEmployeesTitle') }}</h2>
            <p>{{ t('pages.reports.charts.topEmployeesDescription') }}</p>
          </div>
        </header>

        <div class="reports-table">
          <div class="reports-table-row reports-table-head">
            <span>{{ t('common.labels.employee') }}</span>
            <span>{{ t('common.labels.department') }}</span>
            <span>{{ t('common.labels.riskIndex') }}</span>
            <span>{{ EVENT_LABELS.LINK_OPENED }}</span>
            <span>{{ EVENT_LABELS.DATA_SENT }}</span>
            <span>{{ t('common.labels.activity') }}</span>
          </div>
          <div
            v-for="employee in report.topEmployees"
            :key="employee.id"
            class="reports-table-row"
          >
            <span><strong>{{ employee.label }}</strong></span>
            <span>{{ employee.departmentLabel }}</span>
            <span>{{ employee.risk }}</span>
            <span>{{ employee.linkOpenedCount }}</span>
            <span>{{ employee.dataSubmittedCount }}</span>
            <span>{{ employee.lastActivityLabel }}</span>
          </div>
          <div v-if="!report.topEmployees.length" class="reports-empty-row">
            {{ t('pages.reports.charts.topEmployeesEmpty') }}
          </div>
        </div>
      </section>
    </template>
  </section>

  <ReportsPdfDocument
    v-if="isPdfExporting"
    ref="pdfDocumentRef"
    :title="t('pages.reports.title')"
    :description="t('pages.reports.description')"
    :date-range-label="reportDateRangeLabel"
    :campaign-label="selectedCampaignLabel"
    :department-label="selectedDepartmentLabel"
    :generated-at-label="generatedAtLabel"
    :kpis="kpiCards"
    :charts="pdfCharts"
    :top-employees="report.topEmployees"
  />
</template>
