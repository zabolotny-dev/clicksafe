import { createApiClient, getStoredApiBase } from '../api'

const api = createApiClient(() => getStoredApiBase())

export function isOrganizationAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

export function isOrganizationNotFoundError(error) {
  return error?.status === 404
}

export function getOrganization(options = {}) {
  return api.get('/organization', options)
}

export function createOrganization(payload, options = {}) {
  return api.post('/organization', payload, options)
}

export function updateOrganization(payload, options = {}) {
  return api.put('/organization', payload, options)
}
