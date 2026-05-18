<script setup>
defineProps({
  variant: {
    type: String,
    default: 'info',
    validator: (value) => ['info', 'success', 'warning', 'error'].includes(value),
  },
  title: {
    type: String,
    default: '',
  },
  message: {
    type: String,
    default: '',
  },
  items: {
    type: Array,
    default: () => [],
  },
  dismissible: {
    type: Boolean,
    default: false,
  },
})

defineEmits(['dismiss'])
</script>

<template>
  <section
    class="ui-alert"
    :class="`ui-alert--${variant}`"
    :role="variant === 'error' ? 'alert' : 'status'"
  >
    <div class="ui-alert-marker" aria-hidden="true" />
    <div class="ui-alert-copy">
      <h3 v-if="title">{{ title }}</h3>
      <p v-if="message">{{ message }}</p>
      <ul v-if="items.length">
        <li v-for="item in items" :key="item">{{ item }}</li>
      </ul>
      <slot />
    </div>
    <button
      v-if="dismissible"
      type="button"
      class="ui-alert-dismiss"
      @click="$emit('dismiss')"
    >
      Dismiss
    </button>
  </section>
</template>
