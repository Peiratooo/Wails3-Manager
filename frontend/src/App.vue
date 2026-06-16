<template>
    <n-config-provider
        :date-locale="activeNaiveDateLocale"
        :locale="activeNaiveLocale"
        :theme="naiveTheme"
        :theme-overrides="currentNaiveThemeOverrides"
    >
        <n-modal-provider>
            <n-loading-bar-provider>
                <n-message-provider>
                    <n-notification-provider :max="3" placement="top-right">
                        <div class="app-view">
                            <router-view />
                        </div>
                    </n-notification-provider>
                </n-message-provider>
            </n-loading-bar-provider>
        </n-modal-provider>
    </n-config-provider>
</template>

<script setup>
import {useAppStore} from './store'
import {Events, WML} from "@wailsio/runtime"
import {useRoute} from "vue-router";
import {SettingsService} from '../bindings/wails3-manager/core/settings'
import {EnvironmentService} from "../bindings/wails3-manager/core/environment"
import {
    dateDeDE,
    dateEnUS,
    dateFrFR,
    dateJaJP,
    dateKoKR,
    dateZhCN,
    dateZhTW,
    darkTheme,
    deDE,
    enUS,
    frFR,
    jaJP,
    koKR,
    NConfigProvider,
    NLoadingBarProvider,
    NMessageProvider,
    NModalProvider,
    NNotificationProvider,
    zhCN,
    zhTW
} from 'naive-ui'
import {computed, onBeforeUnmount, onMounted, provide, watch} from "vue";
import {
    applyThemeToDocument,
    naiveThemeOverrides,
    themeCssVariables,
    themeNameFromIsDark,
} from './theme'

const route = useRoute()
const store = useAppStore()

let offLogLine = null

const themeName = computed(() => themeNameFromIsDark(store.settings.isDark))
const isDarkTheme = computed(() => store.settings.isDark)
const naiveTheme = computed(() => isDarkTheme.value ? darkTheme : null)
const currentNaiveThemeOverrides = computed(() => naiveThemeOverrides[themeName.value])
const currentThemeCssVariables = computed(() => themeCssVariables[themeName.value])
const naiveLocaleMap = {
    'zh-CN': { locale: zhCN, dateLocale: dateZhCN },
    'zh-TW': { locale: zhTW, dateLocale: dateZhTW },
    'en-US': { locale: enUS, dateLocale: dateEnUS },
    'ja-JP': { locale: jaJP, dateLocale: dateJaJP },
    'ko-KR': { locale: koKR, dateLocale: dateKoKR },
    'fr-FR': { locale: frFR, dateLocale: dateFrFR },
    'de-DE': { locale: deDE, dateLocale: dateDeDE },
}
const activeNaive = computed(() => naiveLocaleMap[store.settings.language] || naiveLocaleMap['zh-CN'])
const activeNaiveLocale = computed(() => activeNaive.value.locale)
const activeNaiveDateLocale = computed(() => activeNaive.value.dateLocale)

const setDarkMode = async (isDark) => {
    const previousSettings = {...store.settings}
    const nextSettings = {
        ...store.settings,
        isDark: Boolean(isDark),
    }
    store.setSettings(nextSettings)
    try {
        const savedSettings = await SettingsService.SaveSettings(nextSettings)
        store.setSettings(savedSettings)
    } catch (error) {
        store.setSettings(previousSettings)
        console.error(error)
    }
}

const toggleTheme = () => {
    return setDarkMode(!isDarkTheme.value)
}

watch(
    themeName,
    (nextThemeName) => {
        applyThemeToDocument(nextThemeName)
    },
    {immediate: true}
)

function initEnvironment() {
    EnvironmentService.CheckEnvironment().then(res=>{
        store.env.data = res
        store.env.loaded = true
    })
}

const formatTimestamp = (timestamp) => {
    if (!timestamp) return '-'

    const date = new Date(timestamp * 1000)

    const pad = (n) => String(n).padStart(2, '0')

    const year = date.getFullYear()
    const month = pad(date.getMonth() + 1)
    const day = pad(date.getDate())
    const hour = pad(date.getHours())
    const minute = pad(date.getMinutes())

    return `${year}-${month}-${day} ${hour}:${minute}`
}

onMounted(()=>{
    offLogLine = Events.On('manager:log-line', (event) => {
        store.appendLogLine(event.data?.line)
    })
    initEnvironment()
    WML.Reload()
})

onBeforeUnmount(() => {
    offLogLine?.()
})

provide("formatTimestamp",formatTimestamp)
provide("store",store)
provide("route",route)
provide("theme", {
    themeName,
    isDarkTheme,
    themeCssVariables: currentThemeCssVariables,
    setDarkMode,
    toggleTheme,
})

defineExpose({
    themeName,
    isDarkTheme,
    themeCssVariables: currentThemeCssVariables,
    setDarkMode,
    toggleTheme,
})
</script>

<style lang="scss" scoped>
.app-view {
    width: 100vw;
    height: 100vh;
    overflow: hidden;
}
</style>
