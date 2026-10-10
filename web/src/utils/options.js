import http from '@/api/http'
import { store, isGrid } from '@/store'
import { money } from './format'

let gridsP = null
let usersP = null

export function loadGrids(force) {
  if (!store.token) return Promise.resolve([])
  if (!gridsP || force) gridsP = http.get('/grids', { params: { size: 500 } }).then(d => (store.grids = d.list)).catch(() => (gridsP = null))
  return gridsP
}
export function loadUsers(force) {
  if (!store.token) return Promise.resolve([])
  if (!usersP || force) usersP = http.get('/users/options').then(d => (store.users = d || [])).catch(() => (usersP = null))
  return usersP
}
export function resetOptionCache() { gridsP = null; usersP = null }

export const gridName = id => store.grids.find(g => g.id === id)?.name || ''
export const userName = id => store.users.find(u => u.id === id)?.real_name || ''

export function optionsOf(f, forForm) {
  if (f.type === 'grid') {
    let gs = store.grids
    if (forForm && isGrid()) gs = gs.filter(g => store.gridIds.includes(g.id))
    return gs.map(g => ({ label: g.name, value: g.id }))
  }
  if (f.type === 'user') {
    return store.users.filter(u => !f.userRole || f.userRole.includes(u.role)).map(u => ({ label: u.real_name + (u.position ? `（${u.position}）` : ''), value: u.id }))
  }
  const opts = typeof f.options === 'function' ? f.options() : f.options || []
  return opts.map(o => (typeof o === 'object' ? o : { label: o, value: o }))
}

export function cellText(f, row) {
  const v = row[f.prop]
  if (f.fmt) return f.fmt(row)
  switch (f.type) {
    case 'grid': return gridName(v)
    case 'user': return row[f.nameProp] || userName(v)
    case 'remote': return row[f.nameProp] || ''
    case 'switch': return v ? '是' : '否'
    case 'money': return v || v === 0 ? money(v, f.unit || '') : ''
    case 'number': return v ?? ''
    case 'progress': return (v || 0) + '%'
    case 'select':
      if (f.options && typeof f.options[0] === 'object') return (f.options.find(o => o.value === v) || {}).label ?? v
      return v ?? ''
    default: return v ?? ''
  }
}
