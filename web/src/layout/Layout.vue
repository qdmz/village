<template>
  <div class="layout" :class="{ mobile: store.isMobile, collapsed }">
    <aside v-if="!store.isMobile" class="sidebar">
      <SideMenu :collapsed="collapsed" />
    </aside>
    <el-drawer v-else v-model="drawer" direction="ltr" size="78%" :with-header="false" class="menu-drawer">
      <SideMenu @navigate="drawer = false" />
    </el-drawer>

    <div class="main">
      <header class="topbar">
        <div class="tb-left">
          <el-button text class="icon-btn" @click="store.isMobile ? (drawer = true) : (collapsed = !collapsed)">
            <el-icon :size="20"><Fold v-if="!collapsed || store.isMobile" /><Expand v-else /></el-icon>
          </el-button>
          <div class="crumb">
            <span class="crumb-village" v-if="!store.isMobile">{{ store.settings.village_name }}</span>
            <el-icon v-if="!store.isMobile"><ArrowRight /></el-icon>
            <b>{{ route.meta.title }}</b>
          </div>
        </div>
        <div class="tb-right">
          <el-tooltip content="数据驾驶舱" v-if="!store.isMobile"><el-button text class="icon-btn" @click="$router.push('/screen')"><el-icon :size="18"><DataBoard /></el-icon></el-button></el-tooltip>
          <el-tooltip content="门户首页" v-if="!store.isMobile"><el-button text class="icon-btn" @click="$router.push('/')"><el-icon :size="18"><HomeFilled /></el-icon></el-button></el-tooltip>
          <el-dropdown trigger="click" @command="setTheme">
            <el-button text class="icon-btn"><span class="theme-dot" /></el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="t in THEMES" :key="t.key" :command="t.key">
                  <span class="theme-swatch" :style="{ background: t.color }" />{{ t.name }}
                  <el-icon v-if="theme === t.key" style="margin-left: 6px"><Check /></el-icon>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-dropdown trigger="click" @command="onUser">
            <div class="user-chip">
              <span class="avatar">{{ (store.user?.real_name || '?').slice(-2) }}</span>
              <div class="u-text" v-if="!store.isMobile"><b>{{ store.user?.real_name }}</b><small>{{ roleName(store.user?.role) }}</small></div>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile"><el-icon><User /></el-icon>个人中心</el-dropdown-item>
                <el-dropdown-item command="portal"><el-icon><HomeFilled /></el-icon>门户首页</el-dropdown-item>
                <el-dropdown-item command="screen"><el-icon><DataBoard /></el-icon>数据驾驶舱</el-dropdown-item>
                <el-dropdown-item command="logout" divided><el-icon><SwitchButton /></el-icon>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>
      <main class="content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in"><component :is="Component" :key="route.fullPath" /></transition>
        </router-view>
      </main>
      <footer v-if="!store.isMobile" class="foot">{{ store.settings.system_title }} · {{ store.settings.footer_text }}</footer>
    </div>

    <nav v-if="store.isMobile" class="tabbar">
      <router-link v-for="t in tabs" :key="t.path" :to="t.path" class="tab" :class="{ on: route.path === t.path, big: t.big }">
        <span class="ti"><el-icon><component :is="t.icon" /></el-icon></span><span>{{ t.label }}</span>
      </router-link>
    </nav>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { store, logout, roleName } from '@/store'
import { THEMES, applyTheme, currentTheme } from '@/utils/theme'
import { resetOptionCache } from '@/utils/options'
import SideMenu from './SideMenu.vue'

const route = useRoute()
const router = useRouter()
const collapsed = ref(false)
const drawer = ref(false)
const theme = ref(currentTheme())
const tabs = [
  { path: '/dashboard', label: '工作台', icon: 'House' },
  { path: '/patrols', label: '巡查', icon: 'Location' },
  { path: '/events', label: '上报', icon: 'Plus', big: true },
  { path: '/tasks', label: '任务', icon: 'Finished' },
  { path: '/profile', label: '我的', icon: 'User' }
]

function setTheme(k) { applyTheme(k); theme.value = k }
async function onUser(cmd) {
  if (cmd === 'logout') {
    await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' })
    logout(); resetOptionCache(); router.replace('/login')
  } else if (cmd === 'portal') router.push('/')
  else router.push('/' + cmd)
}
</script>
