<template>
    <n-modal
        v-model:show="visible"
        class="package-run-modal"
        :mask-closable="finished"
        :close-on-esc="finished"
    >
        <div class="package-run-card">
            <div class="package-run-head">
                <div>
                    <div class="package-run-title">{{ t("project.runBuild") }}</div>
                    <div class="package-run-subtitle">{{ transactionId }}</div>
                </div>

                <n-button
                    :disabled="!finished"
                    @click="visible = false"
                >
                    {{ t("project.close") }}
                </n-button>
            </div>

            <div class="package-run-body">
                <div
                    class="package-status"
                    :class="{ 'is-error': error, 'is-success': finished && !error && artifacts.length, 'is-warning': finished && !error && !artifacts.length }"
                    role="status"
                >
                    <n-spin v-if="building" :size="16" />
                    <span>{{ statusText }}</span>
                </div>

                <div v-if="building && currentCommand" class="package-stage">
                    <n-ellipsis>{{ currentCommand }}</n-ellipsis>
                </div>

                <div
                    v-if="error"
                    class="package-error"
                >
                    {{ error }}
                </div>

                <n-button
                    class="logs-toggle"
                    size="small"
                    quaternary
                    :aria-expanded="logsExpanded"
                    aria-controls="package-run-logs"
                    @click="toggleLogs"
                >
                    {{ logsExpanded ? t("project.hidePackageLogs") : t("project.showPackageLogs") }}
                </n-button>

                <div v-show="logsExpanded" id="package-run-logs" class="package-log-box">
                    <n-scrollbar ref="logScrollbarRef" @scroll="handleLogScroll">
                        <div
                            v-if="logEntries.length"
                            class="package-log-lines"
                        >
                            <div
                                v-for="(entry, index) in logEntries"
                                :key="index"
                                class="package-log-line"
                                :class="`level-${logEntryLevel(entry)}`"
                            >
                                {{ entry.line }}
                            </div>
                        </div>

                        <div
                            v-else
                            class="package-log-empty"
                        >
                            {{ t("settings.noLogs") }}
                        </div>
                    </n-scrollbar>
                </div>

                <n-button v-if="logsExpanded && !followLogs" class="latest-logs-button" size="small" @click="resumeLogFollowing">
                    {{ t("project.latestLogs") }}
                </n-button>

                <div
                    v-if="finished && result"
                    class="package-result"
                >
                    <div v-if="warnings.length" class="package-warnings">
                        <div class="result-title">{{ t("project.packageWarnings") }}</div>
                        <ul>
                            <li v-for="(warning, index) in warnings" :key="index">{{ warning }}</li>
                        </ul>
                    </div>

                    <div class="result-section">
                        <div class="result-title">{{ t("project.packageArtifacts") }}</div>

                        <div v-if="artifacts.length" class="artifact-list">
                            <div v-for="artifact in artifacts" :key="artifact.path" class="artifact-item">
                                <div class="artifact-heading">
                                    <strong>{{ artifact.name }}</strong>
                                    <span>{{ formatSize(artifact.size) }}</span>
                                </div>
                                <div class="artifact-path">{{ artifact.path }}</div>
                                <div class="result-actions">
                                    <n-button size="small" @click="$emit('open-path', artifact.path)">
                                        {{ t("project.revealArtifact") }}
                                    </n-button>
                                    <n-button size="small" @click="copyPath(artifact.path)">
                                        {{ t("project.copyPath") }}
                                    </n-button>
                                </div>
                            </div>
                        </div>
                        <div v-else class="artifact-empty">{{ t("project.packageNoArtifactsHint") }}</div>
                    </div>

                    <div v-if="directories.length" class="result-section">
                        <div class="result-title">{{ t("project.outputDirectories") }}</div>

                        <div
                            v-for="item in directories"
                            :key="item.path"
                            class="result-path"
                        >
                            <span>{{ item.label }}</span>
                            <div class="artifact-path">{{ item.path }}</div>
                            <div class="result-actions">
                                <n-button size="small" @click="$emit('open-path', item.path)">
                                    {{ t("project.openInFinder") }}
                                </n-button>
                                <n-button size="small" @click="copyPath(item.path)">
                                    {{ t("project.copyPath") }}
                                </n-button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </n-modal>
</template>

<script setup>
import { computed, nextTick, ref, watch } from "vue"
import { Clipboard } from "@wailsio/runtime"
import { NButton, NEllipsis, NModal, NScrollbar, NSpin, useMessage } from "naive-ui"
import { useI18n } from "../i18n"
import { logEntryLevel } from "../utils/logs"

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    finished: {
        type: Boolean,
        default: false
    },
    building: {
        type: Boolean,
        default: false
    },
    error: {
        type: String,
        default: ""
    },
    transactionId: {
        type: String,
        default: ""
    },
    logEntries: {
        type: Array,
        default: () => []
    },
    result: {
        type: Object,
        default: null
    },
    directories: {
        type: Array,
        default: () => []
    }
})

const emit = defineEmits(["update:show", "open-path"])
const { t } = useI18n()
const message = useMessage()
const logScrollbarRef = ref(null)
const followLogs = ref(true)
const logsExpanded = ref(true)
const artifacts = computed(() => Array.isArray(props.result?.artifacts) ? props.result.artifacts : [])
const warnings = computed(() => Array.isArray(props.result?.warnings) ? props.result.warnings : [])
const statusText = computed(() => {
    if (props.error) return t("project.packageFailed")
    if (!props.finished) return t("project.packageRunning")
    return artifacts.value.length ? t("project.packageSucceeded") : t("project.packageNoArtifacts")
})
const currentCommand = computed(() => {
    for (let index = props.logEntries.length - 1; index >= 0; index--) {
        const line = String(props.logEntries[index]?.line || "")
        if (line.includes("Running command:")) return line
    }
    return ""
})

const visible = computed({
    get() {
        return props.show
    },
    set(value) {
        if (!value && !props.finished) {
            return
        }
        emit("update:show", value)
    }
})

function scrollLogsToBottom() {
    nextTick(() => {
        if (!followLogs.value || !props.show) return
        logScrollbarRef.value?.scrollTo?.({ top: Number.MAX_SAFE_INTEGER })
    })
}

function handleLogScroll(event) {
    const target = event.target
    followLogs.value = target.scrollHeight - target.scrollTop - target.clientHeight <= 24
}

function resumeLogFollowing() {
    followLogs.value = true
    scrollLogsToBottom()
}

function toggleLogs() {
    logsExpanded.value = !logsExpanded.value
    if (logsExpanded.value) scrollLogsToBottom()
}

function formatSize(value) {
    const size = Number(value)
    if (!Number.isFinite(size) || size < 0) return t("project.sizeUnknown")
    const units = ["B", "KiB", "MiB", "GiB", "TiB"]
    const unit = Math.min(Math.floor(Math.log(Math.max(size, 1)) / Math.log(1024)), units.length - 1)
    return `${(size / (1024 ** unit)).toFixed(unit ? 1 : 0)} ${units[unit]}`
}

async function copyPath(path) {
    try {
        await Clipboard.SetText(path)
        message.success(t("project.pathCopied"))
    } catch {
        message.error(t("project.pathCopyFailed"))
    }
}

watch(
    () => props.logEntries.at(-1),
    () => {
        scrollLogsToBottom()
    },
    { flush: "post" }
)

watch(
    () => props.show,
    (show) => {
        if (show) {
            scrollLogsToBottom()
        }
    },
    { flush: "post" }
)

watch(() => props.transactionId, () => {
    followLogs.value = true
    logsExpanded.value = true
    scrollLogsToBottom()
}, { flush: "post" })

watch(() => props.finished, (finished) => {
    if (finished && !props.error && props.result && followLogs.value) logsExpanded.value = false
}, { flush: "post" })
</script>

<style lang="scss" scoped>
.package-run-card {
    width: min(760px, calc(100vw - 36px));
    max-height: min(720px, calc(100vh - 48px));
    border-radius: 8px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-bg-page);
    overflow: hidden;

    display: flex;
    flex-direction: column;
}

.package-run-head {
    height: 64px;
    min-height: 64px;
    padding: 0 18px;
    box-sizing: border-box;
    border-bottom: 1px solid var(--wm-border-subtle);

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.package-run-title {
    font-size: 16px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.package-run-subtitle {
    margin-top: 4px;
    font-size: 12px;
    color: var(--wm-text-muted);
}

.package-run-body {
    flex: 1;
    min-height: 0;
    padding: 16px 18px 18px;
    box-sizing: border-box;
    overflow-y: auto;
}

.package-status {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 13px;
    color: var(--wm-text-secondary);
}

.package-status.is-success {
    color: var(--wm-color-success);
}

.package-status.is-warning {
    color: var(--wm-color-warning);
}

.package-status.is-error {
    color: var(--wm-color-danger);
}

.package-stage {
    margin-top: 6px;
    font-size: 12px;
    color: var(--wm-text-muted);
}

.latest-logs-button {
    margin-top: 8px;
}

.logs-toggle {
    margin-top: 12px;
}

.package-warnings {
    padding: 10px 12px;
    border-radius: 6px;
    border: 1px solid color-mix(in srgb, var(--wm-color-warning) 40%, transparent);
    background: color-mix(in srgb, var(--wm-color-warning) 12%, transparent);
    font-size: 12px;
    color: var(--wm-color-warning);

    ul {
        margin: 0;
        padding-left: 18px;
        overflow-wrap: anywhere;
    }

    li + li {
        margin-top: 5px;
    }
}

.package-error {
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    margin-top: 12px;
    padding: 10px 12px;
    border-radius: 6px;
    border: 1px solid color-mix(in srgb, var(--wm-color-danger) 40%, transparent);
    color: var(--wm-color-danger);
    background: color-mix(in srgb, var(--wm-color-danger) 12%, transparent);
    font-size: 12px;
}

.package-log-box {
    height: 260px;
    margin-top: 14px;
    border-radius: 8px;
    border: 1px solid var(--wm-border-subtle);
    background: color-mix(in srgb, var(--wm-control-bg) 86%, #000 14%);
    overflow: hidden;
}

.package-log-lines {
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 5px;
}

.package-log-line {
    min-height: 22px;
    padding: 3px 8px;
    border-radius: 4px;
    font-family: Consolas, "Courier New", monospace;
    font-size: 12px;
    line-height: 1.45;
    color: var(--wm-text-secondary);
    white-space: pre-wrap;
    word-break: break-word;
}

.package-log-line.level-command {
    color: var(--wm-text-primary);
    background: color-mix(in srgb, #0284c7 14%, transparent);
}

.package-log-line.level-success {
    color: var(--wm-color-success);
    background: color-mix(in srgb, #16a34a 14%, transparent);
}

.package-log-line.level-warning {
    color: var(--wm-color-warning);
    background: color-mix(in srgb, #ca8a04 16%, transparent);
}

.package-log-line.level-error {
    color: var(--wm-color-danger-hover);
    background: color-mix(in srgb, var(--wm-color-danger) 16%, transparent);
}

.package-log-line.level-package {
    background: color-mix(in srgb, #333333 16%, transparent);
}

.package-log-empty {
    height: 260px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--wm-text-muted);
    font-size: 13px;
}

.package-result {
    margin-top: 14px;
    display: flex;
    gap: 12px;
    flex-direction: column;
    width: 100%;
}

.result-section {
    min-width: 0;
}

.result-title {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--wm-text-muted);
}

.result-path,
.artifact-item {
    transition: 200ms;
    width: 100%;
    min-width: 0;
    min-height: 38px;
    padding: 8px 10px;
    box-sizing: border-box;
    border-radius: 6px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-control-bg);

    display: flex;
    flex-direction: column;
    gap: 5px;

    font-size: 12px;
    color: var(--wm-text-secondary);
}

.result-path + .result-path,
.artifact-item + .artifact-item {
    margin-top: 8px;
}

.result-path span,
.artifact-item span {
    font-size: 11px;
    color: var(--wm-text-muted);
}

.artifact-list {
    max-height: 220px;
    overflow-y: auto;
}

.artifact-heading {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;

    strong {
        overflow-wrap: anywhere;
        font-weight: 500;
        color: var(--wm-text-primary);
    }

    span {
        flex-shrink: 0;
    }
}

.artifact-path {
    overflow-wrap: anywhere;
    user-select: text;
}

.result-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    margin-top: 3px;
}

.artifact-empty {
    min-height: 38px;
    padding: 9px 10px;
    box-sizing: border-box;
    border-radius: 6px;
    border: 1px dashed var(--wm-border-soft);
    color: var(--wm-text-muted);
    font-size: 12px;
}
</style>
