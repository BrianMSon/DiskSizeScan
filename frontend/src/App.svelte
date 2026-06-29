<script>
  import { onMount } from 'svelte'
  import { ListDriveInfo, Scan, Search, Cancel, OpenPath, onProgress } from './wails.js'
  import { contextMenu, closeContextMenu } from './lib/contextmenu.js'
  import { foldersOnly, sortState } from './lib/sort.js'
  import { expandAll, collapseAll } from './lib/treecommand.js'
  import { t } from './lib/i18n.js'
  import { formatBytes, formatCount } from './util.js'
  import TreeNode from './lib/TreeNode.svelte'
  import ColumnHeader from './lib/ColumnHeader.svelte'
  import DriveDashboard from './lib/DriveDashboard.svelte'
  import SearchResults from './lib/SearchResults.svelte'

  let driveInfos = []
  let path = ''
  let scanning = false
  let root = null
  let totals = null
  let drive = null // capacity of the scanned path's volume
  let error = ''
  let progress = { files: 0, bytes: 0, path: '' }

  let query = ''
  let searchResult = null
  let searchTimer

  $: drives = driveInfos.map((d) => d.path)
  // Bars are drawn relative to the whole drive's capacity (fallback: scan root).
  $: denom = drive && drive.total > 0 ? drive.total : root ? root.size : 0
  $: menu = $contextMenu

  // Debounced search whenever the query or sort changes (and a scan exists).
  $: searchKey = `${query.trim()}|${$sortState.by}|${$sortState.asc}`
  $: {
    searchKey
    clearTimeout(searchTimer)
    searchTimer = setTimeout(runSearch, 250)
  }

  async function runSearch() {
    const q = query.trim()
    if (!root || !q) {
      searchResult = null
      return
    }
    try {
      searchResult = await Search(q, $sortState.by, $sortState.asc)
    } catch (e) {
      searchResult = null
    }
  }

  onMount(async () => {
    try {
      driveInfos = await ListDriveInfo()
      if (driveInfos.length) path = driveInfos[0].path
    } catch (e) {
      // Running outside the Wails runtime (e.g. plain `vite`): no bindings.
    }
    onProgress((p) => (progress = p))
  })

  async function startScan() {
    if (!path || scanning) return
    error = ''
    root = null
    totals = null
    drive = null
    query = ''
    searchResult = null
    progress = { files: 0, bytes: 0, path: '' }
    scanning = true
    try {
      const res = await Scan(path)
      root = res.root
      totals = { size: res.totalSize, files: res.totalFiles, ms: res.durationMs }
      drive = { total: res.driveTotal, used: res.driveUsed, free: res.driveFree }
    } catch (e) {
      error = String(e?.message || e)
    } finally {
      scanning = false
    }
  }

  function scanPath(p) {
    path = p
    startScan()
  }

  async function goHome() {
    root = null
    totals = null
    drive = null
    error = ''
    query = ''
    searchResult = null
    try {
      driveInfos = await ListDriveInfo() // refresh capacities
    } catch (e) {
      // outside Wails runtime
    }
  }

  function onKey(e) {
    if (e.key === 'Enter') startScan()
  }

  function menuScan() {
    const n = menu.node
    closeContextMenu()
    if (n && n.isDir) scanPath(n.path)
  }

  function menuExpandAll() {
    const n = menu.node
    closeContextMenu()
    if (n && n.isDir) expandAll(n)
  }

  function menuCollapseAll() {
    const n = menu.node
    closeContextMenu()
    if (n && n.isDir) collapseAll(n)
  }

  function menuOpen() {
    const n = menu.node
    closeContextMenu()
    if (n) OpenPath(n.path)
  }
</script>

<svelte:window on:keydown={(e) => e.key === 'Escape' && closeContextMenu()} />

<main>
  <header>
    <div class="controls">
      <button class="home" on:click={goHome} disabled={scanning} title={$t('homeTip')}>🏠</button>
      <input
        type="text"
        bind:value={path}
        on:keydown={onKey}
        placeholder={$t('pathPlaceholder')}
        disabled={scanning}
      />
      {#if scanning}
        <button class="cancel" on:click={Cancel}>{$t('cancel')}</button>
      {:else}
        <button class="scan" on:click={startScan}>{$t('scan')}</button>
      {/if}
      <label class="opt" title={$t('foldersOnlyTip')}>
        <input type="checkbox" bind:checked={$foldersOnly} />
        {$t('foldersOnly')}
      </label>
    </div>
    {#if drives.length}
      <div class="drives">
        {#each drives as d}
          <button class="drive" class:active={path === d} on:click={() => (path = d)} disabled={scanning}>
            {d}
          </button>
        {/each}
      </div>
    {/if}
  </header>

  <section class="status">
    {#if scanning}
      <!-- live progress is shown in the centered body panel below -->
    {:else if error}
      <span class="err">⚠ {error}</span>
    {:else if totals}
      <span>
        <strong>{formatBytes(totals.size)}</strong>
        {$t('summaryRest', { files: formatCount(totals.files), sec: (totals.ms / 1000).toFixed(2) })}
      </span>
      {#if drive && drive.total > 0}
        <span class="chip">
          {$t('driveChip', {
            pct: ((totals.size / drive.total) * 100).toFixed(1),
            used: formatBytes(drive.used),
            total: formatBytes(drive.total),
          })}
        </span>
      {/if}
    {:else}
      <span class="hint">{$t('startHint')}</span>
    {/if}
  </section>

  {#if root && !scanning}
    <section class="searchbar">
      <span class="ico">🔍</span>
      <input bind:value={query} placeholder={$t('searchPlaceholder')} />
      {#if query}
        <button class="clear" on:click={() => (query = '')} title={$t('clear')}>✕</button>
      {/if}
      {#if searchResult}
        <span class="scount">{$t('searchCount', { n: formatCount(searchResult.total) })}</span>
      {/if}
    </section>
  {/if}

  <section class="body">
    {#if scanning}
      <div class="scanning">
        <div class="spinner-lg" />
        <div class="scan-title">{$t('scanning')}</div>
        <div class="scan-stats">
          <div class="stat">
            <div class="val">{formatCount(progress.files)}</div>
            <div class="lbl">{$t('items')}</div>
          </div>
          <div class="stat">
            <div class="val">{formatBytes(progress.bytes)}</div>
            <div class="lbl">{$t('size')}</div>
          </div>
        </div>
        <div class="scan-path" title={progress.path}>{progress.path || $t('preparing')}</div>
      </div>
    {:else if root}
      <ColumnHeader />
      {#if searchResult}
        <SearchResults
          items={searchResult.items}
          total={searchResult.total}
          truncated={searchResult.truncated}
          {denom}
        />
      {:else}
        {#key root.id}
          <TreeNode node={root} {denom} depth={0} autoExpand={true} />
        {/key}
      {/if}
    {:else}
      <DriveDashboard drives={driveInfos} on:scan={(e) => scanPath(e.detail)} />
    {/if}
  </section>
</main>

{#if menu.visible}
  <div class="menu-overlay" on:click={closeContextMenu} on:keydown={closeContextMenu} on:contextmenu|preventDefault={closeContextMenu} role="presentation">
    <div class="context-menu" style="left:{menu.x}px; top:{menu.y}px" role="menu">
      {#if menu.node?.isDir}
        <button on:click={menuScan}>{$t('ctxScan')}</button>
        <button on:click={menuExpandAll}>{$t('ctxExpandAll')}</button>
        <button on:click={menuCollapseAll}>{$t('ctxCollapseAll')}</button>
      {/if}
      <button on:click={menuOpen}>{$t('ctxOpen')}</button>
    </div>
  </div>
{/if}

<style>
  :global(:root) {
    --bg: #ffffff;
    --fg: #1c1f24;
    --muted: #8a909a;
    --line: #ececef;
    --hover: #f5f7fa;
    --accent: #2f6fed;
    --accent-soft: rgba(47, 111, 237, 0.09);
    --surface: #ffffff;
    --border-strong: #d6d9df;
  }
  :global(html.dark) {
    --bg: #1b1f27;
    --fg: #e7e9ee;
    --muted: #9aa1ac;
    --line: #2b313b;
    --hover: #232a34;
    --accent: #4d8bff;
    --accent-soft: rgba(77, 139, 255, 0.16);
    --surface: #232a34;
    --border-strong: #3a4250;
  }
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    color: var(--fg);
    background: var(--bg);
  }
  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }
  header {
    padding: 12px 12px 8px;
    border-bottom: 1px solid var(--line);
    flex: none;
  }
  .controls {
    display: flex;
    gap: 8px;
  }
  input {
    flex: 1;
    padding: 8px 10px;
    font-size: 14px;
    border: 1px solid var(--border-strong);
    border-radius: 6px;
    outline: none;
    background: var(--surface);
    color: var(--fg);
  }
  input:focus {
    border-color: var(--accent);
  }
  button {
    border: none;
    border-radius: 6px;
    padding: 8px 16px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
  }
  .home {
    background: var(--hover);
    color: var(--fg);
    padding: 8px 12px;
    font-size: 16px;
    line-height: 1;
  }
  .home:hover:not(:disabled) {
    background: var(--line);
  }
  .home:disabled {
    opacity: 0.5;
    cursor: default;
  }
  .scan {
    background: var(--accent);
    color: #fff;
  }
  .cancel {
    background: #e5484d;
    color: #fff;
  }
  .opt {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 13px;
    color: var(--muted);
    white-space: nowrap;
    cursor: pointer;
  }
  .opt input {
    flex: none;
    width: auto;
    padding: 0;
    margin: 0;
    border: none;
    background: none;
    accent-color: var(--accent);
    cursor: pointer;
  }
  .drives {
    display: flex;
    gap: 6px;
    margin-top: 8px;
    flex-wrap: wrap;
  }
  .drive {
    padding: 4px 10px;
    font-size: 12px;
    font-weight: 500;
    background: var(--hover);
    color: var(--fg);
  }
  .drive.active {
    background: var(--accent);
    color: #fff;
  }
  .status {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px;
    font-size: 13px;
    color: var(--fg);
    border-bottom: 1px solid var(--line);
    flex: none;
    min-height: 20px;
  }
  .chip {
    background: var(--hover);
    border-radius: 12px;
    padding: 2px 10px;
    color: var(--muted);
    font-size: 12px;
    white-space: nowrap;
  }
  .hint,
  .err {
    color: var(--muted);
  }
  .err {
    color: #e5484d;
  }
  .body {
    flex: 1;
    overflow: auto;
  }
  .searchbar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    border-bottom: 1px solid var(--line);
    flex: none;
  }
  .searchbar .ico {
    opacity: 0.7;
  }
  .searchbar input {
    padding: 6px 10px;
    font-size: 13px;
  }
  .searchbar .clear {
    background: var(--hover);
    color: var(--muted);
    padding: 4px 9px;
    font-size: 12px;
  }
  .searchbar .clear:hover {
    color: var(--fg);
  }
  .searchbar .scount {
    font-size: 12px;
    color: var(--muted);
    white-space: nowrap;
  }

  /* Scanning panel — fixed, centered layout so nothing reflows as numbers
     and the current path change. */
  .scanning {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 18px;
    padding: 24px;
  }
  .spinner-lg {
    width: 40px;
    height: 40px;
    border: 4px solid var(--line);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  .scan-title {
    font-size: 16px;
    font-weight: 600;
  }
  .scan-stats {
    display: flex;
    gap: 48px;
  }
  .stat {
    min-width: 130px;
    text-align: center;
  }
  .stat .val {
    font-size: 26px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
  }
  .stat .lbl {
    margin-top: 2px;
    font-size: 12px;
    color: var(--muted);
  }
  .scan-path {
    width: 480px;
    max-width: 80vw;
    height: 16px;
    text-align: center;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--muted);
    font-size: 12px;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  .menu-overlay {
    position: fixed;
    inset: 0;
    z-index: 50;
  }
  .context-menu {
    position: absolute;
    min-width: 170px;
    background: var(--surface);
    border: 1px solid var(--line);
    border-radius: 8px;
    box-shadow: 0 6px 24px rgba(0, 0, 0, 0.18);
    padding: 4px;
    display: flex;
    flex-direction: column;
  }
  .context-menu button {
    text-align: left;
    background: transparent;
    color: var(--fg);
    font-weight: 500;
    padding: 8px 10px;
    border-radius: 5px;
  }
  .context-menu button:hover {
    background: var(--hover);
  }
</style>
