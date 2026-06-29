import { writable } from 'svelte/store'

// Current tree sort. Default: largest first.
export const sortState = writable({ by: 'size', asc: false })

// Sensible initial direction per column when first selected.
const defaultAsc = { name: true, files: false, size: false }

export function toggleSort(col) {
  sortState.update((s) =>
    s.by === col ? { by: col, asc: !s.asc } : { by: col, asc: defaultAsc[col] ?? false }
  )
}
