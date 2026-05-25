<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { translateTargetStatus } from '../../i18n'
import { formatDateTime as formatLocalizedDateTime, formatTechnicalId as formatLocalizedTechnicalId, formatMonthDay } from '../../utils/resourceFormat'
import BaseEChart from './BaseEChart.vue'
import { chartTokens } from './chartPalette'

const props = defineProps({
  targets: {
    type: Array,
    default: () => [],
  },
  emptyText: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['select-target'])
const { t } = useI18n()

const tokens = computed(() => chartTokens())

const scheduledTargets = computed(() => props.targets.filter((target) => {
  if (!target?.scheduled_at) {
    return false
  }

  return !Number.isNaN(new Date(target.scheduled_at).getTime())
}))

const nameCounts = computed(() => scheduledTargets.value.reduce((counts, target) => {
  const name = target.employeeName || formatEmployeeName(target)
  counts[name] = (counts[name] || 0) + 1
  return counts
}, {}))

const lanes = computed(() => scheduledTargets.value.map((target) => laneLabel(target)))
const isEmpty = computed(() => scheduledTargets.value.length === 0)
const chartHeight = computed(() => `${Math.min(420, Math.max(260, lanes.value.length * 34 + 120))}px`)
const resolvedEmptyText = computed(() => props.emptyText || t('charts.schedule.empty'))

const chartData = computed(() => scheduledTargets.value.map((target) => ({
  value: [target.scheduled_at, laneLabel(target)],
  target,
  employeeName: target.employeeName || formatEmployeeName(target),
  status: target.status || 'PENDING',
  scheduledAt: target.scheduled_at,
  createdAt: target.created_at,
  token: target.token,
  itemStyle: {
    color: targetStatusColor(target.status),
    borderColor: tokens.value.surface,
    borderWidth: 2,
  },
})))

const option = computed(() => ({
  tooltip: {
    trigger: 'item',
    appendToBody: true,
    borderColor: tokens.value.hairlineCool,
    borderWidth: 1,
    backgroundColor: tokens.value.surface,
    extraCssText: 'box-shadow: 0 10px 28px rgba(31, 22, 51, 0.08); border-radius: 8px;',
    textStyle: {
      color: tokens.value.inkSecondary,
      fontSize: 12,
    },
    formatter: ({ data }) => {
      if (!data) {
        return ''
      }

      return [
        `<strong>${escapeHtml(data.employeeName)}</strong>`,
        `${escapeHtml(t('charts.schedule.status'))}: ${escapeHtml(translateTargetStatus(data.status))}`,
        `${escapeHtml(t('charts.schedule.scheduled'))}: ${escapeHtml(formatDateTime(data.scheduledAt))}`,
        `${escapeHtml(t('charts.schedule.created'))}: ${escapeHtml(formatDateTime(data.createdAt))}`,
        `${escapeHtml(t('charts.schedule.token'))}: ${escapeHtml(formatTechnicalId(data.token))}`,
      ].join('<br />')
    },
  },
  grid: {
    top: 18,
    right: 24,
    bottom: 42,
    left: 124,
    containLabel: true,
  },
  xAxis: {
    type: 'time',
    axisTick: {
      show: false,
    },
    axisLine: {
      lineStyle: {
        color: tokens.value.hairline,
      },
    },
    splitLine: {
      lineStyle: {
        color: tokens.value.hairline,
      },
    },
    axisLabel: {
      color: tokens.value.inkSubtle,
      fontSize: 11,
      hideOverlap: true,
      formatter(value) {
        return formatMonthDay(value)
      },
    },
  },
  yAxis: {
    type: 'category',
    data: lanes.value,
    axisTick: {
      show: false,
    },
    axisLine: {
      lineStyle: {
        color: tokens.value.hairline,
      },
    },
    axisLabel: {
      color: tokens.value.inkSubtle,
      fontSize: 11,
      width: 112,
      overflow: 'truncate',
    },
  },
  series: [
    {
      name: t('charts.schedule.series'),
      type: 'scatter',
      symbolSize: 14,
      data: chartData.value,
      emphasis: {
        scale: true,
        itemStyle: {
          borderColor: tokens.value.ink,
          borderWidth: 2,
        },
      },
      animationDuration: 300,
      animationEasing: 'cubicOut',
    },
  ],
}))

function laneLabel(target) {
  const name = target.employeeName || formatEmployeeName(target)
  if ((nameCounts.value[name] || 0) <= 1) {
    return name
  }

  return `${name} - ${formatTechnicalId(target.id)}`
}

function targetStatusColor(status) {
  switch (status) {
    case 'SENT':
      return tokens.value.info
    case 'FAILED':
      return tokens.value.danger
    case 'OPENED':
      return tokens.value.accentViolet
    case 'REPLIED':
      return tokens.value.info
    case 'CLICKED':
      return tokens.value.warning
    case 'SUBMITTED':
      return tokens.value.danger
    case 'PENDING':
    default:
      return tokens.value.primary
  }
}

function formatEmployeeName(target) {
  return [target?.first_name, target?.last_name].filter(Boolean).join(' ').trim() || t('common.labels.target')
}

function formatTechnicalId(value) {
  return formatLocalizedTechnicalId(value)
}

function formatDateTime(value) {
  return formatLocalizedDateTime(value)
}

function escapeHtml(value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function handleChartClick(params) {
  if (params?.data?.target) {
    emit('select-target', params.data.target)
  }
}
</script>

<template>
  <BaseEChart
    :option="option"
    :empty="isEmpty"
    :empty-text="resolvedEmptyText"
    :height="chartHeight"
    :aria-label="t('charts.aria.campaignSchedule')"
    @chart-click="handleChartClick"
  />
</template>
