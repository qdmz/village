<template>
  <CrudPage :cfg="cfg" :key="mkey" />
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import CrudPage from '@/components/CrudPage.vue'
import { M } from '@/config/modules'
import { store } from '@/store'
import http from '@/api/http'

const props = defineProps({ mkey: String })
const router = useRouter()
const cfg = computed(() => {
  const c = { ...M[props.mkey] }
  if (props.mkey === 'households') c.onDetail = r => router.push('/households/' + r.id)
  return c
})

onMounted(async () => {
  // 部分模块需要的下拉选项
  if (props.mkey === 'partyMembers') {
    const d = await http.get('/party/orgs', { params: { size: 200 } })
    store._orgs = d.list.map(o => ({ label: o.name, value: o.id }))
  }
  if (props.mkey === 'funds') {
    const d = await http.get('/projects', { params: { size: 200 } })
    store._projects = d.list.map(o => ({ label: o.name, value: o.id }))
  }
})
</script>
