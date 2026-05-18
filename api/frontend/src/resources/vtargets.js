import { createApiClient, getStoredApiBase } from '../api'
import { translateEventType } from '../i18n'

export const TARGET_STATUSES = [
  'PENDING',
  'SENT',
  'FAILED',
  'OPENED',
  'CLICKED',
  'SUBMITTED',
]

export const EVENT_TYPES = [
  'MESSAGE_SENT',
  'DELIVERY_FAILED',
  'EMAIL_OPENED',
  'LINK_OPENED',
  'DATA_SENT',
]

export const EVENT_LABELS = new Proxy({}, {
  get(_target, key) {
    return translateEventType(String(key))
  },
})

const api = createApiClient(() => getStoredApiBase())

export function isVTargetAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

export function listVTargets(filters = {}, options = {}) {
  const params = new URLSearchParams()

  params.set('page', String(filters.page ?? 1))
  params.set('rows', String(filters.rows ?? 100))

  if (filters.id) {
    params.set('target_id', filters.id)
  }

  if (filters.campaignId) {
    params.set('campaign_id', filters.campaignId)
  }

  if (filters.employeeId) {
    params.set('employee_id', filters.employeeId)
  }

  if (TARGET_STATUSES.includes(filters.status)) {
    params.set('status', filters.status)
  }

  if (filters.fullName?.trim()) {
    params.set('full_name', filters.fullName.trim())
  }

  return api.get(`/vtarget?${params.toString()}`, options)
}
