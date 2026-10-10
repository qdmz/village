<template>
  <CrudPage ref="page" :cfg="cfg">
    <template #head>
      <el-radio-group v-model="quick" size="small" @change="v => page.setQuery({ status: v })">
        <el-radio-button value="">全部</el-radio-button>
        <el-radio-button value="上报">待受理</el-radio-button>
        <el-radio-button value="处理中">处理中</el-radio-button>
        <el-radio-button value="办结">已办结</el-radio-button>
      </el-radio-group>
    </template>
  </CrudPage>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import CrudPage from '@/components/CrudPage.vue'
import { M } from '@/config/modules'

const router = useRouter()
const page = ref()
const quick = ref('')
const cfg = { ...M.events, desc: '网格事件“上报 → 受理 → 处理中 → 办结 / 退回”全流程闭环，每一步留痕可追溯', onDetail: r => router.push('/events/' + r.id) }
</script>
