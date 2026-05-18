export function unwrapItems(response) {
  return Array.isArray(response?.items) ? response.items : []
}

export async function loadAllPages(fetcher, filters = {}, options = {}) {
  const items = []
  let page = 1
  let total = 0

  do {
    const response = await fetcher({
      ...filters,
      page,
      rows: 100,
    }, options)
    const pageItems = unwrapItems(response)

    total = Number.isFinite(response?.total) ? response.total : items.length + pageItems.length
    items.push(...pageItems)

    if (!pageItems.length) {
      break
    }

    page += 1
  } while (items.length < total)

  return items
}
