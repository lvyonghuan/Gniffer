import { createApp } from 'vue'
import App from './App.vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import axios from 'axios'
import router from './router.js'

axios.defaults.baseURL="http://127.0.0.1:8080"

const app = createApp(App); // 使用 createApp 创建应用程序实例

app.use(ElementPlus);
app.use(router);
app.mount('#app');