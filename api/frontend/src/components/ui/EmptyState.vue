<script setup>
defineProps({
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    default: '',
  },
})
</script>

<template>
  <section class="empty-state">
    <div class="empty-state-mark" aria-hidden="true">
      <span />
    </div>
    <div>
      <h2>{{ title }}</h2>
      <p v-if="description">{{ description }}</p>
    </div>
    <div v-if="$slots.default" class="empty-state-actions">
      <slot />
    </div>
  </section>
</template>
