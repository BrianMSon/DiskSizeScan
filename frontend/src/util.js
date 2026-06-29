const UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']

export function formatBytes(bytes) {
  if (!bytes || bytes <= 0) return '0 B'
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), UNITS.length - 1)
  const v = bytes / Math.pow(1024, i)
  return `${v.toFixed(i === 0 ? 0 : 1)} ${UNITS[i]}`
}

export function formatCount(n) {
  return n.toLocaleString()
}
