import { useRoute, useRouter } from 'vue-router'
import { useSession } from './useSession'
import { isAuthError } from '../utils/resourceFormat'

export function useResourceActions() {
  const route = useRoute()
  const router = useRouter()
  const { csrfToken, loadCurrentSession } = useSession()

  function mutationOptions(extra = {}) {
    const headers = {
      ...(extra.headers || {}),
    }

    if (csrfToken.value) {
      headers['X-CSRF-Token'] = csrfToken.value
    }

    return {
      ...extra,
      headers,
    }
  }

  async function ensureCsrfToken() {
    if (csrfToken.value) {
      return true
    }

    return Boolean(await loadCurrentSession())
  }

  async function handleAuthError(error) {
    if (!isAuthError(error)) {
      return false
    }

    const session = await loadCurrentSession()

    if (!session) {
      await router.push({
        name: 'login',
        query: {
          redirect: route.fullPath,
        },
      })
    }

    return true
  }

  return {
    mutationOptions,
    ensureCsrfToken,
    handleAuthError,
  }
}
