import { createApiClient, getStoredApiBase } from '../api'

const api = createApiClient(() => getStoredApiBase())

function joinUrl(base, path) {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`

  if (!base) {
    return normalizedPath
  }

  return `${base.replace(/\/$/, '')}${normalizedPath}`
}

function formatFieldDetail(item) {
  if (item.error) {
    return `${item.field}: ${item.error}`
  }

  if (item.value !== undefined && item.value !== null) {
    return `${item.field}: ${typeof item.value === 'object' ? JSON.stringify(item.value) : item.value}`
  }

  return item.field
}

function formatApiError(status, payload) {
  if (!payload) {
    return `HTTP ${status}`
  }

  if (typeof payload === 'string') {
    return payload
  }

  const details = Array.isArray(payload.details)
    ? payload.details.map(formatFieldDetail).join('; ')
    : ''

  return [payload.message, details].filter(Boolean).join(' | ') || `HTTP ${status}`
}

async function readErrorPayload(response) {
  const contentType = response.headers.get('content-type') ?? ''
  const text = await response.text()

  if (!text) {
    return null
  }

  if (contentType.includes('application/json')) {
    return JSON.parse(text)
  }

  return text
}

function headerInt(response, name) {
  const raw = response.headers.get(name)
  if (!raw) {
    return 0
  }

  const value = Number.parseInt(raw, 10)
  return Number.isFinite(value) ? value : 0
}

export function getVoiceStatus(options = {}) {
  return api.get('/voice/status', options)
}

export function transcribeVoice(formData, options = {}) {
  return api.post('/voice/transcribe', formData, options)
}

export async function cloneVoice(formData, options = {}) {
  const response = await fetch(joinUrl(getStoredApiBase(), '/voice/clone'), {
    method: 'POST',
    body: formData,
    headers: options.headers || {},
    credentials: 'include',
  })

  if (!response.ok) {
    const payload = await readErrorPayload(response)
    const error = new Error(formatApiError(response.status, payload))
    error.status = response.status
    error.payload = payload
    throw error
  }

  const blob = await response.blob()
  const mimeType = response.headers.get('content-type') || blob.type || 'audio/ogg'

  return {
    blob,
    mimeType,
    extension: response.headers.get('x-voice-extension') || '.ogg',
    durationMs: headerInt(response, 'x-voice-duration-ms'),
    sampleRateHz: headerInt(response, 'x-voice-sample-rate-hz'),
    totalBytes: headerInt(response, 'x-voice-total-bytes') || blob.size,
    requestId: response.headers.get('x-voice-request-id') || '',
  }
}
