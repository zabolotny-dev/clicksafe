<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import EmptyState from '../components/ui/EmptyState.vue'
import PageHeader from '../components/ui/PageHeader.vue'
import SkeletonBlock from '../components/ui/SkeletonBlock.vue'
import UiBadge from '../components/ui/UiBadge.vue'

const props = defineProps({
  titleKey: {
    type: String,
    required: true,
  },
  descriptionKey: {
    type: String,
    default: '',
  },
  sectionKey: {
    type: String,
    default: 'sections.settings',
  },
})

const { t } = useI18n()
const title = computed(() => t(props.titleKey))
const description = computed(() => (props.descriptionKey ? t(props.descriptionKey) : ''))
const section = computed(() => t(props.sectionKey))
</script>

<template>
  <div class="placeholder-page">
    <PageHeader :eyebrow="section" :title="title" :description="description">
      <template #actions>
        <UiBadge variant="info">{{ t('pages.placeholder.milestoneShell') }}</UiBadge>
      </template>
    </PageHeader>

    <div class="placeholder-grid">
      <EmptyState
        :title="t('pages.placeholder.reserved')"
        :description="t('pages.placeholder.description')"
      />

      <section class="placeholder-panel" :aria-label="t('pages.placeholder.panelTitle')">
        <div>
          <span class="panel-kicker">{{ t('pages.placeholder.panelKicker') }}</span>
          <h2>{{ t('pages.placeholder.panelTitle') }}</h2>
          <p>{{ t('pages.placeholder.panelDescription') }}</p>
        </div>
        <SkeletonBlock :rows="5" />
      </section>
    </div>
  </div>
</template>
