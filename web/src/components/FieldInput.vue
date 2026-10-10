<template>
  <el-select v-if="['select', 'grid', 'user'].includes(f.type)" :model-value="modelValue" @update:model-value="set" :placeholder="'请选择' + f.label"
    filterable :clearable="!f.required" :allow-create="f.allowCreate" :disabled="dis" style="width: 100%">
    <el-option v-for="o in optionsOf(f, true)" :key="o.value" :label="o.label" :value="o.value" />
  </el-select>
  <RemoteSelect v-else-if="f.type === 'remote'" :model-value="modelValue" @update:model-value="set" :f="f" :form="form" :disabled="dis" />
  <el-date-picker v-else-if="f.type === 'date'" :model-value="modelValue" @update:model-value="set" type="date" value-format="YYYY-MM-DD"
    :placeholder="'选择' + f.label" :disabled="dis" style="width: 100%" />
  <el-time-picker v-else-if="f.type === 'time'" :model-value="modelValue" @update:model-value="set" format="HH:mm" value-format="HH:mm" :disabled="dis" style="width: 100%" />
  <el-input-number v-else-if="['number', 'money', 'progress'].includes(f.type)" :model-value="modelValue" @update:model-value="set"
    :min="f.min ?? (f.type === 'progress' ? 0 : undefined)" :max="f.max ?? (f.type === 'progress' ? 100 : undefined)" :step="f.step || 1"
    :precision="f.precision" controls-position="right" :disabled="dis" style="width: 100%" />
  <el-switch v-else-if="f.type === 'switch'" :model-value="!!modelValue" @update:model-value="set" :disabled="dis" />
  <el-input v-else-if="f.type === 'textarea'" :model-value="modelValue" @update:model-value="set" type="textarea" :rows="f.rows || 4"
    :maxlength="f.maxlength" show-word-limit :placeholder="f.placeholder || '请输入' + f.label" :disabled="dis" />
  <el-checkbox-group v-else-if="f.type === 'tags'" :model-value="tagArr" @update:model-value="v => set(v.join(','))" :disabled="dis">
    <el-checkbox v-for="o in f.options" :key="o" :value="o">{{ o }}</el-checkbox>
  </el-checkbox-group>
  <ImageUpload v-else-if="f.type === 'images'" :model-value="modelValue" @update:model-value="set" />
  <el-input v-else :model-value="modelValue" @update:model-value="set" :placeholder="f.placeholder || '请输入' + f.label" :maxlength="f.maxlength || 200" :disabled="dis" clearable />
</template>

<script setup>
import { computed } from 'vue'
import { optionsOf } from '@/utils/options'
import RemoteSelect from './RemoteSelect.vue'
import ImageUpload from './ImageUpload.vue'

const props = defineProps({ modelValue: null, f: Object, form: Object })
const emit = defineEmits(['update:modelValue'])
const set = v => { emit('update:modelValue', v); props.f.onChange && props.f.onChange(v, props.form) }
const dis = computed(() => (typeof props.f.disabled === 'function' ? props.f.disabled(props.form) : !!props.f.disabled))
const tagArr = computed(() => String(props.modelValue || '').split(',').filter(Boolean))
</script>
