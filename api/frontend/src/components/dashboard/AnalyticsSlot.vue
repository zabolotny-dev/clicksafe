<script setup>
import SkeletonBlock from '../ui/SkeletonBlock.vue'

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
  <section class="analytics-slot">
    <header class="analytics-slot-header">
      <div>
        <h2>{{ title }}</h2>
        <p v-if="description">{{ description }}</p>
      </div>
    </header>

    <div class="analytics-slot-body">
      <slot>
        <div class="chart-placeholder" aria-hidden="true">
          <SkeletonBlock :rows="5" />
        </div>
      </slot>
    </div>
  </section>
</template>
