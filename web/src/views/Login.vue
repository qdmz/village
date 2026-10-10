<template>
  <div class="login-page">
    <section class="hero">
      <div class="hero-inner">
        <div class="brand-row" @click="$router.push('/')">
          <div class="logo big">{{ store.settings.logo_text || '村' }}</div>
          <div><b>{{ store.settings.village_name }}</b><small>{{ store.settings.town_name || '数字乡村' }}</small></div>
        </div>
        <h1>{{ store.settings.system_title }}</h1>
        <p>党建引领 · 网格治理 · 民生保障 · 产业振兴<br />让村里的每一件事都有人管、有人办、有回音。</p>
        <ul class="feats">
          <li><el-icon><Aim /></el-icon>网格巡查与事件闭环流转</li>
          <li><el-icon><Finished /></el-icon>任务督办、逾期自动预警</li>
          <li><el-icon><DataAnalysis /></el-icon>人口台账与数据驾驶舱</li>
        </ul>
      </div>
      <svg class="hills" viewBox="0 0 1200 220" preserveAspectRatio="none"><path d="M0 160 C200 80 360 200 560 130 S900 60 1200 140 V220 H0Z" fill="rgba(255,255,255,.08)"/><path d="M0 190 C240 120 420 220 700 160 S1000 120 1200 180 V220 H0Z" fill="rgba(255,255,255,.12)"/></svg>
    </section>
    <section class="form-side">
      <div class="form-box">
        <h2>账号登录</h2>
        <p class="muted">管理员、村干部、网格员使用统一入口</p>
        <el-form ref="formRef" :model="form" :rules="rules" size="large" @keyup.enter="submit">
          <el-form-item prop="username"><el-input v-model="form.username" placeholder="登录账号" :prefix-icon="User" autocomplete="username" /></el-form-item>
          <el-form-item prop="password"><el-input v-model="form.password" type="password" placeholder="密码" show-password :prefix-icon="Lock" autocomplete="current-password" /></el-form-item>
          <el-button type="primary" class="w100" :loading="loading" @click="submit">登 录</el-button>
        </el-form>
        <div class="links"><a @click="$router.push('/')"><el-icon><HomeFilled /></el-icon> 返回门户首页</a></div>
        <p class="tip muted">忘记密码请联系系统管理员重置</p>
      </div>
    </section>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import { store, setLogin } from '@/store'
import { resetOptionCache } from '@/utils/options'

const router = useRouter()
const route = useRoute()
const formRef = ref()
const loading = ref(false)
const form = reactive({ username: '', password: '' })
const rules = { username: [{ required: true, message: '请输入账号' }], password: [{ required: true, message: '请输入密码' }] }

async function submit() {
  await formRef.value.validate()
  loading.value = true
  try {
    const d = await http.post('/auth/login', form)
    setLogin(d)
    resetOptionCache()
    ElMessage.success(`欢迎回来，${d.user.real_name}`)
    router.replace(route.query.redirect || '/dashboard')
  } catch { /* 已提示 */ } finally { loading.value = false }
}
</script>

<style scoped lang="scss">
.login-page { min-height: 100vh; display: grid; grid-template-columns: 1.15fr 1fr; background: #fff; }
.hero { position: relative; overflow: hidden; color: #fff; display: flex; align-items: center; padding: 48px 64px;
  background: radial-gradient(1200px 600px at 10% 0%, rgba(163,230,53,.25), transparent 60%), linear-gradient(140deg, var(--vm-primary-deep), var(--vm-primary) 70%, #65a30d); }
.hero-inner { position: relative; z-index: 1; max-width: 560px; }
.brand-row { display: flex; align-items: center; gap: 12px; cursor: pointer; margin-bottom: 48px; }
.brand-row b { display: block; font-size: 18px; letter-spacing: 2px; }
.brand-row small { opacity: .7; }
.logo.big { width: 48px; height: 48px; font-size: 22px; border-radius: 14px; background: rgba(255,255,255,.18); border: 1px solid rgba(255,255,255,.35); display: grid; place-items: center; font-weight: 700; }
h1 { font-size: 34px; line-height: 1.35; margin: 0 0 16px; letter-spacing: 1px; }
.hero p { opacity: .85; line-height: 1.9; margin: 0 0 28px; }
.feats { list-style: none; padding: 0; margin: 0; display: grid; gap: 12px; }
.feats li { display: flex; align-items: center; gap: 10px; background: rgba(255,255,255,.1); border: 1px solid rgba(255,255,255,.16); padding: 10px 14px; border-radius: 12px; width: fit-content; }
.hills { position: absolute; left: 0; right: 0; bottom: 0; width: 100%; height: 160px; }
.form-side { display: flex; align-items: center; justify-content: center; padding: 32px; }
.form-box { width: 100%; max-width: 380px; }
.form-box h2 { font-size: 26px; margin: 0 0 6px; }
.form-box > p { margin: 0 0 28px; }
.w100 { width: 100%; height: 46px; font-size: 16px; letter-spacing: 4px; border-radius: 10px; }
.links { margin-top: 18px; text-align: center; }
.links a { cursor: pointer; display: inline-flex; align-items: center; gap: 4px; }
.tip { text-align: center; font-size: 12px; margin-top: 28px; }
@media (max-width: 860px) {
  .login-page { grid-template-columns: 1fr; }
  .hero { padding: 36px 24px 70px; }
  .brand-row { margin-bottom: 22px; }
  h1 { font-size: 24px; }
  .feats { display: none; }
  .form-side { margin-top: -40px; padding: 0 16px 32px; }
  .form-box { background: #fff; border-radius: 18px; padding: 26px 20px; box-shadow: 0 10px 30px rgba(0,0,0,.08); position: relative; z-index: 2; }
}
</style>
