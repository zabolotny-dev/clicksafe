<script setup>
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { Pause, Play } from 'lucide-vue-next'

const props = defineProps({
  src: { type: String, required: true },
  blob: { type: Blob, default: null },
})

const wrapperEl = ref(null)
const canvasEl = ref(null)
const audioEl = ref(null)

const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const isLoading = ref(false)
const hasError = ref(false)

const BAR_WIDTH = 2
const BAR_GAP = 1
const CANVAS_HEIGHT = 64
// resolved from tokens
const COLOR_PRIMARY = '#422082'
const COLOR_PRIMARY_BORDER = '#d8d0ef'
const COLOR_BG = '#f6f9fc'

let peaks = []
let channelData = null
let resizeObserver = null
let rafId = null

function barCount() {
  return Math.floor((canvasEl.value?.width || 400) / (BAR_WIDTH + BAR_GAP))
}

function samplePeaks(n) {
  if (!channelData || !n) return []
  const step = Math.ceil(channelData.length / n)
  const out = []
  for (let i = 0; i < n; i++) {
    let max = 0
    const start = i * step
    for (let j = 0; j < step; j++) {
      const v = Math.abs(channelData[start + j] || 0)
      if (v > max) max = v
    }
    out.push(max)
  }
  const maxVal = Math.max(...out, 0.001)
  return out.map(v => v / maxVal)
}

function ensureCanvasSize() {
  if (!canvasEl.value || !wrapperEl.value) return false
  const w = wrapperEl.value.offsetWidth
  if (!w) return false
  if (canvasEl.value.width !== w || canvasEl.value.height !== CANVAS_HEIGHT) {
    canvasEl.value.width = w
    canvasEl.value.height = CANVAS_HEIGHT
  }
  return true
}

function draw() {
  if (!canvasEl.value) return
  const canvas = canvasEl.value
  const ctx = canvas.getContext('2d')
  const { width, height } = canvas
  const mid = height / 2
  const progress = duration.value > 0 ? currentTime.value / duration.value : 0
  const progressX = Math.round(progress * width)

  ctx.clearRect(0, 0, width, height)

  if (isLoading.value) {
    // static placeholder bars while decoding
    const n = Math.floor(width / (BAR_WIDTH + BAR_GAP))
    ctx.fillStyle = COLOR_PRIMARY_BORDER
    for (let i = 0; i < n; i++) {
      const x = i * (BAR_WIDTH + BAR_GAP)
      const amp = (Math.sin(i * 0.22) * 0.28 + 0.38) * (mid - 4)
      ctx.fillRect(x, Math.round(mid - amp), BAR_WIDTH, Math.round(amp * 2))
    }
    return
  }

  if (!peaks.length) return

  // bars: played = primary, unplayed = border
  peaks.forEach((peak, i) => {
    const x = i * (BAR_WIDTH + BAR_GAP)
    const amp = Math.max(peak * (mid - 4), 1.5)
    ctx.fillStyle = x < progressX ? COLOR_PRIMARY : COLOR_PRIMARY_BORDER
    ctx.fillRect(x, Math.round(mid - amp), BAR_WIDTH, Math.round(amp * 2))
  })

  // semi-transparent played region overlay
  if (progressX > 0) {
    ctx.fillStyle = 'rgba(66, 32, 130, 0.08)'
    ctx.fillRect(0, 0, progressX, height)
  }

  // playhead line
  if (progress > 0 && progress < 1) {
    ctx.fillStyle = COLOR_PRIMARY
    ctx.fillRect(progressX - 1, 0, 2, height)

    // circle handle at top
    ctx.beginPath()
    ctx.arc(progressX, 6, 5, 0, Math.PI * 2)
    ctx.fillStyle = COLOR_PRIMARY
    ctx.fill()
  }
}

async function loadWaveform(src) {
  if (!src) return
  peaks = []
  channelData = null
  hasError.value = false
  isLoading.value = true
  draw()

  try {
    const buf = props.blob
      ? await props.blob.arrayBuffer()
      : await fetch(src).then(r => r.arrayBuffer())

    const audioCtx = new (window.AudioContext || window.webkitAudioContext)()
    const decoded = await audioCtx.decodeAudioData(buf)
    audioCtx.close()

    channelData = decoded.getChannelData(0)
    // use decoded duration as ground truth — fixes Infinity for Ogg/Opus without headers
    if (!isFinite(duration.value) || !duration.value) {
      duration.value = decoded.duration
    }
    // ensure canvas is sized before computing bar count
    ensureCanvasSize()
    peaks = samplePeaks(barCount())
    // set isLoading=false BEFORE draw so the waveform is actually rendered
    isLoading.value = false
    draw()
  } catch {
    hasError.value = true
    isLoading.value = false
    draw()
  }
}

function onResize() {
  if (!ensureCanvasSize()) return
  if (channelData) {
    peaks = samplePeaks(barCount())
  }
  draw()
}

function seek(event) {
  if (!audioEl.value || !duration.value) return
  const rect = canvasEl.value.getBoundingClientRect()
  const ratio = Math.max(0, Math.min(1, (event.clientX - rect.left) / rect.width))
  audioEl.value.currentTime = ratio * duration.value
  currentTime.value = audioEl.value.currentTime
  draw()
}

function togglePlay() {
  if (!audioEl.value) return
  isPlaying.value ? audioEl.value.pause() : audioEl.value.play()
}

function formatTime(s) {
  if (!s || !isFinite(s) || isNaN(s)) return '--:--'
  const m = Math.floor(s / 60)
  const sec = Math.floor(s % 60)
  return `${m}:${sec.toString().padStart(2, '0')}`
}

function startRaf() {
  if (rafId) return
  function tick() {
    if (!audioEl.value || audioEl.value.paused) {
      rafId = null
      return
    }
    currentTime.value = audioEl.value.currentTime
    draw()
    rafId = requestAnimationFrame(tick)
  }
  rafId = requestAnimationFrame(tick)
}

function stopRaf() {
  if (rafId) {
    cancelAnimationFrame(rafId)
    rafId = null
  }
}

function onTimeUpdate() {
  // used only when paused (seeking via native controls or keyboard)
  if (!isPlaying.value) {
    currentTime.value = audioEl.value?.currentTime || 0
    draw()
  }
}

function onDurationChange() {
  duration.value = audioEl.value?.duration || 0
}

function onPlay() {
  isPlaying.value = true
  startRaf()
}

function onPause() {
  isPlaying.value = false
  stopRaf()
  currentTime.value = audioEl.value?.currentTime || 0
  draw()
}

function onEnded() {
  isPlaying.value = false
  stopRaf()
  currentTime.value = 0
  draw()
}

onMounted(async () => {
  await nextTick()
  ensureCanvasSize()
  resizeObserver = new ResizeObserver(onResize)
  if (wrapperEl.value) resizeObserver.observe(wrapperEl.value)
  loadWaveform(props.src)
})

watch(() => [props.src, props.blob], ([src]) => {
  currentTime.value = 0
  duration.value = 0
  isPlaying.value = false
  audioEl.value?.pause()
  loadWaveform(src)
})

onBeforeUnmount(() => {
  stopRaf()
  resizeObserver?.disconnect()
  audioEl.value?.pause()
})
</script>

<template>
  <div ref="wrapperEl" class="waveform-player">
    <audio
      ref="audioEl"
      :src="src"
      preload="metadata"
      @timeupdate="onTimeUpdate"
      @durationchange="onDurationChange"
      @play="onPlay"
      @pause="onPause"
      @ended="onEnded"
    />

    <canvas
      ref="canvasEl"
      class="waveform-player__canvas"
      :class="{ 'waveform-player__canvas--interactive': duration > 0 }"
      :height="64"
      @click="seek"
    />

    <div class="waveform-player__controls">
      <button
        class="waveform-player__play-btn"
        type="button"
        :aria-label="isPlaying ? 'Пауза' : 'Воспроизвести'"
        @click="togglePlay"
      >
        <Pause v-if="isPlaying" :size="16" stroke-width="2" />
        <Play v-else :size="16" stroke-width="2" />
      </button>

      <span class="waveform-player__time">
        {{ formatTime(currentTime) }}
        <span class="waveform-player__time-sep">/</span>
        {{ formatTime(duration) }}
      </span>

      <span v-if="isLoading" class="waveform-player__hint">Загрузка...</span>
      <span v-else-if="hasError" class="waveform-player__hint waveform-player__hint--error">Не удалось загрузить</span>
    </div>
  </div>
</template>

<style scoped>
.waveform-player {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  min-width: 0;
}

.waveform-player__canvas {
  display: block;
  width: 100%;
  height: 64px;
  border-radius: var(--radius-sm, 6px);
  background: var(--color-canvas-soft, #f6f9fc);
  cursor: default;
  user-select: none;
}

.waveform-player__canvas--interactive {
  cursor: pointer;
}

.waveform-player__controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.waveform-player__play-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: var(--color-primary, #422082);
  color: var(--color-on-primary, #ffffff);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s;
}

.waveform-player__play-btn:hover {
  background: var(--color-primary-hover, #5832a8);
}

.waveform-player__time {
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  color: var(--color-ink-secondary, #3b3348);
  display: flex;
  gap: 4px;
  align-items: center;
}

.waveform-player__time-sep {
  color: var(--color-ink-subtle, #8a8495);
}

.waveform-player__hint {
  font-size: 12px;
  color: var(--color-ink-subtle, #8a8495);
  margin-left: auto;
}

.waveform-player__hint--error {
  color: var(--color-danger, #c92a3a);
}
</style>
