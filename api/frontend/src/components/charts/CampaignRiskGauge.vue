<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseEChart from './BaseEChart.vue'
import { chartTokens, riskValueColor } from './chartPalette'

const props = defineProps({
  value: {
    type: [Number, String],
    default: null,
  },
  emptyText: {
    type: String,
    default: '',
  },
})

const { t } = useI18n()
const numericValue = computed(() => {
  const value = Number(props.value)
  return Number.isFinite(value) ? Math.max(0, Math.min(100, value)) : null
})

const isEmpty = computed(() => numericValue.value === null)
const resolvedEmptyText = computed(() => props.emptyText || t('charts.risk.empty'))

const option = computed(() => {
  const tokens = chartTokens()
  const value = numericValue.value ?? 0
  const currentRiskColor = riskValueColor(value, tokens)

  return {
    tooltip: {
      trigger: 'item',
      appendToBody: true,
      borderColor: tokens.hairlineCool,
      borderWidth: 1,
      backgroundColor: tokens.surface,
      extraCssText: 'box-shadow: 0 10px 28px rgba(31, 22, 51, 0.08); border-radius: 8px;',
      textStyle: {
        color: tokens.inkSecondary,
        fontSize: 12,
      },
      formatter: ({ value: tooltipValue }) => t('charts.risk.tooltip', { value: Math.round(tooltipValue) }),
    },
    series: [
      {
        name: t('charts.risk.title'),
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
          formatter(axisValue) {
            if (axisValue === 25) return t('charts.risk.low')
            if (axisValue === 50) return t('charts.risk.moderate')
            if (axisValue === 75) return t('charts.risk.high')
            return ''
          },
        },
        pointer: {
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
        title: {
          offsetCenter: [0, '10%'],
          fontSize: 13,
          color: tokens.inkMuted,
        },
        detail: {
          valueAnimation: true,
          formatter: (detailValue) => Math.round(detailValue),
          color: tokens.ink,
          fontSize: 30,
          fontWeight: 700,
          offsetCenter: [0, '-18%'],
        },
        data: [
          {
            value,
            name: t('common.labels.riskIndex'),
          },
        ],
        animationDuration: 420,
        animationEasing: 'cubicOut',
      },
    ],
  }
})
</script>

<template>
  <BaseEChart
    :option="option"
    :empty="isEmpty"
    :empty-text="resolvedEmptyText"
    height="300px"
    :aria-label="t('charts.aria.campaignRiskGauge')"
  />
</template>
