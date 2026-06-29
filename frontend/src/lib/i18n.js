import { writable, derived } from 'svelte/store'

const messages = {
  ko: {
    scan: '스캔',
    cancel: '취소',
    pathPlaceholder: '스캔할 경로, 예: C:\\ 또는 /',
    foldersOnly: '폴더만',
    foldersOnlyTip: '폴더만 표시 (파일 숨김)',
    homeTip: '홈으로 (드라이브 목록)',
    drives: '드라이브',
    driveSub: '{used} / {total} 사용 · {free} 여유',
    driveNoInfo: '용량 정보 없음',
    driveHint: '드라이브를 클릭하면 스캔을 시작합니다.',
    startHint: '드라이브를 선택하거나 경로를 입력한 뒤 스캔하세요.',
    scanning: '스캔 중…',
    preparing: '준비 중…',
    loading: '불러오는 중…',
    name: '이름',
    items: '항목',
    usage: '사용률',
    size: '크기',
    summaryRest: '· {files}개 · {sec}s',
    driveChip: '드라이브의 {pct}% · {used}/{total} 사용',
    moreItems: '… {hidden}개 더 있음 (상위 {shown}개만 표시)',
    ctxScan: '📁 이 폴더 스캔',
    ctxExpandAll: '⊞ 모두 펼치기',
    ctxCollapseAll: '⊟ 모두 접기',
    ctxOpen: '↗ 탐색기에서 열기',
    openTip: '탐색기에서 열기',
  },
  en: {
    scan: 'Scan',
    cancel: 'Cancel',
    pathPlaceholder: 'Path to scan, e.g. C:\\ or /',
    foldersOnly: 'Folders',
    foldersOnlyTip: 'Show folders only (hide files)',
    homeTip: 'Home (drive list)',
    drives: 'Drives',
    driveSub: '{used} / {total} used · {free} free',
    driveNoInfo: 'No capacity info',
    driveHint: 'Click a drive to start scanning.',
    startHint: 'Pick a drive or enter a path, then Scan.',
    scanning: 'Scanning…',
    preparing: 'Preparing…',
    loading: 'Loading…',
    name: 'Name',
    items: 'Items',
    usage: 'Usage',
    size: 'Size',
    summaryRest: '· {files} items · {sec}s',
    driveChip: '{pct}% of drive · {used}/{total} used',
    moreItems: '… {hidden} more (showing top {shown})',
    ctxScan: '📁 Scan this folder',
    ctxExpandAll: '⊞ Expand all',
    ctxCollapseAll: '⊟ Collapse all',
    ctxOpen: '↗ Open in file manager',
    openTip: 'Open in file manager',
  },
}

function initialLocale() {
  try {
    const v = localStorage.getItem('lang')
    if (v === 'ko' || v === 'en') return v
  } catch (e) {
    // localStorage unavailable
  }
  return 'ko'
}

export const locale = writable(initialLocale())

locale.subscribe((v) => {
  try {
    localStorage.setItem('lang', v)
  } catch (e) {
    // ignore
  }
})

// $t('key', { var: value }) — reactive on locale.
export const t = derived(locale, ($locale) => (key, vars) => {
  const table = messages[$locale] || messages.ko
  let s = table[key] ?? messages.ko[key] ?? key
  if (vars) {
    for (const k in vars) s = s.split('{' + k + '}').join(vars[k])
  }
  return s
})
