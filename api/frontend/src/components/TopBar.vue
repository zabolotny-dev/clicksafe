<script setup>
import { MagicStick, Refresh } from '@element-plus/icons-vue'

defineProps({
  activeTitle: {
    type: String,
    required: true,
  },
  loading: {
    type: Object,
    required: true,
  },
})

defineEmits(['bootstrap', 'refresh'])
</script>

<template>
  <header class="topbar">
    <div>
      <div class="mini-label">app layer demo</div>
      <h1>{{ activeTitle }}</h1>
    </div>
    <div class="topbar-actions">
      <el-button
        :icon="Refresh"
        :loading="loading.organization || loading.departments || loading.employees || loading.messages"
        @click="$emit('refresh')"
      >
        Обновить
      </el-button>
      <el-button
        type="primary"
        :icon="MagicStick"
        :loading="loading.bootstrap"
        @click="$emit('bootstrap')"
      >
        Демо-цепочка
      </el-button>
    </div>
  </header>
</template>
