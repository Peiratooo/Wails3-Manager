import { createApp } from 'vue'
import { applyThemeToDocument, themeNameFromIsDark } from './theme'
import './styles/themes.css'
import './style.css'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import { useAppStore } from './store'
import { createI18n } from './i18n'
import { Service as SettingsService } from '../bindings/wails3-manager/core/settings'

const defaultSettings = {
	isDark: true,
	language: 'zh-CN',
	recordLogs: true,
}

async function loadSettings() {
	try {
		return await SettingsService.GetSettings()
	} catch (error) {
		console.warn('Load settings failed, using defaults.', error)
		return defaultSettings
	}
}

async function bootstrap() {
	const app = createApp(App)
	const pinia = createPinia()

	app.use(pinia)

	const store = useAppStore()
	const settings = await loadSettings()
	store.setSettings(settings)
	applyThemeToDocument(themeNameFromIsDark(store.settings.isDark))

	app
		.use(router)
		.use(createI18n(store.settings.language))
		.mount('#app')
}

bootstrap()
