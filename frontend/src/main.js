import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import { createI18n } from './i18n'

const app = createApp(App)
app
	.use(router)
	.use(createPinia())
	.use(createI18n())
	.mount('#app')
