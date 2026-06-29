<script>
  import { onMount } from 'svelte'
  import { GetChildren, OpenPath } from '../wails.js'
  import { openContextMenu } from './contextmenu.js'
  import { sortState, foldersOnly } from './sort.js'
  import { treeCommand, reserveRows, isDescendant } from './treecommand.js'
  import { t } from './i18n.js'
  import { formatBytes, formatCount } from '../util.js'

  export let node
  export let denom = 0 // reference size for the bar (the drive's total capacity)
  export let depth = 0
  export let autoExpand = false

  let expanded = false
  let children = null
  let total = 0
  let truncated = false
  let loading = false
  let applied = null
  let lastCmdNonce = 0

  $: percent = denom > 0 ? Math.min((node.size / denom) * 100, 100) : 100
  $: hue = Math.max(0, 120 - percent * 1.2) // green (small) -> red (large)
  $: viewKey = `${$sortState.by}|${$sortState.asc}|${$foldersOnly}`

  async function load() {
    loading = true
    const res = await GetChildren(node.id, $sortState.by, $sortState.asc, $foldersOnly)
    children = res.items
    total = res.total
    truncated = res.truncated
    applied = viewKey
    loading = false
  }

  // Re-fetch visible children whenever the sort order or folders-only changes.
  $: if (expanded && children !== null && !loading && viewKey !== applied) load()

  function commandApplies(cmd) {
    if (cmd.type === 'expandAll') {
      return node.path === cmd.path || isDescendant(node.path, cmd.path)
    }
    return isDescendant(node.path, cmd.path) // collapseAll: descendants only
  }

  async function runCommand(cmd) {
    if (cmd.type === 'collapseAll') {
      expanded = false
      return
    }
    // expandAll — load, then expand if the row budget allows. Cascades to
    // children as they mount and apply the same in-flight command.
    if (!node.isDir || !node.hasChildren) return
    if (children === null) await load()
    if (!children.length) return
    if (!reserveRows(children.length)) return // safety cap reached
    expanded = true
  }

  // Apply an expand/collapse-all command targeting this node's subtree (once).
  $: if ($treeCommand && $treeCommand.nonce !== lastCmdNonce && commandApplies($treeCommand)) {
    lastCmdNonce = $treeCommand.nonce
    runCommand($treeCommand)
  }

  async function toggle() {
    if (!node.isDir || !node.hasChildren) return
    if (children === null) await load()
    expanded = !expanded
  }

  function open(e) {
    e.stopPropagation()
    OpenPath(node.path)
  }

  function onKey(e) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      toggle()
    }
  }

  onMount(() => {
    if (autoExpand && node.isDir && node.hasChildren) toggle()
  })
</script>

<div
  class="row dss-cols"
  class:is-root={depth === 0}
  on:click={toggle}
  on:keydown={onKey}
  on:contextmenu={(e) => openContextMenu(e, node)}
  role="treeitem"
  aria-selected="false"
  tabindex="0"
>
  <span class="name-cell" style="padding-left: {depth * 16}px">
    <span class="caret">
      {#if node.isDir && node.hasChildren}{expanded ? '▾' : '▸'}{/if}
    </span>
    <span class="icon">{node.isDir ? '📁' : '📄'}</span>
    <span class="name" title={node.path}>{node.name}</span>
  </span>
  <span class="meta">{node.isDir ? formatCount(node.files) : ''}</span>
  <span class="bar"><span class="fill" style="width:{percent}%; background:hsl({hue},65%,50%)"></span></span>
  <span class="pct">{percent.toFixed(1)}%</span>
  <span class="size">{formatBytes(node.size)}</span>
  <button class="open" title={$t('openTip')} on:click={open}>📂</button>
</div>

{#if expanded && children}
  {#each children as child (child.id)}
    <svelte:self node={child} {denom} depth={depth + 1} />
  {/each}
  {#if truncated}
    <div class="more" style="padding-left: {(depth + 1) * 16 + 54}px">
      {$t('moreItems', { hidden: formatCount(total - children.length), shown: formatCount(children.length) })}
    </div>
  {/if}
{:else if loading}
  <div class="more" style="padding-left: {(depth + 1) * 16 + 54}px">{$t('loading')}</div>
{/if}

<style>
  .row {
    height: 26px;
    padding: 0 8px;
    font-size: 13px;
    cursor: default;
    border-bottom: 1px solid var(--line);
  }
  .row:hover {
    background: var(--hover);
  }
  /* The top row is always the scan root — mark it distinctly. The left accent
     bar uses inset box-shadow so it doesn't shift content out of column. */
  .row.is-root {
    background: var(--accent-soft);
    box-shadow: inset 3px 0 0 var(--accent);
    font-weight: 600;
  }
  .row.is-root:hover {
    background: var(--accent-soft);
  }
  .name-cell {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    overflow: hidden;
  }
  .caret {
    width: 12px;
    flex: none;
    color: var(--muted);
    text-align: center;
  }
  .icon {
    flex: none;
  }
  .name {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .meta {
    text-align: right;
    color: var(--muted);
    font-size: 11px;
    font-variant-numeric: tabular-nums;
  }
  .bar {
    width: 100%;
    height: 10px;
    background: var(--line);
    border-radius: 3px;
    overflow: hidden;
  }
  .fill {
    display: block;
    height: 100%;
  }
  .pct {
    text-align: right;
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }
  .size {
    text-align: right;
    font-variant-numeric: tabular-nums;
    font-weight: 600;
  }
  .open {
    width: 26px;
    border: none;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    opacity: 0;
    font-size: 13px;
  }
  .row:hover .open {
    opacity: 1;
  }
  .open:hover {
    color: var(--accent);
  }
  .more {
    height: 22px;
    display: flex;
    align-items: center;
    color: var(--muted);
    font-size: 12px;
    font-style: italic;
  }
</style>
