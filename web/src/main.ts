import 'lib-flexible/flexible.js'
import { createApp } from 'vue'
import 'vant/lib/index.css'
import './style/index.less'
import i18n from './language'
import pinia from './pinia'
import App from './App.vue'
import Antd from 'ant-design-vue'
import Vant from 'vant'
import router from './router'
import { Login } from './components/components'

const app = createApp(App)
app.component('Login', Login)
app.use(i18n)
app.use(pinia)
app.use(Vant)
app.use(Antd)
app.use(router)
app.mount('#app')
