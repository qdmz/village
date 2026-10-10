<template><div ref="el" class="chart" :style="height ? { height } : null" /></template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({ option: Object, height: String, dark: Boolean })
const el = ref()
let chart = null
let ro = null

function render() {
  if (!chart || !props.option) return
  chart.setOption(props.option, true)
}
onMounted(async () => {
  await nextTick()
  chart = echarts.init(el.value, props.dark ? 'dark' : null, { renderer: 'canvas' })
  render()
  ro = new ResizeObserver(() => chart && chart.resize())
  ro.observe(el.value)
})
watch(() => props.option, render, { deep: true })
onBeforeUnmount(() => { ro && ro.disconnect(); chart && chart.dispose(); chart = null })
</script>
