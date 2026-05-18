<script setup>
defineProps({
  rows: {
    type: Number,
    default: 3,
  },
})
</script>

<template>
  <div class="skeleton-block" aria-hidden="true">
    <span
      v-for="row in rows"
      :key="row"
      class="skeleton-line"
      :class="`skeleton-line--${((row - 1) % 4) + 1}`"
    />
  </div>
</template>
