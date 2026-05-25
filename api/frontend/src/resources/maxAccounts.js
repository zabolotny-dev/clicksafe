import { createApiClient, getStoredApiBase } from '../api'

const api = createApiClient(() => getStoredApiBase())

export function listMaxAccounts(filters = {}, options = {}) {
  const params = new URLSearchParams()

  params.set('page', String(filters.page ?? 1))
  params.set('rows', String(filters.rows ?? 10))

  if (filters.label?.trim()) {
    params.set('label', filters.label.trim())
  }

  if (filters.phone?.trim()) {
    params.set('phone', filters.phone.trim())
  }

  return api.get(`/max-account?${params.toString()}`, options)
}

export function beginMaxLogin(payload, options = {}) {
  return api.post('/max-account/login', payload, options)
}

export function confirmMaxLogin(attemptId, payload, options = {}) {
  return api.post(`/max-account/login/${encodeURIComponent(attemptId)}/confirm`, payload, options)
}

export function confirmMaxPassword(attemptId, payload, options = {}) {
  return api.post(`/max-account/login/${encodeURIComponent(attemptId)}/password`, payload, options)
}

export function updateMaxAccount(id, payload, options = {}) {
  return api.put(`/max-account/${encodeURIComponent(id)}`, payload, options)
}

export function connectMaxAccount(id, options = {}) {
  return api.post(`/max-account/${encodeURIComponent(id)}/connect`, {}, options)
}

export function disconnectMaxAccount(id, options = {}) {
  return api.post(`/max-account/${encodeURIComponent(id)}/disconnect`, {}, options)
}

export function deleteMaxAccount(id, options = {}) {
  return api.delete(`/max-account/${encodeURIComponent(id)}`, options)
}
