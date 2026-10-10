<template>
  <div class="pd">
    <header class="pd-nav"><div class="pd-wrap"><a @click="$router.push('/')"><el-icon><ArrowLeft /></el-icon> 返回{{ store.settings.village_name }}门户</a></div></header>
    <article class="pd-wrap pd-body" v-loading="!d">
      <template v-if="d">
        <h1>{{ d.title }}</h1>
        <p class="meta">{{ kind === 'notice' ? `${d.category} · ${d.publisher} · ${d.publish_date} · 浏览 ${d.views}` : d.date }}</p>
        <p v-if="d.summary" class="sum">{{ d.summary }}</p>
        <img v-if="d.image" :src="d.image" class="cover" alt="" />
        <div class="pre content">{{ d.content || d.summary }}</div>
      </template>
    </article>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import http from '@/api/http'
import { store } from '@/store'

const route = useRoute()
const kind = route.params.kind
const d = ref(null)
onMounted(async () => { d.value = await http.get(`/public/${kind === 'notice' ? 'notices' : 'items'}/${route.params.id}`) })
</script>

<style scoped>
.pd { min-height: 100vh; background: #f6f8f7; }
.pd-wrap { max-width: 860px; margin: 0 auto; padding: 0 20px; }
.pd-nav { background: #fff; border-bottom: 1px solid var(--vm-border); height: 56px; display: flex; align-items: center; }
.pd-nav a { cursor: pointer; display: inline-flex; align-items: center; gap: 4px; }
.pd-body { background: #fff; margin-top: 20px; border-radius: 16px; padding: 32px 36px; box-shadow: var(--vm-shadow); min-height: 300px; }
h1 { font-size: 26px; text-align: center; margin: 0 0 10px; line-height: 1.4; }
.meta { text-align: center; color: var(--vm-muted); font-size: 13px; padding-bottom: 16px; border-bottom: 1px solid var(--vm-border); }
.sum { background: var(--vm-primary-soft); padding: 12px 16px; border-radius: 10px; color: #3d4f4b; }
.cover { width: 100%; border-radius: 12px; margin: 10px 0; }
.content { font-size: 16px; line-height: 2; color: #26332f; }
@media (max-width: 767px) { .pd-body { padding: 20px 16px; margin-top: 10px; } h1 { font-size: 20px; } }
</style>
