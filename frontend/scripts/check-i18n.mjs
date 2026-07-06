import { fallbackLocale, localeDefinitions, messages } from '../src/i18n/locales.js'

function flattenKeys(value, prefix = '') {
	if (!value || typeof value !== 'object' || Array.isArray(value)) {
		return prefix ? [prefix] : []
	}
	return Object.entries(value).flatMap(([key, child]) => {
		const nextPrefix = prefix ? `${prefix}.${key}` : key
		return flattenKeys(child, nextPrefix)
	})
}

const baseKeys = flattenKeys(messages[fallbackLocale]).sort()
let failed = false

for (const { value } of localeDefinitions) {
	const keys = flattenKeys(messages[value]).sort()
	const keySet = new Set(keys)
	const baseKeySet = new Set(baseKeys)
	const missing = baseKeys.filter((key) => !keySet.has(key))
	const extra = keys.filter((key) => !baseKeySet.has(key))

	if (missing.length || extra.length) {
		failed = true
		console.error(`Locale ${value} does not match ${fallbackLocale}.`)
		if (missing.length) {
			console.error(`  Missing: ${missing.join(', ')}`)
		}
		if (extra.length) {
			console.error(`  Extra: ${extra.join(', ')}`)
		}
	}
}

if (failed) {
	process.exitCode = 1
} else {
	console.log(`i18n key coverage ok for ${localeDefinitions.length} locales.`)
}
