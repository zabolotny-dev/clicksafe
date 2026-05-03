<script setup>
import { Connection } from '@element-plus/icons-vue'

const activeSection = defineModel('activeSection', { type: String, required: true })
const apiBaseUrl = defineModel('apiBaseUrl', { type: String, required: true })

defineProps({
  apiBaseLabel: {
    type: String,
    required: true,
  },
  navItems: {
    type: Array,
    required: true,
  },
})
</script>

<template>
  <aside class="sidebar">
    <div class="brand">
      <div class="brand-mark">CS</div>
      <div>
        <strong>ClickSafe</strong>
        <span>API Console</span>
      </div>
    </div>

    <el-menu
      class="nav-menu"
      :default-active="activeSection"
      @select="activeSection = $event"
    >
      <el-menu-item
        v-for="item in navItems"
        :key="item.key"
        :index="item.key"
      >
        <el-icon><component :is="item.icon" /></el-icon>
        <span>{{ item.label }}</span>
      </el-menu-item>
    </el-menu>

    <div class="api-box">
      <div class="mini-label">API base</div>
      <el-input
        v-model="apiBaseUrl"
        placeholder="same origin"
        clearable
        size="small"
      >
        <template #prefix>
          <el-icon><Connection /></el-icon>
        </template>
      </el-input>
      <small>{{ apiBaseLabel }}</small>
    </div>
  </aside>
</template>
