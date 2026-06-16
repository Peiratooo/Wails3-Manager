<template>
    <n-modal v-model:show="visible" class="settings-modal">
        <div class="settings-card">
            <div class="settings-head">
                <div>
                    <div class="settings-title">{{ t("settings.title") }}</div>
                    <div class="settings-subtitle">{{ t("app.title") }}</div>
                </div>

                <n-button quaternary size="small" @click="visible = false">
                    {{ t("editor.cancel") }}
                </n-button>
            </div>

            <n-scrollbar class="settings-scroll">
                <div class="settings-group">
                    <div class="group-title">{{ t("settings.appearance") }}</div>

                    <div class="setting-row">
                        <div class="setting-copy">
                            <div class="setting-label">{{ t("settings.darkMode") }}</div>
                            <div class="setting-desc">{{ t("settings.darkModeDesc") }}</div>
                        </div>

                        <n-switch
                            :value="store.settings.isDark"
                            :loading="saving"
                            @update:value="updateDarkMode"
                        />
                    </div>

                    <div class="setting-row">
                        <div class="setting-copy">
                            <div class="setting-label">{{ t("settings.language") }}</div>
                            <div class="setting-desc">{{ t("settings.languageDesc") }}</div>
                        </div>

                        <n-select
                            class="language-select"
                            :value="store.settings.language"
                            :options="localeOptions"
                            :disabled="saving"
                            @update:value="updateLanguage"
                        />
                    </div>
                </div>

                <div class="settings-group">
                    <div class="group-head">
                        <div>
                            <div class="group-title">{{ t("settings.logs") }}</div>
                            <div class="group-desc">{{ t("settings.logsDesc") }}</div>
                        </div>

                        <n-button
                            size="small"
                            :disabled="clearingLogs"
                            :loading="clearingLogs"
                            @click="clearLogs"
                        >
                            {{ t("settings.clearLogs") }}
                        </n-button>
                    </div>

                    <div class="setting-row">
                        <div class="setting-copy">
                            <div class="setting-label">{{ t("settings.recordLogs") }}</div>
                            <div class="setting-desc">{{ t("settings.recordLogsDesc") }}</div>
                        </div>

                        <n-switch
                            :value="store.settings.recordLogs"
                            :loading="saving"
                            @update:value="updateRecordLogs"
                        />
                    </div>

                    <div class="log-box">
                        <n-scrollbar>
                            <div v-if="store.logLines.length" class="log-lines">
                                <div
                                    v-for="(line, index) in store.logLines"
                                    :key="index"
                                    class="log-line"
                                >
                                    {{ line }}
                                </div>
                            </div>

                            <div v-else class="empty-log">
                                {{ t("settings.noLogs") }}
                            </div>
                        </n-scrollbar>
                    </div>
                </div>
            </n-scrollbar>
        </div>
    </n-modal>
</template>

<script setup>
import { computed, inject, ref } from "vue"
import { NButton, NModal, NScrollbar, NSelect, NSwitch, useMessage } from "naive-ui"
import { SettingsService } from "../../bindings/wails3-manager/core/settings"
import { useI18n } from "../i18n"

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    }
})

const emit = defineEmits(["update:show"])

const store = inject("store")
const message = useMessage()
const { t, localeOptions, setLocale } = useI18n()

const saving = ref(false)
const clearingLogs = ref(false)

const visible = computed({
    get() {
        return props.show
    },
    set(value) {
        emit("update:show", value)
    }
})

async function saveSettings(nextSettings) {
    if (saving.value) return

    const previousSettings = { ...store.settings }

    saving.value = true
    store.setSettings(nextSettings)
    setLocale(nextSettings.language)

    try {
        const savedSettings = await SettingsService.SaveSettings(nextSettings)
        store.setSettings(savedSettings)
        setLocale(savedSettings.language)
    } catch (error) {
        store.setSettings(previousSettings)
        setLocale(previousSettings.language)
        message.error(error?.message || String(error))
    } finally {
        saving.value = false
    }
}

function updateDarkMode(isDark) {
    saveSettings({
        ...store.settings,
        isDark
    })
}

function updateLanguage(language) {
    saveSettings({
        ...store.settings,
        language
    })
}

function updateRecordLogs(recordLogs) {
    saveSettings({
        ...store.settings,
        recordLogs
    })
}

async function clearLogs() {
    if (clearingLogs.value) return

    clearingLogs.value = true

    try {
        await SettingsService.ClearLogs()
        store.clearLogLines()
    } catch (error) {
        message.error(error?.message || String(error))
    } finally {
        clearingLogs.value = false
    }
}
</script>

<style lang="scss" scoped>
.settings-card {
    width: min(640px, calc(100vw - 32px));
    max-height: min(720px, calc(100vh - 64px));
    box-sizing: border-box;
    border-radius: 8px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-bg-page);


    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.settings-head {
    flex: 0 0 auto;
    min-height: 64px;
    padding: 16px 18px;
    box-sizing: border-box;
    border-bottom: 1px solid var(--wm-border-subtle);

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.settings-title {
    font-size: 16px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.settings-subtitle {
    margin-top: 4px;
    font-size: 12px;
    color: var(--wm-text-muted);
}

.settings-scroll {
    flex: 1;
    min-height: 0;
    max-height: calc(min(720px, 100vh - 64px) - 65px);
}

.settings-group {
    padding: 16px 18px;
}

.settings-group + .settings-group {
    border-top: 1px solid var(--wm-border-subtle);
}

.group-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 14px;
}

.group-title {
    font-size: 14px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.group-desc {
    margin-top: 5px;
    font-size: 12px;
    color: var(--wm-text-muted);
}

.setting-row {
    min-height: 52px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 18px;
}

.group-title + .setting-row,
.group-head + .setting-row {
    margin-top: 12px;
}

.setting-row + .setting-row {
    border-top: 1px solid var(--wm-border-subtle);
}

.setting-copy {
    min-width: 0;
}

.setting-label {
    font-size: 13px;
    color: var(--wm-text-primary);
}

.setting-desc {
    margin-top: 4px;
    font-size: 12px;
    line-height: 1.45;
    color: var(--wm-text-muted);
}

.language-select {
    width: 180px;
    flex: 0 0 auto;
}

.log-box {
    height: 180px;
    margin-top: 12px;
    box-sizing: border-box;
    border-radius: 8px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-control-bg);
    overflow: hidden;
}

.log-lines {
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.log-line {
    font-family: Consolas, "Courier New", monospace;
    font-size: 12px;
    line-height: 1.45;
    white-space: pre-wrap;
    word-break: break-word;
    color: var(--wm-text-secondary);
}

.empty-log {
    height: 180px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--wm-text-muted);
    font-size: 13px;
}

:deep(.n-button) {
    --n-border-radius: 4px !important;
}

@media (max-width: 560px) {
    .settings-head,
    .setting-row,
    .group-head {
        align-items: stretch;
        flex-direction: column;
    }

    .language-select {
        width: 100%;
    }
}
</style>
