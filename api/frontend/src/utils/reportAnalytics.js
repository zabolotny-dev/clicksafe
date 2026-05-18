import { globalT, translateEventType } from '../i18n'
import { EVENT_TYPES } from '../resources/vtargets'
import { employeeName, formatDateTime, formatMonthDay, formatPercent as formatLocalizedPercent, nullableId } from './resourceFormat'

export const REPORT_RISK_SCORES = {
  MESSAGE_SENT: 0,
  DELIVERY_FAILED: 0,
  EMAIL_OPENED: 15,
  LINK_OPENED: 45,
  DATA_SENT: 100,
}

const FUNNEL_STAGES = [
  {
    eventType: 'MESSAGE_SENT',
    reachedBy: ['MESSAGE_SENT', 'EMAIL_OPENED', 'LINK_OPENED', 'DATA_SENT'],
  },
  {
    eventType: 'EMAIL_OPENED',
    reachedBy: ['EMAIL_OPENED', 'LINK_OPENED', 'DATA_SENT'],
  },
  {
    eventType: 'LINK_OPENED',
    reachedBy: ['LINK_OPENED', 'DATA_SENT'],
  },
  {
    eventType: 'DATA_SENT',
    reachedBy: ['DATA_SENT'],
  },
]

const RISK_SIGNAL_TYPES = ['EMAIL_OPENED', 'LINK_OPENED', 'DATA_SENT']

function makeIdMap(items) {
  return new Map((items || []).map((item) => [String(item.id), item]))
}

function dayBounds(value, edge) {
  if (!value) {
    return null
  }

  const date = new Date(`${value}T${edge === 'end' ? '23:59:59.999' : '00:00:00.000'}`)
  return Number.isNaN(date.getTime()) ? null : date
}

function eventTime(event) {
  const date = new Date(event?.occurred_at)
  return Number.isNaN(date.getTime()) ? 0 : date.getTime()
}

function isInDateRange(event, filters) {
  const time = eventTime(event)
  if (!time) {
    return false
  }

  const from = dayBounds(filters.dateFrom, 'start')
  const to = dayBounds(filters.dateTo, 'end')

  if (from && time < from.getTime()) {
    return false
  }

  if (to && time > to.getTime()) {
    return false
  }

  return true
}

function supportedEventsForTarget(target, filters) {
  return Array.isArray(target?.events)
    ? target.events
      .filter((event) => EVENT_TYPES.includes(event?.type))
      .filter((event) => isInDateRange(event, filters))
    : []
}

function uniqueTargetHasEvent(target, events, eventType) {
  return events.some((event) => event.type === eventType)
}

function uniqueTargetReaches(events, eventTypes) {
  return events.some((event) => eventTypes.includes(event.type))
}

function targetRisk(events) {
  return events.reduce((max, event) => Math.max(max, REPORT_RISK_SCORES[event.type] ?? 0), 0)
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

function dateKey(event) {
  const date = new Date(event.occurred_at)
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, '0'),
    String(date.getDate()).padStart(2, '0'),
  ].join('-')
}

function dayLabel(key) {
  const date = new Date(`${key}T00:00:00`)
  if (Number.isNaN(date.getTime())) {
    return key
  }

  return formatMonthDay(date)
}

function labelForCampaign(campaign, fallbackId) {
  return campaign?.label || `${globalT('common.labels.campaign')} ${String(fallbackId).slice(0, 8)}`
}

function labelForDepartment(departmentId, departmentsById) {
  if (!departmentId) {
    return globalT('common.selection.noDepartment')
  }

  return departmentsById.get(String(departmentId))?.label || globalT('common.labels.department')
}

function csvCell(value) {
  const text = String(value ?? '')
  return /[",\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text
}

function employeeDisplayName(target, employee) {
  const fromEmployee = employeeName(employee)
  if (fromEmployee && fromEmployee !== '-') {
    return fromEmployee
  }

  return [target?.first_name, target?.last_name].filter(Boolean).join(' ').trim() || globalT('common.labels.employee')
}

export function buildReportAnalytics({ campaigns, vtargets, employees, departments, filters }) {
  const campaignsById = makeIdMap(campaigns)
  const employeesById = makeIdMap(employees)
  const departmentsById = makeIdMap(departments)

  const scopedTargets = (vtargets || [])
    .filter((target) => !filters.campaignId || String(target.campaign_id) === filters.campaignId)
    .filter((target) => {
      if (!filters.departmentId) {
        return true
      }

      const employee = employeesById.get(String(target.employee_id))
      return nullableId(employee?.department_id) === filters.departmentId
    })
    .map((target) => ({
      target,
      employee: employeesById.get(String(target.employee_id)) || null,
      events: supportedEventsForTarget(target, filters),
    }))

  const targetsWithEvents = scopedTargets.filter((item) => item.events.length)
  const totalEvents = scopedTargets.reduce((sum, item) => sum + item.events.length, 0)
  const totalRisk = scopedTargets.reduce((sum, item) => sum + targetRisk(item.events), 0)
  const companyRiskValue = scopedTargets.length && totalEvents
    ? Math.round(totalRisk / scopedTargets.length)
    : null

  const sentTargets = scopedTargets.filter((item) => uniqueTargetReaches(item.events, FUNNEL_STAGES[0].reachedBy)).length
  const openedTargets = scopedTargets.filter((item) => uniqueTargetReaches(item.events, FUNNEL_STAGES[1].reachedBy)).length
  const clickedTargets = scopedTargets.filter((item) => uniqueTargetReaches(item.events, FUNNEL_STAGES[2].reachedBy)).length
  const submittedTargets = scopedTargets.filter((item) => uniqueTargetReaches(item.events, FUNNEL_STAGES[3].reachedBy)).length

  const campaignIds = new Set(scopedTargets.map((item) => String(item.target.campaign_id)))
  const employeeIds = new Set(scopedTargets.map((item) => String(item.target.employee_id)))

  const eventDistribution = EVENT_TYPES.map((eventType) => ({
    eventType,
    label: translateEventType(eventType),
    count: scopedTargets.reduce((sum, item) => sum + item.events.filter((event) => event.type === eventType).length, 0),
  }))

  const campaignRows = Array.from(campaignIds)
    .map((campaignId) => {
      const campaignTargets = scopedTargets.filter((item) => String(item.target.campaign_id) === campaignId)
      const eventCounts = Object.fromEntries(EVENT_TYPES.map((eventType) => [
        eventType,
        campaignTargets.filter((item) => uniqueTargetHasEvent(item.target, item.events, eventType)).length,
      ]))
      const risk = campaignTargets.length && campaignTargets.some((item) => item.events.length)
        ? Math.round(campaignTargets.reduce((sum, item) => sum + targetRisk(item.events), 0) / campaignTargets.length)
        : null

      return {
        id: campaignId,
        label: labelForCampaign(campaignsById.get(campaignId), campaignId),
        targetCount: campaignTargets.length,
        risk,
        eventCounts,
      }
    })
    .sort((left, right) => (right.risk ?? -1) - (left.risk ?? -1) || right.targetCount - left.targetCount)

  const employeeGroups = new Map()
  scopedTargets.forEach((item) => {
    const employeeId = String(item.target.employee_id)
    const employee = item.employee
    const departmentId = nullableId(employee?.department_id)
    const current = employeeGroups.get(employeeId) || {
      id: employeeId,
      label: employeeDisplayName(item.target, employee),
      departmentId,
      departmentLabel: labelForDepartment(departmentId, departmentsById),
      targetCount: 0,
      eventCount: 0,
      riskSignalCount: 0,
      riskTotal: 0,
      linkOpenedCount: 0,
      dataSubmittedCount: 0,
      lastActivity: null,
    }

    current.targetCount += 1
    current.eventCount += item.events.length
    current.riskTotal += targetRisk(item.events)
    current.riskSignalCount += item.events.filter((event) => RISK_SIGNAL_TYPES.includes(event.type)).length
    current.linkOpenedCount += item.events.filter((event) => event.type === 'LINK_OPENED').length
    current.dataSubmittedCount += item.events.filter((event) => event.type === 'DATA_SENT').length

    const latest = item.events.slice().sort((left, right) => eventTime(right) - eventTime(left))[0]
    if (latest && eventTime(latest) > eventTime(current.lastActivity)) {
      current.lastActivity = latest
    }

    employeeGroups.set(employeeId, current)
  })

  const employeeRows = Array.from(employeeGroups.values())
    .map((employee) => ({
      ...employee,
      risk: employee.targetCount ? Math.round(employee.riskTotal / employee.targetCount) : 0,
      lastActivityLabel: employee.lastActivity ? formatDateTime(employee.lastActivity.occurred_at) : '—',
    }))
    .sort((left, right) => right.risk - left.risk || right.dataSubmittedCount - left.dataSubmittedCount || right.linkOpenedCount - left.linkOpenedCount)

  const departmentGroups = new Map()
  employeeRows.forEach((employee) => {
    const key = employee.departmentId || 'none'
    const current = departmentGroups.get(key) || {
      id: key,
      label: employee.departmentLabel,
      employeeCount: 0,
      riskTotal: 0,
    }

    current.employeeCount += 1
    current.riskTotal += employee.risk
    departmentGroups.set(key, current)
  })

  const departmentRows = Array.from(departmentGroups.values())
    .map((department) => ({
      ...department,
      risk: department.employeeCount ? Math.round(department.riskTotal / department.employeeCount) : 0,
    }))
    .sort((left, right) => right.risk - left.risk || right.employeeCount - left.employeeCount)

  const funnelStages = FUNNEL_STAGES.map((stage) => ({
    eventType: stage.eventType,
    label: translateEventType(stage.eventType),
    count: scopedTargets.filter((item) => uniqueTargetReaches(item.events, stage.reachedBy)).length,
  }))

  const trendGroups = new Map()
  scopedTargets.forEach((item) => {
    item.events.forEach((event) => {
      const key = dateKey(event)
      const current = trendGroups.get(key) || {
        key,
        label: dayLabel(key),
        value: 0,
      }

      current.value += REPORT_RISK_SCORES[event.type] ?? 0
      trendGroups.set(key, current)
    })
  })

  const trendBuckets = Array.from(trendGroups.values()).sort((left, right) => left.key.localeCompare(right.key))
  const trendMarkAreas = trendBuckets
    .map((bucket, index) => {
      if (index === 0 || bucket.value <= trendBuckets[index - 1].value) {
        return null
      }

      return [
        { xAxis: trendBuckets[index - 1].label },
        { xAxis: bucket.label },
      ]
    })
    .filter(Boolean)

  return {
    filters,
    targets: scopedTargets,
    targetsWithEvents,
    hasSupportedEvents: totalEvents > 0,
    companyRiskValue,
    kpis: {
      companyRisk: formatMetric(companyRiskValue),
      campaignsAnalyzed: String(campaignIds.size),
      employeesTested: String(employeeIds.size),
      emailOpenRate: formatPercent(percent(openedTargets, sentTargets)),
      linkClickRate: formatPercent(percent(clickedTargets, sentTargets)),
      dataSubmittedRate: formatPercent(percent(submittedTargets, sentTargets)),
      totalEvents: String(totalEvents),
    },
    campaignRows,
    departmentRows,
    employeeRows,
    topEmployees: employeeRows.slice(0, 8),
    eventDistribution,
    funnelStages,
    trendBuckets,
    trendMarkAreas,
  }
}

export function serializeReportCsv(report) {
  const rows = []
  rows.push([globalT('pages.reports.csv.section'), globalT('pages.reports.csv.field'), globalT('pages.reports.csv.value')])
  rows.push([globalT('pages.reports.csv.filters'), globalT('common.labels.dateFrom'), report.filters.dateFrom || ''])
  rows.push([globalT('pages.reports.csv.filters'), globalT('common.labels.dateTo'), report.filters.dateTo || ''])
  rows.push([globalT('pages.reports.csv.filters'), globalT('common.labels.campaign'), report.filters.campaignId || ''])
  rows.push([globalT('pages.reports.csv.filters'), globalT('common.labels.department'), report.filters.departmentId || ''])
  rows.push(['KPI', globalT('common.labels.riskIndex'), report.kpis.companyRisk])
  rows.push(['KPI', globalT('pages.reports.kpis.campaignsAnalyzed.label'), report.kpis.campaignsAnalyzed])
  rows.push(['KPI', globalT('pages.reports.kpis.employeesTested.label'), report.kpis.employeesTested])
  rows.push(['KPI', globalT('pages.reports.kpis.emailOpenRate.label'), report.kpis.emailOpenRate])
  rows.push(['KPI', globalT('pages.reports.kpis.linkClickRate.label'), report.kpis.linkClickRate])
  rows.push(['KPI', globalT('pages.reports.kpis.dataSubmittedRate.label'), report.kpis.dataSubmittedRate])
  rows.push(['KPI', globalT('pages.reports.kpis.totalEvents.label'), report.kpis.totalEvents])

  rows.push([])
  rows.push([globalT('common.labels.campaign'), globalT('common.labels.targets'), globalT('common.labels.riskIndex'), ...EVENT_TYPES.map((eventType) => translateEventType(eventType))])
  report.campaignRows.forEach((campaign) => {
    rows.push([
      campaign.label,
      campaign.targetCount,
      campaign.risk ?? '',
      ...EVENT_TYPES.map((eventType) => campaign.eventCounts[eventType] || 0),
    ])
  })

  rows.push([])
  rows.push([globalT('common.labels.department'), globalT('common.labels.employees'), globalT('common.labels.riskIndex')])
  report.departmentRows.forEach((department) => {
    rows.push([department.label, department.employeeCount, department.risk])
  })

  rows.push([])
  rows.push([globalT('common.labels.employee'), globalT('common.labels.department'), globalT('common.labels.riskIndex'), translateEventType('LINK_OPENED'), translateEventType('DATA_SENT'), globalT('common.labels.activity')])
  report.employeeRows.forEach((employee) => {
    rows.push([
      employee.label,
      employee.departmentLabel,
      employee.risk,
      employee.linkOpenedCount,
      employee.dataSubmittedCount,
      employee.lastActivityLabel,
    ])
  })

  return rows.map((row) => row.map(csvCell).join(',')).join('\n')
}
