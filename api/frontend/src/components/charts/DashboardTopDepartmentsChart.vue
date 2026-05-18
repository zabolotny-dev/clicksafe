<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseEChart from './BaseEChart.vue'
import { chartTokens } from './chartPalette'

const props = defineProps({
  departments: {
    type: Array,
    default: () => [],
  },
  emptyText: {
    type: String,
    default: '',
  },
})

const { t } = useI18n()
const tokens = computed(() => chartTokens())
const isEmpty = computed(() => !props.departments.length)
const resolvedEmptyText = computed(() => props.emptyText || t('charts.departments.empty'))

function riskColor(value) {
  if (value >= 70) {
    return tokens.value.danger
  }

  if (value >= 45) {
    return tokens.value.warning
  }

  return tokens.value.success
}

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
    formatter(params) {
      const row = props.departments[params.dataIndex]
      return [
        `<strong>${row?.label || params.name}</strong>`,
        `${t('charts.departments.riskIndex')}: ${params.value}`,
        `${t('charts.departments.employees')}: ${row?.employeeCount ?? t('common.placeholder')}`,
      ].join('<br/>')
    },
  },
  grid: {
    top: 12,
    right: 20,
    bottom: 28,
    left: 12,
    containLabel: true,
  },
  xAxis: {
    type: 'value',
    max: 100,
    splitLine: { lineStyle: { color: tokens.value.hairline } },
    axisLabel: { color: tokens.value.inkSubtle, fontSize: 11 },
  },
  yAxis: {
    type: 'category',
    inverse: true,
    data: props.departments.map((department) => department.label),
    axisLabel: {
      color: tokens.value.inkSecondary,
      fontSize: 12,
      width: 120,
      overflow: 'truncate',
    },
    axisTick: { show: false },
    axisLine: { lineStyle: { color: tokens.value.hairline } },
  },
  series: [
    {
      name: t('charts.departments.riskIndex'),
      type: 'bar',
      barMaxWidth: 20,
      data: props.departments.map((department) => ({
        value: department.risk,
        itemStyle: {
          color: riskColor(department.risk),
          borderRadius: [0, 5, 5, 0],
        },
      })),
      label: {
        show: true,
        position: 'right',
        color: tokens.value.inkMuted,
        fontSize: 11,
        formatter: '{c}',
      },
      animationDuration: 360,
      animationEasing: 'cubicOut',
    },
  ],
}))
</script>

<template>
  <BaseEChart
    :option="option"
    :empty="isEmpty"
    :empty-text="resolvedEmptyText"
    height="260px"
    :aria-label="t('charts.aria.dashboardTopDepartments')"
  />
</template>
