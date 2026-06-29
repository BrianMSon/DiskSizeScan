import { writable } from 'svelte/store'

// Shared right-click menu state. TreeNode (any depth) opens it; App renders it.
export const contextMenu = writable({ visible: false, x: 0, y: 0, node: null })

export function openContextMenu(e, node) {
  e.preventDefault()
  contextMenu.set({ visible: true, x: e.clientX, y: e.clientY, node })
}

export function closeContextMenu() {
  contextMenu.update((m) => ({ ...m, visible: false }))
}
