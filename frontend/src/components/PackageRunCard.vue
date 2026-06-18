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
                <n-progress
                    type="line"
                    :percentage="progress"
                    :processing="building"
                    :status="error ? 'error' : 'success'"
                />

                <div
                    v-if="error"
                    class="package-error"
                >
                    {{ error }}
                </div>

                <div class="package-log-box">
                    <n-scrollbar ref="logScrollbarRef">
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

                <div
                    v-if="finished && result"
                    class="package-result"
                >
                    <div class="result-section">
                        <div class="result-title">{{ t("project.outputDirectories") }}</div>

                        <div
                            v-for="item in directories"
                            :key="item.path"
                            class="result-path"
                            @click="$emit('open-path', item.path)"
                        >
                            <span>{{ item.label }}</span>
                            <n-ellipsis>{{ item.path }}</n-ellipsis>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </n-modal>
</template>

<script setup>
import { computed, nextTick, ref, watch } from "vue"
import { NButton, NEllipsis, NModal, NProgress, NScrollbar } from "naive-ui"
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
    progress: {
        type: Number,
        default: 0
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
const logScrollbarRef = ref(null)

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
        logScrollbarRef.value?.scrollTo?.({ top: Number.MAX_SAFE_INTEGER })
    })
}

watch(
    () => props.logEntries.length,
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
}

.package-error {
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
    color: #7dd3fc;
    background: color-mix(in srgb, #0284c7 14%, transparent);
}

.package-log-line.level-success {
    color: #86efac;
    background: color-mix(in srgb, #16a34a 14%, transparent);
}

.package-log-line.level-warning {
    color: #facc15;
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

.result-path {
    cursor: pointer;
}

.result-path + .result-path,
.artifact-item + .artifact-item {
    margin-top: 8px;
}

.result-path:hover {
    border-color: var(--wm-color-primary-border);
    color: var(--wm-color-primary-hover);
}

.result-path span,
.artifact-item span {
    font-size: 11px;
    color: var(--wm-text-muted);
}

.artifact-list {
    max-height: 108px;
    overflow-y: auto;
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
