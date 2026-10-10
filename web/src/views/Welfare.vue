<template>
  <CrudPage ref="page" :cfg="cfg">
    <template #top>
      <div class="w-cards">
        <div v-for="t in types" :key="t.name" class="w-card" :class="{ on: cur === t.name }" @click="pick(t.name)">
          <el-icon :style="{ color: t.color }"><component :is="t.icon" /></el-icon>
          <div><b>{{ stat(t.name).n }}</b><span>{{ t.name }}</span></div>
          <small v-if="stat(t.name).sum">月 {{ money(stat(t.name).sum) }} 元</small>
        </div>
      </div>
    </template>
  </CrudPage>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import http from '@/api/http'
import CrudPage from '@/components/CrudPage.vue'
import { M } from '@/config/modules'
import { money } from '@/utils/format'

const page = ref()
const cfg = { ...M.welfare, desc: '低保特困、优抚对象、残疾人、留守老人/儿童/妇女等保障对象动态台账' }
const rows = ref([])
const cur = ref('')
const types = [
  { name: '低保', icon: 'Wallet', color: '#f59e0b' }, { name: '特困供养', icon: 'House', color: '#ef4444' },
  { name: '优抚对象', icon: 'Medal', color: '#2563eb' }, { name: '残疾人', icon: 'FirstAidKit', color: '#8b5cf6' },
  { name: '留守老人', icon: 'UserFilled', color: '#0d9488' }, { name: '留守儿童', icon: 'Sunny', color: '#16a34a' },
  { name: '留守妇女', icon: 'Female', color: '#ec4899' }, { name: '临时救助', icon: 'Help', color: '#64748b' }
]
const stat = n => rows.value.find(r => r.type === n) || { n: 0, sum: 0 }
function pick(n) { cur.value = cur.value === n ? '' : n; page.value.setQuery({ type: cur.value }) }
onMounted(async () => { rows.value = await http.get('/welfare/stats') })
</script>

<style scoped>
.w-cards { display: grid; grid-template-columns: repeat(8, 1fr); gap: 10px; margin-bottom: 14px; }
.w-card { background: #fff; border: 1px solid var(--vm-border); border-radius: 12px; padding: 12px; cursor: pointer; display: flex; flex-direction: column; gap: 4px; box-shadow: var(--vm-shadow); transition: .15s; }
.w-card:hover, .w-card.on { border-color: var(--vm-primary); background: var(--vm-primary-soft); }
.w-card .el-icon { font-size: 20px; }
.w-card div { display: flex; align-items: baseline; gap: 6px; }
.w-card b { font-size: 22px; }
.w-card span { font-size: 13px; color: var(--vm-muted); }
.w-card small { font-size: 11px; color: var(--vm-muted); }
@media (max-width: 1200px) { .w-cards { grid-template-columns: repeat(4, 1fr); } }
@media (max-width: 767px) { .w-cards { grid-template-columns: repeat(2, 1fr); } }
</style>
