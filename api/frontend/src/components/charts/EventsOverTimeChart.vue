<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { translateEventType } from '../../i18n'
import { EVENT_TYPES } from '../../resources/vtargets'
import BaseEChart from './BaseEChart.vue'
import { chartTokens, eventChartColor } from './chartPalette'

const props = defineProps({
  buckets: {
    type: Array,
    default: () => [],
  },
  emptyText: {
    type: String,
    default: '',
  },
})

const { t, locale } = useI18n()
const tokens = computed(() => chartTokens())
const resolvedEmptyText = computed(() => props.emptyText || t('charts.eventsOverTime.empty'))
const eventLabels = computed(() => {
  locale.value

  return Object.fromEntries(EVENT_TYPES.map((eventType) => [
    eventType,
    translateEventType(eventType),
  ]))
})

const isEmpty = computed(() => !props.buckets.some((bucket) => (
  EVENT_TYPES.some((eventType) => Number(bucket.counts?.[eventType] ?? 0) > 0)
)))

const option = computed(() => ({
  color: EVENT_TYPES.map((eventType) => eventChartColor(eventType, tokens.value)),
  tooltip: {
    trigger: 'axis',
    appendToBody: true,
    borderColor: tokens.value.hairlineCool,
    borderWidth: 1,
    backgroundColor: tokens.value.surface,
    extraCssText: 'box-shadow: 0 10px 28px rgba(31, 22, 51, 0.08); border-radius: 8px;',
    textStyle: {
      color: tokens.value.inkSecondary,
      fontSize: 12,
    },
    axisPointer: {
      type: 'line',
      lineStyle: {
        color: tokens.value.hairlineStrong,
        width: 1,
        type: 'dashed',
      },
    },
  },
  legend: {
    bottom: 0,
    left: 'center',
    itemWidth: 8,
    itemHeight: 8,
    icon: 'roundRect',
    textStyle: {
      color: tokens.value.inkMuted,
      fontSize: 11,
      fontWeight: 500,
    },
    itemGap: 12,
  },
  grid: {
    top: 10,
    right: 12,
    bottom: 48,
    left: 34,
    containLabel: true,
  },
  xAxis: {
    type: 'category',
    data: props.buckets.map((bucket) => bucket.label),
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
      margin: 10,
    },
  },
  yAxis: {
    type: 'value',
    minInterval: 1,
    splitLine: {
      lineStyle: {
        color: tokens.value.hairline,
      },
    },
    axisLabel: {
      color: tokens.value.inkSubtle,
      fontSize: 11,
    },
    axisLine: {
      show: false,
    },
    axisTick: {
      show: false,
    },
  },
  series: EVENT_TYPES.map((eventType) => ({
    name: eventLabels.value[eventType],
    type: 'bar',
    stack: 'events',
    barMaxWidth: 24,
    barCategoryGap: '48%',
    itemStyle: {
      color: eventChartColor(eventType, tokens.value),
      borderColor: tokens.value.surface,
      borderWidth: 1,
    },
    data: props.buckets.map((bucket) => Number(bucket.counts?.[eventType] ?? 0)),
    animationDuration: 340,
    animationEasing: 'cubicOut',
  })),
}))
</script>

<template>
  <BaseEChart
    :option="option"
    :empty="isEmpty"
    :empty-text="resolvedEmptyText"
    height="260px"
    :aria-label="t('charts.aria.eventsOverTime')"
  />
</template>
