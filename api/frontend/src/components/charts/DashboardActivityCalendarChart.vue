<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseEChart from './BaseEChart.vue'
import { chartTokens } from './chartPalette'

const props = defineProps({
  year: {
    type: Number,
    required: true,
  },
  days: {
    type: Array,
    default: () => [],
  },
  emptyText: {
    type: String,
    default: '',
  },
})

const { t, tm } = useI18n()
const tokens = computed(() => chartTokens())
const maxValue = computed(() => Math.max(...props.days.map((day) => Number(day[1] || 0)), 1))
const isEmpty = computed(() => !props.days.some((day) => Number(day[1] || 0) > 0))
const resolvedEmptyText = computed(() => props.emptyText || t('charts.calendar.empty'))
const calendarDays = computed(() => tm('charts.calendar.days'))
const calendarMonths = computed(() => tm('charts.calendar.months'))

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
      const [date, count] = params.value || []
      return `${date}<br/>${t('charts.calendar.tooltip', { count: count || 0 })}`
    },
  },
  visualMap: {
    show: false,
    min: 0,
    max: maxValue.value,
    inRange: {
      color: [
        tokens.value.canvasMuted,
        tokens.value.primarySoft,
        tokens.value.accentViolet,
        tokens.value.warning,
        tokens.value.danger,
      ],
    },
  },
  calendar: {
    top: 28,
    left: 42,
    right: 24,
    bottom: 22,
    range: String(props.year),
    cellSize: ['auto', 18],
    itemStyle: {
      color: tokens.value.canvasMuted,
      borderWidth: 2,
      borderColor: tokens.value.surface,
      borderRadius: 3,
    },
    splitLine: {
      lineStyle: {
        color: tokens.value.hairline,
        width: 1,
      },
    },
    yearLabel: {
      show: false,
    },
    monthLabel: {
      nameMap: calendarMonths.value,
      color: tokens.value.inkMuted,
      fontSize: 11,
      margin: 8,
    },
    dayLabel: {
      firstDay: 1,
      nameMap: calendarDays.value,
      color: tokens.value.inkSubtle,
      fontSize: 10,
    },
  },
  series: [
    {
      name: t('charts.calendar.series'),
      type: 'heatmap',
      coordinateSystem: 'calendar',
      data: props.days,
      emphasis: {
        itemStyle: {
          borderColor: tokens.value.primary,
          borderWidth: 1,
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
    height="286px"
    :aria-label="t('charts.aria.dashboardActivityCalendar')"
  />
</template>
