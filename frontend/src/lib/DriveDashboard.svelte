<script>
  import { createEventDispatcher } from 'svelte'
  import { formatBytes } from '../util.js'
  import { t, locale } from './i18n.js'
  import { theme, toggleTheme } from './theme.js'

  export let drives = []

  const dispatch = createEventDispatcher()
  const usedPct = (d) => (d.total > 0 ? (d.used / d.total) * 100 : 0)
  const hue = (p) => Math.max(0, 120 - p * 1.2)
</script>

<div class="dash">
  <div class="dash-head">
    <h2>{$t('drives')}</h2>
    <div class="settings">
      <button class="theme" on:click={toggleTheme} title={$theme === 'dark' ? $t('darkOff') : $t('darkOn')}>
        {$theme === 'dark' ? '☀️' : '🌙'}
      </button>
      <select class="lang" bind:value={$locale} title="Language">
        <option value="ko">한국어</option>
        <option value="en">English</option>
      </select>
    </div>
  </div>
  <div class="cards">
    {#each drives as d}
      <button class="card" on:click={() => dispatch('scan', d.path)}>
        <div class="top">
          <span class="path">{d.path}</span>
          <span class="pct">{usedPct(d).toFixed(0)}%</span>
        </div>
        <div class="bar">
          <span class="fill" style="width:{usedPct(d)}%; background:hsl({hue(usedPct(d))},65%,50%)"></span>
        </div>
        <div class="sub">
          {#if d.total > 0}
            {$t('driveSub', {
              used: formatBytes(d.used),
              total: formatBytes(d.total),
              free: formatBytes(d.free),
            })}
          {:else}
            {$t('driveNoInfo')}
          {/if}
        </div>
      </button>
    {/each}
  </div>
  <p class="hint">{$t('driveHint')}</p>
</div>

<style>
  .dash {
    padding: 20px;
    max-width: 720px;
    margin: 0 auto;
  }
  .dash-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 4px 0 14px;
  }
  h2 {
    font-size: 15px;
    font-weight: 600;
    margin: 0;
    color: var(--fg);
  }
  .settings {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .theme {
    border: 1px solid var(--border-strong);
    border-radius: 6px;
    padding: 5px 9px;
    font-size: 15px;
    line-height: 1;
    background: var(--surface);
    cursor: pointer;
  }
  .theme:hover {
    background: var(--hover);
  }
  .lang {
    border: 1px solid var(--border-strong);
    border-radius: 6px;
    padding: 6px 8px;
    font-size: 13px;
    background: var(--surface);
    color: var(--fg);
    cursor: pointer;
  }
  .cards {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 12px;
  }
  .card {
    text-align: left;
    border: 1px solid var(--line);
    border-radius: 10px;
    padding: 14px;
    background: var(--surface);
    color: var(--fg);
    cursor: pointer;
    transition: border-color 0.12s, box-shadow 0.12s;
  }
  .card:hover {
    border-color: var(--accent);
    box-shadow: 0 2px 8px rgba(47, 111, 237, 0.12);
  }
  .top {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    margin-bottom: 8px;
  }
  .path {
    font-size: 16px;
    font-weight: 700;
  }
  .pct {
    font-size: 13px;
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }
  .bar {
    height: 12px;
    background: var(--line);
    border-radius: 4px;
    overflow: hidden;
  }
  .fill {
    display: block;
    height: 100%;
  }
  .sub {
    margin-top: 8px;
    font-size: 12px;
    color: var(--muted);
  }
  .hint {
    color: var(--muted);
    font-size: 13px;
    margin-top: 16px;
  }
</style>
