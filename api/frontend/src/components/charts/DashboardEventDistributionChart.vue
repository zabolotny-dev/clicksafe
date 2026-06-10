<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseEChart from './BaseEChart.vue'
import { chartTokens, eventChartColor } from './chartPalette'

const props = defineProps({
  items: {
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
const resolvedEmptyText = computed(() => props.emptyText || t('charts.eventDistribution.empty'))

const chartData = computed(() => props.items.map((item) => ({
  name: item.label,
  value: Number(item.count ?? 0),
  itemStyle: {
    color: eventChartColor(item.eventType, tokens.value),
  },
})))

const isEmpty = computed(() => !chartData.value.some((item) => item.value > 0))

const option = computed(() => ({
  color: props.items.map((item) => eventChartColor(item.eventType, tokens.value)),
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
    formatter: ({ name, value, percent }) => `${name}: ${value} (${percent}%)`,
  },
  legend: {
    bottom: 0,
    left: 'center',
    itemWidth: 8,
    itemHeight: 8,
    icon: 'circle',
    itemGap: 10,
    textStyle: {
      color: tokens.value.inkMuted,
      fontSize: 11,
      fontWeight: 500,
    },
  },
  series: [
    {
      name: t('charts.eventDistribution.series'),
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['50%', '42%'],
      avoidLabelOverlap: true,
      itemStyle: {
        borderRadius: 6,
        borderColor: tokens.value.surface,
        borderWidth: 2,
      },
      label: {
        show: false,
      },
      labelLine: {
        show: false,
      },
      emphasis: {
        scale: true,
        scaleSize: 4,
        label: { show: false },
      },
      data: chartData.value,
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
    :aria-label="t('charts.aria.dashboardEventDistribution')"
  />
</template>
