import { ref, watchEffect } from 'vue'

const STORAGE_KEY = 'sing-panel-theme'

const theme = ref(localStorage.getItem(STORAGE_KEY) || 'system')
const isDark = ref(false)

let mediaQuery = null
let mediaHandler = null

const applyTheme = () => {
  let dark = false
  if (theme.value === 'dark') {
    dark = true
  } else if (theme.value === 'light') {
    dark = false
  } else {
    dark = window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
}

const startSystemListener = () => {
  stopSystemListener()
  mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  mediaHandler = () => {
    if (theme.value === 'system') applyTheme()
  }
  mediaQuery.addEventListener('change', mediaHandler)
}

const stopSystemListener = () => {
  if (mediaQuery && mediaHandler) {
    mediaQuery.removeEventListener('change', mediaHandler)
    mediaQuery = null
    mediaHandler = null
  }
}

export function useTheme() {
  const setTheme = (val) => {
    theme.value = val
    localStorage.setItem(STORAGE_KEY, val)
    if (val === 'system') {
      startSystemListener()
    } else {
      stopSystemListener()
    }
    applyTheme()
  }

  // Initialize on first use
  if (typeof window !== 'undefined') {
    applyTheme()
    if (theme.value === 'system') {
      startSystemListener()
    }
  }

  return { theme, isDark, setTheme }
}
