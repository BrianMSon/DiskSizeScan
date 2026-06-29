import { writable } from 'svelte/store'

// Broadcast expand/collapse-all commands to the recursive tree. Each TreeNode
// applies a command once (tracked by nonce) when it targets the node's subtree.
// Freshly mounted children pick up an in-flight command via the same reactive
// check, so an expand-all cascades down as children load.
export const treeCommand = writable(null) // { type, path, nonce }

let nonce = 0

// Safety cap: bound how many rows an expand-all may add so the DOM can't blow
// up on a huge drive. Expand-all stops descending once this is exhausted.
const EXPAND_BUDGET = 15000
let budget = 0

export function expandAll(node) {
  budget = EXPAND_BUDGET
  treeCommand.set({ type: 'expandAll', path: node.path, nonce: ++nonce })
}

export function collapseAll(node) {
  treeCommand.set({ type: 'collapseAll', path: node.path, nonce: ++nonce })
}

// Reserve room for n more rows; false when the budget is exhausted.
export function reserveRows(n) {
  if (budget < n) return false
  budget -= n
  return true
}

// True when `path` sits strictly under `ancestor` in the filesystem tree.
export function isDescendant(path, ancestor) {
  if (path === ancestor || !path.startsWith(ancestor)) return false
  const rest = path.slice(ancestor.length)
  return rest[0] === '\\' || rest[0] === '/' || ancestor.endsWith('\\') || ancestor.endsWith('/')
}
