import { globalT, translateEventType } from '../i18n'
import { EVENT_TYPES } from '../resources/vtargets'
import { employeeName, formatDateRange as formatLocalizedDateRange, formatDateTime, formatPercent as formatLocalizedPercent, nullableId } from './resourceFormat'

export const DASHBOARD_RISK_SCORES = {
  MESSAGE_SENT: 0,
  DELIVERY_FAILED: 0,
  EMAIL_OPENED: 15,
  LINK_OPENED: 45,
  DATA_SENT: 100,
}

const SENT_OR_LATER_EVENTS = ['MESSAGE_SENT', 'EMAIL_OPENED', 'LINK_OPENED', 'DATA_SENT']

function makeIdMap(items) {
  return new Map((items || []).map((item) => [String(item.id), item]))
}

function eventTime(event) {
  const date = new Date(event?.occurred_at)
  return Number.isNaN(date.getTime()) ? 0 : date.getTime()
}

function eventYear(event) {
  const time = eventTime(event)
  return time ? new Date(time).getFullYear() : null
}

function dateKey(date) {
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, '0'),
    String(date.getDate()).padStart(2, '0'),
  ].join('-')
}

function supportedEventsForTarget(target) {
  return Array.isArray(target?.events)
    ? target.events.filter((event) => EVENT_TYPES.includes(event?.type) && eventTime(event))
    : []
}

function targetRisk(events) {
  return events.reduce((max, event) => Math.max(max, DASHBOARD_RISK_SCORES[event.type] ?? 0), 0)
}

function hasAnyEvent(events, eventTypes) {
  return events.some((event) => eventTypes.includes(event.type))
}

function percent(numerator, denominator) {
  if (!denominator) {
    return null
  }

  return Math.round((numerator / denominator) * 100)
}

function formatPercent(value) {
  return formatLocalizedPercent(value)
}

function formatMetric(value) {
  return value == null ? globalT('common.placeholder') : String(value)
}

function campaignLabel(campaign, fallbackId) {
  return campaign?.label || `${globalT('common.labels.campaign')} ${String(fallbackId || '').slice(0, 8)}`
}

function formatDateRange(campaign) {
  return formatLocalizedDateRange(campaign?.date_from, campaign?.date_to)
}

function displayEmployeeName(target, employee) {
  const fromTarget = [target?.first_name, target?.last_name].filter(Boolean).join(' ').trim()
  const fromEmployee = employeeName(employee)

  if (fromTarget) {
    return fromTarget
  }

  return fromEmployee && fromEmployee !== globalT('common.placeholder') ? fromEmployee : globalT('common.labels.employee')
}

function departmentLabel(employee, departmentsById) {
  const departmentId = nullableId(employee?.department_id)

  if (!departmentId) {
    return globalT('common.selection.noDepartment')
  }

  return departmentsById.get(departmentId)?.label || globalT('common.labels.department')
}

function yearEventsForTarget(target, year) {
  return supportedEventsForTarget(target).filter((event) => eventYear(event) === year)
}

function eventsByTypeCount(scopedTargets, eventType) {
  return scopedTargets.reduce(
    (sum, item) => sum + item.events.filter((event) => event.type === eventType).length,
    0,
  )
}

function targetCountWithEvent(scopedTargets, eventTypes) {
  return scopedTargets.filter((item) => hasAnyEvent(item.events, eventTypes)).length
}

function buildCalendarDays(year, events) {
  const counts = new Map()
  events.forEach(({ event }) => {
    const time = eventTime(event)
    if (!time) {
      return
    }

    const key = dateKey(new Date(time))
    counts.set(key, (counts.get(key) || 0) + 1)
  })

  const days = []
  const cursor = new Date(year, 0, 1)
  const end = new Date(year, 11, 31)

  while (cursor <= end) {
    const key = dateKey(cursor)
    days.push([key, counts.get(key) || 0])
    cursor.setDate(cursor.getDate() + 1)
  }

  return days
}

function buildCampaignRows(campaigns, vtargets) {
  return (campaigns || [])
    .filter((campaign) => campaign.status === 'ACTIVE')
    .map((campaign) => {
      const targets = (vtargets || []).filter((target) => String(target.campaign_id) === String(campaign.id))
      const scopedTargets = targets.map((target) => ({
        target,
        events: supportedEventsForTarget(target),
      }))
      const sentTargets = targetCountWithEvent(scopedTargets, SENT_OR_LATER_EVENTS)
      const openedTargets = targetCountWithEvent(scopedTargets, ['EMAIL_OPENED'])
      const clickedTargets = targetCountWithEvent(scopedTargets, ['LINK_OPENED'])
      const submittedTargets = targetCountWithEvent(scopedTargets, ['DATA_SENT'])

      return {
        id: campaign.id,
        label: campaignLabel(campaign, campaign.id),
        status: campaign.status,
        dateRange: formatDateRange(campaign),
        targetCount: targets.length,
        openRate: formatPercent(percent(openedTargets, sentTargets)),
        clickRate: formatPercent(percent(clickedTargets, sentTargets)),
        submitRate: formatPercent(percent(submittedTargets, sentTargets)),
      }
    })
}

export function buildDashboardAnalytics({
  campaigns,
  vtargets,
  employees,
  departments,
  selectedYear,
}) {
  const year = Number(selectedYear) || new Date().getFullYear()
  const campaignsById = makeIdMap(campaigns)
  const employeesById = makeIdMap(employees)
  const departmentsById = makeIdMap(departments)

  const allSupportedEvents = (vtargets || []).flatMap((target) => (
    supportedEventsForTarget(target).map((event, eventIndex) => ({
      event,
      target,
      eventIndex,
    }))
  ))

  const availableYears = Array.from(new Set([
    new Date().getFullYear(),
    ...allSupportedEvents.map(({ event }) => eventYear(event)).filter(Boolean),
  ])).sort((left, right) => right - left)

  const scopedTargets = (vtargets || [])
    .map((target) => ({
      target,
      employee: employeesById.get(String(target.employee_id)) || null,
      events: yearEventsForTarget(target, year),
    }))

  const targetsWithEvents = scopedTargets.filter((item) => item.events.length)
  const scopedEvents = scopedTargets.flatMap((item) => (
    item.events.map((event, eventIndex) => ({
      event,
      target: item.target,
      employee: item.employee,
      eventIndex,
    }))
  ))

  const activeCampaigns = (campaigns || []).filter((campaign) => campaign.status === 'ACTIVE')
  const sentTargets = targetCountWithEvent(scopedTargets, SENT_OR_LATER_EVENTS)
  const openedTargets = targetCountWithEvent(scopedTargets, ['EMAIL_OPENED'])
  const clickedTargets = targetCountWithEvent(scopedTargets, ['LINK_OPENED'])
  const submittedTargets = targetCountWithEvent(scopedTargets, ['DATA_SENT'])
  const totalRisk = targetsWithEvents.reduce((sum, item) => sum + targetRisk(item.events), 0)
  const riskIndex = targetsWithEvents.length ? Math.round(totalRisk / targetsWithEvents.length) : null

  const eventDistribution = EVENT_TYPES.map((eventType) => ({
    eventType,
    label: translateEventType(eventType),
    count: eventsByTypeCount(scopedTargets, eventType),
  }))

  const employeeGroups = new Map()
  scopedTargets.forEach((item) => {
    if (!item.events.length) {
      return
    }

    const employeeId = String(item.target.employee_id || item.target.id || 'unknown')
    const current = employeeGroups.get(employeeId) || {
      id: employeeId,
      employee: item.employee,
      departmentLabel: departmentLabel(item.employee, departmentsById),
      targetCount: 0,
      riskTotal: 0,
    }

    current.targetCount += 1
    current.riskTotal += targetRisk(item.events)
    employeeGroups.set(employeeId, current)
  })

  const departmentGroups = new Map()
  Array.from(employeeGroups.values()).forEach((employee) => {
    const risk = employee.targetCount ? Math.round(employee.riskTotal / employee.targetCount) : 0
    const current = departmentGroups.get(employee.departmentLabel) || {
      label: employee.departmentLabel,
      employeeCount: 0,
      riskTotal: 0,
    }

    current.employeeCount += 1
    current.riskTotal += risk
    departmentGroups.set(employee.departmentLabel, current)
  })

  const topDepartments = Array.from(departmentGroups.values())
    .map((department) => ({
      ...department,
      risk: department.employeeCount ? Math.round(department.riskTotal / department.employeeCount) : 0,
    }))
    .sort((left, right) => right.risk - left.risk || right.employeeCount - left.employeeCount)
    .slice(0, 3)

  const recentEvents = scopedEvents
    .slice()
    .sort((left, right) => eventTime(right.event) - eventTime(left.event))
    .slice(0, 8)
    .map((item) => {
      const campaign = campaignsById.get(String(item.target.campaign_id))
      return {
        key: [
          item.target?.id || item.target?.token || 'target',
          item.event.type,
          item.event.occurred_at,
          item.eventIndex,
        ].join(':'),
        eventType: item.event.type,
        label: translateEventType(item.event.type),
        employeeLabel: displayEmployeeName(item.target, item.employee),
        campaignLabel: campaignLabel(campaign, item.target.campaign_id),
        occurredAt: item.event.occurred_at,
        occurredAtLabel: formatDateTime(item.event.occurred_at),
        targetStatus: item.target?.status || '',
      }
    })

  return {
    selectedYear: year,
    availableYears,
    hasActivity: scopedEvents.length > 0,
    calendarDays: buildCalendarDays(year, scopedEvents),
    eventDistribution,
    topDepartments,
    activeCampaignRows: buildCampaignRows(campaigns, vtargets),
    recentEvents,
    kpis: {
      activeCampaigns: String(activeCampaigns.length),
      riskIndex: formatMetric(riskIndex),
      openRate: formatPercent(percent(openedTargets, sentTargets)),
      clickRate: formatPercent(percent(clickedTargets, sentTargets)),
      submitRate: formatPercent(percent(submittedTargets, sentTargets)),
    },
  }
}
