<template>
    <div class="wails3-config">
        <div class="config-head">
            <div>
                <div class="config-title">{{ t("editor.basicConfig") }}</div>
                <div class="config-desc">{{ t("editor.basicConfigDesc") }}</div>
            </div>

            <div class="head-actions">
                <div class="change-state">
                    <span class="state-dot" :class="{ active: isDirty }"></span>
                    <span>{{ isDirty ? t("editor.hasChanges") : t("editor.noChanges") }}</span>
                </div>

                <n-button
                    type="primary"
                    :disabled="!isDirty || saving"
                    :loading="saving"
                    @click="saveWails3Cfg"
                >
                    {{ t("editor.save") }}
                </n-button>
            </div>
        </div>

        <div class="config-body">
            <div class="icon-panel">
                <div class="panel-head">
                    <div class="panel-title">{{ t("editor.appIcon") }}</div>
<!--                    <div class="panel-desc">{{ t("editor.iconPanelDesc") }}</div>-->
                </div>

                <div class="icon-area">
                    <div class="icon-stage">
                        <n-image
                            v-if="currentIcon"
                            class="product-icon"
                            object-fit="contain"
                            :src="currentIcon"
                            preview-disabled
                        />

                        <div v-else class="icon-empty">
                            <n-icon size="36">
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                    <path d="M26 4H6a2 2 0 0 0-2 2v20a2 2 0 0 0 2 2h20a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2zM6 6h20v13.17l-4.59-4.58a2 2 0 0 0-2.82 0L16 17.17l-5.59-5.58a2 2 0 0 0-2.82 0L6 13.17zm0 20v-9.99l3-3L19.99 24H6zm20 0h-3.17l-5.42-5.41L20 18l6 6z" fill="currentColor"></path>
                                </svg>
                            </n-icon>
                        </div>

                        <div v-if="icon.newIcon" class="new-badge">
                            {{ t("editor.newIcon") }}
                        </div>
                    </div>
                </div>

                <div class="icon-footer">

                    <div class="icon-actions">
                        <n-button
                            block
                            :disabled="saving"
                            @click="chooseIcon"
                        >
                            {{ icon.newIcon ? t("editor.reselectIcon") : t("editor.selectIcon") }}
                        </n-button>

                        <n-button
                            v-if="icon.newIcon"
                            block
                            quaternary
                            :disabled="saving"
                            @click="cancelIcon"
                        >
                            {{ t("editor.cancel") }}
                        </n-button>
                    </div>
                </div>
            </div>

            <div class="info-panel">
                <div class="panel-head">
                    <div class="panel-title">{{ t("editor.basicInfo") }}</div>
<!--                    <div class="panel-desc">{{ t("editor.basicInfoDesc") }}</div>-->
                </div>

                <div class="form-grid">
                    <div class="input-area">
                        <div class="label">{{ t("editor.productName") }}</div>
                        <n-input
                            v-model:value="cfg.productName"
                            :placeholder="t('editor.productNamePlaceholder')"
                        />
                    </div>

                    <div class="input-area">
                        <div class="label">{{ t("editor.version") }}</div>
                        <n-input
                            v-model:value="cfg.version"
                            :placeholder="t('editor.versionPlaceholder')"
                        />
                    </div>

                    <div class="input-area">
                        <div class="label">{{ t("editor.companyName") }}</div>
                        <n-input
                            v-model:value="cfg.companyName"
                            :placeholder="t('editor.companyNamePlaceholder')"
                        />
                    </div>

                    <div class="input-area">
                        <div class="label">{{ t("editor.productIdentifier") }}</div>
                        <n-input
                            v-model:value="cfg.productIdentifier"
                            :placeholder="t('editor.productIdentifierPlaceholder')"
                        />
                    </div>

                    <div class="input-area full">
                        <div class="label">{{ t("editor.description") }}</div>
                        <n-input
                            v-model:value="cfg.description"
                            type="textarea"
                            :autosize="{ minRows: 4, maxRows: 7 }"
                            :placeholder="t('editor.descriptionPlaceholder')"
                        />
                    </div>

                    <div class="input-area full">
                        <div class="label">{{ t("editor.copyright") }}</div>
                        <n-input
                            v-model:value="cfg.copyright"
                            :placeholder="t('editor.copyrightPlaceholder')"
                        />
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { useI18n } from "../../i18n"
import { computed, onMounted, ref } from "vue"
import { NButton, NIcon, NImage, NInput, useMessage } from "naive-ui"
import { ProjectService } from "../../../bindings/wails3-manager/core/project"
import { AppService } from "../../../bindings/wails3-manager/desktop"

const props = defineProps({
    wails3Cfg: {
        type: Object,
        required: true
    },
})

const { t } = useI18n()
const message = useMessage()

const cfg = ref({})
const originalCfgJson = ref("")

const icon = ref({
    oldIcon: "",
    newIcon: "",
})

const saving = ref(false)

const currentIcon = computed(() => {
    return icon.value.newIcon || icon.value.oldIcon
})

const isDirty = computed(() => {
    return JSON.stringify(cfg.value) !== originalCfgJson.value || !!icon.value.newIcon
})

onMounted(async () => {
    cfg.value = cloneData(props.wails3Cfg.project.wailsConfig.info || {})
    originalCfgJson.value = JSON.stringify(cfg.value)

    if (props.wails3Cfg.iconPath) {
        icon.value.oldIcon = await imageUrlToBase64("/local/file?" + props.wails3Cfg.iconPath)
    }
})

function chooseIcon() {
    AppService.ChooseIcon().then((res)=>{
        if (res) {
            setNewIconByUrl("/local/file?"+res)
        }
    })
}

async function setNewIconByUrl(url) {
    icon.value.newIcon = await imageUrlToBase64(url)
}

function cancelIcon() {
    icon.value.newIcon = ""
}

async function saveWails3Cfg() {
    if (!isDirty.value || saving.value) return

    try {
        saving.value = true

        const nextProject = cloneData(props.wails3Cfg)
        nextProject.project.wailsConfig.info = cloneData(cfg.value)

        await ProjectService.SaveProject(nextProject)

        if (icon.value.newIcon) {
            await ProjectService.ReplaceProjectIcon(
                props.wails3Cfg.projectDir,
                icon.value.newIcon
            )

            icon.value.oldIcon = icon.value.newIcon
            icon.value.newIcon = ""
        }

        props.wails3Cfg.project.wailsConfig.info = cloneData(cfg.value)
        originalCfgJson.value = JSON.stringify(cfg.value)

        message.success(t("editor.saveSuccess"))
    } catch (error) {
        message.error(error?.message || String(error))
    } finally {
        saving.value = false
    }
}

function cloneData(data) {
    return JSON.parse(JSON.stringify(data || {}))
}

function imageUrlToBase64(url) {
    return new Promise(async (resolve, reject) => {
        try {
            if (!url) {
                reject(new Error(t("editor.imageUrlRequired")))
                return
            }

            if (url.startsWith("data:image")) {
                resolve(url)
                return
            }

            const response = await fetch(url)

            if (!response.ok) {
                reject(new Error(t("editor.imageLoadFailed", { status: response.status })))
                return
            }

            const blob = await response.blob()
            const reader = new FileReader()

            reader.onloadend = () => {
                resolve(reader.result)
            }

            reader.onerror = () => {
                reject(new Error(t("editor.imageBase64Failed")))
            }

            reader.readAsDataURL(blob)
        } catch (error) {
            reject(error)
        }
    })
}


</script>

<style lang="scss" scoped>
.wails3-config {
    width: 100%;
    height: 100%;
    min-height: 0;
    box-sizing: border-box;

    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.config-head {
    flex: 0 0 auto;

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.config-title {
    font-size: 17px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.config-desc {
    margin-top: 6px;
    font-size: 12px;
    color: var(--wm-text-muted);
}

.head-actions {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 14px;
}

.change-state {
    display: flex;
    align-items: center;
    gap: 8px;

    font-size: 12px;
    color: var(--wm-text-muted);
    white-space: nowrap;
}

.state-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--wm-text-muted);
}

.state-dot.active {
    background: var(--wm-color-primary);
    box-shadow: 0 0 0 4px var(--wm-color-primary-shadow);
}

.config-body {
    flex: 1;
    min-height: 0;
    margin-top: 18px;

    display: grid;
    grid-template-columns: 280px minmax(0, 1fr);
    gap: 18px;
    overflow: hidden;
}

.icon-panel,
.info-panel {
    min-height: 0;
    box-sizing: border-box;
    border-radius: 18px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-surface-2);
}

.icon-panel {
    padding: 18px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.info-panel {
    padding: 18px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.panel-head {
    flex: 0 0 auto;
}

.panel-title {
    font-size: 15px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.panel-desc {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.5;
    color: var(--wm-text-muted);
}

.icon-area {
    flex: 1;

    display: flex;
    align-items: center;
    justify-content: center;
}

.icon-stage {
    position: relative;
    width: 220px;
    height: 220px;
    border-radius: 26px;
    overflow: hidden;

    display: flex;
    align-items: center;
    justify-content: center;

    background:
        radial-gradient(circle at 30% 18%, var(--wm-color-primary-shadow), transparent 58%),
        var(--wm-control-bg);
    border: 1px solid var(--wm-border-soft);
}

.product-icon {
    width: 168px;
    height: 168px;
}

:deep(.product-icon img) {
    width: 168px;
    height: 168px;
    object-fit: contain;
}

.icon-empty {
    width: 86px;
    height: 86px;
    border-radius: 26px;

    display: flex;
    align-items: center;
    justify-content: center;

    color: var(--wm-text-muted);
    background: var(--wm-control-bg);
    border: 1px solid var(--wm-border-soft);
}

.new-badge {
    position: absolute;
    left: 12px;
    bottom: 12px;
    z-index: 2;

    height: 22px;
    padding: 0 9px;
    border-radius: 999px;

    font-size: 11px;
    line-height: 22px;

    color: var(--wm-text-inverse);
    background: linear-gradient(
        180deg,
        var(--wm-color-logo-start) 0%,
        var(--wm-color-primary) 100%
    );
    box-shadow: var(--wm-shadow-primary);
}

.icon-footer {
    flex: 0 0 auto;
    padding-top: 16px;
}

.icon-status {
    margin-bottom: 12px;

    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;

    font-size: 12px;
    color: var(--wm-text-muted);
}

.icon-status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--wm-text-muted);
}

.icon-status-dot.active {
    background: var(--wm-color-primary);
    box-shadow: 0 0 0 4px var(--wm-color-primary-shadow);
}

.icon-actions {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.form-grid {
    flex: 1;
    min-height: 0;
    margin-top: 6px;
    overflow-y: auto;

    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    align-content: start;
    gap: 16px;
    padding-right: 4px;
}

.form-grid::-webkit-scrollbar {
    width: 6px;
}

.form-grid::-webkit-scrollbar-thumb {
    border-radius: 999px;
    background: var(--wm-border-soft);
}

.input-area {
    min-width: 0;
}

.input-area.full {
    grid-column: 1 / -1;
}

.label {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--wm-text-muted);
}

:deep(.n-input) {
    --n-color: var(--wm-control-bg) !important;
    --n-color-focus: var(--wm-control-bg-hover) !important;
    --n-border: 1px solid var(--wm-border-soft) !important;
    --n-border-hover: 1px solid var(--wm-color-primary-border) !important;
    --n-border-focus: 1px solid var(--wm-border-strong) !important;
    --n-box-shadow-focus: 0 0 0 2px var(--wm-color-primary-shadow) !important;
    --n-text-color: var(--wm-text-secondary) !important;
    --n-placeholder-color: var(--wm-placeholder) !important;
}

:deep(.n-button) {
    --n-border-radius: 4px !important;
    --n-text-color: var(--wm-text-secondary) !important;
    --n-text-color-hover: var(--wm-color-primary-hover) !important;
    --n-text-color-pressed: var(--wm-color-primary-pressed) !important;
    --n-color: var(--wm-control-bg) !important;
    --n-color-hover: var(--wm-control-bg-hover) !important;
    --n-color-pressed: var(--wm-control-bg-active) !important;
    --n-border: 1px solid var(--wm-border-soft) !important;
    --n-border-hover: 1px solid var(--wm-color-primary-border) !important;
    --n-border-pressed: 1px solid var(--wm-border-strong) !important;

    backdrop-filter: blur(12px);
}

:deep(.n-button--primary-type) {
    --n-color: linear-gradient(180deg, var(--wm-color-logo-start) 0%, var(--wm-color-primary) 100%) !important;
    --n-color-hover: linear-gradient(180deg, var(--wm-color-primary-hover) 0%, var(--wm-color-primary) 100%) !important;
    --n-color-pressed: linear-gradient(180deg, var(--wm-color-primary) 0%, var(--wm-color-primary-pressed) 100%) !important;
    --n-border: 1px solid var(--wm-color-primary-border) !important;
    --n-border-hover: 1px solid var(--wm-border-strong) !important;
    --n-border-pressed: 1px solid var(--wm-color-primary-border) !important;
    --n-text-color: var(--wm-text-inverse) !important;
    --n-text-color-hover: var(--wm-text-inverse) !important;
    --n-text-color-pressed: var(--wm-text-inverse) !important;
}

:deep(.n-button--disabled) {
    opacity: 0.5;
    box-shadow: none;
}
</style>
