<template>
    <n-config-provider :date-locale="dateZhCN" :locale="zhCN">
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
import {useRoute} from 'vue-router'
import {WML} from "@wailsio/runtime"
import {
    dateZhCN,
    NConfigProvider,
    NLoadingBarProvider,
    NMessageProvider,
    NModalProvider,
    NNotificationProvider,
    useMessage,
    useNotification,
    zhCN
} from 'naive-ui'
import {onMounted,provide} from "vue";

const store = useAppStore()
const route = useRoute()


onMounted(()=>{
    WML.Reload()
})

provide("store",store)
</script>

<style lang="scss" scoped>

</style>