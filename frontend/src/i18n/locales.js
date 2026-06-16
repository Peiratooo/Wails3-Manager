import enUS from './messages/en-US.js'
import zhCN from './messages/zh-CN.js'
import zhTW from './messages/zh-TW.js'
import jaJP from './messages/ja-JP.js'
import koKR from './messages/ko-KR.js'
import frFR from './messages/fr-FR.js'
import deDE from './messages/de-DE.js'

export const fallbackLocale = 'zh-CN'

export const localeDefinitions = [
	{ label: '简体中文', value: 'zh-CN', message: zhCN },
	{ label: '繁體中文', value: 'zh-TW', message: zhTW },
	{ label: 'English', value: 'en-US', message: enUS },
	{ label: '日本語', value: 'ja-JP', message: jaJP },
	{ label: '한국어', value: 'ko-KR', message: koKR },
	{ label: 'Français', value: 'fr-FR', message: frFR },
	{ label: 'Deutsch', value: 'de-DE', message: deDE },
]

export const localeOptions = localeDefinitions.map(({ label, value }) => ({ label, value }))

export const messages = Object.fromEntries(
	localeDefinitions.map(({ value, message }) => [value, message])
)

const localeAliases = {
	zh: 'zh-CN',
	'zh-cn': 'zh-CN',
	'zh-hans': 'zh-CN',
	'zh-sg': 'zh-CN',
	'zh-tw': 'zh-TW',
	'zh-hk': 'zh-TW',
	'zh-mo': 'zh-TW',
	en: 'en-US',
	'en-us': 'en-US',
	'en-gb': 'en-US',
	ja: 'ja-JP',
	'ja-jp': 'ja-JP',
	ko: 'ko-KR',
	'ko-kr': 'ko-KR',
	fr: 'fr-FR',
	'fr-fr': 'fr-FR',
	de: 'de-DE',
	'de-de': 'de-DE',
}

export function resolveLocale(locale) {
	if (!locale) {
		return fallbackLocale
	}
	if (messages[locale]) {
		return locale
	}
	const normalized = String(locale).trim().replace('_', '-').toLowerCase()
	if (localeAliases[normalized]) {
		return localeAliases[normalized]
	}
	const language = normalized.split('-')[0]
	return localeAliases[language] || fallbackLocale
}
