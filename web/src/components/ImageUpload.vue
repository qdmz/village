<template>
  <div class="img-up">
    <div v-for="(u, i) in list" :key="u" class="img-item">
      <el-image :src="u" :preview-src-list="list" :initial-index="i" fit="cover" preview-teleported />
      <span class="del" @click="del(i)"><el-icon><Close /></el-icon></span>
    </div>
    <label v-if="list.length < max" class="img-add" :class="{ busy }">
      <input type="file" accept="image/*" capture="environment" multiple hidden @change="pick" />
      <el-icon v-if="!busy"><Camera /></el-icon><el-icon v-else class="is-loading"><Loading /></el-icon>
      <span>{{ busy ? '上传中' : '拍照/上传' }}</span>
    </label>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '@/api/http'
import { parseImages } from '@/utils/format'

const props = defineProps({ modelValue: String, max: { type: Number, default: 6 } })
const emit = defineEmits(['update:modelValue'])
const list = computed(() => parseImages(props.modelValue))
const busy = ref(false)

// 大图在前端压缩到 1600px 以内，节省流量（网格员多用手机拍照）
function compress(file) {
  return new Promise(resolve => {
    if (file.size < 600 * 1024 || !/image\/(jpeg|png|webp)/.test(file.type)) return resolve(file)
    const img = new Image()
    img.onload = () => {
      const s = Math.min(1, 1600 / Math.max(img.width, img.height))
      const c = document.createElement('canvas')
      c.width = img.width * s; c.height = img.height * s
      c.getContext('2d').drawImage(img, 0, 0, c.width, c.height)
      c.toBlob(b => resolve(b ? new File([b], file.name.replace(/\.\w+$/, '.jpg'), { type: 'image/jpeg' }) : file), 'image/jpeg', 0.82)
    }
    img.onerror = () => resolve(file)
    img.src = URL.createObjectURL(file)
  })
}

async function pick(e) {
  const files = [...e.target.files].slice(0, props.max - list.value.length)
  e.target.value = ''
  busy.value = true
  const out = [...list.value]
  try {
    for (const f of files) {
      const r = await api.upload('/upload', await compress(f))
      out.push(r.url)
    }
  } catch { ElMessage.error('图片上传失败') } finally {
    busy.value = false
    emit('update:modelValue', JSON.stringify(out))
  }
}
function del(i) {
  const out = [...list.value]
  out.splice(i, 1)
  emit('update:modelValue', JSON.stringify(out))
}
</script>
