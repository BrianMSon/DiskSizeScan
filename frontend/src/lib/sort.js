import { writable } from 'svelte/store'

// Current tree sort. Default: largest first.
export const sortState = writable({ by: 'size', asc: false })

// Show only directories (hide files) in the tree.
export const foldersOnly = writable(false)

// Sensible initial direction per column when first selected.
const defaultAsc = { name: true, files: false, size: false, usage: false }

export function toggleSort(col) {
  sortState.update((s) =>
    s.by === col ? { by: col, asc: !s.asc } : { by: col, asc: defaultAsc[col] ?? false }
  )
}
