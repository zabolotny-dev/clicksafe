import { createApiClient, getStoredApiBase } from '../api'

const api = createApiClient(() => getStoredApiBase())

export function isDepartmentAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

export function listDepartments(filters = {}, options = {}) {
  const params = new URLSearchParams()

  params.set('page', String(filters.page ?? 1))
  params.set('rows', String(filters.rows ?? 100))

  if (filters.id?.trim()) {
    params.set('id', filters.id.trim())
  }

  if (filters.label?.trim()) {
    params.set('label', filters.label.trim())
  }

  return api.get(`/department?${params.toString()}`, options)
}

export function getDepartment(id, options = {}) {
  return api.get(`/department/${encodeURIComponent(id)}`, options)
}

export function createDepartment(payload, options = {}) {
  return api.post('/department', payload, options)
}

export function importDepartments(file, options = {}) {
  const formData = new FormData()
  formData.append('file', file)

  return api.post('/department/import', formData, options)
}

export function updateDepartment(id, payload, options = {}) {
  return api.put(`/department/${encodeURIComponent(id)}`, payload, options)
}

export function deleteDepartment(id, options = {}) {
  return api.delete(`/department/${encodeURIComponent(id)}`, options)
}
