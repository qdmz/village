<template>
  <div>
    <div class="page-head"><div><h2>乡村治理 · 治理积分</h2><p class="muted">六维积分：爱党爱国、邻里和睦、人居环境、勤劳致富、移风易俗、遵纪守法</p></div></div>
    <div class="dims">
      <div v-for="(dm, i) in dims" :key="dm" class="dim" :style="{ '--c': colors[i] }">
        <span class="ch">{{ chars[i] }}</span>
        <div><b>{{ dm }}</b><small>累计 {{ dimTotal(dm) }} 分</small></div>
      </div>
    </div>
    <div class="grid-21 mt">
      <div class="card">
        <div class="card-title"><span>积分排行榜</span><span class="muted" style="font-weight: 400; font-size: 12px">按家庭户累计积分</span></div>
        <el-table :data="rank.slice(0, 15)" size="small" stripe>
          <el-table-column label="名次" width="64"><template #default="{ $index }"><span class="rk" :class="'r' + ($index + 1)">{{ $index + 1 }}</span></template></el-table-column>
          <el-table-column prop="household_name" label="户主"><template #default="{ row }">{{ row.household_name }} 户</template></el-table-column>
          <el-table-column prop="grid_name" label="网格" min-width="140" v-if="!store.isMobile" />
          <el-table-column prop="times" label="记录" width="64" />
          <el-table-column prop="total" label="积分" width="80"><template #default="{ row }"><b class="pt">{{ row.total }}</b></template></el-table-column>
        </el-table>
      </div>
      <div class="card"><div class="card-title"><span>各维度积分</span></div><Chart :option="dimOpt" height="360px" /></div>
    </div>
    <div class="card mt"><CrudPage embedded :cfg="M.points" @saved="load" /></div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import http from '@/api/http'
import CrudPage from '@/components/CrudPage.vue'
import Chart from '@/components/Chart.vue'
import { M, OPT } from '@/config/modules'
import { store } from '@/store'
import { themeColor } from '@/utils/theme'

const dims = OPT.dimensions
const chars = ['爱', '和', '美', '勤', '德', '治']
const colors = ['#c8102e', '#16a34a', '#2563eb', '#d97706', '#7c3aed', '#0891b2']
const rank = ref([])
const dimRows = ref([])
const dimTotal = d => dimRows.value.find(x => x.dimension === d)?.total || 0
const dimOpt = computed(() => ({
  tooltip: {},
  radar: { indicator: dims.map(d => ({ name: d, max: Math.max(50, ...dimRows.value.map(x => x.total)) * 1.1 })), radius: '64%', splitArea: { areaStyle: { color: ['#fff', '#f6f9f8'] } } },
  series: [{ type: 'radar', areaStyle: { opacity: 0.25 }, data: [{ name: '积分', value: dims.map(dimTotal) }] }],
  color: [themeColor()]
}))
async function load() {
  const d = await http.get('/points/rank')
  rank.value = d.rank
  dimRows.value = d.dimensions
}
onMounted(load)
</script>

<style scoped>
.dims { display: grid; grid-template-columns: repeat(6, 1fr); gap: 10px; }
.dim { background: #fff; border-radius: 12px; padding: 12px; display: flex; gap: 10px; align-items: center; border: 1px solid var(--vm-border); box-shadow: var(--vm-shadow); }
.dim .ch { width: 40px; height: 40px; border-radius: 50%; display: grid; place-items: center; color: #fff; font-weight: 700; font-size: 18px; background: var(--c); flex-shrink: 0; }
.dim b { display: block; font-size: 14px; }
.dim small { color: var(--vm-muted); font-size: 12px; }
.rk { display: inline-grid; place-items: center; width: 24px; height: 24px; border-radius: 50%; background: #eef2f1; font-size: 12px; font-weight: 600; }
.rk.r1 { background: #fbbf24; color: #fff; } .rk.r2 { background: #94a3b8; color: #fff; } .rk.r3 { background: #d97706; color: #fff; }
.pt { color: var(--vm-primary); }
@media (max-width: 1100px) { .dims { grid-template-columns: repeat(3, 1fr); } }
@media (max-width: 767px) { .dims { grid-template-columns: repeat(2, 1fr); } }
</style>
