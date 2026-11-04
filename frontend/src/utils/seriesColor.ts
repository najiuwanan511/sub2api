/**
 * DESIGN_1 group series colours.
 *
 * Groups are coloured by spend rank (highest first), cycling through
 * series-1..5 by position modulo 5. Groups not in a ranked list fall back
 * to abs(id) % 5. The same mapping must be used for the group table and
 * for usage charts on the same page, so a group keeps one colour.
 */
export function seriesIndexByRank(rank: number): number {
  return ((rank % 5) + 5) % 5 + 1
}

export function seriesIndexById(id: number): number {
  return (Math.abs(id) % 5) + 1
}

/** Tailwind class for a series colour, e.g. "bg-series-3". */
export function seriesClass(index: number): string {
  const safe = ((index - 1) % 5) + 1
  return `bg-series-${safe}`
}

/** CSS variable name for a series colour, e.g. "--c-series-3". */
export function seriesVar(index: number): string {
  const safe = ((index - 1) % 5) + 1
  return `--c-series-${safe}`
}
