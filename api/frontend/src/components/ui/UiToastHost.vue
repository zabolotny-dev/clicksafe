<script setup>
import { useI18n } from 'vue-i18n'
import { useNotifications } from '../../composables/useNotifications'

const { notifications, dismissNotification } = useNotifications()
const { t } = useI18n()

function toastRole(variant) {
  return variant === 'error' || variant === 'warning' ? 'alert' : 'status'
}
</script>

<template>
  <Teleport to="body">
    <TransitionGroup
      tag="section"
      name="ui-toast-list"
      class="ui-toast-host"
      :aria-label="t('common.notifications.host')"
      aria-live="polite"
    >
      <article
        v-for="notification in notifications"
        :key="notification.id"
        class="ui-toast"
        :class="`ui-toast--${notification.variant}`"
        :role="toastRole(notification.variant)"
      >
        <div class="ui-toast-marker" aria-hidden="true" />
        <div class="ui-toast-copy">
          <h2>{{ notification.title }}</h2>
          <p v-if="notification.message">{{ notification.message }}</p>
        </div>
        <button
          type="button"
          class="ui-toast-dismiss"
          :aria-label="t('common.notifications.dismissLabel', { title: notification.title })"
          @click="dismissNotification(notification.id)"
        >
          {{ t('common.actions.dismiss') }}
        </button>
        <div
          v-if="notification.autoDismiss"
          class="ui-toast-progress"
          :style="{ '--toast-duration': `${notification.duration}ms` }"
          aria-hidden="true"
        >
          <span />
        </div>
      </article>
    </TransitionGroup>
  </Teleport>
</template>
