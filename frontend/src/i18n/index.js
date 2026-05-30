import { useI18n as useVueI18n, createI18n as createVueI18n } from 'vue-i18n'
import enUS from './messages/en-US'
import zhCN from './messages/zh-CN'

export const localeOptions = [
	{ label: '简体中文', value: 'zh-CN' },
	{ label: 'English', value: 'en-US' },
]

const storageKey = 'wails-builder.locale'
const messages = {
	'zh-CN': zhCN,
	'en-US': enUS,
}

function loadInitialLocale() {
	const saved = localStorage.getItem(storageKey)
	if (saved && messages[saved]) {
		return saved
	}
	const language = navigator.language || 'zh-CN'
	return language.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en-US'
}

export function createI18n() {
	return createVueI18n({
		legacy: false,
		globalInjection: true,
		locale: loadInitialLocale(),
		fallbackLocale: 'zh-CN',
		messages,
	})
}

export function useI18n() {
	const i18n = useVueI18n()
	const setLocale = (locale) => {
		if (!messages[locale]) {
			return
		}
		i18n.locale.value = locale
		localStorage.setItem(storageKey, locale)
		document.documentElement.lang = locale
	}
	const ta = (key) => i18n.tm(key)
	return { ...i18n, localeOptions, setLocale, ta }
}
