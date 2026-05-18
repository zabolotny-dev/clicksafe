import { computed, readonly, ref } from 'vue'
import { createApiClient, getStoredApiBase } from '../api'
import { translateAuthValidationError, translateKnownError } from '../i18n'

const session = ref(null)
const csrfToken = ref('')
const isReady = ref(false)
const isLoading = ref(false)
const error = ref('')
let sessionRequest = null

const api = createApiClient(
  () => getStoredApiBase(),
  undefined,
  () => csrfToken.value,
)

const isAuthenticated = computed(() => Boolean(session.value))

function isAuthError(err) {
  return err?.status === 401 || err?.status === 403
}

function normalizeError(err) {
  return translateAuthValidationError(err) || translateKnownError(err, 'errors.requestFailed')
}

function setSession(payload) {
  session.value = payload
  csrfToken.value = payload?.csrf_token ?? ''
}

function clearSession() {
  session.value = null
  csrfToken.value = ''
}

async function loadCurrentSession() {
  if (sessionRequest) {
    return sessionRequest
  }

  isLoading.value = true
  error.value = ''

  sessionRequest = api.get('/me')
    .then((payload) => {
      setSession(payload)
      return session.value
    })
    .catch((err) => {
      clearSession()

      if (!isAuthError(err)) {
        error.value = normalizeError(err)
      }

      return null
    })
    .finally(() => {
      isReady.value = true
      isLoading.value = false
      sessionRequest = null
    })

  return sessionRequest
}

async function login(credentials) {
  isLoading.value = true
  error.value = ''

  try {
    const payload = await api.post('/login', {
      login: credentials.login,
      password: credentials.password,
    }, { skipCsrf: true })

    setSession(payload)
    isReady.value = true
    return session.value
  } catch (err) {
    clearSession()
    isReady.value = true
    error.value = normalizeError(err)
    return null
  } finally {
    isLoading.value = false
  }
}

async function logout() {
  isLoading.value = true
  error.value = ''

  try {
    await api.post('/logout', {})
  } catch {
    // Backend logout failure must not keep the frontend session alive.
  } finally {
    clearSession()
    isReady.value = true
    isLoading.value = false
  }
}

function clearError() {
  error.value = ''
}

export function useSession() {
  return {
    session: readonly(session),
    csrfToken: readonly(csrfToken),
    isAuthenticated,
    isReady: readonly(isReady),
    isLoading: readonly(isLoading),
    error: readonly(error),
    loadCurrentSession,
    login,
    logout,
    clearError,
  }
}
