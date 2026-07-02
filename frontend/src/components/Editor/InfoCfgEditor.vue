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
                </div>

                <div class="form-grid">
                    <div
                        v-for="field in infoFields"
                        :key="field.path"
                        class="input-area"
                        :class="{ full: field.full }"
                    >
                        <div class="label">{{ field.label }}</div>
                        <n-input
                            v-model:value="cfg[field.path]"
                            v-bind="field.props || {}"
                            :status="visibleFieldError(field.path) ? 'error' : undefined"
                            :placeholder="field.placeholder"
                            @blur="touchField(field.path)"
                        />
                        <div
                            v-if="visibleFieldError(field.path)"
                            class="field-error"
                        >
                            {{ visibleFieldError(field.path) }}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { useI18n } from "../../i18n"
import { computed, onMounted, reactive, ref } from "vue"
import { NButton, NIcon, NImage, NInput, useMessage } from "naive-ui"
import { Service as ProjectService } from "../../../bindings/wails3-manager/core/project"
import { AppService } from "../../../bindings/wails3-manager/desktop"
import { localFileUrl } from "../../utils/files"

const props = defineProps({
    wails3Cfg: {
        type: Object,
        required: true
    },
})
const emit = defineEmits(["saved"])

const { t } = useI18n()
const message = useMessage()

const cfg = ref({})
const originalCfgJson = ref("")
const touchedFields = reactive({})

const icon = ref({
    oldIcon: "",
    newIcon: "",
})

const saving = ref(false)
const requiredFields = [
    "productName",
    "version",
    "companyName",
    "productIdentifier",
    "description",
    "copyright"
]

const infoFields = computed(() => [
    {
        path: "productName",
        label: t("editor.productName"),
        placeholder: t("editor.productNamePlaceholder")
    },
    {
        path: "version",
        label: t("editor.version"),
        placeholder: t("editor.versionPlaceholder")
    },
    {
        path: "companyName",
        label: t("editor.companyName"),
        placeholder: t("editor.companyNamePlaceholder")
    },
    {
        path: "productIdentifier",
        label: t("editor.productIdentifier"),
        placeholder: t("editor.productIdentifierPlaceholder")
    },
    {
        path: "description",
        label: t("editor.description"),
        placeholder: t("editor.descriptionPlaceholder"),
        full: true,
        props: {
            type: "textarea",
            autosize: {
                minRows: 3,
                maxRows: 5
            }
        }
    },
    {
        path: "comments",
        label: t("createProject.fields.productComments.label"),
        placeholder: t("createProject.fields.productComments.placeholder"),
        full: true,
        props: {
            type: "textarea",
            autosize: {
                minRows: 2,
                maxRows: 4
            }
        }
    },
    {
        path: "copyright",
        label: t("editor.copyright"),
        placeholder: t("editor.copyrightPlaceholder"),
        full: true
    }
])

const currentIcon = computed(() => {
    return icon.value.newIcon || icon.value.oldIcon
})

const isDirty = computed(() => {
    return JSON.stringify(cfg.value) !== originalCfgJson.value || !!icon.value.newIcon
})

const firstFieldError = computed(() => {
    for (const field of infoFields.value) {
        const error = fieldError(field.path)
        if (error) {
            return error
        }
    }
    return ""
})

onMounted(async () => {
    cfg.value = cloneData(props.wails3Cfg.project.wailsConfig.info || {})
    originalCfgJson.value = JSON.stringify(cfg.value)

    if (props.wails3Cfg.iconPath) {
        icon.value.oldIcon = await imageUrlToBase64(localFileUrl(props.wails3Cfg.iconPath))
    }
})

function chooseIcon() {
    AppService.ChooseIcon().then((res)=>{
        if (res) {
            setNewIconByUrl(localFileUrl(res))
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
    if (saving.value) return
    if (firstFieldError.value) {
        touchAllFields()
        message.error(firstFieldError.value)
        return
    }
    if (!isDirty.value) return

    try {
        saving.value = true

        const nextProject = cloneData(props.wails3Cfg)
        nextProject.project.wailsConfig.info = cloneData(cfg.value)

        let savedProject = await ProjectService.SaveProject(nextProject)
        let iconChanged = false

        if (icon.value.newIcon) {
            savedProject = await ProjectService.ReplaceProjectIcon(
                props.wails3Cfg.projectDir,
                icon.value.newIcon
            )

            icon.value.oldIcon = icon.value.newIcon
            icon.value.newIcon = ""
            iconChanged = true
        }

        props.wails3Cfg.project = cloneData(savedProject.project)
        props.wails3Cfg.importedAt = savedProject.importedAt
        props.wails3Cfg.lastOpenedAt = savedProject.lastOpenedAt
        originalCfgJson.value = JSON.stringify(cfg.value)
        emit("saved", {
            record: cloneData(savedProject),
            iconChanged
        })

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

function touchField(path) {
    touchedFields[path] = true
}

function touchAllFields() {
    for (const field of infoFields.value) {
        touchField(field.path)
    }
}

function visibleFieldError(path) {
    if (!touchedFields[path]) {
        return ""
    }
    return fieldError(path)
}

function fieldError(path) {
    const value = String(cfg.value[path] ?? "").trim()
    if (requiredFields.includes(path) && !value) {
        return t("editor.requiredField", {
            field: fieldLabel(path)
        })
    }
    if (
        path === "productIdentifier" &&
        value &&
        !/^[A-Za-z][A-Za-z0-9]*(?:\.[A-Za-z0-9][A-Za-z0-9-]*)+$/.test(value)
    ) {
        return t("createProject.validation.identifierInvalid")
    }
    return ""
}

function fieldLabel(path) {
    return infoFields.value.find(field => field.path === path)?.label || path
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

defineExpose({
    save: saveWails3Cfg
})

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

.field-error {
    margin-top: 6px;
    font-size: 11px;
    line-height: 16px;
    color: var(--wm-color-danger);
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
    --n-color: var(--wm-color-primary) !important;
    --n-color-hover: var(--wm-color-primary-hover) !important;
    --n-color-pressed: var(--wm-color-primary-pressed) !important;
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
