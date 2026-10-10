<template>
  <el-select :model-value="modelValue" @update:model-value="change" filterable remote :remote-method="search" :loading="loading"
    :placeholder="'输入姓名/户号搜索' + f.label" clearable :disabled="disabled" style="width: 100%">
    <el-option v-for="o in opts" :key="o.id" :label="label(o)" :value="o.id">
      <span>{{ label(o) }}</span><span class="opt-sub">{{ sub(o) }}</span>
    </el-option>
  </el-select>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import http from '@/api/http'

const props = defineProps({ modelValue: null, f: Object, form: Object, disabled: Boolean })
const emit = defineEmits(['update:modelValue'])
const opts = ref([])
const loading = ref(false)
const label = o => o[props.f.labelKey || 'name']
const sub = o => (props.f.subKey ? o[props.f.subKey] : '')

async function search(kw) {
  loading.value = true
  try {
    const d = await http.get(props.f.url, { params: { kw, size: 30 } })
    opts.value = d.list
  } finally { loading.value = false }
}
function change(v) {
  emit('update:modelValue', v || 0)
  const o = opts.value.find(x => x.id === v)
  if (props.f.nameProp && props.form) props.form[props.f.nameProp] = o ? label(o) : ''
  props.f.onPick && props.f.onPick(o, props.form)
}
onMounted(async () => {
  await search('')
  if (props.modelValue && !opts.value.find(o => o.id === props.modelValue) && props.form?.[props.f.nameProp]) {
    opts.value.unshift({ id: props.modelValue, [props.f.labelKey || 'name']: props.form[props.f.nameProp] })
  }
})
</script>
