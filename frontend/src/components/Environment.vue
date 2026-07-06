<template>
    <div class="env-entry">
        <n-button
            class="env-button"
            :class="{ 'has-issue': hasIssue }"
            quaternary
            :aria-label="t('environment.title')"
            :title="t('environment.title')"
            @click="store.panels.environment = true"
        >
            <n-icon size="20">
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 32 32"
                    aria-hidden="true"
                >
                    <path
                        d="M6 6h20a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2zm0 2v12h20V8z"
                        fill="currentColor"
                    />
                    <path
                        d="M13 24h6v2h5v2H8v-2h5z"
                        fill="currentColor"
                    />
                </svg>
            </n-icon>

            <span
                v-if="issueCount"
                class="env-badge"
                aria-hidden="true"
            >
                {{ issueCount }}
            </span>
        </n-button>

        <n-modal
            v-model:show="store.panels.environment"
            class="env-modal"
            :auto-focus="false"
        >
            <section
                class="env-card"
                role="dialog"
                aria-modal="true"
                aria-labelledby="env-dialog-title"
            >
                <header class="env-head">
                    <div class="env-logo" aria-hidden="true">
                        <n-icon size="22">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                viewBox="0 0 32 32"
                            >
                                <path
                                    d="M6 6h20a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2zm0 2v12h20V8z"
                                    fill="currentColor"
                                />
                                <path
                                    d="M13 24h6v2h5v2H8v-2h5z"
                                    fill="currentColor"
                                />
                            </svg>
                        </n-icon>
                    </div>

                    <div class="env-heading">
                        <div class="env-title-line">
                            <h2
                                id="env-dialog-title"
                                class="env-title"
                            >
                                {{ t("environment.title") }}
                            </h2>

                            <div
                                class="env-state"
                                :class="{
                                    'is-ready': isReady,
                                    'has-error': hasIssue
                                }"
                                aria-live="polite"
                            >
                                <span class="state-dot"></span>

                                <span>
                                    {{
                                        isReady
                                            ? t("environment.readyFull")
                                            : t("environment.notReadyFull")
                                    }}
                                </span>
                            </div>
                        </div>

                        <p class="env-platform">
                            {{ platformName }}

                            <span class="separator">·</span>

                            {{
                                report.arch ||
                                t("environment.unknown")
                            }}
                        </p>
                    </div>

                    <n-button
                        class="env-close"
                        quaternary
                        circle
                        :aria-label="t('editor.cancel')"
                        :title="t('editor.cancel')"
                        @click="store.panels.environment = false"
                    >
                        <n-icon size="17">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                viewBox="0 0 32 32"
                                aria-hidden="true"
                            >
                                <path
                                    d="M17.4 16 24 9.4 22.6 8 16 14.6 9.4 8 8 9.4l6.6 6.6L8 22.6 9.4 24l6.6-6.6 6.6 6.6 1.4-1.4z"
                                    fill="currentColor"
                                />
                            </svg>
                        </n-icon>
                    </n-button>
                </header>

                <div class="env-body">
                    <div
                        v-if="env.error"
                        class="env-error-message"
                        role="alert"
                    >
                        <n-icon size="16">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                viewBox="0 0 32 32"
                                aria-hidden="true"
                            >
                                <path
                                    d="M16 3 2.5 27h27zm0 5.1L25.2 25H6.8z"
                                    fill="currentColor"
                                />
                                <path
                                    d="M15 12h2v7h-2zm0 9h2v2h-2z"
                                    fill="currentColor"
                                />
                            </svg>
                        </n-icon>

                        <span>{{ env.error }}</span>
                    </div>

                    <div
                        v-if="checks.length"
                        class="env-list"
                    >
                        <article
                            v-for="item in checks"
                            :key="item.id"
                            class="tool-row"
                            :class="getToolClass(item)"
                            :title="getStatusTitle(item)"
                        >
                            <div class="tool-icon">
                                <img
                                    :src="`/env-icons/${item.id}.png`"
                                    :alt="item.name || item.id"
                                    draggable="false"
                                />
                            </div>

                            <div class="tool-content">
                                <div class="tool-heading">
                                    <span class="tool-name">
                                        {{ item.name || item.id }}
                                    </span>

                                    <span class="tool-kind">
                                        {{
                                            item.required
                                                ? t("environment.required")
                                                : t("environment.optional")
                                        }}
                                    </span>
                                </div>

                                <div class="tool-version">
                                    {{
                                        item.version ||
                                        item.path ||
                                        "-"
                                    }}
                                </div>

                                <div class="tool-command">
                                    <code>
                                        {{ item.command || "-" }}
                                    </code>
                                </div>
                            </div>

                            <div class="tool-status">
                                <n-icon size="15">
                                    <svg
                                        v-if="item.found"
                                        xmlns="http://www.w3.org/2000/svg"
                                        viewBox="0 0 32 32"
                                        aria-hidden="true"
                                    >
                                        <path
                                            d="m13 22.4-6.7-6.7-1.4 1.4 8.1 8.1 15-15-1.4-1.4z"
                                            fill="currentColor"
                                        />
                                    </svg>

                                    <svg
                                        v-else-if="item.required"
                                        xmlns="http://www.w3.org/2000/svg"
                                        viewBox="0 0 32 32"
                                        aria-hidden="true"
                                    >
                                        <path
                                            d="M17.4 16 24 9.4 22.6 8 16 14.6 9.4 8 8 9.4l6.6 6.6L8 22.6 9.4 24l6.6-6.6 6.6 6.6 1.4-1.4z"
                                            fill="currentColor"
                                        />
                                    </svg>

                                    <svg
                                        v-else
                                        xmlns="http://www.w3.org/2000/svg"
                                        viewBox="0 0 32 32"
                                        aria-hidden="true"
                                    >
                                        <path
                                            d="M15 8h2v12h-2zm1 15a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3z"
                                            fill="currentColor"
                                        />
                                    </svg>
                                </n-icon>

                                <span class="tool-status-text">
                                    {{ getStatusTitle(item) }}
                                </span>
                            </div>
                        </article>
                    </div>

                    <div
                        v-else
                        class="env-empty"
                    >
                        <div
                            class="empty-icon"
                            aria-hidden="true"
                        >
                            <n-icon size="24">
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    viewBox="0 0 32 32"
                                >
                                    <path
                                        d="M6 6h20a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2zm0 2v12h20V8z"
                                        fill="currentColor"
                                    />
                                    <path
                                        d="M13 24h6v2h5v2H8v-2h5z"
                                        fill="currentColor"
                                    />
                                </svg>
                            </n-icon>
                        </div>

                        <span>
                            {{
                                env.error ||
                                t("environment.noData")
                            }}
                        </span>
                    </div>
                </div>
            </section>
        </n-modal>
    </div>
</template>

<script setup>
import {
    NButton,
    NIcon,
    NModal
} from "naive-ui"
import {
    computed,
    inject,
    unref
} from "vue"
import { useI18n } from "../i18n"

const { t } = useI18n()
const store = inject("store")

const env = computed(() => {
    return unref(store?.env) || {}
})

const report = computed(() => {
    return env.value.data || {}
})

const checks = computed(() => {
    if (!Array.isArray(report.value?.checks)) {
        return []
    }

    return report.value.checks
})

const requiredMissingCount = computed(() => {
    return checks.value.reduce(
        (count, item) => {
            if (
                item.required &&
                !item.found
            ) {
                return count + 1
            }

            return count
        },
        0
    )
})

const isReady = computed(() => {
    if (
        !env.value.loaded ||
        env.value.error
    ) {
        return false
    }

    if (env.value.passed) {
        return true
    }

    return (
        checks.value.length > 0 &&
        requiredMissingCount.value === 0
    )
})

const hasIssue = computed(() => {
    return Boolean(
        env.value.loaded &&
        !isReady.value
    )
})

const issueCount = computed(() => {
    if (!hasIssue.value) {
        return 0
    }

    return (
        requiredMissingCount.value ||
        1
    )
})

const platformName = computed(() => {
    const platformNames = {
        windows: "Windows",
        darwin: "macOS",
        linux: "Linux"
    }

    return (
        platformNames[report.value?.os] ||
        report.value?.os ||
        t("environment.unknown")
    )
})

function getToolClass(item) {
    return {
        "is-found": item.found,
        "is-required-missing":
            !item.found &&
            item.required,
        "is-optional-missing":
            !item.found &&
            !item.required
    }
}

function getStatusTitle(item) {
    if (item.found) {
        return t("environment.found")
    }

    if (item.required) {
        return t(
            "environment.requiredMissing"
        )
    }

    return t(
        "environment.optionalMissing"
    )
}
</script>

<style lang="scss" scoped>
.env-entry {
    display: inline-flex;
    align-items: center;
}

.env-entry :deep(.n-button.env-button) {
    position: relative;
    width: 34px;
    height: 34px;
    padding: 0;
    overflow: visible;
    color: var(--wm-text-secondary) !important;

    --n-border-radius: 10px !important;
    --n-color: var(--wm-control-bg) !important;
    --n-color-hover: var(--wm-control-bg-hover) !important;
    --n-color-pressed: var(--wm-control-bg-active) !important;
    --n-border: 1px solid var(--wm-border-subtle) !important;
    --n-border-hover: 1px solid var(--wm-border-soft) !important;
    --n-border-pressed: 1px solid var(--wm-color-primary-border) !important;
}

.env-entry :deep(.n-button.env-button:hover) {
    color: var(--wm-text-primary) !important;
}

.env-entry :deep(.n-button.env-button.has-issue) {
    color: var(--wm-color-danger) !important;

    --n-color:
        rgba(
            var(--wm-color-danger-rgb),
            0.08
        ) !important;

    --n-color-hover:
        rgba(
            var(--wm-color-danger-rgb),
            0.13
        ) !important;

    --n-border:
        1px solid
        rgba(
            var(--wm-color-danger-rgb),
            0.24
        ) !important;

    --n-border-hover:
        1px solid
        rgba(
            var(--wm-color-danger-rgb),
            0.42
        ) !important;
}

.env-badge {
    position: absolute;
    top: -5px;
    right: -5px;
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    box-sizing: border-box;
    border: 2px solid var(--wm-bg-page);
    border-radius: 999px;
    color: #ffffff;
    background: var(--wm-color-danger);
    font-size: 9px;
    font-weight: 600;
    line-height: 12px;
    text-align: center;
    box-shadow:
        0 3px 8px
        rgba(
            var(--wm-color-danger-rgb),
            0.24
        );
}

.env-card {
    width: min(
        680px,
        calc(100vw - 32px)
    );
    max-height: min(
        720px,
        calc(100vh - 40px)
    );
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-sizing: border-box;
    border: 1px solid var(--wm-border-soft);
    border-radius: 14px;
    background: var(--wm-bg-page);
    box-shadow: var(--wm-shadow-page);
}

.env-head {
    position: relative;
    flex: 0 0 auto;
    min-height: 82px;
    padding: 18px 18px 18px 20px;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    gap: 13px;
    border-bottom:
        1px solid
        var(--wm-border-subtle);
    background:
        linear-gradient(
            180deg,
            var(--wm-surface-2),
            transparent
        );
}

.env-logo {
    width: 42px;
    height: 42px;
    min-width: 42px;
    display: flex;
    align-items: center;
    justify-content: center;
    border:
        1px solid
        var(--wm-color-primary-border);
    border-radius: 11px;
    color: var(--wm-color-primary-hover);
    background:
        rgba(
            var(--wm-color-primary-rgb),
            0.07
        );
    box-shadow:
        inset 0 1px 0
        rgba(255, 255, 255, 0.035),
        0 8px 20px
        rgba(
            var(--wm-color-primary-rgb),
            0.07
        );
}

.env-heading {
    flex: 1;
    min-width: 0;
}

.env-title-line {
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 9px;
}

.env-title {
    min-width: 0;
    margin: 0;
    color: var(--wm-text-primary);
    font-size: 16px;
    font-weight: 600;
    line-height: 22px;
    letter-spacing: 0.01em;
}

.env-state {
    height: 22px;
    padding: 0 8px;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    box-sizing: border-box;
    border:
        1px solid
        var(--wm-border-soft);
    border-radius: 999px;
    color: var(--wm-text-muted);
    background: var(--wm-control-bg);
    font-size: 11px;
    line-height: 20px;
    white-space: nowrap;
}

.env-state.is-ready {
    color: var(--wm-color-success);
    border-color:
        rgba(
            var(--wm-color-success-rgb),
            0.18
        );
    background:
        rgba(
            var(--wm-color-success-rgb),
            0.055
        );
}

.env-state.has-error {
    color: var(--wm-color-danger);
    border-color:
        rgba(
            var(--wm-color-danger-rgb),
            0.22
        );
    background:
        rgba(
            var(--wm-color-danger-rgb),
            0.075
        );
}

.state-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: currentColor;
    opacity: 0.78;
}

.env-platform {
    margin: 5px 0 0;
    overflow: hidden;
    color: var(--wm-text-muted);
    font-size: 12px;
    line-height: 18px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.separator {
    margin: 0 4px;
    opacity: 0.62;
}

.env-entry :deep(.n-button.env-close) {
    width: 30px;
    height: 30px;
    min-width: 30px;
    color: var(--wm-text-muted) !important;

    --n-color-hover:
        var(--wm-control-bg-hover) !important;

    --n-color-pressed:
        var(--wm-control-bg-active) !important;
}

.env-entry :deep(.n-button.env-close:hover) {
    color: var(--wm-text-primary) !important;
}

.env-body {
    flex: 1;
    min-height: 0;
    padding: 14px 18px 18px;
    box-sizing: border-box;
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color:
        var(--wm-border-soft)
        transparent;
}

.env-body::-webkit-scrollbar {
    width: 6px;
}

.env-body::-webkit-scrollbar-track {
    background: transparent;
}

.env-body::-webkit-scrollbar-thumb {
    border-radius: 999px;
    background: var(--wm-border-soft);
}

.env-error-message {
    margin-bottom: 12px;
    padding: 10px 12px;
    display: flex;
    align-items: flex-start;
    gap: 8px;
    box-sizing: border-box;
    border:
        1px solid
        rgba(
            var(--wm-color-danger-rgb),
            0.22
        );
    border-radius: 9px;
    color: var(--wm-color-danger);
    background:
        rgba(
            var(--wm-color-danger-rgb),
            0.07
        );
    font-size: 12px;
    line-height: 18px;
}

.env-error-message :deep(.n-icon) {
    flex: 0 0 auto;
    margin-top: 1px;
}

.env-list {
    overflow: hidden;
    border:
        1px solid
        var(--wm-border-subtle);
    border-radius: 11px;
    background: var(--wm-surface-1);
}

.tool-row {
    position: relative;
    min-width: 0;
    min-height: 68px;
    padding: 11px 12px;
    display: grid;
    grid-template-columns:
        40px
        minmax(0, 1fr)
        auto;
    align-items: center;
    gap: 11px;
    box-sizing: border-box;
    transition:
        background-color 160ms ease,
        border-color 160ms ease;
}

.tool-row + .tool-row {
    border-top:
        1px solid
        var(--wm-border-subtle);
}

.tool-row:hover {
    background: var(--wm-surface-hover);
}

.tool-row.is-found {
    background: transparent;
}

.tool-row.is-found:hover {
    background: var(--wm-surface-hover);
}

.tool-row.is-required-missing {
    background:
        linear-gradient(
            90deg,
            rgba(
                var(--wm-color-danger-rgb),
                0.05
            ),
            transparent 40%
        );
}

.tool-row.is-optional-missing {
    background:
        linear-gradient(
            90deg,
            rgba(
                var(--wm-color-warning-rgb),
                0.04
            ),
            transparent 40%
        );
}

.tool-icon {
    width: 40px;
    height: 40px;
    min-width: 40px;
    padding: 7px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    overflow: hidden;
    border:
        1px solid
        var(--wm-border-subtle);
    border-radius: 10px;
    background: var(--wm-control-bg);
}

.tool-icon img {
    width: 100%;
    height: 100%;
    display: block;
    object-fit: contain;
    user-select: none;
}

.tool-content {
    min-width: 0;
}

.tool-heading {
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 7px;
}

.tool-name {
    min-width: 0;
    overflow: hidden;
    color: var(--wm-text-primary);
    font-size: 13px;
    font-weight: 550;
    line-height: 18px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.tool-kind {
    flex: 0 0 auto;
    padding: 1px 5px;
    border:
        1px solid
        var(--wm-border-subtle);
    border-radius: 5px;
    color: var(--wm-text-muted);
    background: var(--wm-control-bg);
    font-size: 9px;
    line-height: 14px;
}

.tool-version {
    margin-top: 2px;
    overflow: hidden;
    color: var(--wm-text-secondary);
    font-size: 12px;
    line-height: 17px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.tool-command {
    margin-top: 2px;
    overflow: hidden;
    color: var(--wm-text-muted);
    font-size: 10px;
    line-height: 15px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.tool-command code {
    font-family:
        "SFMono-Regular",
        "Cascadia Code",
        "Roboto Mono",
        Consolas,
        monospace;
}

.tool-status {
    min-width: 30px;
    height: 28px;
    padding: 0 8px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
    box-sizing: border-box;
    border:
        1px solid
        var(--wm-border-soft);
    border-radius: 8px;
    color: var(--wm-text-muted);
    background: var(--wm-control-bg);
    font-size: 11px;
    white-space: nowrap;
}

.tool-row.is-found .tool-status {
    color: var(--wm-color-success);
    border-color:
        rgba(
            var(--wm-color-success-rgb),
            0.16
        );
    background:
        rgba(
            var(--wm-color-success-rgb),
            0.045
        );
}

.tool-row.is-required-missing .tool-status {
    color: var(--wm-color-danger);
    border-color:
        rgba(
            var(--wm-color-danger-rgb),
            0.22
        );
    background:
        rgba(
            var(--wm-color-danger-rgb),
            0.075
        );
}

.tool-row.is-optional-missing .tool-status {
    color: var(--wm-color-warning);
    border-color:
        rgba(
            var(--wm-color-warning-rgb),
            0.18
        );
    background:
        rgba(
            var(--wm-color-warning-rgb),
            0.055
        );
}

.tool-status-text {
    line-height: 16px;
}

.env-empty {
    min-height: 190px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 10px;
    border:
        1px dashed
        var(--wm-border-soft);
    border-radius: 11px;
    color: var(--wm-text-muted);
    background: var(--wm-surface-1);
    font-size: 12px;
    text-align: center;
}

.empty-icon {
    width: 44px;
    height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
    border:
        1px solid
        var(--wm-border-subtle);
    border-radius: 12px;
    color: var(--wm-text-muted);
    background: var(--wm-control-bg);
}

@media (max-width: 600px) {
    .env-card {
        width: calc(100vw - 20px);
        max-height: calc(100vh - 20px);
        border-radius: 12px;
    }

    .env-head {
        min-height: 74px;
        padding: 15px;
        gap: 10px;
    }

    .env-logo {
        width: 38px;
        height: 38px;
        min-width: 38px;
    }

    .env-title-line {
        align-items: flex-start;
        flex-direction: column;
        gap: 4px;
    }

    .env-body {
        padding: 13px;
    }

    .tool-row {
        grid-template-columns:
            38px
            minmax(0, 1fr)
            30px;
        padding: 10px;
    }

    .tool-icon {
        width: 38px;
        height: 38px;
        min-width: 38px;
    }

    .tool-status {
        width: 30px;
        min-width: 30px;
        padding: 0;
    }

    .tool-status-text {
        display: none;
    }
}

@media (max-width: 420px) {
    .env-state {
        display: none;
    }

    .tool-kind {
        display: none;
    }
}

@media (prefers-reduced-motion: reduce) {
    .tool-row {
        transition: none;
    }
}
</style>
