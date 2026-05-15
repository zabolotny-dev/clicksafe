<script setup>
import { MagicStick, Refresh, SwitchButton, User } from '@element-plus/icons-vue'

defineProps({
  activeTitle: {
    type: String,
    required: true,
  },
  loading: {
    type: Object,
    required: true,
  },
  sessionLabel: {
    type: String,
    required: true,
  },
})

defineEmits(['bootstrap', 'refresh', 'logout'])
</script>

<template>
  <header class="topbar">
    <div>
      <div class="mini-label">admin workspace</div>
      <h1>{{ activeTitle }}</h1>
    </div>
    <div class="topbar-actions">
      <el-tag
        class="session-tag"
        effect="plain"
        size="large"
      >
        <el-icon><User /></el-icon>
        {{ sessionLabel }}
      </el-tag>
      <el-button
        :icon="Refresh"
        :loading="loading.organization || loading.departments || loading.employees || loading.messages || loading.landings || loading.campaigns || loading.targets"
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
      <el-button
        :icon="SwitchButton"
        :loading="loading.auth"
        @click="$emit('logout')"
      >
        Выйти
      </el-button>
    </div>
  </header>
</template>
