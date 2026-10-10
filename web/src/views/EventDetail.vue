<template>
  <div v-loading="!d">
    <div class="page-head">
      <div><el-button text @click="$router.back()"><el-icon><ArrowLeft /></el-icon>返回</el-button></div>
    </div>
    <template v-if="d">
      <div class="card ev-head">
        <div class="ev-title">
          <h2>{{ e.title }}</h2>
          <div class="tags">
            <el-tag :type="statusType(e.status)" effect="dark" round>{{ e.status }}</el-tag>
            <el-tag :type="statusType(e.urgency)" effect="light" round>{{ e.urgency }}</el-tag>
            <el-tag effect="plain" round>{{ e.category }}</el-tag>
          </div>
        </div>
        <div class="flow-actions" v-if="d.actions.length">
          <el-button v-if="has('accept')" type="primary" @click="act('accept')">受理</el-button>
          <el-button v-if="has('process')" type="primary" @click="open('process')">{{ e.status === '处理中' ? '改派' : '交办处理' }}</el-button>
          <el-button v-if="has('finish')" type="success" @click="open('finish')">办结</el-button>
          <el-button v-if="has('reject')" type="danger" plain @click="open('reject')">退回</el-button>
          <el-button v-if="has('resubmit')" type="warning" @click="act('resubmit')">重新上报</el-button>
        </div>
      </div>

      <div class="grid-21 mt">
        <div class="card">
          <div class="card-title"><span>事件信息</span></div>
          <el-descriptions :column="store.isMobile ? 1 : 2" border>
            <el-descriptions-item label="事件编号">{{ e.event_no }}</el-descriptions-item>
            <el-descriptions-item label="所属网格">{{ d.grid_name }}</el-descriptions-item>
            <el-descriptions-item label="发生位置" :span="store.isMobile ? 1 : 2">{{ e.location || '—' }}</el-descriptions-item>
            <el-descriptions-item label="上报人">{{ e.reporter }}</el-descriptions-item>
            <el-descriptions-item label="上报时间">{{ fmtTime(e.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="处理人">{{ e.handler || '—' }}</el-descriptions-item>
            <el-descriptions-item label="办结时间">{{ fmtTime(e.finished_at) || '—' }}</el-descriptions-item>
            <el-descriptions-item label="事件描述" :span="store.isMobile ? 1 : 2"><span class="pre">{{ e.description }}</span></el-descriptions-item>
            <el-descriptions-item v-if="e.result" label="办理结果" :span="store.isMobile ? 1 : 2"><span class="pre">{{ e.result }}</span></el-descriptions-item>
          </el-descriptions>
          <div class="imgs" v-if="imgs.length">
            <el-image v-for="(u, i) in imgs" :key="u" :src="u" :preview-src-list="imgs" :initial-index="i" fit="cover" preview-teleported />
          </div>
        </div>
        <div class="card">
          <div class="card-title"><span>流转记录</span></div>
          <el-timeline>
            <el-timeline-item v-for="f in [...d.flows].reverse()" :key="f.id" :timestamp="fmtTime(f.created_at)" :type="dot(f.to_status)" placement="top">
              <b>{{ f.action }}</b> <span class="muted">· {{ f.operator }}</span>
              <div class="flow-c" v-if="f.comment">{{ f.comment }}</div>
            </el-timeline-item>
          </el-timeline>
        </div>
      </div>
    </template>

    <el-dialog v-model="dlg" :title="{ process: '交办处理', finish: '办结事件', reject: '退回事件' }[mode]" width="480px" :fullscreen="store.isMobile">
      <el-form label-position="top">
        <el-form-item v-if="mode === 'process'" label="处理人" required>
          <el-select v-model="fm.handler_id" filterable placeholder="选择处理人" style="width: 100%">
            <el-option v-for="u in store.users" :key="u.id" :label="u.real_name + (u.position ? '（' + u.position + '）' : '')" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="mode === 'finish'" label="办理结果" required><el-input v-model="fm.result" type="textarea" :rows="4" placeholder="说明处置过程与结果" /></el-form-item>
        <el-form-item :label="mode === 'reject' ? '退回原因' : '备注说明'" :required="mode === 'reject'"><el-input v-model="fm.comment" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dlg = false">取消</el-button><el-button type="primary" :loading="busy" @click="act(mode)">确定</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import { store } from '@/store'
import { fmtTime, statusType, parseImages } from '@/utils/format'
import { loadUsers } from '@/utils/options'

const route = useRoute()
const d = ref(null)
const e = computed(() => d.value.event)
const imgs = computed(() => parseImages(e.value.images))
const has = a => d.value.actions.includes(a)
const dot = s => ({ 办结: 'success', 退回: 'danger', 处理中: 'primary', 受理: 'primary' }[s] || 'warning')
const dlg = ref(false)
const mode = ref('')
const busy = ref(false)
const fm = ref({})

async function load() { d.value = await http.get('/events/' + route.params.id) }
function open(m) { mode.value = m; fm.value = { handler_id: e.value.handler_id || null }; dlg.value = true }
async function act(a) {
  busy.value = true
  try {
    await http.post(`/events/${e.value.id}/flow`, { action: a, ...fm.value })
    ElMessage.success('操作成功')
    dlg.value = false
    fm.value = {}
    load()
  } catch { /* */ } finally { busy.value = false }
}
onMounted(() => { load(); loadUsers() })
</script>

<style scoped>
.ev-head { display: flex; justify-content: space-between; align-items: center; gap: 12px; flex-wrap: wrap; }
.ev-head h2 { margin: 0 0 8px; font-size: 20px; }
.tags { display: flex; gap: 6px; }
.flow-actions { display: flex; gap: 8px; flex-wrap: wrap; }
.flow-actions .el-button + .el-button { margin-left: 0; }
.imgs { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 14px; }
.imgs .el-image { width: 110px; height: 110px; border-radius: 10px; }
.flow-c { margin-top: 4px; color: #3d4f4b; font-size: 13px; background: #f6f9f8; padding: 6px 10px; border-radius: 8px; }
</style>
