<template>
  <div>
    <div class="page-head"><div><h2>村情与网格</h2><p class="muted">村 → 村民小组 → 网格 三级组织架构与网格员分工</p></div></div>
    <div class="grid-21">
      <div class="card" v-loading="!tree">
        <div class="card-title"><span>{{ v.name }} 概况</span><el-button v-if="isManager()" link type="primary" @click="editV">编辑村情</el-button></div>
        <div class="v-nums">
          <div><b>{{ v.area || '—' }}</b><span>面积(km²)</span></div>
          <div><b>{{ v.farmland || '—' }}</b><span>耕地(亩)</span></div>
          <div><b>{{ groups.length }}</b><span>村民小组</span></div>
          <div><b>{{ gridCount }}</b><span>网格</span></div>
        </div>
        <p class="pre intro">{{ v.intro }}</p>
        <p class="muted small">所属乡镇：{{ v.town || '—' }} · 书记：{{ v.secretary || '—' }} · 主任：{{ v.director || '—' }} · 电话：{{ v.phone || '—' }}</p>
      </div>
      <div class="card">
        <div class="card-title"><span>组织架构</span></div>
        <el-tree v-if="tree" :data="tree" node-key="key" default-expand-all :expand-on-click-node="false" class="org-tree">
          <template #default="{ data }">
            <div class="node" :class="data.type">
              <el-icon><OfficeBuilding v-if="data.type === 'village'" /><Collection v-else-if="data.type === 'group'" /><Grid v-else /></el-icon>
              <span class="nl">{{ data.label }}</span>
              <span v-if="data.type === 'grid'" class="nm">{{ data.worker || '未分配' }} · {{ data.households }}户{{ data.people }}人</span>
              <span v-if="data.type === 'group' && data.leader" class="nm">组长 {{ data.leader }}</span>
            </div>
          </template>
        </el-tree>
      </div>
    </div>
    <el-tabs class="card mt" v-model="tab">
      <el-tab-pane label="网格与网格员" name="grids"><CrudPage embedded :cfg="M.grids" @saved="reload" /></el-tab-pane>
      <el-tab-pane label="村民小组" name="groups"><CrudPage embedded :cfg="M.groups" @saved="reload" /></el-tab-pane>
    </el-tabs>

    <el-dialog v-model="vd" title="编辑村情信息" width="720px" :fullscreen="store.isMobile">
      <el-form :model="vf" label-width="90px" :label-position="store.isMobile ? 'top' : 'right'">
        <el-row :gutter="12">
          <el-col :span="store.isMobile ? 24 : 12"><el-form-item label="村名"><el-input v-model="vf.name" /></el-form-item></el-col>
          <el-col :span="store.isMobile ? 24 : 12"><el-form-item label="所属乡镇"><el-input v-model="vf.town" /></el-form-item></el-col>
          <el-col :span="store.isMobile ? 24 : 12"><el-form-item label="面积(km²)"><el-input-number v-model="vf.area" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="store.isMobile ? 24 : 12"><el-form-item label="耕地(亩)"><el-input-number v-model="vf.farmland" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="store.isMobile ? 24 : 12"><el-form-item label="支部书记"><el-input v-model="vf.secretary" /></el-form-item></el-col>
          <el-col :span="store.isMobile ? 24 : 12"><el-form-item label="村委主任"><el-input v-model="vf.director" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="办公地址"><el-input v-model="vf.address" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="联系电话"><el-input v-model="vf.phone" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="村情简介"><el-input v-model="vf.intro" type="textarea" :rows="5" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <template #footer><el-button @click="vd = false">取消</el-button><el-button type="primary" @click="saveV">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import CrudPage from '@/components/CrudPage.vue'
import { M } from '@/config/modules'
import { store, isManager } from '@/store'
import { loadGrids } from '@/utils/options'

const tree = ref(null)
const tab = ref('grids')
const v = computed(() => tree.value?.[0]?.data || {})
const groups = computed(() => (tree.value?.[0]?.children || []).filter(g => g.id))
const gridCount = computed(() => (tree.value?.[0]?.children || []).reduce((n, g) => n + g.children.length, 0))
const vd = ref(false)
const vf = ref({})

async function reload() {
  tree.value = await http.get('/org/tree')
  store._groups = groups.value.map(g => ({ label: g.label, value: g.id }))
  loadGrids(true)
}
function editV() { vf.value = { ...v.value }; vd.value = true }
async function saveV() {
  if (vf.value.id) await http.put('/villages/' + vf.value.id, vf.value)
  else await http.post('/villages', vf.value)
  ElMessage.success('已保存')
  vd.value = false
  reload()
}
onMounted(reload)
</script>

<style scoped>
.v-nums { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; margin-bottom: 12px; }
.v-nums div { background: var(--vm-primary-soft); border-radius: 12px; padding: 12px; display: flex; flex-direction: column; align-items: center; }
.v-nums b { font-size: 22px; color: var(--vm-primary-deep); }
.v-nums span { font-size: 12px; color: var(--vm-muted); }
.intro { margin: 8px 0; color: #3d4f4b; }
.small { font-size: 12px; }
.org-tree { --el-tree-node-content-height: 36px; }
.node { display: flex; align-items: center; gap: 6px; font-size: 14px; width: 100%; }
.node.village { font-weight: 700; }
.node.group { font-weight: 600; }
.nl { flex: 0 1 auto; }
.nm { margin-left: auto; color: var(--vm-muted); font-size: 12px; padding-right: 6px; }
@media (max-width: 767px) { .v-nums { grid-template-columns: repeat(2, 1fr); } .nm { display: none; } }
</style>
