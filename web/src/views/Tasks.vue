<template>
  <CrudPage ref="page" :cfg="cfg">
    <template #detail="{ row, reload }">
      <template v-if="row && row.task">
        <h3 class="t-title">{{ row.task.title }}</h3>
        <div class="t-tags">
          <el-tag :type="statusType(row.task.status)" effect="dark" round>{{ row.task.status }}</el-tag>
          <el-tag :type="statusType(row.task.priority)" round>{{ row.task.priority }}优先级</el-tag>
          <el-tag v-if="row.task.overdue" type="danger" effect="plain" round>已逾期</el-tag>
        </div>
        <el-descriptions :column="1" border class="mt">
          <el-descriptions-item label="责任人">{{ row.task.assignee_name }}</el-descriptions-item>
          <el-descriptions-item label="交办人">{{ row.task.creator_name }}</el-descriptions-item>
          <el-descriptions-item label="截止日期">{{ row.task.deadline }}</el-descriptions-item>
          <el-descriptions-item label="类别">{{ row.task.category || '—' }}</el-descriptions-item>
          <el-descriptions-item label="任务要求"><span class="pre">{{ row.task.content || '—' }}</span></el-descriptions-item>
          <el-descriptions-item label="当前进度"><el-progress :percentage="row.task.progress" :status="row.task.progress >= 100 ? 'success' : ''" /></el-descriptions-item>
        </el-descriptions>

        <div v-if="canUpdate(row.task)" class="card mt upd">
          <div class="card-title"><span>更新进展</span></div>
          <el-slider v-model="pf.progress" :step="5" show-stops :marks="{ 0: '0', 50: '50%', 100: '完成' }" />
          <el-input v-model="pf.content" type="textarea" :rows="3" placeholder="本次工作进展情况" class="mt" />
          <el-button type="primary" class="mt" :loading="busy" @click="submit(row, reload)">提交进展</el-button>
        </div>

        <div class="card-title mt"><span>进展记录</span></div>
        <el-timeline v-if="row.updates.length">
          <el-timeline-item v-for="u in row.updates" :key="u.id" :timestamp="fmtTime(u.created_at)" placement="top" :type="u.progress >= 100 ? 'success' : 'primary'">
            <b>{{ u.operator }}</b> 更新进度至 <b>{{ u.progress }}%</b>
            <div class="muted">{{ u.content }}</div>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无进展记录" :image-size="60" />
      </template>
    </template>
  </CrudPage>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import CrudPage from '@/components/CrudPage.vue'
import { M } from '@/config/modules'
import { store, isManager } from '@/store'
import { fmtTime, statusType } from '@/utils/format'

const page = ref()
const cfg = { ...M.tasks, fetchDetail: true, desc: '交办任务、跟踪进度，逾期任务自动标红预警' }
const pf = reactive({ progress: 0, content: '' })
const busy = ref(false)
const canUpdate = t => isManager() || t.assignee_id === store.user.id

async function submit(row, reload) {
  if (!pf.content.trim()) return ElMessage.warning('请填写进展情况')
  busy.value = true
  try {
    await http.post(`/tasks/${row.task.id}/progress`, { progress: pf.progress, content: pf.content })
    ElMessage.success('已提交')
    const d = await http.get('/tasks/' + row.task.id)
    Object.assign(row, d)
    pf.content = ''
    reload()
  } catch { /* */ } finally { busy.value = false }
}
</script>

<style scoped>
.t-title { margin: 0 0 8px; font-size: 18px; }
.t-tags { display: flex; gap: 6px; flex-wrap: wrap; }
.upd { background: var(--vm-primary-soft); }
.upd .el-slider { padding: 0 10px 14px; }
</style>
