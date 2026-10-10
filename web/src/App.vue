<template>
  <router-view />
</template>

<script setup>
import { onMounted } from 'vue'
import http from '@/api/http'
import { store } from '@/store'

onMounted(async () => {
  try {
    const s = await http.get('/public/settings', { silent: true })
    Object.assign(store.settings, s)
    document.title = s.system_title + ' · ' + s.village_name
  } catch { /* 使用默认值 */ }
})
</script>
