<template>
  <div class="screen">
    <header class="sc-head">
      <div class="sc-side"><a @click="$router.push('/dashboard')"><el-icon><Back /></el-icon> 返回工作台</a></div>
      <h1>{{ store.settings.village_name }} · 乡村振兴数据驾驶舱</h1>
      <div class="sc-side right">{{ now }}</div>
    </header>
    <div class="sc-body" v-if="s && d">
      <section class="col">
        <div class="panel"><h3>人口结构</h3>
          <div class="kv"><div><b>{{ d.cards.households }}</b><span>户数</span></div><div><b>{{ d.cards.villagers }}</b><span>人口</span></div><div><b>{{ s.party.total }}</b><span>党员</span></div></div>
          <Chart :option="ageOpt" height="190px" dark />
        </div>
        <div class="panel"><h3>就业与外出务工</h3><Chart :option="empOpt" height="200px" dark /></div>
        <div class="panel"><h3>民生保障对象</h3><Chart :option="welOpt" height="210px" dark /></div>
      </section>

      <section class="col mid">
        <div class="big-nums">
          <div v-for="n in bigNums" :key="n.l" class="bn"><b>{{ n.v }}</b><span>{{ n.l }}</span></div>
        </div>
        <div class="panel"><h3>网格分布（户数 / 人口 / 事件）</h3><Chart :option="gridOpt" height="260px" dark /></div>
        <div class="panel"><h3>近 30 天事件上报趋势</h3><Chart :option="trendOpt" height="200px" dark /></div>
        <div class="panel"><h3>最新事件</h3>
          <div class="ticker">
            <div class="tk-in" :style="{ animationDuration: s.latest_events.length * 3 + 's' }">
              <div v-for="(e, i) in [...s.latest_events, ...s.latest_events]" :key="i" class="tk"><em :class="'st-' + e.status">{{ e.status }}</em><span>{{ e.title }}</span><small>{{ e.category }} · {{ fmtTime(e.created_at) }}</small></div>
            </div>
          </div>
        </div>
      </section>

      <section class="col">
        <div class="panel"><h3>事件处理状态</h3><Chart :option="evOpt" height="200px" dark /></div>
        <div class="panel"><h3>产业项目进度</h3>
          <div v-for="p in s.projects" :key="p.name" class="pj"><div class="pj-t"><span>{{ p.name }}</span><b>{{ p.progress }}%</b></div><div class="bar"><i :style="{ width: p.progress + '%' }" /></div></div>
        </div>
        <div class="panel"><h3>治理积分榜</h3>
          <ol class="rk"><li v-for="(r, i) in s.rank" :key="r.household_id"><span :class="'n' + (i + 1)">{{ i + 1 }}</span>{{ r.household_name }} 户<em>{{ r.total }}</em></li></ol>
        </div>
      </section>
    </div>
    <div v-else class="loading" v-loading="true" element-loading-background="transparent" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import dayjs from 'dayjs'
import http from '@/api/http'
import { store } from '@/store'
import { wan, fmtTime } from '@/utils/format'
import Chart from '@/components/Chart.vue'

const s = ref(null)
const d = ref(null)
const now = ref('')
let timer
const C = ['#2dd4bf', '#a3e635', '#fbbf24', '#60a5fa', '#f87171', '#c084fc', '#22d3ee', '#f472b6']
const base = { backgroundColor: 'transparent', color: C, textStyle: { color: '#cfe9e4' } }
const axis = { axisLine: { lineStyle: { color: '#2b4b47' } }, splitLine: { lineStyle: { color: '#1d3a36', type: 'dashed' } }, axisLabel: { color: '#9cc7bf', fontSize: 11 } }
const pieOpt = (data, rose) => ({ ...base, tooltip: { trigger: 'item' }, legend: { bottom: 0, textStyle: { color: '#9cc7bf', fontSize: 11 }, itemWidth: 8, itemHeight: 8 },
  series: [{ type: 'pie', radius: rose ? ['15%', '65%'] : ['42%', '65%'], roseType: rose ? 'area' : undefined, center: ['50%', '42%'], label: { color: '#cfe9e4', fontSize: 11 }, itemStyle: { borderColor: '#071a18', borderWidth: 2 }, data: data.filter(x => x.value > 0) }] })

const bigNums = computed(() => [
  { v: d.value.cards.grids, l: '治理网格' }, { v: d.value.cards.events_open, l: '在办事件' },
  { v: d.value.cards.task_rate.toFixed(0) + '%', l: '任务完成率' }, { v: d.value.funds ? wan(d.value.funds.balance) : '—', l: '集体资金结余(万)' },
  { v: s.value.meetings_year, l: '本年会议活动' }, { v: s.value.leaders, l: '致富带头人' }
])
const ageOpt = computed(() => ({ ...base, tooltip: {}, grid: { left: 6, right: 6, top: 10, bottom: 0, containLabel: true },
  xAxis: { type: 'category', data: d.value.population.age.map(x => x.name), ...axis }, yAxis: { type: 'value', ...axis },
  series: [{ type: 'bar', barWidth: 16, data: d.value.population.age.map(x => x.value), itemStyle: { borderRadius: [4, 4, 0, 0], color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: '#2dd4bf' }, { offset: 1, color: '#0f766e' }] } } }] }))
const empOpt = computed(() => pieOpt(d.value.population.employment, true))
const welOpt = computed(() => pieOpt(s.value.welfare))
const evOpt = computed(() => pieOpt(d.value.events.status))
const gridOpt = computed(() => ({ ...base, tooltip: { trigger: 'axis' }, legend: { top: 0, textStyle: { color: '#9cc7bf' } }, grid: { left: 6, right: 6, top: 30, bottom: 0, containLabel: true },
  xAxis: { type: 'category', data: s.value.grids.map(g => g.name.replace(/（.*）/, '')), ...axis }, yAxis: { type: 'value', ...axis },
  series: [{ name: '户数', type: 'bar', data: s.value.grids.map(g => g.households), barWidth: 12 }, { name: '人口', type: 'bar', data: s.value.grids.map(g => g.people), barWidth: 12 },
    { name: '事件', type: 'line', smooth: true, data: s.value.grids.map(g => g.events) }] }))
const trendOpt = computed(() => ({ ...base, tooltip: { trigger: 'axis' }, grid: { left: 6, right: 10, top: 14, bottom: 0, containLabel: true },
  xAxis: { type: 'category', data: s.value.event_trend.days, ...axis, boundaryGap: false }, yAxis: { type: 'value', minInterval: 1, ...axis },
  series: [{ type: 'line', smooth: true, symbol: 'none', data: s.value.event_trend.counts, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(45,212,191,.5)' }, { offset: 1, color: 'rgba(45,212,191,0)' }] } }, lineStyle: { width: 2 } }] }))

async function load() {
  const [a, b] = await Promise.all([http.get('/screen'), http.get('/dashboard')])
  s.value = a; d.value = b
}
onMounted(() => {
  load()
  const tick = () => (now.value = dayjs().format('YYYY年MM月DD日 dddd HH:mm:ss'))
  tick()
  timer = setInterval(tick, 1000)
})
onBeforeUnmount(() => clearInterval(timer))
</script>

<style scoped>
.screen { min-height: 100vh; color: #e6fffa; background: radial-gradient(1200px 600px at 50% -10%, #0f3f3a 0%, #071a18 60%), #071a18; padding: 0 16px 16px; }
.sc-head { height: 72px; display: flex; align-items: center; justify-content: space-between; position: relative;
  background: linear-gradient(90deg, transparent, rgba(45,212,191,.12), transparent); border-bottom: 1px solid rgba(45,212,191,.25); }
.sc-head h1 { margin: 0; font-size: 28px; letter-spacing: 4px; background: linear-gradient(180deg, #fff, #7ee8d6); -webkit-background-clip: text; background-clip: text; color: transparent; text-align: center; }
.sc-side { width: 280px; color: #9cc7bf; font-size: 14px; }
.sc-side.right { text-align: right; font-variant-numeric: tabular-nums; }
.sc-side a { color: #9cc7bf; cursor: pointer; display: inline-flex; align-items: center; gap: 4px; }
.sc-body { display: grid; grid-template-columns: 1fr 1.6fr 1fr; gap: 14px; margin-top: 14px; }
.col { display: flex; flex-direction: column; gap: 14px; min-width: 0; }
.panel { background: linear-gradient(180deg, rgba(20,60,56,.55), rgba(10,34,31,.55)); border: 1px solid rgba(45,212,191,.18); border-radius: 10px; padding: 12px 14px; position: relative; }
.panel::before { content: ''; position: absolute; left: -1px; top: -1px; width: 14px; height: 14px; border-left: 2px solid #2dd4bf; border-top: 2px solid #2dd4bf; border-radius: 10px 0 0 0; }
.panel h3 { margin: 0 0 8px; font-size: 15px; font-weight: 600; color: #b8f3e8; padding-left: 8px; border-left: 3px solid #2dd4bf; line-height: 1; }
.kv { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; margin-bottom: 4px; }
.kv div { text-align: center; background: rgba(45,212,191,.08); border-radius: 8px; padding: 6px 0; }
.kv b { display: block; font-size: 22px; color: #5eead4; font-variant-numeric: tabular-nums; }
.kv span { font-size: 12px; color: #9cc7bf; }
.big-nums { display: grid; grid-template-columns: repeat(6, 1fr); gap: 10px; }
.bn { text-align: center; padding: 14px 4px; border-radius: 10px; background: linear-gradient(180deg, rgba(45,212,191,.16), rgba(45,212,191,.03)); border: 1px solid rgba(45,212,191,.25); }
.bn b { display: block; font-size: 26px; font-weight: 700; color: #fde68a; font-variant-numeric: tabular-nums; text-shadow: 0 0 12px rgba(253,230,138,.35); }
.bn span { font-size: 12px; color: #9cc7bf; }
.pj { margin-bottom: 10px; }
.pj-t { display: flex; justify-content: space-between; font-size: 13px; margin-bottom: 4px; color: #cfe9e4; }
.pj-t b { color: #a3e635; }
.bar { height: 8px; border-radius: 4px; background: rgba(255,255,255,.08); overflow: hidden; }
.bar i { display: block; height: 100%; border-radius: 4px; background: linear-gradient(90deg, #0d9488, #a3e635); }
.rk { list-style: none; margin: 0; padding: 0; }
.rk li { display: flex; align-items: center; gap: 10px; padding: 6px 0; font-size: 13px; border-bottom: 1px dashed rgba(45,212,191,.15); }
.rk li span { width: 20px; height: 20px; border-radius: 4px; display: grid; place-items: center; background: rgba(255,255,255,.1); font-size: 12px; }
.rk .n1 { background: #f59e0b; } .rk .n2 { background: #64748b; } .rk .n3 { background: #b45309; }
.rk em { margin-left: auto; font-style: normal; color: #fde68a; font-weight: 700; }
.ticker { height: 150px; overflow: hidden; }
.tk-in { animation: scroll linear infinite; }
.tk { display: flex; align-items: center; gap: 10px; padding: 7px 0; font-size: 13px; border-bottom: 1px dashed rgba(45,212,191,.12); }
.tk span { flex: 1; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
.tk small { color: #7fa8a1; flex-shrink: 0; }
.tk em { font-style: normal; font-size: 11px; padding: 1px 6px; border-radius: 4px; background: #f59e0b33; color: #fbbf24; flex-shrink: 0; }
.tk em.st-办结 { background: #16a34a33; color: #86efac; } .tk em.st-处理中 { background: #2563eb33; color: #93c5fd; } .tk em.st-退回 { background: #dc262633; color: #fca5a5; }
@keyframes scroll { from { transform: translateY(0); } to { transform: translateY(-50%); } }
.loading { height: 70vh; }
@media (max-width: 1100px) {
  .sc-body { grid-template-columns: 1fr; }
  .sc-head h1 { font-size: 18px; letter-spacing: 1px; }
  .sc-side { width: auto; font-size: 12px; }
  .sc-side.right { display: none; }
  .big-nums { grid-template-columns: repeat(3, 1fr); }
  .mid { order: -1; }
}
</style>
