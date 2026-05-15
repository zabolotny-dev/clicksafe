<script setup>
import { Connection, Lock, User } from '@element-plus/icons-vue'

const apiBaseUrl = defineModel('apiBaseUrl', { type: String, required: true })

defineProps({
  apiBaseLabel: {
    type: String,
    required: true,
  },
  form: {
    type: Object,
    required: true,
  },
  loading: {
    type: Object,
    required: true,
  },
  lastError: {
    type: Object,
    default: null,
  },
})

defineEmits(['clear-error', 'login'])
</script>

<template>
  <main class="login-page">
    <section class="login-panel">
      <div class="brand login-brand">
        <div class="brand-mark">CS</div>
        <div>
          <strong>ClickSafe</strong>
          <span>Admin Panel</span>
        </div>
      </div>

      <el-alert
        v-if="lastError"
        class="login-alert"
        :title="lastError.title"
        :description="lastError.message"
        type="error"
        show-icon
        :closable="true"
        @close="$emit('clear-error')"
      />

      <el-form
        class="login-form"
        label-position="top"
        @keyup.enter="$emit('login')"
      >
        <el-form-item label="Логин">
          <el-input
            v-model="form.login"
            autocomplete="username"
            autofocus
            size="large"
          >
            <template #prefix>
              <el-icon><User /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="Пароль">
          <el-input
            v-model="form.password"
            autocomplete="current-password"
            show-password
            size="large"
            type="password"
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="API base">
          <el-input
            v-model="apiBaseUrl"
            clearable
            placeholder="same origin"
          >
            <template #prefix>
              <el-icon><Connection /></el-icon>
            </template>
          </el-input>
          <small class="field-note">{{ apiBaseLabel }}</small>
        </el-form-item>

        <el-button
          class="full-width"
          native-type="button"
          size="large"
          type="primary"
          :loading="loading.auth"
          @click="$emit('login')"
        >
          Войти
        </el-button>
      </el-form>
    </section>
  </main>
</template>
