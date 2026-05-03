const API_BASE_STORAGE_KEY = 'clicksafe.apiBaseUrl'

export function getStoredApiBase() {
  return localStorage.getItem(API_BASE_STORAGE_KEY) ?? import.meta.env.VITE_API_BASE_URL ?? ''
}

export function storeApiBase(value) {
  localStorage.setItem(API_BASE_STORAGE_KEY, value.trim())
}

function joinUrl(base, path) {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`

  if (!base) {
    return normalizedPath
  }

  return `${base.replace(/\/$/, '')}${normalizedPath}`
}

async function readResponse(response) {
  const contentType = response.headers.get('content-type') ?? ''
  const text = await response.text()

  if (response.status === 204 || !text) {
    return null
  }

  if (contentType.includes('application/json')) {
    return JSON.parse(text)
  }

  return text
}

function formatApiError(status, payload) {
  if (!payload) {
    return `HTTP ${status}`
  }

  if (typeof payload === 'string') {
    return payload
  }

  const details = Array.isArray(payload.details)
    ? payload.details.map((item) => `${item.field}: ${item.error}`).join('; ')
    : ''

  return [payload.message, details].filter(Boolean).join(' | ') || `HTTP ${status}`
}

export function createApiClient(getBaseUrl, onTrace) {
  async function request(path, options = {}) {
    const method = options.method ?? 'GET'
    const startedAt = performance.now()
    const headers = { ...(options.headers ?? {}) }

    if (options.body && !(options.body instanceof FormData)) {
      headers['Content-Type'] = 'application/json'
    }

    const response = await fetch(joinUrl(getBaseUrl(), path), {
      ...options,
      headers,
    })
    const payload = await readResponse(response)
    const elapsedMs = Math.round(performance.now() - startedAt)

    onTrace?.({
      method,
      path,
      status: response.status,
      ok: response.ok,
      elapsedMs,
    })

    if (!response.ok) {
      const error = new Error(formatApiError(response.status, payload))
      error.status = response.status
      error.payload = payload
      throw error
    }

    return payload
  }

  return {
    get: (path, options) => request(path, { ...options, method: 'GET' }),
    delete: (path, options) => request(path, { ...options, method: 'DELETE' }),
    post: (path, body, options) =>
      request(path, {
        ...options,
        method: 'POST',
        body: body instanceof FormData ? body : JSON.stringify(body ?? {}),
      }),
    put: (path, body, options) =>
      request(path, {
        ...options,
        method: 'PUT',
        body: body instanceof FormData ? body : JSON.stringify(body ?? {}),
      }),
  }
}
