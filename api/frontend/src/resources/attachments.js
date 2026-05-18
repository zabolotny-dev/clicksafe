import { createApiClient, getStoredApiBase } from '../api'

const api = createApiClient(() => getStoredApiBase())

function joinUrl(base, path) {
  if (!base) {
    return path
  }

  return `${base.replace(/\/$/, '')}${path}`
}

export function isAttachmentAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

export function listAttachments(filters = {}, options = {}) {
  const params = new URLSearchParams()

  params.set('page', String(filters.page ?? 1))
  params.set('rows', String(filters.rows ?? 100))

  if (filters.id?.trim()) {
    params.set('id', filters.id.trim())
  }

  if (filters.label?.trim()) {
    params.set('label', filters.label.trim())
  }

  if (filters.type?.trim()) {
    params.set('type', filters.type.trim())
  }

  return api.get(`/attachment?${params.toString()}`, options)
}

export function getAttachmentContent(id, options = {}) {
  return api.get(`/attachment/${encodeURIComponent(id)}`, options)
}

export function getRenderedAttachmentContent(id, targetId, options = {}) {
  return api.get(`/attachment/${encodeURIComponent(id)}/render/${encodeURIComponent(targetId)}`, options)
}

export function getAttachmentUrl(id) {
  return joinUrl(getStoredApiBase(), `/attachment/${encodeURIComponent(id)}`)
}

export function uploadAttachment(file, payload = {}, options = {}) {
  const formData = new FormData()
  formData.append('file', file)

  if (payload.public !== undefined) {
    formData.append('public', String(Boolean(payload.public)))
  }

  return api.post('/attachment', formData, options)
}

export function updateAttachment(id, payload, options = {}) {
  const body = {}

  if (payload.label !== undefined) {
    body.label = payload.label
  }

  if (payload.public !== undefined) {
    body.public = String(Boolean(payload.public))
  }

  return api.put(`/attachment/${encodeURIComponent(id)}`, body, options)
}

export function deleteAttachment(id, options = {}) {
  return api.delete(`/attachment/${encodeURIComponent(id)}`, options)
}
