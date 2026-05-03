<script setup>
import { Delete } from '@element-plus/icons-vue'

defineProps({
  traces: {
    type: Array,
    required: true,
  },
})

defineEmits(['clear'])
</script>

<template>
  <section class="panel trace-panel">
    <div class="panel-heading">
      <div>
        <div class="mini-label">network</div>
        <h2>Последние запросы</h2>
      </div>
      <el-button :icon="Delete" @click="$emit('clear')" />
    </div>

    <div v-if="traces.length" class="trace-list">
      <div
        v-for="trace in traces"
        :key="trace.id"
        class="trace-row"
        :class="{ failed: !trace.ok }"
      >
        <span>{{ trace.at }}</span>
        <el-tag size="small" effect="plain">{{ trace.method }}</el-tag>
        <code>{{ trace.path }}</code>
        <strong>{{ trace.status }}</strong>
        <span>{{ trace.elapsedMs }} ms</span>
      </div>
    </div>
    <el-empty v-else description="Запросов пока нет" />
  </section>
</template>
