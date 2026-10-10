<template>
  <div v-loading="!d">
    <div class="welcome card" v-if="d">
      <div>
        <h2>{{ greet }}，{{ store.user.real_name }}</h2>
        <p class="muted">{{ today }} · {{ roleName(store.user.role) }}{{ store.user.position ? ' · ' + store.user.position : '' }}<span v-if="isGrid()"> · 负责 {{ d.cards.grids }} 个网格</span></p>
      </div>
      <div class="quick">
        <el-button type="primary" @click="$router.push({ path: '/events', query: { new: 1 } })"><el-icon><Promotion /></el-icon><span>上报事件</span></el-button>
        <el-button @click="$router.push({ path: '/patrols', query: { new: 1 } })"><el-icon><Location /></el-icon><span>记录巡查</span></el-button>
        <el-button @click="$router.push({ path: '/worklogs' })"><el-icon><EditPen /></el-icon><span>写日志</span></el-button>
        <el-button v-if="!store.isMobile" @click="$router.push('/screen')"><el-icon><DataBoard /></el-icon><span>驾驶舱</span></el-button>
      </div>
    </div>

    <template v-if="d">
      <div class="stat-grid mt">
        <StatCard label="户数" :value="d.cards.households" unit="户" icon="House" to="/households">共 {{ d.cards.grids }} 个网格</StatCard>
        <StatCard label="人口" :value="d.cards.villagers" unit="人" icon="User" color="#3b82f6" to="/villagers">党员 {{ d.cards.party_members }} 名</StatCard>
        <StatCard label="待办结事件" :value="d.cards.events_open" unit="件" icon="Warning" color="#f59e0b" to="/events">累计 {{ d.cards.events_total }} 件</StatCard>
        <StatCard label="任务完成率" :value="d.cards.task_rate.toFixed(0)" unit="%" icon="Finished" color="#84cc16" to="/tasks">逾期 <b>{{ d.cards.tasks_overdue }}</b> 项 / 共 {{ d.cards.tasks_total }} 项</StatCard>
        <StatCard label="本月巡查" :value="d.cards.patrols_month" unit="次" icon="Location" color="#06b6d4" to="/patrols" />
        <StatCard v-if="d.funds" label="集体资金结余" :value="wan(d.funds.balance)" unit="万元" icon="Wallet" color="#8b5cf6" to="/funds">项目 {{ d.cards.projects }} 个</StatCard>
      </div>

      <div class="grid-21 mt">
        <div class="card">
          <div class="card-title"><span>我的待办</span><el-tag round>{{ d.todos.length }}</el-tag></div>
          <ul class="list-plain" v-if="d.todos.length">
            <li v-for="t in d.todos" :key="t.type + t.id" @click="$router.push(t.type === 'task' ? { path: '/tasks', query: { id: t.id } } : '/events/' + t.id)">
              <el-tag size="small" :type="t.type === 'task' ? 'primary' : 'warning'" effect="dark" round>{{ t.type === 'task' ? '任务' : '事件' }}</el-tag>
              <span class="t">{{ t.title }}</span>
              <el-tag v-if="t.overdue" size="small" type="danger" effect="plain">已逾期</el-tag>
              <el-tag v-else size="small" :type="statusType(t.status)" effect="plain">{{ t.status }}</el-tag>
              <span class="d">{{ t.deadline ? '截止 ' + t.deadline : fmtTime(t.created_at) }}</span>
            </li>
          </ul>
          <el-empty v-else description="暂无待办，干得漂亮！" :image-size="80" />
        </div>
        <div class="card">
          <div class="card-title"><span>最新公告</span><a @click="$router.push('/notices')">更多</a></div>
          <ul class="list-plain">
            <li v-for="n in d.notices" :key="n.id" @click="openNotice(n)">
              <el-tag size="small" effect="plain" :type="n.top ? 'danger' : 'info'">{{ n.top ? '置顶' : n.category.slice(0, 2) }}</el-tag>
              <span class="t">{{ n.title }}</span><span class="d">{{ n.publish_date.slice(5) }}</span>
            </li>
          </ul>
        </div>
      </div>

      <div class="grid-2 mt">
        <div class="card"><div class="card-title"><span>人口结构 · 年龄段</span></div><Chart :option="ageOpt" /></div>
        <div class="card"><div class="card-title"><span>特殊人群分布</span></div><Chart :option="tagOpt" /></div>
        <div class="card"><div class="card-title"><span>事件处理状态</span></div><Chart :option="evOpt" /></div>
        <div class="card"><div class="card-title"><span>任务完成情况</span></div><Chart :option="taskOpt" /></div>
      </div>
      <div class="card mt" v-if="d.funds">
        <div class="card-title"><span>近 12 个月资金收支（万元）</span><span class="muted" style="font-weight: 400; font-size: 13px">累计收入 {{ wan(d.funds.total_income) }} · 支出 {{ wan(d.funds.total_expense) }}</span></div>
        <Chart :option="fundOpt" height="300px" />
      </div>
    </template>

    <el-dialog v-model="nv" :title="notice?.title" width="680px" :fullscreen="store.isMobile">
      <p class="muted" v-if="notice">{{ notice.category }} · {{ notice.publisher }} · {{ notice.publish_date }}</p>
      <div class="pre">{{ notice?.content }}</div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import dayjs from 'dayjs'
import http from '@/api/http'
import { store, roleName, isGrid } from '@/store'
import { wan, fmtTime, statusType } from '@/utils/format'
import { pie, bar } from '@/utils/chart'
import StatCard from '@/components/StatCard.vue'
import Chart from '@/components/Chart.vue'

const d = ref(null)
const nv = ref(false)
const notice = ref(null)
const today = dayjs().format('YYYY年M月D日 dddd')
const greet = computed(() => { const h = new Date().getHours(); return h < 11 ? '早上好' : h < 13 ? '中午好' : h < 18 ? '下午好' : '晚上好' })

const ageOpt = computed(() => bar(d.value.population.age.map(x => x.name), [{ name: '人数', data: d.value.population.age.map(x => x.value) }]))
const tagOpt = computed(() => pie(d.value.population.tags, { rose: true }))
const evOpt = computed(() => pie(d.value.events.status))
const taskOpt = computed(() => ({
  ...pie(d.value.tasks.status),
  title: { text: d.value.tasks.rate.toFixed(0) + '%', subtext: '完成率', left: 'center', top: '34%', textStyle: { fontSize: 24 }, subtextStyle: { fontSize: 12 } }
}))
const fundOpt = computed(() => bar(d.value.funds.months.map(m => m.slice(2)), [
  { name: '收入', data: d.value.funds.income.map(v => +v.toFixed(2)) },
  { name: '支出', data: d.value.funds.expense.map(v => +v.toFixed(2)), itemStyle: { color: '#f59e0b', borderRadius: [6, 6, 0, 0] } }
]))

async function openNotice(n) { notice.value = await http.get('/notices/' + n.id); nv.value = true }
onMounted(async () => { d.value = await http.get('/dashboard') })
</script>

<style scoped>
.welcome { display: flex; justify-content: space-between; align-items: center; gap: 12px; flex-wrap: wrap;
  background: linear-gradient(120deg, var(--vm-primary-soft), #fff 60%); }
.welcome h2 { margin: 0 0 4px; font-size: 20px; }
.welcome p { margin: 0; font-size: 13px; }
.quick { display: flex; gap: 8px; flex-wrap: wrap; }
.quick .el-button + .el-button { margin-left: 0; }
.quick .el-icon + span { margin-left: 4px; }
@media (max-width: 767px) { .quick { width: 100%; display: grid; grid-template-columns: repeat(3, 1fr); } }
</style>
