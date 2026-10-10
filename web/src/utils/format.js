import dayjs from 'dayjs'

export const fmtTime = t => (t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '')
export const fmtDate = t => (t ? dayjs(t).format('YYYY-MM-DD') : '')
export const money = (v, unit = '') => (v === null || v === undefined ? '' : Number(v).toLocaleString('zh-CN', { maximumFractionDigits: 2 }) + unit)
export const wan = v => (Number(v || 0) / 10000).toLocaleString('zh-CN', { maximumFractionDigits: 2 })
export const ageOf = birth => (birth ? dayjs().diff(dayjs(birth), 'year') : '')
export const parseImages = s => { try { const a = JSON.parse(s || '[]'); return Array.isArray(a) ? a : [] } catch { return [] } }

export const statusType = s => ({
  上报: 'warning', 受理: 'primary', 处理中: '', 办结: 'success', 退回: 'danger',
  待办: 'info', 进行中: 'primary', 已完成: 'success',
  已发布: 'success', 草稿: 'info', 在保: 'success', 已退出: 'info',
  待调解: 'warning', 调解中: 'primary', 调解成功: 'success', 调解失败: 'danger',
  规划中: 'info', 建设中: 'warning', 已完工: 'success', 运营中: 'success',
  收入: 'success', 支出: 'danger', 红榜: 'danger', 黑榜: 'info',
  高: 'danger', 中: 'warning', 低: 'info', 一般: 'info', 紧急: 'warning', 特急: 'danger',
  在用: 'success', 闲置: 'warning', 出租: 'primary', 承包: 'primary', 报废: 'info'
}[s] ?? 'info')
