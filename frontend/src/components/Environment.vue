<template>
    <div class="env-container">
        <div class="env-head">
            <div class="env-logo">
                <slot name="icon">
                    <n-icon size="26">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                            <path d="M6 6h20a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2zm0 2v12h20V8z" fill="currentColor"></path>
                            <path d="M13 24h6v2h5v2H8v-2h5z" fill="currentColor"></path>
                        </svg>
                    </n-icon>
                </slot>
            </div>

            <div class="env-main">
                <div class="env-title-row">
                    <div class="env-title">{{ t('environment.system') }}</div>

                    <div class="env-ready" :class="{ ok: isReady, error: !isReady }">
                        <span class="ready-dot"></span>
                        <span>{{ isReady ? t('environment.readyFull') : t('environment.notReadyFull') }}</span>
                    </div>
                </div>

                <div class="env-desc">
                    {{ platformName }} · {{ report.arch || t('environment.unknown') }}
                </div>
            </div>

            <div class="env-ring" :class="{ ok: isReady, error: !isReady }">
                <n-progress
                    class="env-progress"
                    type="circle"
                    :percentage="progressPercent"
                    :show-indicator="false"
                    :stroke-width="6"
                    :color="progressColor"
                    :rail-color="progressRailColor"
                />

                <div class="ring-text">
                    {{ foundCount }}/{{ checks.length }}
                </div>
            </div>
        </div>

        <div class="env-list" v-if="checks.length">
            <div
                class="tool-card"
                v-for="item in checks"
                :key="item.id"
                :class="{
                    ok: item.found,
                    error: !item.found && item.required,
                    warn: !item.found && !item.required
                }"
                :title="getStatusTitle(item)"
            >
                <div class="tool-icon">
                    <slot :name="`tool-${item.id}`" :item="item">
                        <div class="tool-icon-placeholder">
                            <img style="width: 100%;height: 100%;object-fit: contain" :src="'/env-icons/'+item.id+'.png'" alt="">
                        </div>
                    </slot>
                </div>

                <div class="tool-content">
                    <div class="tool-name">
                        {{ item.name || item.id }}
                    </div>

                    <div class="tool-meta">
                        <span>{{ item.command || '-' }}</span>
                        <span class="meta-dot"></span>
                        <span>{{ item.required ? t('environment.required') : t('environment.optional') }}</span>
                    </div>

                    <div class="tool-version">
                        {{ item.version || '-' }}
                    </div>
                </div>

                <div class="tool-status">
                    <n-icon size="16">
                        <svg v-if="item.found" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                            <path d="M13 22.4L6.3 15.7L4.9 17.1L13 25.2L28 10.2L26.6 8.8z" fill="currentColor"></path>
                        </svg>

                        <svg v-else-if="item.required" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                            <path d="M17.4 16L24 9.4L22.6 8L16 14.6L9.4 8L8 9.4L14.6 16L8 22.6L9.4 24L16 17.4L22.6 24L24 22.6z" fill="currentColor"></path>
                        </svg>

                        <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                            <path d="M17 8h-2v12h2z" fill="currentColor"></path>
                            <path d="M16 23a1.5 1.5 0 1 0 0 3a1.5 1.5 0 0 0 0-3z" fill="currentColor"></path>
                        </svg>
                    </n-icon>
                </div>
            </div>
        </div>

        <div class="empty" v-else>
            {{ t('environment.noData') }}
        </div>
    </div>
</template>

<script setup>
import { NIcon, NProgress } from "naive-ui"
import { computed, inject, unref } from "vue"
import { useI18n } from "../i18n"

const { t } = useI18n()
const store = inject("store")

const report = computed(() => {
    const env = unref(store?.env)

    if (!env) {
        return {}
    }

    return env.data || env
})

const checks = computed(() => {
    if (!Array.isArray(report.value?.checks)) {
        return []
    }

    return report.value.checks
})

const foundCount = computed(() => {
    return checks.value.filter(item => item.found).length
})

const requiredMissingCount = computed(() => {
    return checks.value.filter(item => item.required && !item.found).length
})

const isReady = computed(() => {
    return checks.value.length > 0 && requiredMissingCount.value === 0
})

const progressPercent = computed(() => {
    if (!checks.value.length) {
        return 0
    }

    return Math.round((foundCount.value / checks.value.length) * 100)
})

const progressColor = computed(() => {
    return isReady.value ? "var(--wm-color-primary)" : "var(--wm-color-danger)"
})

const progressRailColor = computed(() => {
    return "var(--wm-control-bg)"
})

const platformName = computed(() => {
    const names = {
        windows: "Windows",
        darwin: "macOS",
        linux: "Linux"
    }

    return names[report.value?.os] || report.value?.os || t('environment.unknown')
})

function getStatusTitle(item) {
    if (item.found) {
        return t('environment.found')
    }

    if (item.required) {
        return t('environment.requiredMissing')
    }

    return t('environment.optionalMissing')
}

function getToolLetter(item) {
    const name = item.name || item.id || "?"
    return name.slice(0, 1).toUpperCase()
}
</script>

<style lang="scss" scoped>
.env-container {
    width: 100%;
    height: 100%;
    min-height: 0;
    box-sizing: border-box;
    padding: 18px;
    border-radius: 18px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-surface-1);
    backdrop-filter: blur(18px);

    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.env-head {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 14px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--wm-border-subtle);
}

.env-logo {
    width: 52px;
    height: 52px;
    min-width: 52px;
    border-radius: 17px;

    display: flex;
    align-items: center;
    justify-content: center;

    color: var(--wm-color-primary-hover);
    background:
        radial-gradient(circle at 35% 20%, var(--wm-color-primary-shadow), transparent 58%),
        var(--wm-control-bg);
    border: 1px solid var(--wm-color-primary-border);
}

.env-main {
    min-width: 0;
    flex: 1;
}

.env-title-row {
    display: flex;
    align-items: center;
    gap: 9px;
}

.env-title {
    font-size: 16px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.env-ready {
    height: 24px;
    padding: 0 8px;
    border-radius: 999px;

    display: flex;
    align-items: center;
    gap: 6px;

    font-size: 12px;
    line-height: 20px;
    color: var(--wm-text-muted);
    background: var(--wm-control-bg);
    border: 1px solid var(--wm-border-soft);
}

.env-ready.ok {
    color: var(--wm-color-primary-hover);
    background: var(--wm-color-primary-shadow);
    border-color: var(--wm-color-primary-border);
}

.env-ready.error {
    color: var(--wm-color-danger);
    background: rgba(217, 45, 32, 0.10);
    border-color: rgba(217, 45, 32, 0.22);
}

.ready-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: currentColor;
}

.env-desc {
    margin-top: 6px;
    font-size: 12px;
    color: var(--wm-text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.env-ring {
    position: relative;
    width: 48px;
    height: 48px;
    min-width: 48px;

    display: flex;
    align-items: center;
    justify-content: center;
}

.env-progress {
    width: 48px;
    height: 48px;
}

.env-progress :deep(svg) {
    width: 48px;
    height: 48px;
}

.ring-text {
    position: absolute;
    inset: 0;
    pointer-events: none;

    display: flex;
    align-items: center;
    justify-content: center;

    font-size: 11px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.env-list {
    flex: 1;
    min-height: 0;
    margin-top: 14px;
    overflow-y: auto;

    display: flex;
    flex-direction: column;
    gap: 10px;
    padding-right: 6px;
}

.env-list::-webkit-scrollbar {
    width: 3px;
}

.env-list::-webkit-scrollbar-thumb {
    border-radius: 999px;
    background: var(--wm-border-soft);
}

.tool-card {
    min-width: 0;
    padding: 12px;
    border-radius: 15px;
    background: var(--wm-surface-2);
    border: 1px solid var(--wm-border-subtle);

    display: flex;
    align-items: center;
    gap: 12px;

    transition:
        background 0.18s ease,
        border-color 0.18s ease,
        transform 0.18s ease;
}

.tool-card:hover {
    background: var(--wm-surface-hover);
    border-color: var(--wm-border-soft);
}

.tool-card.error {
    border-color: rgba(217, 45, 32, 0.24);
}

.tool-card.warn {
    opacity: 0.76;
}

.tool-icon {
    width: 42px;
    height: 42px;
    min-width: 42px;
    border-radius: 13px;

    display: flex;
    align-items: center;
    justify-content: center;

    background: var(--wm-control-bg);
    border: 1px solid var(--wm-border-soft);
    overflow: hidden;
}

.tool-icon-placeholder {
    display: flex;
    box-sizing: border-box;
    padding: 6px;
    color: var(--wm-text-muted);
}

.tool-content {
    min-width: 0;
    flex: 1;
}

.tool-name {
    min-width: 0;
    font-size: 13px;
    color: var(--wm-text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.tool-meta {
    min-width: 0;


    display: flex;
    align-items: center;
    gap: 6px;

    font-size: 10px;
    color: var(--wm-text-muted);
    overflow: hidden;
    white-space: nowrap;
}

.tool-meta span:first-child {
    overflow: hidden;
    text-overflow: ellipsis;
}

.meta-dot {
    width: 3px;
    height: 3px;
    min-width: 3px;
    border-radius: 50%;
    background: var(--wm-text-muted);
    opacity: 0.7;
}

.tool-version {

    font-size: 12px;
    color: var(--wm-text-secondary);

    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.tool-status {
    width: 28px;
    height: 28px;
    min-width: 28px;
    border-radius: 9px;

    display: flex;
    align-items: center;
    justify-content: center;
    align-self: center;

    color: var(--wm-text-muted);
    background: var(--wm-control-bg);
    border: 1px solid var(--wm-border-soft);
}

.tool-card.ok .tool-status {
    color: var(--wm-color-primary-hover);
    background: var(--wm-color-primary-shadow);
    border-color: var(--wm-color-primary-border);
}

.tool-card.error .tool-status {
    color: var(--wm-color-danger);
    background: rgba(217, 45, 32, 0.10);
    border-color: rgba(217, 45, 32, 0.24);
}

.empty {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;

    color: var(--wm-text-muted);
    font-size: 13px;
}
</style>