import { describe, expect, it, beforeEach } from 'vitest'
import {
  getUiStyle,
  setUiStyle,
  initUiStyle,
  UI_STYLE_STORAGE_KEY
} from '../uiStyle'

describe('uiStyle', () => {
  beforeEach(() => {
    window.localStorage.clear()
    document.documentElement.removeAttribute('data-ui')
  })

  it('defaults to modern when nothing is stored', () => {
    expect(getUiStyle()).toBe('modern')
  })

  it('persists and applies classic without reloading when asked', () => {
    setUiStyle('classic', { reload: false })
    expect(window.localStorage.getItem(UI_STYLE_STORAGE_KEY)).toBe('classic')
    expect(document.documentElement.getAttribute('data-ui')).toBe('classic')
    expect(getUiStyle()).toBe('classic')
  })

  it('falls back to modern for an invalid stored value', () => {
    window.localStorage.setItem(UI_STYLE_STORAGE_KEY, 'neon')
    expect(getUiStyle()).toBe('modern')
  })

  it('initUiStyle sets the data-ui attribute', () => {
    expect(initUiStyle()).toBe('modern')
    expect(document.documentElement.getAttribute('data-ui')).toBe('modern')
  })
})
