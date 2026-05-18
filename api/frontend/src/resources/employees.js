import { createApiClient, getStoredApiBase } from '../api'

const api = createApiClient(() => getStoredApiBase())

export function isEmployeeAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

export function listEmployees(filters = {}, options = {}) {
  const params = new URLSearchParams()

  params.set('page', String(filters.page ?? 1))
  params.set('rows', String(filters.rows ?? 100))

  if (filters.id?.trim()) {
    params.set('id', filters.id.trim())
  }

  if (filters.departmentId) {
    params.set('department_id', filters.departmentId)
  }

  if (filters.fullName?.trim()) {
    params.set('full_name', filters.fullName.trim())
  }

  if (filters.email?.trim()) {
    params.set('email', filters.email.trim())
  }

  if (filters.phone?.trim()) {
    params.set('phone', filters.phone.trim())
  }

  return api.get(`/employee?${params.toString()}`, options)
}

export function getEmployee(id, options = {}) {
  return api.get(`/employee/${encodeURIComponent(id)}`, options)
}

export function createEmployee(payload, options = {}) {
  return api.post('/employee', payload, options)
}

export function importEmployees(file, options = {}) {
  const formData = new FormData()
  formData.append('file', file)

  return api.post('/employee/import', formData, options)
}

export function updateEmployee(id, payload, options = {}) {
  return api.put(`/employee/${encodeURIComponent(id)}`, payload, options)
}

export function deleteEmployee(id, options = {}) {
  return api.delete(`/employee/${encodeURIComponent(id)}`, options)
}
