import { useI18n as useVueI18n, createI18n as createVueI18n } from 'vue-i18n'
import { fallbackLocale, localeOptions, messages, resolveLocale } from './locales.js'

function loadInitialLocale(preferredLocale) {
	return resolveLocale(preferredLocale || navigator.language)
}

export function createI18n(preferredLocale) {
	const locale = loadInitialLocale(preferredLocale)
	document.documentElement.lang = locale
	return createVueI18n({
		legacy: false,
		globalInjection: true,
		locale,
		fallbackLocale,
		messages,
	})
}

export function useI18n() {
	const i18n = useVueI18n()
	const setLocale = (locale) => {
		const nextLocale = resolveLocale(locale)
		i18n.locale.value = nextLocale
		document.documentElement.lang = nextLocale
	}
	const ta = (key) => i18n.tm(key)
	return { ...i18n, localeOptions, setLocale, ta }
}

export { fallbackLocale, localeOptions, messages, resolveLocale }
