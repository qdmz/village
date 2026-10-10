<template>
  <div v-loading="!d">
    <div class="page-head"><el-button text @click="$router.back()"><el-icon><ArrowLeft /></el-icon>返回户籍台账</el-button></div>
    <template v-if="d">
      <div class="card hh-head">
        <div class="hh-avatar">{{ h.head_name.slice(0, 1) }}</div>
        <div class="hh-info">
          <h2>{{ h.head_name }} 户 <el-tag :type="statusType(h.type)" round>{{ h.type }}</el-tag></h2>
          <p class="muted">户号 {{ h.household_no }} · {{ gridName(h.grid_id) }} · {{ h.address }}</p>
        </div>
        <div class="hh-nums">
          <div><b>{{ d.members.length }}</b><span>家庭人口</span></div>
          <div><b>{{ money(h.income) }}</b><span>年收入(元)</span></div>
          <div><b>{{ d.visits.length }}</b><span>走访次数</span></div>
        </div>
      </div>
      <el-tabs v-model="tab" class="card mt">
        <el-tab-pane label="家庭成员" name="m">
          <CrudPage embedded :cfg="memberCfg" :fixed-query="{ household_id: h.id }" @saved="load" />
        </el-tab-pane>
        <el-tab-pane label="走访记录" name="v">
          <el-button type="primary" @click="$router.push({ path: '/worklogs', query: { new: 1 } })" class="mb">新增走访</el-button>
          <el-timeline v-if="d.visits.length">
            <el-timeline-item v-for="v in d.visits" :key="v.id" :timestamp="v.log_date + ' · ' + v.user_name" placement="top">
              <b>{{ v.title || v.type }}</b>
              <div class="pre">{{ v.content }}</div>
              <div v-if="v.demand" class="muted">诉求：{{ v.demand }}</div>
            </el-timeline-item>
          </el-timeline>
          <el-empty v-else description="暂无走访记录" />
        </el-tab-pane>
        <el-tab-pane label="基本信息" name="i">
          <el-descriptions :column="store.isMobile ? 1 : 2" border>
            <el-descriptions-item label="联系电话">{{ h.phone || '—' }}</el-descriptions-item>
            <el-descriptions-item label="住房情况">{{ h.housing || '—' }}</el-descriptions-item>
            <el-descriptions-item label="备注" :span="2">{{ h.remark || '—' }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import http from '@/api/http'
import CrudPage from '@/components/CrudPage.vue'
import { M } from '@/config/modules'
import { store } from '@/store'
import { money, statusType } from '@/utils/format'
import { gridName, loadGrids } from '@/utils/options'

const route = useRoute()
const d = ref(null)
const tab = ref('m')
const h = computed(() => d.value.household)
const memberCfg = {
  ...M.villagers, importable: false, filters: [], exportable: false,
  fields: M.villagers.fields.map(f => (f.prop === 'household_id' ? { ...f, form: false } : f)).filter(f => f.prop !== 'grid_id')
}
async function load() { d.value = await http.get('/households/' + route.params.id) }
onMounted(() => { loadGrids(); load() })
</script>

<style scoped>
.hh-head { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
.hh-avatar { width: 60px; height: 60px; border-radius: 16px; display: grid; place-items: center; font-size: 26px; font-weight: 700; color: #fff; background: linear-gradient(135deg, var(--vm-primary), #84cc16); }
.hh-info { flex: 1; min-width: 220px; }
.hh-info h2 { margin: 0 0 6px; font-size: 20px; display: flex; align-items: center; gap: 8px; }
.hh-info p { margin: 0; }
.hh-nums { display: flex; gap: 26px; }
.hh-nums div { display: flex; flex-direction: column; align-items: center; }
.hh-nums b { font-size: 20px; }
.hh-nums span { font-size: 12px; color: var(--vm-muted); }
.mb { margin-bottom: 12px; }
</style>
