<script>
  import { createEventDispatcher } from 'svelte'
  import { formatBytes } from '../util.js'

  export let drives = []

  const dispatch = createEventDispatcher()
  const usedPct = (d) => (d.total > 0 ? (d.used / d.total) * 100 : 0)
  const hue = (p) => Math.max(0, 120 - p * 1.2)
</script>

<div class="dash">
  <h2>드라이브</h2>
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
            {formatBytes(d.used)} / {formatBytes(d.total)} 사용 · {formatBytes(d.free)} 여유
          {:else}
            용량 정보 없음
          {/if}
        </div>
      </button>
    {/each}
  </div>
  <p class="hint">드라이브를 클릭하면 스캔을 시작합니다.</p>
</div>

<style>
  .dash {
    padding: 20px;
    max-width: 720px;
    margin: 0 auto;
  }
  h2 {
    font-size: 15px;
    font-weight: 600;
    margin: 4px 0 14px;
    color: var(--fg);
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
    background: #fff;
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
