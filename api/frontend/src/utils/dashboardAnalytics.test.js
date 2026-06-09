import { describe, it, expect, vi } from 'vitest'

vi.mock('../i18n', () => ({
  globalT: (key) => `[${key}]`,
  translateEventType: (type) => type,
  resolveLocaleTag: () => 'en-US',
  hasTranslation: () => true,
  translateKnownError: (error, fallbackKey) => fallbackKey || 'errors.requestFailed',
}))

vi.mock('../resources/vtargets', () => ({
  EVENT_TYPES: [
    'MESSAGE_SENT',
    'DELIVERY_FAILED',
    'EMAIL_OPENED',
    'LINK_OPENED',
    'DATA_SENT',
    'MESSAGE_READ',
    'MESSAGE_REPLIED',
  ],
}))

import { buildDashboardAnalytics, DASHBOARD_RISK_SCORES } from './dashboardAnalytics'

function makeTarget(overrides = {}) {
  return {
    id: overrides.id ?? 'target-1',
    campaign_id: overrides.campaign_id ?? 'campaign-1',
    employee_id: overrides.employee_id ?? 'emp-1',
    first_name: overrides.first_name ?? 'Ivan',
    last_name: overrides.last_name ?? 'Ivanov',
    status: overrides.status ?? 'SENT',
    events: overrides.events ?? [],
    ...overrides,
  }
}

function makeEvent(type, occurredAt = '2025-03-15T10:00:00Z') {
  return { type, occurred_at: occurredAt }
}

function makeCampaign(overrides = {}) {
  return {
    id: overrides.id ?? 'campaign-1',
    label: overrides.label ?? 'Test Campaign',
    status: overrides.status ?? 'ACTIVE',
    date_from: overrides.date_from ?? '2025-01-01',
    date_to: overrides.date_to ?? '2025-12-31',
    ...overrides,
  }
}

describe('DASHBOARD_RISK_SCORES', () => {
  it('has zero risk for sent and failed events', () => {
    expect(DASHBOARD_RISK_SCORES['MESSAGE_SENT']).toBe(0)
    expect(DASHBOARD_RISK_SCORES['DELIVERY_FAILED']).toBe(0)
  })

  it('has highest risk for DATA_SENT', () => {
    expect(DASHBOARD_RISK_SCORES['DATA_SENT']).toBe(100)
  })

  it('ranks link click above email open', () => {
    expect(DASHBOARD_RISK_SCORES['LINK_OPENED']).toBeGreaterThan(DASHBOARD_RISK_SCORES['EMAIL_OPENED'])
  })
})

describe('buildDashboardAnalytics — empty input', () => {
  it('returns safe defaults for empty datasets', () => {
    const result = buildDashboardAnalytics({
      campaigns: [],
      vtargets: [],
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    expect(result.hasActivity).toBe(false)
    expect(result.kpis.activeCampaigns).toBe('0')
    expect(result.kpis.riskIndex).toBe('[common.placeholder]')
    expect(result.activeCampaignRows).toEqual([])
    expect(result.recentEvents).toEqual([])
    expect(result.topDepartments).toEqual([])
  })
})

describe('buildDashboardAnalytics — KPI calculations', () => {
  it('counts active campaigns correctly', () => {
    const result = buildDashboardAnalytics({
      campaigns: [
        makeCampaign({ id: '1', status: 'ACTIVE' }),
        makeCampaign({ id: '2', status: 'ACTIVE' }),
        makeCampaign({ id: '3', status: 'DRAFT' }),
      ],
      vtargets: [],
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    expect(result.kpis.activeCampaigns).toBe('2')
  })

  it('computes open rate as percent of sent targets that opened', () => {
    const targets = [
      makeTarget({ id: 't1', events: [makeEvent('MESSAGE_SENT'), makeEvent('EMAIL_OPENED')] }),
      makeTarget({ id: 't2', events: [makeEvent('MESSAGE_SENT')] }),
    ]

    const result = buildDashboardAnalytics({
      campaigns: [makeCampaign()],
      vtargets: targets,
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    expect(result.kpis.openRate).toBe('50%')
  })

  it('computes click rate for LINK_OPENED events', () => {
    const targets = [
      makeTarget({ id: 't1', events: [makeEvent('MESSAGE_SENT'), makeEvent('LINK_OPENED')] }),
      makeTarget({ id: 't2', events: [makeEvent('MESSAGE_SENT')] }),
      makeTarget({ id: 't3', events: [makeEvent('MESSAGE_SENT')] }),
      makeTarget({ id: 't4', events: [makeEvent('MESSAGE_SENT')] }),
    ]

    const result = buildDashboardAnalytics({
      campaigns: [makeCampaign()],
      vtargets: targets,
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    expect(result.kpis.clickRate).toBe('25%')
  })

  it('computes risk index as average risk across targets with events', () => {
    const targets = [
      makeTarget({ id: 't1', events: [makeEvent('DATA_SENT')] }),    // risk 100
      makeTarget({ id: 't2', events: [makeEvent('EMAIL_OPENED')] }), // risk 15
    ]

    const result = buildDashboardAnalytics({
      campaigns: [makeCampaign()],
      vtargets: targets,
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    expect(result.kpis.riskIndex).toBe('58')
  })
})

describe('buildDashboardAnalytics — year filtering', () => {
  it('only counts events from the selected year', () => {
    const targets = [
      makeTarget({
        id: 't1',
        events: [
          makeEvent('EMAIL_OPENED', '2025-06-01T10:00:00Z'),
          makeEvent('LINK_OPENED', '2024-06-01T10:00:00Z'),
        ],
      }),
    ]

    const result = buildDashboardAnalytics({
      campaigns: [makeCampaign()],
      vtargets: targets,
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    // Only EMAIL_OPENED from 2025 should count for risk
    expect(result.kpis.riskIndex).toBe('15')
  })
})

describe('buildDashboardAnalytics — event distribution', () => {
  it('counts events by type for all event types', () => {
    const targets = [
      makeTarget({ id: 't1', events: [makeEvent('EMAIL_OPENED'), makeEvent('EMAIL_OPENED')] }),
      makeTarget({ id: 't2', events: [makeEvent('LINK_OPENED')] }),
    ]

    const result = buildDashboardAnalytics({
      campaigns: [makeCampaign()],
      vtargets: targets,
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    const emailOpened = result.eventDistribution.find((d) => d.eventType === 'EMAIL_OPENED')
    const linkOpened = result.eventDistribution.find((d) => d.eventType === 'LINK_OPENED')
    const messageSent = result.eventDistribution.find((d) => d.eventType === 'MESSAGE_SENT')

    expect(emailOpened.count).toBe(2)
    expect(linkOpened.count).toBe(1)
    expect(messageSent.count).toBe(0)
  })
})

describe('buildDashboardAnalytics — recent events', () => {
  it('returns up to 8 most recent events sorted newest first', () => {
    const events = Array.from({ length: 10 }, (_, i) => ({
      type: 'EMAIL_OPENED',
      occurred_at: `2025-0${Math.min(i + 1, 9)}-${String(i + 1).padStart(2, '0')}T00:00:00Z`,
    }))

    const targets = [makeTarget({ id: 't1', events })]

    const result = buildDashboardAnalytics({
      campaigns: [makeCampaign()],
      vtargets: targets,
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    expect(result.recentEvents).toHaveLength(8)
    // newest event should be first
    expect(result.recentEvents[0].occurredAt >= result.recentEvents[1].occurredAt).toBe(true)
  })
})

describe('buildDashboardAnalytics — available years', () => {
  it('always includes current year', () => {
    const result = buildDashboardAnalytics({
      campaigns: [],
      vtargets: [],
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    expect(result.availableYears).toContain(new Date().getFullYear())
  })

  it('collects years from event timestamps', () => {
    const targets = [
      makeTarget({ events: [makeEvent('EMAIL_OPENED', '2023-01-15T10:00:00Z')] }),
    ]

    const result = buildDashboardAnalytics({
      campaigns: [makeCampaign()],
      vtargets: targets,
      employees: [],
      departments: [],
      selectedYear: 2023,
    })

    expect(result.availableYears).toContain(2023)
  })
})

describe('buildDashboardAnalytics — active campaign rows', () => {
  it('only includes ACTIVE campaigns', () => {
    const result = buildDashboardAnalytics({
      campaigns: [
        makeCampaign({ id: '1', status: 'ACTIVE' }),
        makeCampaign({ id: '2', status: 'DRAFT' }),
        makeCampaign({ id: '3', status: 'COMPLETED' }),
      ],
      vtargets: [],
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    expect(result.activeCampaignRows).toHaveLength(1)
    expect(result.activeCampaignRows[0].id).toBe('1')
  })

  it('includes target count per campaign', () => {
    const targets = [
      makeTarget({ id: 't1', campaign_id: 'campaign-1' }),
      makeTarget({ id: 't2', campaign_id: 'campaign-1' }),
    ]

    const result = buildDashboardAnalytics({
      campaigns: [makeCampaign({ id: 'campaign-1', status: 'ACTIVE' })],
      vtargets: targets,
      employees: [],
      departments: [],
      selectedYear: 2025,
    })

    expect(result.activeCampaignRows[0].targetCount).toBe(2)
  })
})
