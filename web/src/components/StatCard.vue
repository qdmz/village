<template>
  <div class="stat" :class="{ link: !!to }" @click="to && $router.push(to)">
    <div class="s-ico" :style="{ background: color + '1a', color }"><el-icon><component :is="icon" /></el-icon></div>
    <div class="s-label">{{ label }}</div>
    <div class="s-val">{{ value }}<small v-if="unit">{{ unit }}</small></div>
    <div class="s-sub" v-if="$slots.default"><slot /></div>
  </div>
</template>
<script setup>
defineProps({ label: String, value: [String, Number], unit: String, icon: { type: String, default: 'DataLine' }, color: { type: String, default: '#0d9488' }, to: String })
</script>
