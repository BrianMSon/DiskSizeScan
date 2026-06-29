<script>
  import { onMount } from 'svelte'
  import { GetChildren, OpenPath } from '../wails.js'
  import { openContextMenu } from './contextmenu.js'
  import { sortState } from './sort.js'
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
  let appliedSort = null

  $: percent = denom > 0 ? Math.min((node.size / denom) * 100, 100) : 100
  $: hue = Math.max(0, 120 - percent * 1.2) // green (small) -> red (large)
  $: sortKey = `${$sortState.by}|${$sortState.asc}`

  async function load() {
    loading = true
    const res = await GetChildren(node.id, $sortState.by, $sortState.asc)
    children = res.items
    total = res.total
    truncated = res.truncated
    appliedSort = sortKey
    loading = false
  }

  // Re-fetch (re-sort) visible children whenever the sort order changes.
  $: if (expanded && children !== null && !loading && sortKey !== appliedSort) load()

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
  <button class="open" title="탐색기에서 열기" on:click={open}>↗</button>
</div>

{#if expanded && children}
  {#each children as child (child.id)}
    <svelte:self node={child} {denom} depth={depth + 1} />
  {/each}
  {#if truncated}
    <div class="more" style="padding-left: {(depth + 1) * 16 + 54}px">
      … {formatCount(total - children.length)}개 더 있음 (상위 {formatCount(children.length)}개만 표시)
    </div>
  {/if}
{:else if loading}
  <div class="more" style="padding-left: {(depth + 1) * 16 + 54}px">불러오는 중…</div>
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
