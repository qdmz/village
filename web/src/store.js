import { reactive } from 'vue'

const saved = JSON.parse(localStorage.getItem('vm_user') || 'null')

export const store = reactive({
  token: localStorage.getItem('vm_token') || '',
  user: saved,
  gridIds: JSON.parse(localStorage.getItem('vm_grids') || '[]'),
  settings: { village_name: '青山村', system_title: '乡村振兴村居日常工作管理系统', logo_text: '青', footer_text: '' },
  isMobile: window.innerWidth < 768,
  grids: [],
  users: []
})

window.addEventListener('resize', () => { store.isMobile = window.innerWidth < 768 })

export function setLogin(data) {
  store.token = data.token
  store.user = data.user
  store.gridIds = data.grid_ids || []
  localStorage.setItem('vm_token', data.token)
  localStorage.setItem('vm_user', JSON.stringify(data.user))
  localStorage.setItem('vm_grids', JSON.stringify(store.gridIds))
}

export function logout() {
  store.token = ''
  store.user = null
  store.grids = []
  store.users = []
  localStorage.removeItem('vm_token')
  localStorage.removeItem('vm_user')
  localStorage.removeItem('vm_grids')
}

export const role = () => store.user?.role
export const isAdmin = () => role() === 'admin'
export const isManager = () => role() === 'admin' || role() === 'cadre'
export const isGrid = () => role() === 'grid'
export const roleName = r => ({ admin: '管理员', cadre: '村干部', grid: '网格员' }[r] || r)
