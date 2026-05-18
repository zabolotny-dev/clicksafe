<script setup>
defineProps({
  label: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    default: '',
  },
  value: {
    type: String,
    default: '—',
  },
  icon: {
    type: [Object, Function],
    default: null,
  },
})
</script>

<template>
  <article class="metric-card">
    <header class="metric-card-header">
      <span v-if="icon" class="metric-icon" aria-hidden="true">
        <component :is="icon" :size="16" stroke-width="1.8" />
      </span>
      <span class="metric-label">{{ label }}</span>
      <slot name="label-actions" />
    </header>
    <strong class="metric-value">{{ value }}</strong>
    <p v-if="description" class="metric-description">{{ description }}</p>
  </article>
</template>
