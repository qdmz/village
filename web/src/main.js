import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import * as Icons from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'
import 'dayjs/locale/zh-cn'
import dayjs from 'dayjs'
import App from './App.vue'
import router from './router'
import './styles/main.scss'
import { applyTheme, currentTheme } from './utils/theme'

dayjs.locale('zh-cn')
applyTheme(currentTheme())
const app = createApp(App)
for (const [k, c] of Object.entries(Icons)) app.component(k, c)
app.use(ElementPlus, { locale: zhCn }).use(router).mount('#app')
