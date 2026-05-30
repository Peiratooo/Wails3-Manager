import { createApp } from 'vue'
import { applyThemeToDocument, getInitialThemeName } from './theme'
import './styles/themes.css'
import './style.css'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import { createI18n } from './i18n'

applyThemeToDocument(getInitialThemeName())

const app = createApp(App)
app
	.use(router)
	.use(createPinia())
	.use(createI18n())
	.mount('#app')
