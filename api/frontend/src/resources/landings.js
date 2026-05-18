import { createApiClient, getStoredApiBase } from '../api'

const api = createApiClient(() => getStoredApiBase())

export function isLandingAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

export function listLandings(filters = {}, options = {}) {
  const params = new URLSearchParams()

  params.set('page', String(filters.page ?? 1))
  params.set('rows', String(filters.rows ?? 100))

  if (filters.id?.trim()) {
    params.set('id', filters.id.trim())
  }

  if (filters.label?.trim()) {
    params.set('label', filters.label.trim())
  }

  return api.get(`/landing?${params.toString()}`, options)
}

export function getLanding(id, options = {}) {
  return api.get(`/landing/${encodeURIComponent(id)}`, options)
}

export function createLanding(payload, options = {}) {
  return api.post('/landing', payload, options)
}

export function updateLanding(id, payload, options = {}) {
  return api.put(`/landing/${encodeURIComponent(id)}`, payload, options)
}

export function deleteLanding(id, options = {}) {
  return api.delete(`/landing/${encodeURIComponent(id)}`, options)
}
