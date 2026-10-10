<template>
  <div>
    <div class="page-head"><div><h2>系统设置</h2><p class="muted">村名与标识、主题配色、数据备份与恢复</p></div></div>
    <div class="grid-2">
      <div class="card">
        <div class="card-title"><span>基础信息</span><el-tag v-if="!isAdmin()" size="small" type="info">仅管理员可修改</el-tag></div>
        <el-form :model="s" label-width="96px" :disabled="!isAdmin()" :label-position="store.isMobile ? 'top' : 'right'">
          <el-form-item label="系统名称"><el-input v-model="s.system_title" maxlength="40" /></el-form-item>
          <el-form-item label="村名"><el-input v-model="s.village_name" maxlength="20" /></el-form-item>
          <el-form-item label="所属乡镇"><el-input v-model="s.town_name" maxlength="30" /></el-form-item>
          <el-form-item label="Logo 文字"><el-input v-model="s.logo_text" maxlength="2" placeholder="1-2 个字，如 青" style="width: 140px" />
            <span class="logo-prev">{{ s.logo_text }}</span></el-form-item>
          <el-form-item label="服务电话"><el-input v-model="s.contact_phone" maxlength="30" /></el-form-item>
          <el-form-item label="页脚文字"><el-input v-model="s.footer_text" maxlength="60" /></el-form-item>
          <el-form-item v-if="isAdmin()"><el-button type="primary" @click="save">保存设置</el-button></el-form-item>
        </el-form>
      </div>
      <div>
        <div class="card">
          <div class="card-title"><span>主题配色</span></div>
          <div class="themes">
            <div v-for="t in THEMES" :key="t.key" class="theme" :class="{ on: theme === t.key }" @click="setTheme(t.key)">
              <span class="sw" :style="{ background: `linear-gradient(135deg, ${t.color}, ${t.color}aa)` }"><el-icon v-if="theme === t.key"><Check /></el-icon></span>
              <span>{{ t.name }}</span>
            </div>
          </div>
          <p class="muted small">主题保存在当前浏览器，不影响其他用户。</p>
        </div>
        <div class="card mt" v-if="isAdmin()">
          <div class="card-title"><span>数据备份与恢复</span></div>
          <p class="muted small">备份为 JSON 文件，包含全部业务数据、用户（密码为加密散列）与系统设置。恢复将<b>覆盖</b>当前全部数据，请谨慎操作。</p>
          <div class="bk">
            <el-button type="primary" :loading="bk" @click="backup"><el-icon><Download /></el-icon><span>下载完整备份</span></el-button>
            <el-button type="danger" plain :loading="rs" @click="$refs.rf.click()"><el-icon><Upload /></el-icon><span>从备份恢复</span></el-button>
            <input ref="rf" type="file" accept=".json,application/json" hidden @change="restore" />
          </div>
        </div>
        <div class="card mt">
          <div class="card-title"><span>关于</span></div>
          <p class="small">乡村振兴村居日常工作管理系统 v0.1.0<br /><span class="muted">Go (Gin + GORM) · Vue 3 · Element Plus · ECharts</span></p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import http from '@/api/http'
import { store, isAdmin, logout } from '@/store'
import { THEMES, applyTheme, currentTheme } from '@/utils/theme'
import { downloadBlob } from '@/utils/excel'

const s = reactive({})
const theme = ref(currentTheme())
const bk = ref(false)
const rs = ref(false)
function setTheme(k) { applyTheme(k); theme.value = k }
async function save() {
  const d = await http.put('/settings', s)
  Object.assign(store.settings, d)
  ElMessage.success('已保存')
}
async function backup() {
  bk.value = true
  try {
    const blob = await http.get('/system/backup', { responseType: 'blob' })
    downloadBlob(blob, `village-backup-${new Date().toISOString().slice(0, 10)}.json`)
  } finally { bk.value = false }
}
async function restore(e) {
  const f = e.target.files[0]
  e.target.value = ''
  if (!f) return
  await ElMessageBox.confirm(`将使用「${f.name}」覆盖当前全部数据（包括用户账号），此操作不可撤销。确定继续？`, '恢复数据', { type: 'error', confirmButtonText: '确定恢复' })
  rs.value = true
  try {
    const fd = new FormData()
    fd.append('file', f)
    await http.post('/system/restore', fd)
    await ElMessageBox.alert('数据已恢复，请重新登录。', '完成')
    logout(); location.hash = '#/login'
  } finally { rs.value = false }
}
onMounted(async () => Object.assign(s, await http.get('/settings')))
</script>

<style scoped>
.themes { display: grid; grid-template-columns: repeat(5, 1fr); gap: 10px; }
.theme { display: flex; flex-direction: column; align-items: center; gap: 6px; cursor: pointer; font-size: 13px; padding: 10px 0; border-radius: 12px; border: 1px solid transparent; }
.theme.on { border-color: var(--vm-primary); background: var(--vm-primary-soft); }
.sw { width: 40px; height: 40px; border-radius: 12px; display: grid; place-items: center; color: #fff; font-size: 18px; }
.small { font-size: 13px; line-height: 1.8; }
.bk { display: flex; gap: 10px; flex-wrap: wrap; }
.bk .el-button + .el-button { margin-left: 0; }
.bk .el-icon + span { margin-left: 4px; }
.logo-prev { display: inline-grid; place-items: center; width: 34px; height: 34px; border-radius: 10px; margin-left: 10px; color: #fff; font-weight: 700; background: linear-gradient(135deg, var(--vm-primary), #84cc16); }
@media (max-width: 767px) { .themes { grid-template-columns: repeat(3, 1fr); } }
</style>
