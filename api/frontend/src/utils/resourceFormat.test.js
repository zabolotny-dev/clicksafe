import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../i18n', () => ({
  globalT: (key) => `[${key}]`,
  resolveLocaleTag: () => 'en-US',
  translateKnownError: (error, fallbackKey) => fallbackKey || 'errors.requestFailed',
  hasTranslation: () => true,
}))

import {
  itemId,
  nullableId,
  isAuthError,
  normalizeAttributes,
  attributesToRows,
  formatNullable,
  employeeName,
  formatAttributes,
  formatKilobytes,
  importErrorInfo,
} from './resourceFormat'

describe('itemId', () => {
  it('returns empty string for falsy input', () => {
    expect(itemId(null)).toBe('')
    expect(itemId(undefined)).toBe('')
    expect(itemId('')).toBe('')
  })

  it('returns string as-is', () => {
    expect(itemId('abc-123')).toBe('abc-123')
  })

  it('extracts UUID from object with UUID field', () => {
    expect(itemId({ UUID: 'aaa-bbb' })).toBe('aaa-bbb')
    expect(itemId({ uuid: 'ccc-ddd' })).toBe('ccc-ddd')
  })

  it('extracts id from object with id field', () => {
    expect(itemId({ id: 42 })).toBe('42')
  })

  it('prefers UUID over id', () => {
    expect(itemId({ UUID: 'from-uuid', id: 'from-id' })).toBe('from-uuid')
  })
})

describe('nullableId', () => {
  it('returns empty string for falsy input', () => {
    expect(nullableId(null)).toBe('')
    expect(nullableId(undefined)).toBe('')
    expect(nullableId('')).toBe('')
  })

  it('returns string as-is', () => {
    expect(nullableId('some-id')).toBe('some-id')
  })

  it('returns empty string when Valid is false', () => {
    expect(nullableId({ Valid: false, UUID: 'abc' })).toBe('')
    expect(nullableId({ valid: false, uuid: 'abc' })).toBe('')
  })

  it('extracts UUID when Valid is true', () => {
    expect(nullableId({ Valid: true, UUID: 'my-id' })).toBe('my-id')
  })
})

describe('isAuthError', () => {
  it('returns true for 401', () => {
    expect(isAuthError({ status: 401 })).toBe(true)
  })

  it('returns true for 403', () => {
    expect(isAuthError({ status: 403 })).toBe(true)
  })

  it('returns false for other status codes', () => {
    expect(isAuthError({ status: 400 })).toBe(false)
    expect(isAuthError({ status: 500 })).toBe(false)
    expect(isAuthError({ status: 200 })).toBe(false)
  })

  it('returns false for null/undefined', () => {
    expect(isAuthError(null)).toBe(false)
    expect(isAuthError(undefined)).toBe(false)
  })
})

describe('normalizeAttributes', () => {
  it('converts rows to object, trimming keys and values', () => {
    const result = normalizeAttributes([
      { key: ' Name ', value: ' Alice ' },
      { key: 'Role', value: 'Admin' },
    ])
    expect(result).toEqual({ Name: 'Alice', Role: 'Admin' })
  })

  it('skips rows with empty key or value', () => {
    const result = normalizeAttributes([
      { key: '', value: 'orphan' },
      { key: 'Valid', value: '' },
      { key: 'OK', value: 'yes' },
    ])
    expect(result).toEqual({ OK: 'yes' })
  })

  it('returns empty object for empty input', () => {
    expect(normalizeAttributes([])).toEqual({})
  })
})

describe('attributesToRows', () => {
  it('converts object to key/value rows', () => {
    const result = attributesToRows({ Role: 'Admin', Dept: 'HR' })
    expect(result).toEqual([
      { key: 'Role', value: 'Admin' },
      { key: 'Dept', value: 'HR' },
    ])
  })

  it('returns one empty row for null/empty input', () => {
    expect(attributesToRows(null)).toEqual([{ key: '', value: '' }])
    expect(attributesToRows({})).toEqual([{ key: '', value: '' }])
  })
})

describe('formatNullable', () => {
  it('returns placeholder for null and empty string', () => {
    expect(formatNullable(null)).toBe('[common.placeholder]')
    expect(formatNullable(undefined)).toBe('[common.placeholder]')
    expect(formatNullable('')).toBe('[common.placeholder]')
  })

  it('returns string representation of value', () => {
    expect(formatNullable('hello')).toBe('hello')
    expect(formatNullable(42)).toBe('42')
  })
})

describe('employeeName', () => {
  it('joins first and last name', () => {
    expect(employeeName({ first_name: 'Ivan', last_name: 'Ivanov' })).toBe('Ivan Ivanov')
  })

  it('handles missing last name', () => {
    expect(employeeName({ first_name: 'Ivan' })).toBe('Ivan')
  })

  it('handles missing first name', () => {
    expect(employeeName({ last_name: 'Ivanov' })).toBe('Ivanov')
  })

  it('returns placeholder for null employee', () => {
    expect(employeeName(null)).toBe('[common.placeholder]')
    expect(employeeName({})).toBe('[common.placeholder]')
  })
})

describe('formatAttributes', () => {
  it('formats attributes as key: value pairs', () => {
    expect(formatAttributes({ Role: 'Admin', Dept: 'HR' })).toBe('Role: Admin, Dept: HR')
  })

  it('returns placeholder for empty attributes', () => {
    expect(formatAttributes({})).toBe('[common.placeholder]')
    expect(formatAttributes(null)).toBe('[common.placeholder]')
  })

  it('skips empty values', () => {
    expect(formatAttributes({ Role: 'Admin', Empty: '' })).toBe('Role: Admin')
  })
})

describe('formatKilobytes', () => {
  it('converts bytes to kilobytes', () => {
    const result = formatKilobytes(2048)
    expect(result).toBe('2 KB')
  })

  it('returns placeholder for null/NaN', () => {
    expect(formatKilobytes(null)).toBe('[common.placeholder]')
    expect(formatKilobytes(undefined)).toBe('[common.placeholder]')
  })
})

describe('importErrorInfo', () => {
  it('extracts row errors from error payload', () => {
    const error = {
      payload: {
        message: 'Import failed',
        details: [
          { field: 'csv_row', value: { row: 5, field: 'email', error: 'invalid format' } },
          { field: 'other_field', value: {} },
        ],
      },
    }

    const info = importErrorInfo(error, (rowError) => `Row ${rowError.row}: ${rowError.error}`)
    expect(info.message).toBe('Import failed')
    expect(info.items).toEqual(['Row 5: invalid format'])
  })

  it('returns fallback message when payload has none', () => {
    const info = importErrorInfo({}, () => '')
    expect(info.message).toBe('[errors.requestFailed]')
  })
})
