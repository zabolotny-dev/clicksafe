import { createApiClient, getStoredApiBase } from '../api'

const api = createApiClient(() => getStoredApiBase())

export function isMessageAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

export function listMessages(filters = {}, options = {}) {
  const params = new URLSearchParams()

  params.set('page', String(filters.page ?? 1))
  params.set('rows', String(filters.rows ?? 100))

  if (filters.id?.trim()) {
    params.set('id', filters.id.trim())
  }

  if (filters.label?.trim()) {
    params.set('label', filters.label.trim())
  }

  if (filters.fromEmail?.trim()) {
    params.set('from_email', filters.fromEmail.trim())
  }

  if (filters.fromName?.trim()) {
    params.set('from_name', filters.fromName.trim())
  }

  if (filters.subject?.trim()) {
    params.set('subject', filters.subject.trim())
  }

  return api.get(`/message?${params.toString()}`, options)
}

export function getMessage(id, options = {}) {
  return api.get(`/message/${encodeURIComponent(id)}`, options)
}

export function createMessage(payload, options = {}) {
  return api.post('/message', payload, options)
}

export function updateMessage(id, payload, options = {}) {
  return api.put(`/message/${encodeURIComponent(id)}`, payload, options)
}

export function deleteMessage(id, options = {}) {
  return api.delete(`/message/${encodeURIComponent(id)}`, options)
}
