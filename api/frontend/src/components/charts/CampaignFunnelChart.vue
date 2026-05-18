<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseEChart from './BaseEChart.vue'
import { chartTokens, eventChartColor } from './chartPalette'

const props = defineProps({
  stages: {
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
const resolvedEmptyText = computed(() => props.emptyText || t('charts.funnel.empty'))

const chartData = computed(() => props.stages.map((stage) => ({
  name: stage.label,
  value: Number(stage.count ?? stage.value ?? 0),
  itemStyle: {
    color: eventChartColor(stage.eventType, tokens.value),
  },
})))

const isEmpty = computed(() => !chartData.value.some((item) => item.value > 0))
const maxValue = computed(() => Math.max(...chartData.value.map((item) => item.value), 1))

const option = computed(() => ({
  color: props.stages.map((stage) => eventChartColor(stage.eventType, tokens.value)),
  tooltip: {
    trigger: 'item',
    appendToBody: true,
    borderColor: tokens.value.hairline,
    borderWidth: 1,
    backgroundColor: tokens.value.surface,
    textStyle: {
      color: tokens.value.inkSecondary,
      fontSize: 12,
    },
    formatter: ({ name, value }) => `${name}: ${value}`,
  },
  grid: {
    top: 12,
    right: 42,
    bottom: 12,
    left: 112,
    containLabel: false,
  },
  xAxis: {
    type: 'value',
    max: maxValue.value,
    axisLabel: {
      show: false,
    },
    axisLine: {
      show: false,
    },
    axisTick: {
      show: false,
    },
    splitLine: {
      show: false,
    },
  },
  yAxis: {
    type: 'category',
    inverse: true,
    data: chartData.value.map((item) => item.name),
    axisTick: {
      show: false,
    },
    axisLine: {
      show: false,
    },
    axisLabel: {
      color: tokens.value.inkSecondary,
      fontSize: 12,
      fontWeight: 600,
      margin: 12,
    },
  },
  series: [
    {
      name: t('charts.funnel.series'),
      type: 'bar',
      barWidth: 22,
      showBackground: true,
      backgroundStyle: {
        color: tokens.value.primarySoft,
        borderRadius: [0, 6, 6, 0],
      },
      data: chartData.value.map((item) => ({
        ...item,
        label: {
          color: item.value > 0 ? tokens.value.ink : tokens.value.inkSubtle,
        },
      })),
      label: {
        show: true,
        position: 'right',
        distance: 8,
        fontSize: 12,
        fontWeight: 700,
        formatter: ({ value }) => value,
      },
      itemStyle: {
        borderRadius: [0, 6, 6, 0],
      },
      emphasis: {
        focus: 'self',
        itemStyle: {
          opacity: 0.92,
        },
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
    height="280px"
    :aria-label="t('charts.aria.campaignFunnel')"
  />
</template>
