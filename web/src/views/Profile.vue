<template>
  <div class="profile">
    <div class="card me">
      <div class="big-avatar">{{ (store.user?.real_name || '').slice(-2) }}</div>
      <div>
        <h2>{{ store.user?.real_name }}</h2>
        <p class="muted">{{ store.user?.username }} · {{ roleName(store.user?.role) }}{{ store.user?.position ? ' · ' + store.user.position : '' }}</p>
      </div>
    </div>
    <div class="grid-2 mt">
      <div class="card">
        <div class="card-title"><span>个人信息</span></div>
        <el-form label-width="80px">
          <el-form-item label="姓名"><el-input v-model="pf.real_name" /></el-form-item>
          <el-form-item label="手机号"><el-input v-model="pf.phone" /></el-form-item>
          <el-button type="primary" @click="saveProfile">保存</el-button>
        </el-form>
      </div>
      <div class="card">
        <div class="card-title"><span>修改密码</span></div>
        <el-form ref="pwRef" :model="pw" :rules="rules" label-width="80px">
          <el-form-item label="原密码" prop="old_password"><el-input v-model="pw.old_password" type="password" show-password /></el-form-item>
          <el-form-item label="新密码" prop="new_password"><el-input v-model="pw.new_password" type="password" show-password placeholder="至少 8 位，含字母和数字" /></el-form-item>
          <el-form-item label="确认密码" prop="confirm"><el-input v-model="pw.confirm" type="password" show-password /></el-form-item>
          <el-button type="primary" @click="savePw">修改密码</el-button>
        </el-form>
      </div>
    </div>
    <div class="card mt m-links" v-if="store.isMobile">
      <div class="ml" @click="$router.push('/worklogs')"><el-icon><Notebook /></el-icon>工作日志</div>
      <div class="ml" @click="$router.push('/households')"><el-icon><House /></el-icon>户籍台账</div>
      <div class="ml" @click="$router.push('/notices')"><el-icon><Bell /></el-icon>通知公告</div>
      <div class="ml" @click="$router.push('/points')"><el-icon><Medal /></el-icon>治理积分</div>
      <div class="ml" @click="$router.push('/settings')"><el-icon><Brush /></el-icon>主题设置</div>
      <div class="ml" @click="$router.push('/')"><el-icon><HomeFilled /></el-icon>门户首页</div>
    </div>
    <el-button v-if="store.isMobile" class="mt logout" type="danger" plain @click="out">退出登录</el-button>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import { store, roleName, logout } from '@/store'

const router = useRouter()
const pf = reactive({ real_name: store.user?.real_name, phone: store.user?.phone })
const pw = reactive({ old_password: '', new_password: '', confirm: '' })
const pwRef = ref()
const rules = {
  old_password: [{ required: true, message: '请输入原密码' }],
  new_password: [{ required: true, message: '请输入新密码' }, { min: 8, message: '至少 8 位' }],
  confirm: [{ validator: (_, v, cb) => (v === pw.new_password ? cb() : cb(new Error('两次输入不一致'))) }]
}
async function saveProfile() {
  await http.put('/me/profile', pf)
  Object.assign(store.user, pf)
  localStorage.setItem('vm_user', JSON.stringify(store.user))
  ElMessage.success('已保存')
}
async function savePw() {
  await pwRef.value.validate()
  await http.put('/me/password', pw)
  ElMessage.success('密码已修改，请重新登录')
  logout(); router.replace('/login')
}
function out() { logout(); router.replace('/login') }
</script>

<style scoped>
.me { display: flex; align-items: center; gap: 16px; background: linear-gradient(120deg, var(--vm-primary-soft), #fff); }
.me h2 { margin: 0 0 4px; }
.me p { margin: 0; }
.big-avatar { width: 64px; height: 64px; border-radius: 50%; display: grid; place-items: center; color: #fff; font-size: 20px; font-weight: 700; background: linear-gradient(135deg, var(--vm-primary), var(--vm-primary-deep)); }
.m-links { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
.ml { display: flex; flex-direction: column; align-items: center; gap: 6px; font-size: 13px; padding: 10px 0; }
.ml .el-icon { font-size: 22px; color: var(--vm-primary); }
.logout { width: 100%; }
</style>
