import { writable } from 'svelte/store'

function initialTheme() {
  try {
    const v = localStorage.getItem('theme')
    if (v === 'light' || v === 'dark') return v
  } catch (e) {
    // localStorage unavailable
  }
  return 'light'
}

export const theme = writable(initialTheme())

// Persist and apply the theme by toggling a class on <html>. The subscriber
// runs immediately on first import, so importing this module activates it.
theme.subscribe((v) => {
  try {
    localStorage.setItem('theme', v)
  } catch (e) {
    // ignore
  }
  if (typeof document !== 'undefined') {
    document.documentElement.classList.toggle('dark', v === 'dark')
  }
})

export function toggleTheme() {
  theme.update((v) => (v === 'dark' ? 'light' : 'dark'))
}
