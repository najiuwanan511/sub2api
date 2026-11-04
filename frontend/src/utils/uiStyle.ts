/**
 * UI style switch for the DESIGN_1 (Money Tile) design system.
 *
 * This file did NOT exist in upstream Wei-Shaw/sub2api at the time the
 * theme branch was created; it is new in the theme/money-tile branch.
 *
 * Two styles are supported:
 * - "modern"  (default): the Metro tile design system in docs/design-system.md
 * - "classic": the upstream look (teal primary, rounded corners, shadows,
 *   gradients) — preserved untouched.
 *
 * The choice is stored per browser in localStorage (key below) and applied
 * as <html data-ui="modern|classic">. Switching styles reloads the whole
 * page, as the spec requires, so that every component re-renders with the
 * correct CSS variables.
 */
export type UiStyle = 'modern' | 'classic'

export const UI_STYLE_STORAGE_KEY = 'sub2api-ui-style'
export const DEFAULT_UI_STYLE: UiStyle = 'modern'

function isUiStyle(value: unknown): value is UiStyle {
  return value === 'modern' || value === 'classic'
}

/** Read the stored style. Falls back to "modern" when unset or invalid. */
export function getUiStyle(): UiStyle {
  try {
    const stored = window.localStorage.getItem(UI_STYLE_STORAGE_KEY)
    if (isUiStyle(stored)) return stored
  } catch {
    // localStorage unavailable (private mode, SSR): fall back to default.
  }
  return DEFAULT_UI_STYLE
}

export function isModernUi(): boolean {
  return getUiStyle() === 'modern'
}

/**
 * Apply a style to <html data-ui> and persist it. When the style actually
 * changes, the page is reloaded (spec behaviour) unless reload is false.
 */
export function setUiStyle(style: UiStyle, options: { reload?: boolean } = {}): void {
  const { reload = true } = options
  const previous = getUiStyle()
  try {
    window.localStorage.setItem(UI_STYLE_STORAGE_KEY, style)
  } catch {
    // Persisting is best-effort; the attribute still applies for this page.
  }
  document.documentElement.setAttribute('data-ui', style)
  if (reload && previous !== style) {
    window.location.reload()
  }
}

/** Toggle between modern and classic (reloads the page). */
export function toggleUiStyle(): UiStyle {
  const next: UiStyle = getUiStyle() === 'modern' ? 'classic' : 'modern'
  setUiStyle(next)
  return next
}

/**
 * Call this as early as possible (main.ts, before the app mounts) so the
 * first paint already uses the right style and there is no theme flash.
 */
export function initUiStyle(): UiStyle {
  const style = getUiStyle()
  document.documentElement.setAttribute('data-ui', style)
  return style
}
