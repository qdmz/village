<template>
  <div>
    <div class="page-head"><div><h2>党建引领 · 党组织架构</h2><p class="muted">支部—党小组—村级组织，党员联户包片</p></div></div>
    <div class="stat-grid">
      <StatCard label="党员总数" :value="stats.total" unit="名" icon="Flag" color="#c8102e" to="/party/members" />
      <StatCard label="流动党员" :value="stats.mobile" unit="名" icon="Position" color="#f59e0b" />
      <StatCard label="预备党员" :value="stats.probation" unit="名" icon="Avatar" color="#3b82f6" />
      <StatCard label="本年三会一课" :value="stats.meetings" unit="次" icon="Calendar" color="#16a34a" to="/meetings" />
    </div>
    <div class="card mt">
      <div class="card-title"><span>组织架构图</span></div>
      <div class="org-chart" v-if="tree.length">
        <div v-for="root in tree" :key="root.id" class="oc-root">
          <div class="oc-node root" @click="edit(root.data)"><b>{{ root.label }}</b><small>书记：{{ root.secretary }}</small><small v-if="root.members">{{ root.members }}</small></div>
          <div class="oc-children">
            <div v-for="c in root.children" :key="c.id" class="oc-node" :class="{ side: c.type !== '党小组' }" @click="edit(c.data)">
              <b>{{ c.label }}</b>
              <small>{{ c.type === '党小组' ? '组长' : '负责人' }}：{{ c.secretary || '—' }}{{ c.count ? ` · ${c.count}名党员` : '' }}</small>
              <small class="duty">{{ c.duty }}</small>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="card mt"><CrudPage embedded :cfg="cfg" @saved="load" /></div>
    <div class="card mt">
      <div class="card-title"><span>近期党建活动</span><a @click="$router.push('/meetings')">全部</a></div>
      <el-timeline>
        <el-timeline-item v-for="m in acts" :key="m.id" :timestamp="m.meet_date + ' · ' + m.type" placement="top" color="#c8102e">
          <b>{{ m.title }}</b><div class="muted">{{ m.content || m.topics }}</div>
        </el-timeline-item>
      </el-timeline>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import http from '@/api/http'
import CrudPage from '@/components/CrudPage.vue'
import StatCard from '@/components/StatCard.vue'
import { store } from '@/store'

const tree = ref([])
const acts = ref([])
const stats = reactive({ total: 0, mobile: 0, probation: 0, meetings: 0 })
const cfg = {
  title: '党组织', api: '/party/orgs', writeRoles: ['admin', 'cadre'], exportable: false,
  fields: [
    { prop: 'name', label: '组织名称', required: true, minWidth: 200 },
    { prop: 'type', label: '类型', type: 'select', options: ['党委', '党总支', '党支部', '党小组', '村级组织'], default: '党小组', width: 100, tag: true },
    { prop: 'parent_id', label: '上级组织', type: 'select', options: () => store._orgs || [], default: 0, fmt: r => (store._orgs || []).find(o => o.value === r.parent_id)?.label || '—', minWidth: 160 },
    { prop: 'secretary', label: '书记/组长', width: 100 },
    { prop: 'members', label: '委员/成员', list: false },
    { prop: 'duty', label: '职责分工', type: 'textarea', rows: 2, minWidth: 200 },
    { prop: 'sort', label: '排序', type: 'number', default: 0, list: false }
  ]
}
function edit() { /* 架构图仅展示，编辑在下方列表进行 */ }
async function load() {
  tree.value = await http.get('/party/orgs/tree')
  const all = await http.get('/party/orgs', { params: { size: 200 } })
  store._orgs = [{ label: '（无上级）', value: 0 }, ...all.list.map(o => ({ label: o.name, value: o.id }))]
}
onMounted(async () => {
  load()
  const s = await http.get('/screen')
  Object.assign(stats, { total: s.party.total, mobile: s.party.mobile, probation: s.party.probation, meetings: s.meetings_year })
  const m = await http.get('/meetings', { params: { size: 6 } })
  acts.value = m.list
})
</script>

<style scoped>
.org-chart { overflow-x: auto; padding: 6px 0 4px; }
.oc-root { display: flex; flex-direction: column; align-items: center; min-width: 760px; }
.oc-node { background: #fff; border: 1px solid #f3c6cd; border-radius: 12px; padding: 10px 14px; display: flex; flex-direction: column; gap: 2px; min-width: 150px; max-width: 220px; position: relative; }
.oc-node b { font-size: 14px; }
.oc-node small { color: var(--vm-muted); font-size: 12px; }
.oc-node .duty { color: #3d4f4b; }
.oc-node.root { background: linear-gradient(135deg, #c8102e, #e04848); color: #fff; border: none; align-items: center; max-width: 420px; }
.oc-node.root small { color: rgba(255,255,255,.85); }
.oc-node.side { border-color: var(--vm-border); background: #f8faf9; }
.oc-children { display: flex; gap: 12px; margin-top: 26px; position: relative; padding-top: 0; flex-wrap: nowrap; }
.oc-children::before { content: ''; position: absolute; top: -13px; left: 8%; right: 8%; border-top: 2px solid #f3c6cd; }
.oc-children > .oc-node::before { content: ''; position: absolute; top: -14px; left: 50%; height: 13px; border-left: 2px solid #f3c6cd; }
.oc-root > .oc-node.root::after { content: ''; position: absolute; bottom: -14px; left: 50%; height: 13px; border-left: 2px solid #f3c6cd; }
</style>
