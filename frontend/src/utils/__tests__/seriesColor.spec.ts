import { describe, expect, it } from 'vitest'
import { seriesIndexByRank, seriesIndexById, seriesClass } from '../seriesColor'

describe('seriesColor', () => {
  it('cycles series 1..5 by rank', () => {
    expect([0, 1, 2, 3, 4, 5].map(seriesIndexByRank)).toEqual([1, 2, 3, 4, 5, 1])
  })

  it('falls back to abs(id) % 5 + 1', () => {
    expect(seriesIndexById(-7)).toBe(3)
    expect(seriesIndexById(10)).toBe(1)
  })

  it('builds a Tailwind series class', () => {
    expect(seriesClass(3)).toBe('bg-series-3')
  })
})
