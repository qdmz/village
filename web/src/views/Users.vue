<template>
  <CrudPage :cfg="cfg">
    <template #row-actions="{ row }">
      <el-button link type="warning" size="small" @click="reset(row)">重置密码</el-button>
    </template>
  </CrudPage>
</template>

<script setup>
import { ElMessageBox, ElMessage } from 'element-plus'
import http from '@/api/http'
import CrudPage from '@/components/CrudPage.vue'
import { M } from '@/config/modules'

const cfg = { ...M.users, desc: '管理员、村干部、网格员三类角色；网格员仅能查看和维护所负责网格的数据', actionWidth: 230 }
async function reset(row) {
  await ElMessageBox.confirm(`为「${row.real_name}」生成新的随机密码？`, '重置密码', { type: 'warning' })
  const d = await http.post(`/users/${row.id}/reset-password`, {})
  ElMessageBox.alert(`新密码：<b style="font-size:18px;letter-spacing:1px">${d.password}</b><br/><span style="color:#888">请妥善转告本人并提醒登录后修改。此密码仅显示一次。</span>`, '密码已重置', { dangerouslyUseHTMLString: true })
  ElMessage.success('已重置')
}
</script>
