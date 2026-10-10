import axios from 'axios'
import { ElMessage } from 'element-plus'
import { store, logout } from '@/store'

const http = axios.create({ baseURL: './api', timeout: 30000 })

http.interceptors.request.use(cfg => {
  if (store.token) cfg.headers.Authorization = 'Bearer ' + store.token
  return cfg
})

http.interceptors.response.use(
  res => {
    if (res.config.responseType === 'blob' || res.config.responseType === 'text') return res.data
    const d = res.data
    if (d && d.code === 0) return d.data
    ElMessage.error((d && d.msg) || '请求失败')
    return Promise.reject(d)
  },
  err => {
    const st = err.response?.status
    const msg = err.response?.data?.msg || (st ? `请求失败（${st}）` : '网络异常，请检查网络连接')
    if (st === 401 && store.token) {
      logout()
      ElMessage.warning(msg)
      location.hash = '#/login'
    } else if (!err.config?.silent) {
      ElMessage.error(msg)
    }
    return Promise.reject(err.response?.data || err)
  }
)

export default http

export const api = {
  list: (url, params) => http.get(url, { params }),
  get: (url, id) => http.get(`${url}/${id}`),
  create: (url, data) => http.post(url, data),
  update: (url, id, data) => http.put(`${url}/${id}`, data),
  remove: (url, id) => http.delete(`${url}/${id}`),
  exportText: (url, params) => http.get(`${url}/export`, { params, responseType: 'text' }),
  upload: (url, file) => {
    const fd = new FormData()
    fd.append('file', file)
    return http.post(url, fd)
  }
}
