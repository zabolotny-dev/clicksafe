<script setup>
import { useI18n } from 'vue-i18n'
import { translateEventType } from '../../i18n'

const { t } = useI18n()

const events = [
  'MESSAGE_SENT',
  'DELIVERY_FAILED',
  'EMAIL_OPENED',
  'LINK_OPENED',
  'DATA_SENT',
]
</script>

<template>
  <section class="event-timeline" aria-labelledby="recent-events-title">
    <header class="dashboard-panel-header">
      <div>
        <h2 id="recent-events-title">{{ t('pages.dashboard.recentEventsTitle') }}</h2>
        <p>{{ t('pages.dashboard.recentEventsDescription') }}</p>
      </div>
    </header>

    <ol class="event-list">
      <li v-for="event in events" :key="event" class="event-item">
        <span class="event-dot" aria-hidden="true" />
        <code>{{ translateEventType(event) }}</code>
        <span class="event-placeholder">—</span>
      </li>
    </ol>
  </section>
</template>
