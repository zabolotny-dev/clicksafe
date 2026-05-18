<script setup>
import {
  BarChart,
  FunnelChart,
  GaugeChart,
  HeatmapChart,
  LineChart,
  PieChart,
  ScatterChart,
} from 'echarts/charts'
import {
  CalendarComponent,
  DatasetComponent,
  GridComponent,
  LegendComponent,
  MarkAreaComponent,
  TooltipComponent,
  VisualMapComponent,
} from 'echarts/components'
import * as echarts from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { nextTick, onBeforeUnmount, onMounted, ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'

echarts.use([
  BarChart,
  FunnelChart,
  GaugeChart,
  HeatmapChart,
  LineChart,
  PieChart,
  ScatterChart,
  CalendarComponent,
  DatasetComponent,
  GridComponent,
  LegendComponent,
  MarkAreaComponent,
  TooltipComponent,
  VisualMapComponent,
  CanvasRenderer,
])

const props = defineProps({
  option: {
    type: Object,
    required: true,
  },
  empty: {
    type: Boolean,
    default: false,
  },
  emptyText: {
    type: String,
    default: '',
  },
  height: {
    type: String,
    default: '320px',
  },
  ariaLabel: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['chart-click'])
const { t } = useI18n()

let chart = null
let resizeObserver = null
let resizeFrame = 0
let renderFrame = 0

const chartRef = ref(null)
const rootRef = ref(null)
const resolvedEmptyText = computed(() => props.emptyText || t('charts.empty'))
const resolvedAriaLabel = computed(() => props.ariaLabel || t('charts.aria.chart'))

function canMeasureChart() {
  const element = chartRef.value
  if (!element?.isConnected) {
    return false
  }

  const rect = element.getBoundingClientRect()
  return rect.width > 0 && rect.height > 0
}

function scheduleResize() {
  if (resizeFrame) {
    cancelAnimationFrame(resizeFrame)
  }

  resizeFrame = requestAnimationFrame(() => {
    resizeFrame = 0
    if (chart && canMeasureChart()) {
      chart.resize()
    }
  })
}

function ensureChart() {
  if (!chartRef.value || props.empty || !canMeasureChart()) {
    return null
  }

  if (!chart) {
    chart = echarts.init(chartRef.value, null, {
      renderer: 'canvas',
    })
    chart.on('click', (params) => {
      emit('chart-click', params)
    })
  }

  return chart
}

function renderChart() {
  if (props.empty) {
    chart?.clear()
    return
  }

  const instance = ensureChart()
  if (!instance) {
    return
  }

  instance.setOption(props.option, true)
  scheduleResize()
}

function scheduleRender() {
  if (renderFrame) {
    cancelAnimationFrame(renderFrame)
  }

  renderFrame = requestAnimationFrame(() => {
    renderFrame = 0
    renderChart()
  })
}

function resizeNow() {
  const instance = chart || ensureChart()
  if (!instance || !canMeasureChart()) {
    return false
  }

  instance.resize()
  return true
}

function getDataURL(options = {}) {
  if (props.empty) {
    return null
  }

  const instance = chart || ensureChart()
  if (!instance || !canMeasureChart()) {
    return null
  }

  try {
    instance.resize()
    return instance.getDataURL({
      type: 'png',
      pixelRatio: 2,
      backgroundColor: '#ffffff',
      ...options,
    })
  } catch {
    return null
  }
}

defineExpose({
  getDataURL,
  resizeNow,
})

onMounted(() => {
  scheduleRender()

  if (typeof ResizeObserver !== 'undefined' && rootRef.value) {
    resizeObserver = new ResizeObserver(() => {
      if (!chart && !props.empty) {
        scheduleRender()
      }

      scheduleResize()
    })
    resizeObserver.observe(rootRef.value)
  }

  window.addEventListener('resize', scheduleResize)
})

watch(
  () => [props.option, props.empty],
  () => {
    nextTick(scheduleRender)
  },
  {
    deep: true,
  },
)

onBeforeUnmount(() => {
  if (resizeFrame) {
    cancelAnimationFrame(resizeFrame)
  }

  if (renderFrame) {
    cancelAnimationFrame(renderFrame)
  }

  resizeObserver?.disconnect()
  window.removeEventListener('resize', scheduleResize)
  chart?.off('click')
  chart?.dispose()
  chart = null
})
</script>

<template>
  <div ref="rootRef" class="base-echart" :style="{ '--chart-height': height }">
    <div
      v-show="!empty"
      ref="chartRef"
      class="base-echart-canvas"
      role="img"
      :aria-label="resolvedAriaLabel"
    />
    <div v-if="empty" class="base-echart-empty">
      {{ resolvedEmptyText }}
    </div>
  </div>
</template>

<style scoped>
.base-echart {
  display: grid;
  min-height: var(--chart-height);
  width: 100%;
}

.base-echart-canvas {
  min-height: var(--chart-height);
  width: 100%;
}

.base-echart-empty {
  display: grid;
  place-items: center;
  min-height: var(--chart-height);
  padding: 20px;
  color: var(--color-ink-muted);
  background: var(--color-canvas-muted);
  border: 1px dashed var(--color-hairline-strong);
  border-radius: var(--radius-md);
  font-size: 13px;
  line-height: 1.45;
  text-align: center;
}
</style>
