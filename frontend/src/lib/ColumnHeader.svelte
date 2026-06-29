<script>
  import { sortState, toggleSort } from './sort.js'
  import { t } from './i18n.js'

  $: s = $sortState

  function keySort(e, col) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      toggleSort(col)
    }
  }
</script>

<div class="thead dss-cols">
  <span
    class="h"
    class:active={s.by === 'name'}
    on:click={() => toggleSort('name')}
    on:keydown={(e) => keySort(e, 'name')}
    role="button"
    tabindex="0"
  >
    {$t('name')}
    <span class="arr" class:on={s.by === 'name'}>{s.by === 'name' && s.asc ? '▲' : '▼'}</span>
  </span>

  <span
    class="h right"
    class:active={s.by === 'files'}
    on:click={() => toggleSort('files')}
    on:keydown={(e) => keySort(e, 'files')}
    role="button"
    tabindex="0"
  >
    {$t('items')}
    <span class="arr" class:on={s.by === 'files'}>{s.by === 'files' && s.asc ? '▲' : '▼'}</span>
  </span>

  <span
    class="h center"
    class:active={s.by === 'usage'}
    on:click={() => toggleSort('usage')}
    on:keydown={(e) => keySort(e, 'usage')}
    role="button"
    tabindex="0"
  >
    {$t('usage')}
    <span class="arr" class:on={s.by === 'usage'}>{s.by === 'usage' && s.asc ? '▲' : '▼'}</span>
  </span>
  <span></span>

  <span
    class="h right"
    class:active={s.by === 'size'}
    on:click={() => toggleSort('size')}
    on:keydown={(e) => keySort(e, 'size')}
    role="button"
    tabindex="0"
  >
    {$t('size')}
    <span class="arr" class:on={s.by === 'size'}>{s.by === 'size' && s.asc ? '▲' : '▼'}</span>
  </span>

  <span></span>
</div>

<style>
  .thead {
    position: sticky;
    top: 0;
    z-index: 1;
    height: 30px;
    padding: 0 8px;
    background: var(--bg);
    border-bottom: 1px solid var(--line);
    font-size: 12px;
    color: var(--muted);
    user-select: none;
  }
  .h {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    cursor: pointer;
    border-radius: 4px;
    padding: 2px 4px;
    white-space: nowrap;
  }
  .h.right {
    justify-content: flex-end;
  }
  .h.center {
    justify-content: center;
  }
  .h:hover {
    color: var(--fg);
    background: var(--hover);
  }
  .h.active {
    color: var(--accent);
    font-weight: 600;
  }
  /* Triangle only on the active column (accent, pointing the sort direction).
     Inactive columns keep the reserved space so nothing shifts. */
  .arr {
    font-size: 9px;
    line-height: 1;
    opacity: 0;
  }
  .arr.on {
    opacity: 1;
    color: var(--accent);
  }
</style>
