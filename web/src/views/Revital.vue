<template>
  <div v-loading="!fs">
    <div class="page-head"><div><h2>乡村振兴 · 概览</h2><p class="muted">产业项目、集体经济收入、三资（资金·资产·资源）与致富带头人</p></div></div>
    <template v-if="fs">
      <div class="stat-grid">
        <StatCard label="本年集体收入" :value="wan(yearIncome)" unit="万元" icon="Coin" color="#16a34a" to="/funds" />
        <StatCard label="资金结余" :value="wan(sum.balance)" unit="万元" icon="Wallet" color="#0d9488" to="/funds" />
        <StatCard label="集体资产总值" :value="fs.asset_value.toFixed(1)" unit="万元" icon="OfficeBuilding" color="#8b5cf6" to="/assets" />
        <StatCard label="产业项目" :value="projects.length" unit="个" icon="TrendCharts" color="#f59e0b" to="/projects">在建 {{ projects.filter(p => p.status === '建设中').length }} 个</StatCard>
        <StatCard label="致富带头人" :value="leaders.length" unit="人" icon="Trophy" color="#ef4444" to="/leaders">带动 {{ leaders.reduce((n, l) => n + l.drive, 0) }} 户</StatCard>
      </div>
      <div class="grid-2 mt">
        <div class="card"><div class="card-title"><span>集体经济收支（按年度，万元）</span></div><Chart :option="yearOpt" /></div>
        <div class="card"><div class="card-title"><span>本年收入构成</span></div><Chart :option="catOpt" /></div>
      </div>
      <div class="card mt">
        <div class="card-title"><span>产业项目进度</span><a @click="$router.push('/projects')">管理项目</a></div>
        <div class="proj-list">
          <div v-for="p in projects" :key="p.id" class="proj">
            <div class="pj-top"><b>{{ p.name }}</b><el-tag size="small" :type="statusType(p.status)" round>{{ p.status }}</el-tag></div>
            <el-progress :percentage="p.progress" :stroke-width="10" :status="p.progress >= 100 ? 'success' : ''" />
            <div class="muted pj-sub">{{ p.type }} · 负责人 {{ p.leader }} · 预算 {{ p.budget }} 万 / 已投 {{ p.invested }} 万</div>
            <div class="pj-benefit">{{ p.benefit }}</div>
          </div>
        </div>
      </div>
      <div class="card mt">
        <div class="card-title"><span>致富带头人</span><a @click="$router.push('/leaders')">全部</a></div>
        <div class="leaders">
          <div v-for="l in leaders" :key="l.id" class="leader">
            <div class="la">{{ l.name.slice(0, 1) }}</div>
            <div class="lt"><b>{{ l.name }} <el-tag v-if="l.is_party" size="small" type="danger" effect="plain">党员</el-tag></b><small>{{ l.entity }}</small><p>{{ l.story }}</p>
              <span class="muted">带动 {{ l.drive }} 户 · 年产值 {{ l.income }} 万元</span></div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import http from '@/api/http'
import StatCard from '@/components/StatCard.vue'
import Chart from '@/components/Chart.vue'
import { wan, statusType } from '@/utils/format'
import { pie, bar } from '@/utils/chart'

const fs = ref(null)
const sum = ref({})
const projects = ref([])
const leaders = ref([])
const yearIncome = computed(() => fs.value.years.find(y => y.year === String(new Date().getFullYear()))?.income || 0)
const yearOpt = computed(() => bar(fs.value.years.map(y => y.year), [
  { name: '收入', data: fs.value.years.map(y => +(y.income / 10000).toFixed(2)) },
  { name: '支出', data: fs.value.years.map(y => +(y.expense / 10000).toFixed(2)), itemStyle: { color: '#f59e0b', borderRadius: [6, 6, 0, 0] } }
]))
const catOpt = computed(() => pie(fs.value.categories.map(c => ({ name: c.name, value: +(c.value / 10000).toFixed(2) }))))
onMounted(async () => {
  const [a, b, c, d] = await Promise.all([http.get('/funds/stats'), http.get('/funds/summary'), http.get('/projects', { params: { size: 50 } }), http.get('/leaders', { params: { size: 50 } })])
  fs.value = a; sum.value = b; projects.value = c.list; leaders.value = d.list
})
</script>

<style scoped>
.proj-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 12px; }
.proj { border: 1px solid var(--vm-border); border-radius: 12px; padding: 14px; background: #fbfdfc; }
.pj-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; gap: 8px; }
.pj-sub { font-size: 12px; margin-top: 6px; }
.pj-benefit { font-size: 13px; margin-top: 6px; color: #3d4f4b; }
.leaders { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 12px; }
.leader { display: flex; gap: 12px; border: 1px solid var(--vm-border); border-radius: 12px; padding: 14px; }
.la { width: 46px; height: 46px; border-radius: 50%; flex-shrink: 0; display: grid; place-items: center; color: #fff; font-weight: 700; font-size: 18px; background: linear-gradient(135deg, #f59e0b, #ef4444); }
.lt b { display: flex; align-items: center; gap: 6px; }
.lt small { color: var(--vm-primary); }
.lt p { margin: 6px 0; font-size: 13px; color: #3d4f4b; line-height: 1.6; }
.lt span { font-size: 12px; }
</style>
