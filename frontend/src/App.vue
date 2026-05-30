<template>
    <n-config-provider
        :date-locale="dateZhCN"
        :locale="zhCN"
        :theme="naiveTheme"
        :theme-overrides="currentNaiveThemeOverrides"
    >
        <n-modal-provider>
            <n-loading-bar-provider>
                <n-message-provider>
                    <n-notification-provider :max="3" placement="top-right">
                        <router-view />
                    </n-notification-provider>
                </n-message-provider>
            </n-loading-bar-provider>
        </n-modal-provider>
    </n-config-provider>
</template>

<script setup>
import {useAppStore} from './store'
import {WML} from "@wailsio/runtime"
import {
    dateZhCN,
    darkTheme,
    NConfigProvider,
    NLoadingBarProvider,
    NMessageProvider,
    NModalProvider,
    NNotificationProvider,
    zhCN
} from 'naive-ui'
import {computed, onMounted, provide, ref, watch} from "vue";
import {
    applyThemeToDocument,
    getInitialThemeName,
    naiveThemeOverrides,
    normalizeThemeName,
    storeThemeName,
    themeCssVariables,
    THEME_NAMES,
} from './theme'

const store = useAppStore()

const themeName = ref(getInitialThemeName())
const isDarkTheme = computed(() => themeName.value === THEME_NAMES.DARK)
const naiveTheme = computed(() => isDarkTheme.value ? darkTheme : null)
const currentNaiveThemeOverrides = computed(() => naiveThemeOverrides[themeName.value])
const currentThemeCssVariables = computed(() => themeCssVariables[themeName.value])

const setTheme = (nextThemeName) => {
    themeName.value = normalizeThemeName(nextThemeName)
}

const toggleTheme = () => {
    setTheme(isDarkTheme.value ? THEME_NAMES.LIGHT : THEME_NAMES.DARK)
}

watch(
    themeName,
    (nextThemeName) => {
        applyThemeToDocument(nextThemeName)
        storeThemeName(nextThemeName)
        store.setThemeName(nextThemeName)
    },
    {immediate: true}
)

onMounted(()=>{
    WML.Reload()
})

provide("store",store)
provide("theme", {
    themeName,
    isDarkTheme,
    themeCssVariables: currentThemeCssVariables,
    setTheme,
    toggleTheme,
})

defineExpose({
    themeName,
    isDarkTheme,
    themeCssVariables: currentThemeCssVariables,
    setTheme,
    toggleTheme,
})
</script>

<style lang="scss" scoped>

</style>
