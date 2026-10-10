// 主题色切换：动态生成 Element Plus 色阶变量
export const THEMES = [
  { key: 'teal', name: '青山绿', color: '#0d9488' },
  { key: 'green', name: '田园绿', color: '#16a34a' },
  { key: 'blue', name: '政务蓝', color: '#2563eb' },
  { key: 'red', name: '党建红', color: '#c8102e' },
  { key: 'amber', name: '丰收金', color: '#b45309' }
]

function mix(c1, c2, w) {
  const p = s => [1, 3, 5].map(i => parseInt(s.slice(i, i + 2), 16))
  const a = p(c1), b = p(c2)
  return '#' + a.map((v, i) => Math.round(v * (1 - w) + b[i] * w).toString(16).padStart(2, '0')).join('')
}

export function applyTheme(key) {
  const t = THEMES.find(x => x.key === key) || THEMES[0]
  const el = document.documentElement
  const c = t.color
  el.style.setProperty('--el-color-primary', c)
  for (const n of [3, 5, 7, 8, 9]) el.style.setProperty(`--el-color-primary-light-${n}`, mix(c, '#ffffff', n / 10))
  el.style.setProperty('--el-color-primary-dark-2', mix(c, '#000000', 0.2))
  el.style.setProperty('--vm-primary', c)
  el.style.setProperty('--vm-primary-soft', mix(c, '#ffffff', 0.9))
  el.style.setProperty('--vm-primary-deep', mix(c, '#0b1f1c', 0.45))
  document.querySelector('meta[name=theme-color]')?.setAttribute('content', c)
  localStorage.setItem('vm_theme', t.key)
  return t
}

export const currentTheme = () => localStorage.getItem('vm_theme') || 'teal'
export const themeColor = () => (THEMES.find(x => x.key === currentTheme()) || THEMES[0]).color
