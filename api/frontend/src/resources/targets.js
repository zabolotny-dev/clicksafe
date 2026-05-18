import { createApiClient, getStoredApiBase } from '../api'

const api = createApiClient(() => getStoredApiBase())

export function isTargetAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

export function createTarget(payload, options = {}) {
  return api.post('/target', {
    employee_id: payload.employee_id,
    campaign_id: payload.campaign_id,
  }, options)
}

export function scheduleTarget(id, scheduledAt, options = {}) {
  return api.put(`/target/${encodeURIComponent(id)}/schedule`, {
    scheduled_at: scheduledAt,
  }, options)
}

export function distributeCampaignTargets(campaignId, options = {}) {
  return api.put(`/target/campaign/${encodeURIComponent(campaignId)}/distribute`, undefined, options)
}

export function deleteTarget(id, options = {}) {
  return api.delete(`/target/${encodeURIComponent(id)}`, options)
}
