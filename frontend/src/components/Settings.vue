<template>
    <n-modal v-model:show="store.panels.settings" class="settings-modal">
        <div class="settings-card">
            <div class="settings-head">
                <div>
                    <div class="settings-title">{{ t("settings.title") }}</div>
                    <div class="settings-subtitle">{{ t("app.title") }}</div>
                </div>

                <n-button quaternary size="small" @click="store.panels.settings = false">
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
                        <div class="log-terminal-head">
                            <span>{{ t("settings.logs") }}</span>
                            <div class="log-head-actions">
                                <span>{{ store.logEntries.length }}</span>
                                <label class="follow-toggle">
                                    <span>{{ t("settings.followLatestLog") }}</span>
                                    <n-switch
                                        v-model:value="followLatestLogs"
                                        size="small"
                                    />
                                </label>
                            </div>
                        </div>

                        <n-scrollbar
                            ref="logScrollbarRef"
                            class="log-scroll"
                        >
                            <div v-if="store.logEntries.length" class="log-lines">
                                <template
                                    v-for="(entry, index) in store.logEntries"
                                    :key="index"
                                >
                                    <div
                                        v-if="isPackageTransactionStart(entry, index)"
                                        class="transaction-divider"
                                    >
                                        <span>{{ entry.transactionTitle }}</span>
                                        <span>{{ entry.transactionId }}</span>
                                    </div>

                                    <div
                                        class="log-line"
                                        :class="[
                                            `level-${logEntryLevel(entry)}`,
                                            { package: entry.transactionId }
                                        ]"
                                    >
                                        <span class="log-dot"></span>
                                        <span class="log-text">{{ entry.line }}</span>
                                        <span
                                            v-if="entry.transactionType"
                                            class="log-tag"
                                        >
                                            {{ entry.transactionType }}
                                        </span>
                                    </div>
                                </template>
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
import { inject, nextTick, ref, watch } from "vue"
import { NButton, NModal, NScrollbar, NSelect, NSwitch, useMessage } from "naive-ui"
import { Service as SettingsService } from "../../bindings/wails3-manager/core/settings"
import { useI18n } from "../i18n"
import { logEntryLevel } from "../utils/logs"

const store = inject("store")
const message = useMessage()
const { t, localeOptions, setLocale } = useI18n()

const saving = ref(false)
const clearingLogs = ref(false)
const followLatestLogs = ref(true)
const logScrollbarRef = ref(null)

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

function isPackageTransactionStart(entry, index) {
    return entry.transactionType === "package" &&
        entry.transactionId &&
        (index === 0 || store.logEntries[index - 1].transactionId !== entry.transactionId)
}

function scrollLogsToBottom() {
    if (!followLatestLogs.value) {
        return
    }
    nextTick(() => {
        logScrollbarRef.value?.scrollTo?.({ top: Number.MAX_SAFE_INTEGER })
    })
}

watch(
    () => store.logEntries.length,
    () => {
        scrollLogsToBottom()
    },
    { flush: "post" }
)

watch(
    () => store.panels.settings,
    (show) => {
        if (show) {
            scrollLogsToBottom()
        }
    },
    { flush: "post" }
)

watch(followLatestLogs, (enabled) => {
    if (enabled) {
        scrollLogsToBottom()
    }
})
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
    height: 240px;
    margin-top: 12px;
    box-sizing: border-box;
    border-radius: 8px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-control-bg);
    overflow: hidden;

    display: flex;
    flex-direction: column;
}

.log-terminal-head {
    flex: 0 0 auto;
    height: 36px;
    padding: 0 12px;
    box-sizing: border-box;
    border-bottom: 1px solid var(--wm-border-subtle);

    display: flex;
    align-items: center;
    justify-content: space-between;

    font-size: 11px;
    color: var(--wm-text-muted);
    background: var(--wm-surface-2);
}

.log-head-actions {
    display: flex;
    align-items: center;
    gap: 12px;
}

.follow-toggle {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--wm-text-muted);
}

.log-scroll {
    flex: 1;
    min-height: 0;
}

.log-lines {
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.log-line {
    position: relative;
    min-height: 22px;
    padding: 3px 8px 3px 22px;
    box-sizing: border-box;
    border-radius: 4px;

    display: flex;
    align-items: flex-start;
    gap: 8px;

    font-family: Consolas, "Courier New", monospace;
    font-size: 12px;
    line-height: 1.45;
    color: var(--wm-text-secondary);
}

.log-line.package {
    background: color-mix(in srgb, var(--wm-color-primary-shadow) 46%, transparent);
}

.log-line.level-command {
    color: #7dd3fc;
    background: color-mix(in srgb, #0284c7 14%, transparent);
}

.log-line.level-success {
    color: #86efac;
    background: color-mix(in srgb, #16a34a 14%, transparent);
}

.log-line.level-warning {
    color: #facc15;
    background: color-mix(in srgb, #ca8a04 16%, transparent);
}

.log-line.level-error {
    color: var(--wm-color-danger-hover);
    background: color-mix(in srgb, var(--wm-color-danger) 16%, transparent);
}

.log-dot {
    position: absolute;
    left: 9px;
    top: 10px;
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--wm-text-muted);
}

.log-line.package .log-dot {
    background: var(--wm-color-primary);
}

.log-line.level-command .log-dot {
    background: #38bdf8;
}

.log-line.level-success .log-dot {
    background: #22c55e;
}

.log-line.level-warning .log-dot {
    background: #eab308;
}

.log-line.level-error .log-dot {
    background: var(--wm-color-danger-hover);
}

.log-text {
    flex: 1;
    min-width: 0;
    white-space: pre-wrap;
    word-break: break-word;
}

.log-tag {
    flex: 0 0 auto;
    height: 18px;
    padding: 0 6px;
    border-radius: 4px;
    border: 1px solid var(--wm-color-primary-border);

    font-size: 10px;
    line-height: 17px;
    color: var(--wm-color-primary-hover);
    background: var(--wm-color-primary-shadow);
}

.transaction-divider {
    height: 28px;
    margin: 4px 0;
    padding: 0 8px;
    box-sizing: border-box;
    border-radius: 4px;
    border: 1px solid var(--wm-color-primary-border);

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;

    font-size: 11px;
    color: var(--wm-color-primary-hover);
    background: var(--wm-color-primary-shadow);
}

.transaction-divider span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.empty-log {
    height: 204px;
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
