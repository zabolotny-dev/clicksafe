import { createApiClient, getStoredApiBase } from '../api'

export const CAMPAIGN_STATUSES = [
  'DRAFT',
  'ACTIVE',
  'PAUSED',
  'COMPLETED',
  'CANCELED',
]

const api = createApiClient(() => getStoredApiBase())

function toRFC3339Date(value, edge) {
  if (!value) {
    return ''
  }

  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match) {
    return ''
  }

  const year = Number(match[1])
  const month = Number(match[2])
  const day = Number(match[3])
  const date = edge === 'end'
    ? new Date(year, month - 1, day, 23, 59, 59, 999)
    : new Date(year, month - 1, day, 0, 0, 0, 0)

  if (
    date.getFullYear() !== year
    || date.getMonth() !== month - 1
    || date.getDate() !== day
  ) {
    return ''
  }

  return date.toISOString()
}

export function isCampaignAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

export function isCampaignNotFoundError(error) {
  return error?.status === 404
}

export function listCampaigns(filters = {}, options = {}) {
  const params = new URLSearchParams()
  const label = filters.label?.trim()
  const dateFrom = toRFC3339Date(filters.dateFrom, 'start')
  const dateTo = toRFC3339Date(filters.dateTo, 'end')

  params.set('page', String(filters.page ?? 1))
  params.set('rows', String(filters.rows ?? 100))

  if (label) {
    params.set('label', label)
  }

  if (CAMPAIGN_STATUSES.includes(filters.status)) {
    params.set('status', filters.status)
  }

  if (dateFrom) {
    params.set('date_from', dateFrom)
  }

  if (dateTo) {
    params.set('date_to', dateTo)
  }

  return api.get(`/campaign?${params.toString()}`, options)
}

export function createCampaign(payload, options = {}) {
  return api.post('/campaign', payload, options)
}

export function getCampaign(id, options = {}) {
  return api.get(`/campaign/${encodeURIComponent(id)}`, options)
}

export function updateCampaign(id, payload, options = {}) {
  return api.put(`/campaign/${encodeURIComponent(id)}`, payload, options)
}

export function startCampaign(id, options = {}) {
  return api.put(`/campaign/${encodeURIComponent(id)}/start`, {}, options)
}

export function pauseCampaign(id, options = {}) {
  return api.put(`/campaign/${encodeURIComponent(id)}/pause`, {}, options)
}

export function cancelCampaign(id, options = {}) {
  return api.put(`/campaign/${encodeURIComponent(id)}/cancel`, {}, options)
}

export function deleteCampaign(id, options = {}) {
  return api.delete(`/campaign/${encodeURIComponent(id)}`, options)
}
