<script setup>
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import UiButton from '../components/ui/UiButton.vue'
import UiInput from '../components/ui/UiInput.vue'
import { useSession } from '../composables/useSession'

const route = useRoute()
const router = useRouter()
const { error, isLoading, login, clearError } = useSession()
const { t } = useI18n()

const form = reactive({
  login: '',
  password: '',
})
const localError = ref('')

const canSubmit = computed(() => Boolean(form.login.trim() && form.password && !isLoading.value))
const visibleError = computed(() => localError.value || error.value)

function safeRedirectTarget(value) {
  const target = Array.isArray(value) ? value[0] : value

  if (!target || typeof target !== 'string') {
    return '/dashboard'
  }

  if (!target.startsWith('/') || target.startsWith('//') || target.includes('\\')) {
    return '/dashboard'
  }

  if (target === '/login' || target.startsWith('/login?')) {
    return '/dashboard'
  }

  return target
}

function resetErrors() {
  localError.value = ''
  clearError()
}

async function submit() {
  resetErrors()

  if (!form.login.trim() || !form.password) {
    localError.value = t('auth.enterCredentials')
    return
  }

  const result = await login({
    login: form.login.trim(),
    password: form.password,
  })

  if (result) {
    form.password = ''
    await router.replace(safeRedirectTarget(route.query.redirect))
  }
}
</script>

<template>
  <main class="auth-page">
    <section class="auth-card" :aria-label="t('app.name')">
      <div class="auth-brand">
        <strong>ClickSafe</strong>
      </div>

      <form class="auth-form" novalidate @submit.prevent="submit">
        <div v-if="visibleError" class="auth-error" role="alert">
          {{ visibleError }}
        </div>

        <UiInput
          id="login"
          v-model="form.login"
          :label="t('auth.login')"
          name="login"
          autocomplete="username"
          autofocus
          :disabled="isLoading"
          @update:model-value="resetErrors"
        />

        <UiInput
          id="password"
          v-model="form.password"
          :label="t('auth.password')"
          name="password"
          type="password"
          autocomplete="current-password"
          :disabled="isLoading"
          @update:model-value="resetErrors"
        />

        <UiButton
          class="auth-submit"
          type="submit"
          :loading="isLoading"
          :disabled="!canSubmit"
        >
          {{ t('auth.submit') }}
        </UiButton>
      </form>
    </section>
  </main>
</template>
