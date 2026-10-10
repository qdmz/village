<template>
  <CrudPage ref="page" :cfg="cfg" @saved="loadSum">
    <template #top>
      <div class="stat-grid fs">
        <StatCard label="收入合计（当前筛选）" :value="money(sum.income)" unit="元" icon="Top" color="#16a34a" />
        <StatCard label="支出合计（当前筛选）" :value="money(sum.expense)" unit="元" icon="Bottom" color="#f59e0b" />
        <StatCard label="结余" :value="money(sum.balance)" unit="元" icon="Wallet" color="#0d9488" />
      </div>
    </template>
  </CrudPage>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import http from '@/api/http'
import CrudPage from '@/components/CrudPage.vue'
import StatCard from '@/components/StatCard.vue'
import { M } from '@/config/modules'
import { store } from '@/store'
import { money } from '@/utils/format'

const page = ref()
const sum = ref({ income: 0, expense: 0, balance: 0 })
const cfg = { ...M.funds, desc: '村集体资金收支逐笔登记，支持按科目、日期筛选汇总，财务公开有据可依' }
async function loadSum() { sum.value = await http.get('/funds/summary') }
onMounted(async () => {
  loadSum()
  const d = await http.get('/projects', { params: { size: 200 } })
  store._projects = d.list.map(o => ({ label: o.name, value: o.id }))
})
</script>
<style scoped>.fs { margin-bottom: 14px; }</style>
