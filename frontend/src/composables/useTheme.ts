import { computed, ref } from 'vue'

export const themeOptions = [
  { name: '深空', value: 'dark', swatch: 'oklch(85% .11 205)' },
  { name: '晨光', value: 'morning', swatch: 'oklch(68% .12 78)' },
  { name: '森屿', value: 'forest', swatch: 'oklch(82% .13 132)' },
  { name: '梅雾', value: 'plum', swatch: 'oklch(68% .14 325)' },
]

const themeStorageKey = 'navigation.theme.override'
const activeTheme = ref(themeOptions[0].value)

function hasTheme(theme: string | null | undefined) {
  return themeOptions.some(option => option.value === theme)
}

export function normalizedTheme(theme: string | null | undefined) {
  return hasTheme(theme) ? String(theme) : themeOptions[0].value
}

function storedTheme() {
  try {
    const value = localStorage.getItem(themeStorageKey)
    return hasTheme(value) ? value : null
  } catch {
    return null
  }
}

export function useTheme() {
  const currentTheme = computed(() => themeOptions.find(theme => theme.value === activeTheme.value) || themeOptions[0])

  function applyTheme(theme: string | null | undefined, persist = false) {
    activeTheme.value = normalizedTheme(theme)
    document.documentElement.dataset.theme = activeTheme.value
    if (persist) {
      try {
        localStorage.setItem(themeStorageKey, activeTheme.value)
      } catch {
        // localStorage 不可用时只应用到当前会话。
      }
    }
  }

  function applyDefaultTheme(defaultTheme: string | null | undefined) {
    applyTheme(storedTheme() || defaultTheme)
  }

  return {
    activeTheme,
    currentTheme,
    themeOptions,
    applyTheme,
    applyDefaultTheme,
  }
}
