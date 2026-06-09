import { describe, it, expect, vi } from 'vitest'
import { unwrapItems, loadAllPages } from './resourcePagination'

describe('unwrapItems', () => {
  it('returns items array from response', () => {
    expect(unwrapItems({ items: ['a', 'b'], total: 2 })).toEqual(['a', 'b'])
  })

  it('returns empty array when items is missing', () => {
    expect(unwrapItems({ total: 5 })).toEqual([])
    expect(unwrapItems({})).toEqual([])
  })

  it('returns empty array for null/undefined response', () => {
    expect(unwrapItems(null)).toEqual([])
    expect(unwrapItems(undefined)).toEqual([])
  })

  it('returns empty array when items is not an array', () => {
    expect(unwrapItems({ items: 'nope' })).toEqual([])
    expect(unwrapItems({ items: null })).toEqual([])
  })
})

describe('loadAllPages', () => {
  it('fetches a single page when all items fit', async () => {
    const fetcher = vi.fn().mockResolvedValueOnce({ items: ['a', 'b'], total: 2 })

    const result = await loadAllPages(fetcher)
    expect(result).toEqual(['a', 'b'])
    expect(fetcher).toHaveBeenCalledOnce()
    expect(fetcher).toHaveBeenCalledWith({ page: 1, rows: 100 }, {})
  })

  it('fetches multiple pages until total reached', async () => {
    const fetcher = vi.fn()
      .mockResolvedValueOnce({ items: ['a', 'b'], total: 3 })
      .mockResolvedValueOnce({ items: ['c'], total: 3 })

    const result = await loadAllPages(fetcher)
    expect(result).toEqual(['a', 'b', 'c'])
    expect(fetcher).toHaveBeenCalledTimes(2)
    expect(fetcher).toHaveBeenNthCalledWith(1, { page: 1, rows: 100 }, {})
    expect(fetcher).toHaveBeenNthCalledWith(2, { page: 2, rows: 100 }, {})
  })

  it('stops early when page returns no items', async () => {
    const fetcher = vi.fn()
      .mockResolvedValueOnce({ items: ['a'], total: 100 })
      .mockResolvedValueOnce({ items: [], total: 100 })

    const result = await loadAllPages(fetcher)
    expect(result).toEqual(['a'])
    expect(fetcher).toHaveBeenCalledTimes(2)
  })

  it('passes filters and options to the fetcher', async () => {
    const fetcher = vi.fn().mockResolvedValueOnce({ items: ['x'], total: 1 })
    const filters = { campaignId: 'abc-123' }
    const options = { signal: 'test-signal' }

    await loadAllPages(fetcher, filters, options)
    expect(fetcher).toHaveBeenCalledWith({ campaignId: 'abc-123', page: 1, rows: 100 }, options)
  })

  it('returns empty array for empty first page', async () => {
    const fetcher = vi.fn().mockResolvedValueOnce({ items: [], total: 0 })

    const result = await loadAllPages(fetcher)
    expect(result).toEqual([])
    expect(fetcher).toHaveBeenCalledOnce()
  })

  it('handles missing total by counting items', async () => {
    const fetcher = vi.fn()
      .mockResolvedValueOnce({ items: ['a', 'b'] })
      .mockResolvedValueOnce({ items: [] })

    const result = await loadAllPages(fetcher)
    expect(result).toEqual(['a', 'b'])
  })
})
