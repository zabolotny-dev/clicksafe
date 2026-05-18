import { resolveLocaleTag, translateKnownError, globalT } from '../i18n'

export function itemId(value) {
  if (!value) {
    return ''
  }

  if (typeof value === 'string') {
    return value
  }

  if (value.UUID || value.uuid) {
    return String(value.UUID || value.uuid)
  }

  return value.id ? String(value.id) : ''
}

export function nullableId(value) {
  if (!value) {
    return ''
  }

  if (typeof value === 'string') {
    return value
  }

  if (value.Valid === false || value.valid === false) {
    return ''
  }

  return String(value.UUID || value.uuid || value.id || '')
}

export function formatTechnicalId(value) {
  const text = nullableId(value) || itemId(value)

  if (!text) {
    return globalT('common.placeholder')
  }

  return text.length > 16 ? `${text.slice(0, 8)}...${text.slice(-4)}` : text
}

export function formatNullable(value) {
  if (value == null || value === '') {
    return globalT('common.placeholder')
  }

  return String(value)
}

export function formatDateTime(value, locale) {
  if (!value) {
    return globalT('common.placeholder')
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return globalT('common.placeholder')
  }

  return new Intl.DateTimeFormat(resolveLocaleTag(locale), {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

export function formatDate(value, locale) {
  if (!value) {
    return globalT('common.placeholder')
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return globalT('common.placeholder')
  }

  return new Intl.DateTimeFormat(resolveLocaleTag(locale), {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
  }).format(date)
}

export function formatMonthDay(value, locale) {
  if (!value) {
    return globalT('common.placeholder')
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return globalT('common.placeholder')
  }

  return new Intl.DateTimeFormat(resolveLocaleTag(locale), {
    month: 'short',
    day: '2-digit',
  }).format(date)
}

export function formatNumber(value, locale, options = {}) {
  if (value == null || Number.isNaN(Number(value))) {
    return globalT('common.placeholder')
  }

  return new Intl.NumberFormat(resolveLocaleTag(locale), options).format(Number(value))
}

export function formatPercent(value, locale, options = {}) {
  if (value == null || Number.isNaN(Number(value))) {
    return globalT('common.placeholder')
  }

  const minimumFractionDigits = options.minimumFractionDigits ?? 0
  const maximumFractionDigits = options.maximumFractionDigits ?? minimumFractionDigits

  return new Intl.NumberFormat(resolveLocaleTag(locale), {
    style: 'percent',
    minimumFractionDigits,
    maximumFractionDigits,
  }).format(Number(value) / 100)
}

export function formatKilobytes(value, locale) {
  if (value == null || Number.isNaN(Number(value))) {
    return globalT('common.placeholder')
  }

  return `${formatNumber(Math.round(Number(value) / 1024), locale)} KB`
}

export function formatDateRange(from, to, locale) {
  const fromLabel = formatDate(from, locale)
  const toLabel = formatDate(to, locale)

  if (
    fromLabel === globalT('common.placeholder')
    && toLabel === globalT('common.placeholder')
  ) {
    return globalT('common.placeholder')
  }

  return `${fromLabel} — ${toLabel}`
}

export function employeeName(employee) {
  return [employee?.first_name, employee?.last_name].filter(Boolean).join(' ').trim() || globalT('common.placeholder')
}

export function formatAttributes(attributes) {
  const entries = Object.entries(attributes || {}).filter(([key, value]) => key && value !== '')

  if (!entries.length) {
    return globalT('common.placeholder')
  }

  return entries.map(([key, value]) => `${key}: ${value}`).join(', ')
}

export function normalizeAttributes(rows) {
  return rows.reduce((acc, row) => {
    const key = row.key.trim()
    const value = row.value.trim()

    if (key && value) {
      acc[key] = value
    }

    return acc
  }, {})
}

export function attributesToRows(attributes) {
  const entries = Object.entries(attributes || {})

  if (!entries.length) {
    return [{ key: '', value: '' }]
  }

  return entries.map(([key, value]) => ({ key, value }))
}

export function isAuthError(error) {
  return error?.status === 401 || error?.status === 403
}

export function importErrorInfo(error, formatRowError) {
  const rowErrors = importRowErrors(error)
  const message = error?.payload?.message || error?.message || globalT('errors.requestFailed')

  return {
    message,
    items: rowErrors.map((rowError) => formatRowError(rowError)),
  }
}

function importRowErrors(error) {
  const details = Array.isArray(error?.payload?.details) ? error.payload.details : []

  return details
    .filter((detail) => detail?.field === 'csv_row')
    .map((detail) => {
      const value = detail?.value && typeof detail.value === 'object' ? detail.value : {}

      return {
        row: value.row ?? value.Row ?? '',
        field: value.field ?? value.Field ?? '',
        error: value.error ?? value.Err ?? detail.error ?? '',
      }
    })
    .filter((rowError) => rowError.row || rowError.field || rowError.error)
}

export function errorMessage(error, fallbackKey = 'errors.requestFailed') {
  return translateKnownError(error, fallbackKey)
}
