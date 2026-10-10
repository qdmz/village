<template>
  <div class="crud">
    <div class="page-head" v-if="!embedded">
      <div>
        <h2>{{ cfg.title }}</h2>
        <p v-if="cfg.desc" class="muted">{{ cfg.desc }}</p>
      </div>
      <div class="head-extra"><slot name="head" /></div>
    </div>
    <slot name="top" :reload="load" />

    <div class="card toolbar">
      <div class="filters">
        <el-input v-model="q.kw" :placeholder="cfg.searchPlaceholder || '关键字搜索'" clearable class="f-kw" @keyup.enter="search" @clear="search">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <template v-for="f in visibleFilters" :key="f.prop">
          <el-select v-if="f.type === 'select' || f.type === 'grid' || f.type === 'user'" v-model="q[f.prop]" :placeholder="f.label" clearable class="f-item" @change="search" filterable>
            <el-option v-for="o in optionsOf(f)" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
          <el-date-picker v-else-if="f.type === 'daterange'" v-model="q._range" type="daterange" value-format="YYYY-MM-DD" range-separator="至"
            start-placeholder="开始" end-placeholder="结束" class="f-range" @change="search" />
          <el-checkbox v-else-if="f.type === 'check'" v-model="q[f.prop]" true-value="1" false-value="" border @change="search">{{ f.label }}</el-checkbox>
        </template>
        <el-button v-if="!isMobile" type="primary" plain @click="search"><el-icon><Search /></el-icon><span>查询</span></el-button>
        <el-button v-if="!isMobile" @click="reset">重置</el-button>
      </div>
      <div class="actions">
        <slot name="toolbar" />
        <el-button v-if="canCreate" type="primary" @click="openForm()"><el-icon><Plus /></el-icon><span>{{ cfg.createText || '新增' }}</span></el-button>
        <el-dropdown v-if="cfg.importable && canCreate" trigger="click" @command="onImport">
          <el-button><el-icon><Upload /></el-icon><span v-if="!isMobile">导入</span></el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="file">选择 Excel / CSV 文件</el-dropdown-item>
              <el-dropdown-item command="tpl">下载导入模板</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-dropdown v-if="cfg.exportable !== false" trigger="click" @command="onExport">
          <el-button :loading="exporting"><el-icon><Download /></el-icon><span v-if="!isMobile">导出</span></el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="xlsx">导出 Excel (.xlsx)</el-dropdown-item>
              <el-dropdown-item command="csv">导出 CSV</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <input ref="fileInput" type="file" accept=".xlsx,.xls,.csv" hidden @change="doImport" />
      </div>
    </div>

    <!-- 桌面端表格 -->
    <div v-if="!isMobile" class="card table-card">
      <el-table :data="rows" v-loading="loading" stripe :row-class-name="cfg.rowClass" @row-dblclick="openDetail" empty-text="暂无数据">
        <el-table-column type="index" label="#" width="56" :index="i => (q.page - 1) * q.size + i + 1" />
        <el-table-column v-for="f in listFields" :key="f.prop" :prop="f.prop" :label="f.label" :width="f.width" :min-width="f.minWidth || 100"
          :show-overflow-tooltip="f.type !== 'progress'">
          <template #default="{ row }"><CellValue :f="f" :row="row" /></template>
        </el-table-column>
        <el-table-column label="操作" :width="cfg.actionWidth || 170" fixed="right">
          <template #default="{ row }">
            <slot name="row-actions" :row="row" :reload="load" />
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="canEditRow(row)" link type="primary" @click="openForm(row)">编辑</el-button>
            <el-button v-if="canDeleteRow(row)" link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 移动端卡片 -->
    <div v-else class="m-list" v-loading="loading">
      <div v-for="row in rows" :key="row.id" class="m-card" :class="cfg.rowClass && cfg.rowClass({ row })" @click="openDetail(row)">
        <div class="m-top">
          <div class="m-title">{{ mobileTitle(row) }}</div>
          <el-tag v-for="t in mobileTags" :key="t.prop" size="small" :type="statusType(row[t.prop])" effect="light" round>{{ cellText(t, row) }}</el-tag>
        </div>
        <div class="m-meta">
          <span v-for="f in mobileMeta" :key="f.prop"><em>{{ f.label }}</em>{{ cellText(f, row) || '—' }}</span>
        </div>
        <div class="m-actions" @click.stop>
          <slot name="row-actions" :row="row" :reload="load" />
          <el-button v-if="canEditRow(row)" size="small" @click="openForm(row)">编辑</el-button>
          <el-button v-if="canDeleteRow(row)" size="small" type="danger" plain @click="remove(row)">删除</el-button>
        </div>
      </div>
      <el-empty v-if="!loading && !rows.length" description="暂无数据" />
    </div>

    <div class="pager">
      <el-pagination v-model:current-page="q.page" v-model:page-size="q.size" :total="total" :page-sizes="[10, 20, 50, 100]"
        :layout="isMobile ? 'total, prev, pager, next' : 'total, sizes, prev, pager, next, jumper'" :pager-count="isMobile ? 5 : 7"
        background @current-change="load" @size-change="search" />
    </div>

    <!-- 表单 -->
    <el-dialog v-model="formVisible" :title="(form.id ? '编辑' : '新增') + (cfg.itemName || cfg.title)" :width="cfg.formWidth || '760px'"
      :fullscreen="isMobile" destroy-on-close :close-on-click-modal="false" class="vm-dialog">
      <el-form ref="formRef" :model="form" :rules="rules" :label-width="isMobile ? 'auto' : (cfg.labelWidth || '100px')" :label-position="isMobile ? 'top' : 'right'">
        <el-row :gutter="16">
          <template v-for="f in formFields" :key="f.prop">
            <el-col v-if="!f.showIf || f.showIf(form)" :span="isMobile ? 24 : (f.span || (['textarea', 'images'].includes(f.type) ? 24 : 12))">
              <el-form-item :label="f.label" :prop="f.prop">
                <FieldInput v-model="form[f.prop]" :f="f" :form="form" />
              </el-form-item>
            </el-col>
          </template>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 详情 -->
    <el-drawer v-model="detailVisible" :title="(cfg.itemName || cfg.title) + '详情'" :size="isMobile ? '100%' : (cfg.detailWidth || '560px')" destroy-on-close>
      <slot name="detail" :row="detail" :reload="load" :close="() => (detailVisible = false)">
        <el-descriptions :column="1" border v-if="detail">
          <el-descriptions-item v-for="f in detailFields" :key="f.prop" :label="f.label" label-class-name="desc-label">
            <CellValue :f="f" :row="detail" full />
          </el-descriptions-item>
        </el-descriptions>
      </slot>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api/http'
import { store, role } from '@/store'
import { exportList, importFile, downloadTemplate } from '@/utils/excel'
import { statusType } from '@/utils/format'
import { loadGrids, loadUsers, optionsOf, cellText } from '@/utils/options'
import CellValue from './CellValue.vue'
import FieldInput from './FieldInput.vue'

const props = defineProps({ cfg: { type: Object, required: true }, embedded: Boolean, fixedQuery: { type: Object, default: () => ({}) } })
const emit = defineEmits(['saved'])
const cfg = props.cfg
const isMobile = computed(() => store.isMobile)

const q = reactive({ page: 1, size: 20, kw: '', _range: null, ...(cfg.defaultQuery || {}) })
const rows = ref([])
const total = ref(0)
const loading = ref(false)
const exporting = ref(false)

const visibleFilters = computed(() => (cfg.filters || []).filter(f => !f.roles || f.roles.includes(role())))
const listFields = computed(() => cfg.fields.filter(f => f.list !== false && (!f.roles || f.roles.includes(role()))))
const formFields = computed(() => cfg.fields.filter(f => f.form !== false && f.type !== 'computed' && (!f.formRoles || f.formRoles.includes(role()))))
const detailFields = computed(() => cfg.fields.filter(f => f.detail !== false))
const mobileTags = computed(() => cfg.fields.filter(f => f.mobile === 'tag'))
const mobileMeta = computed(() => cfg.fields.filter(f => f.mobile === 'meta'))
const mobileTitle = row => (cfg.mobileTitle ? cfg.mobileTitle(row) : row[cfg.fields[0].prop])

const has = roles => !roles || roles.includes(role())
const canCreate = computed(() => cfg.creatable !== false && has(cfg.writeRoles))
const canEditRow = row => cfg.editable !== false && has(cfg.writeRoles) && (!cfg.canEdit || cfg.canEdit(row))
const canDeleteRow = row => cfg.deletable !== false && has(cfg.deleteRoles || cfg.writeRoles) && (!cfg.canDelete || cfg.canDelete(row))

function params() {
  const p = { ...props.fixedQuery }
  for (const [k, v] of Object.entries(q)) {
    if (k === '_range') continue
    if (v !== '' && v !== null && v !== undefined) p[k] = v
  }
  if (q._range) { p.start = q._range[0]; p.end = q._range[1] }
  return p
}

async function load() {
  loading.value = true
  try {
    const d = await api.list(cfg.api, params())
    rows.value = d.list
    total.value = d.total
  } catch { /* 已提示 */ } finally { loading.value = false }
}
function search() { q.page = 1; load() }
function reset() {
  for (const k of Object.keys(q)) if (!['page', 'size'].includes(k)) q[k] = k === '_range' ? null : ''
  Object.assign(q, cfg.defaultQuery || {})
  search()
}

// 表单
const formVisible = ref(false)
const saving = ref(false)
const formRef = ref()
const form = ref({})
const rules = computed(() => {
  const r = {}
  for (const f of cfg.fields) if (f.required) r[f.prop] = [{ required: true, message: `请${['select', 'grid', 'user', 'date', 'remote'].includes(f.type) ? '选择' : '填写'}${f.label}`, trigger: ['blur', 'change'] }]
  return r
})
function openForm(row) {
  const base = {}
  for (const f of cfg.fields) if (f.default !== undefined) base[f.prop] = typeof f.default === 'function' ? f.default() : f.default
  form.value = row ? JSON.parse(JSON.stringify(row)) : { ...base, ...(props.fixedQuery || {}) }
  if (row && cfg.loadForEdit) {
    api.get(cfg.api, row.id).then(d => { form.value = cfg.loadForEdit(d) })
  }
  formVisible.value = true
}
async function save() {
  await formRef.value.validate()
  saving.value = true
  try {
    const data = cfg.beforeSave ? cfg.beforeSave({ ...form.value }) : form.value
    if (form.value.id) await api.update(cfg.api, form.value.id, data)
    else await api.create(cfg.api, data)
    ElMessage.success('保存成功')
    formVisible.value = false
    emit('saved')
    load()
  } catch { /* 已提示 */ } finally { saving.value = false }
}
async function remove(row) {
  await ElMessageBox.confirm(`确定删除「${mobileTitle(row)}」吗？删除后不可恢复。`, '删除确认', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
  await api.remove(cfg.api, row.id)
  ElMessage.success('已删除')
  load()
}

// 详情
const detailVisible = ref(false)
const detail = ref(null)
async function openDetail(row) {
  if (cfg.onDetail) return cfg.onDetail(row)
  detail.value = row
  detailVisible.value = true
  if (cfg.fetchDetail) {
    try { detail.value = await api.get(cfg.api, row.id) } catch { /* */ }
  }
}

// 导入导出
const fileInput = ref()
async function onExport(fmt) {
  exporting.value = true
  try { await exportList(cfg.api, params(), cfg.exportTitle || cfg.title, fmt); ElMessage.success('已导出，请查看浏览器下载') } catch { /* */ } finally { exporting.value = false }
}
function onImport(cmd) {
  if (cmd === 'tpl') downloadTemplate(cfg.title, cfg.importable.headers, cfg.importable.sample)
  else fileInput.value.click()
}
async function doImport(e) {
  const file = e.target.files[0]
  e.target.value = ''
  if (!file) return
  try { await importFile(cfg.api, file); search() } catch { /* */ }
}

const route = useRoute()
onMounted(async () => {
  await Promise.all([loadGrids(), loadUsers()])
  load()
  if (props.embedded) return
  if (route.query.new && canCreate.value) openForm()
  if (route.query.id) openDetail({ id: Number(route.query.id) })
})
watch(() => props.fixedQuery, search, { deep: true })
function setQuery(o) { Object.assign(q, o); search() }
defineExpose({ reload: load, openForm, openDetail, setQuery })
</script>
