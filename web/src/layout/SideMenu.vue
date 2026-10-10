<template>
  <div class="side">
    <div class="brand" @click="go('/dashboard')">
      <div class="logo">{{ store.settings.logo_text || '村' }}</div>
      <div class="brand-text" v-if="!collapsed">
        <b>{{ store.settings.village_name }}</b>
        <small>乡村振兴 · 村居管理</small>
      </div>
    </div>
    <el-scrollbar class="menu-scroll">
      <el-menu :default-active="active" :collapse="collapsed" :collapse-transition="false" router @select="emit('navigate')" class="side-menu">
        <template v-for="g in menus" :key="g.title">
          <el-menu-item v-if="g.path" :index="g.path">
            <el-icon><component :is="g.icon" /></el-icon><template #title>{{ g.title }}</template>
          </el-menu-item>
          <el-sub-menu v-else :index="g.title">
            <template #title><el-icon><component :is="g.icon" /></el-icon><span>{{ g.title }}</span></template>
            <el-menu-item v-for="c in g.children" :key="c.path" :index="c.path">{{ c.title }}</el-menu-item>
          </el-sub-menu>
        </template>
      </el-menu>
    </el-scrollbar>
    <div class="side-foot" v-if="!collapsed">
      <a @click="go('/screen')"><el-icon><DataBoard /></el-icon>数据驾驶舱</a>
      <a @click="go('/')"><el-icon><HomeFilled /></el-icon>门户首页</a>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { store } from '@/store'

defineProps({ collapsed: Boolean })
const emit = defineEmits(['navigate'])
const route = useRoute()
const router = useRouter()
const AC = ['admin', 'cadre']

const ALL = [
  { title: '工作台', path: '/dashboard', icon: 'Odometer' },
  { title: '村情与网格', path: '/org', icon: 'MapLocation' },
  { title: '户籍人口', icon: 'User', children: [{ title: '户籍台账', path: '/households' }, { title: '人口台账', path: '/villagers' }] },
  { title: '网格治理', icon: 'Aim', children: [{ title: '网格巡查', path: '/patrols' }, { title: '事件上报', path: '/events' }, { title: '矛盾纠纷调解', path: '/disputes' }] },
  { title: '任务督办', path: '/tasks', icon: 'Finished' },
  { title: '工作日志', path: '/worklogs', icon: 'Notebook' },
  { title: '党建引领', icon: 'Flag', children: [{ title: '党组织架构', path: '/party/orgs' }, { title: '党员花名册', path: '/party/members' }, { title: '三会一课', path: '/meetings' }] },
  { title: '民生保障', path: '/welfare', icon: 'FirstAidKit' },
  { title: '乡村治理', icon: 'Medal', children: [{ title: '治理积分', path: '/points' }, { title: '红黑榜', path: '/honors' }, { title: '村规民约', path: '/rules' }] },
  { title: '乡村振兴', icon: 'TrendCharts', children: [
    { title: '振兴概览', path: '/revital', roles: AC }, { title: '产业项目', path: '/projects', roles: AC }, { title: '资金台账', path: '/funds', roles: AC },
    { title: '集体资产', path: '/assets', roles: AC }, { title: '致富带头人', path: '/leaders' }] },
  { title: '通知公告', path: '/notices', icon: 'Bell' },
  { title: '门户管理', path: '/portal-admin', icon: 'Monitor', roles: AC },
  { title: '系统管理', icon: 'Setting', children: [
    { title: '用户管理', path: '/users', roles: ['admin'] }, { title: '操作日志', path: '/oplogs', roles: ['admin'] },
    { title: '系统设置', path: '/settings' }, { title: '个人中心', path: '/profile' }] }
]

const ok = x => !x.roles || x.roles.includes(store.user?.role)
const menus = computed(() => ALL.filter(ok).map(g => (g.children ? { ...g, children: g.children.filter(ok) } : g)).filter(g => !g.children || g.children.length))
const active = computed(() => '/' + route.path.split('/').filter(Boolean).slice(0, route.path.startsWith('/party') ? 2 : 1).join('/'))
function go(p) { router.push(p); emit('navigate') }
</script>
