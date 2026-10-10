import { createRouter, createWebHashHistory } from 'vue-router'
import { store } from '@/store'
import Layout from '@/layout/Layout.vue'

const AC = ['admin', 'cadre']
const crud = (path, key, title, extra = {}) => ({ path, component: () => import('@/views/CrudView.vue'), props: { mkey: key }, meta: { title, ...extra } })

const routes = [
  { path: '/', component: () => import('@/views/Portal.vue'), meta: { public: true } },
  { path: '/p/:kind/:id', component: () => import('@/views/PortalDetail.vue'), meta: { public: true } },
  { path: '/login', component: () => import('@/views/Login.vue'), meta: { public: true } },
  { path: '/screen', component: () => import('@/views/Screen.vue'), meta: { title: '数据驾驶舱' } },
  {
    path: '/',
    component: Layout,
    children: [
      { path: 'dashboard', component: () => import('@/views/Dashboard.vue'), meta: { title: '工作台' } },
      { path: 'org', component: () => import('@/views/OrgGrid.vue'), meta: { title: '村情与网格' } },
      crud('households', 'households', '户籍台账'),
      { path: 'households/:id', component: () => import('@/views/HouseholdDetail.vue'), meta: { title: '家庭户详情' } },
      crud('villagers', 'villagers', '人口台账'),
      crud('patrols', 'patrols', '网格巡查'),
      { path: 'events', component: () => import('@/views/Events.vue'), meta: { title: '事件上报' } },
      { path: 'events/:id', component: () => import('@/views/EventDetail.vue'), meta: { title: '事件详情' } },
      { path: 'tasks', component: () => import('@/views/Tasks.vue'), meta: { title: '任务督办' } },
      crud('worklogs', 'worklogs', '工作日志'),
      { path: 'party/orgs', component: () => import('@/views/PartyOrgs.vue'), meta: { title: '党组织架构' } },
      crud('party/members', 'partyMembers', '党员花名册'),
      crud('meetings', 'meetings', '三会一课'),
      { path: 'welfare', component: () => import('@/views/Welfare.vue'), meta: { title: '民生保障' } },
      crud('disputes', 'disputes', '矛盾纠纷调解'),
      crud('rules', 'rules', '村规民约'),
      crud('honors', 'honors', '红黑榜'),
      { path: 'points', component: () => import('@/views/Points.vue'), meta: { title: '治理积分' } },
      { path: 'revital', component: () => import('@/views/Revital.vue'), meta: { title: '振兴概览', roles: AC } },
      crud('projects', 'projects', '产业项目', { roles: AC }),
      { path: 'funds', component: () => import('@/views/Funds.vue'), meta: { title: '资金台账', roles: AC } },
      crud('assets', 'assets', '集体资产', { roles: AC }),
      crud('leaders', 'leaders', '致富带头人'),
      crud('notices', 'notices', '通知公告'),
      crud('portal-admin', 'portal', '门户内容', { roles: AC }),
      { path: 'users', component: () => import('@/views/Users.vue'), meta: { title: '用户管理', roles: ['admin'] } },
      crud('oplogs', 'oplogs', '操作日志', { roles: ['admin'] }),
      { path: 'settings', component: () => import('@/views/Settings.vue'), meta: { title: '系统设置' } },
      { path: 'profile', component: () => import('@/views/Profile.vue'), meta: { title: '个人中心' } }
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' }
]

const router = createRouter({ history: createWebHashHistory(), routes, scrollBehavior: () => ({ top: 0 }) })

router.beforeEach(to => {
  if (to.meta.public) return true
  if (!store.token) return { path: '/login', query: { redirect: to.fullPath } }
  if (to.meta.roles && !to.meta.roles.includes(store.user?.role)) return '/dashboard'
  return true
})

export default router
