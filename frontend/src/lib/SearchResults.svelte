<script>
  import { OpenPath } from '../wails.js'
  import { openContextMenu } from './contextmenu.js'
  import { t } from './i18n.js'
  import { formatBytes, formatCount, formatDate } from '../util.js'

  export let items = []
  export let total = 0
  export let truncated = false
  export let denom = 0

  const usedPct = (size) => (denom > 0 ? Math.min((size / denom) * 100, 100) : 0)
  const hue = (p) => Math.max(0, 120 - p * 1.2)

  function open(e, node) {
    e.stopPropagation()
    OpenPath(node.path)
  }
</script>

{#if items.length === 0}
  <div class="empty">{$t('noResults')}</div>
{:else}
  <div class="list" role="list">
    {#each items as node (node.id)}
    <div class="row dss-cols" role="listitem" on:contextmenu={(e) => openContextMenu(e, node)}>
      <span class="name-cell">
        <span class="icon">{node.isDir ? '📁' : '📄'}</span>
        <span class="name" title={node.path}>{node.name}</span>
        <span class="path" title={node.path}>{node.path}</span>
      </span>
      <span class="meta">{node.isDir ? formatCount(node.files) : ''}</span>
      <span class="bar"><span class="fill" style="width:{usedPct(node.size)}%; background:hsl({hue(usedPct(node.size))},65%,50%)"></span></span>
      <span class="pct">{usedPct(node.size).toFixed(1)}%</span>
      <span class="date">{formatDate(node.modTime)}</span>
      <span class="size">{formatBytes(node.size)}</span>
      <button class="open" title={$t('openTip')} on:click={(e) => open(e, node)}>📂</button>
    </div>
    {/each}
  </div>
  {#if truncated}
    <div class="more">{$t('searchCapped', { shown: formatCount(items.length) })} · {$t('searchCount', { n: formatCount(total) })}</div>
  {/if}
{/if}

<style>
  .row {
    height: 26px;
    padding: 0 8px;
    font-size: 13px;
    border-bottom: 1px solid var(--line);
  }
  .row:hover {
    background: var(--hover);
  }
  .name-cell {
    display: flex;
    align-items: baseline;
    gap: 8px;
    min-width: 0;
    overflow: hidden;
  }
  .icon {
    flex: none;
    align-self: center;
  }
  .name {
    flex: none;
    max-width: 45%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .path {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--muted);
    font-size: 11px;
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
  .date {
    text-align: right;
    color: var(--muted);
    font-size: 11px;
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
  .more,
  .empty {
    display: flex;
    align-items: center;
    height: 30px;
    padding: 0 12px;
    color: var(--muted);
    font-size: 12px;
    font-style: italic;
  }
</style>
