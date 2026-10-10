<template>
  <el-tag v-if="f.tag && text" :type="statusType(text)" effect="light" size="small" round>{{ text }}</el-tag>
  <span v-else-if="f.type === 'tags'" class="tag-list">
    <el-tag v-for="t in tagList" :key="t" size="small" :type="tagColor(t)" effect="plain" round>{{ t }}</el-tag>
    <span v-if="!tagList.length" class="muted">—</span>
  </span>
  <el-progress v-else-if="f.type === 'progress'" :percentage="Number(row[f.prop] || 0)" :stroke-width="8" :status="row[f.prop] >= 100 ? 'success' : ''" />
  <span v-else-if="f.type === 'images'" class="img-cell">
    <el-image v-for="(u, i) in imgs" :key="i" :src="u" :preview-src-list="imgs" :initial-index="i" fit="cover" preview-teleported class="thumb" />
    <span v-if="!imgs.length" class="muted">无</span>
  </span>
  <span v-else-if="f.type === 'textarea' && full" class="pre">{{ text }}</span>
  <span v-else :class="{ muted: !text && text !== 0 }">{{ text === '' || text === null || text === undefined ? '—' : text }}</span>
</template>

<script setup>
import { computed } from 'vue'
import { statusType, parseImages } from '@/utils/format'
import { cellText } from '@/utils/options'

const props = defineProps({ f: Object, row: Object, full: Boolean })
const text = computed(() => cellText(props.f, props.row))
const tagList = computed(() => String(props.row[props.f.prop] || '').split(',').filter(Boolean))
const imgs = computed(() => parseImages(props.row[props.f.prop]))
const tagColor = t => ({ 低保: 'warning', 五保: 'danger', 残疾: 'danger', 脱贫监测户: 'warning', 留守老人: 'primary', 留守儿童: 'success', 退役军人: '' }[t] ?? 'info')
</script>
