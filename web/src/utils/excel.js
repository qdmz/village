// Excel(xlsx)/CSV 导入导出：后端统一输出 CSV，前端用 SheetJS 转为 xlsx；导入时把 xlsx 转为 CSV 上传
import * as XLSX from 'xlsx'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api/http'

function save(blob, name) {
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = name
  document.body.appendChild(a)
  a.click()
  setTimeout(() => { URL.revokeObjectURL(a.href); a.remove() }, 500)
}

const stamp = () => new Date().toISOString().slice(0, 10).replace(/-/g, '')

export async function exportList(url, params, title, format = 'xlsx') {
  const text = await api.exportText(url, params)
  if (format === 'csv') {
    save(new Blob([text], { type: 'text/csv;charset=utf-8' }), `${title}_${stamp()}.csv`)
    return
  }
  const wb = XLSX.read(text.replace(/^\uFEFF/, ''), { type: 'string', raw: true })
  const ws = wb.Sheets[wb.SheetNames[0]]
  const rows = XLSX.utils.sheet_to_json(ws, { header: 1 })
  ws['!cols'] = (rows[0] || []).map((_, i) => ({ wch: Math.min(40, Math.max(8, ...rows.slice(0, 200).map(r => String(r[i] ?? '').length * 1.6))) }))
  wb.SheetNames[0] = title.slice(0, 30)
  wb.Sheets[wb.SheetNames[0]] = ws
  XLSX.writeFile(wb, `${title}_${stamp()}.xlsx`)
}

export function downloadTemplate(title, headers, sample) {
  const ws = XLSX.utils.aoa_to_sheet(sample ? [headers, sample] : [headers])
  ws['!cols'] = headers.map(h => ({ wch: Math.max(12, h.length * 2.2) }))
  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, '导入模板')
  XLSX.writeFile(wb, `${title}_导入模板.xlsx`)
}

export async function importFile(url, file) {
  let csvText
  if (/\.csv$/i.test(file.name)) {
    csvText = await file.text()
  } else {
    const buf = await file.arrayBuffer()
    const wb = XLSX.read(buf, { type: 'array', cellDates: false, raw: false })
    const ws = wb.Sheets[wb.SheetNames[0]]
    // 以文本方式读取，避免身份证号被转成科学计数
    csvText = XLSX.utils.sheet_to_csv(ws, { rawNumbers: true, blankrows: false })
  }
  const blob = new Blob(['\uFEFF' + csvText], { type: 'text/csv' })
  const res = await api.upload(`${url}/import`, new File([blob], 'import.csv'))
  if (res.failed > 0) {
    await ElMessageBox.alert(
      `<p>成功导入 <b>${res.success}</b> 条，失败 <b>${res.failed}</b> 条：</p><ul style="max-height:260px;overflow:auto;padding-left:18px">${res.errors.slice(0, 50).map(e => `<li>${e.replace(/</g, '&lt;')}</li>`).join('')}</ul>`,
      '导入结果', { dangerouslyUseHTMLString: true }
    )
  } else {
    ElMessage.success(`成功导入 ${res.success} 条数据`)
  }
  return res
}

export function downloadBlob(blob, name) { save(blob, name) }
