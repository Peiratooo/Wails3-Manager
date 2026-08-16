<template>
    <div class="project-view">
        <div class="container" v-if="loaded">
            <div class="head">
                <div class="project-info">
                    <div class="icon">
                        <img
                            v-if="iconSrc && !iconLoadError"
                            :src="iconSrc"
                            alt="icon"
                            @error="iconLoadError = true"
                        >

                        <div v-else class="icon-placeholder">
                            {{ projectName.slice(0, 1).toUpperCase() }}
                        </div>
                    </div>

                    <div class="info">


                        <div class="name">
                            <div>
                                {{ projectName }}
                            </div>
                            <span v-if="projectVersion" class="version">{{ projectVersion }}</span>
                        </div>

                        <div class="path-line">
                            <n-ellipsis class="path">
                               {{t("project.importTime")}}: {{ formatTimestamp(wails3Cfg.importedAt) }}
                            </n-ellipsis>
                        </div>
                    </div>
                </div>

                <div class="head-actions">
                    <Environment />
                    <n-button class="settings-button" circle quaternary @click="store.panels.settings = true">
                        <n-icon size="20">
                            <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 32 32"><g fill="none"><path d="M16 11a5 5 0 1 0 0 10a5 5 0 0 0 0-10zm-3 5a3 3 0 1 1 6 0a3 3 0 0 1-6 0zm-.16 13.628c1.035.247 2.096.372 3.16.372a13.643 13.643 0 0 0 3.156-.375a1.478 1.478 0 0 0 1.13-1.276l.234-2.13a1.471 1.471 0 0 1 2.066-1.2l1.955.856a1.472 1.472 0 0 0 1.671-.345a14.245 14.245 0 0 0 3.156-5.443a1.478 1.478 0 0 0-.535-1.627l-1.729-1.275a1.481 1.481 0 0 1 .003-2.396l1.72-1.27a1.474 1.474 0 0 0 .537-1.63a14.199 14.199 0 0 0-3.157-5.443a1.48 1.48 0 0 0-1.674-.345l-1.946.856a1.483 1.483 0 0 1-2.067-1.2l-.236-2.12a1.476 1.476 0 0 0-1.147-1.283a15.123 15.123 0 0 0-3.127-.363a15.395 15.395 0 0 0-3.146.363a1.469 1.469 0 0 0-1.147 1.28l-.237 2.122a1.493 1.493 0 0 1-2.073 1.206l-1.946-.857a1.493 1.493 0 0 0-1.67.35a14.245 14.245 0 0 0-3.16 5.446a1.478 1.478 0 0 0 .536 1.625l1.725 1.272a1.488 1.488 0 0 1 0 2.397L3.167 18.47a1.477 1.477 0 0 0-.535 1.63a14.253 14.253 0 0 0 3.16 5.45a1.458 1.458 0 0 0 1.077.465c.203 0 .404-.042.591-.123l1.955-.859a1.485 1.485 0 0 1 2.065 1.2l.235 2.126a1.476 1.476 0 0 0 1.125 1.27zm5.501-1.866a11.638 11.638 0 0 1-4.677 0l-.195-1.74a3.48 3.48 0 0 0-1.14-2.208a3.534 3.534 0 0 0-3.718-.6l-1.606.7a12.237 12.237 0 0 1-2.348-4.05l1.424-1.052a3.488 3.488 0 0 0 0-5.616L4.66 12.147a12.243 12.243 0 0 1 2.348-4.046l1.6.7a3.45 3.45 0 0 0 1.4.294a3.5 3.5 0 0 0 3.467-3.108l.194-1.747c.774-.15 1.56-.23 2.347-.24c.782.01 1.562.09 2.33.24l.186 1.74a3.48 3.48 0 0 0 1.137 2.216a3.525 3.525 0 0 0 3.727.6l1.6-.7a12.212 12.212 0 0 1 2.35 4.047l-1.423 1.046a3.48 3.48 0 0 0 0 5.62l1.422 1.05A12.273 12.273 0 0 1 25 23.901l-1.6-.7a3.473 3.473 0 0 0-4.866 2.81l-.193 1.75z" fill="currentColor"></path></g></svg>
                        </n-icon>
                    </n-button>
                    <n-button class="close-button" circle quaternary @click="backToHome">
                        <n-icon size="22">
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                <path d="M24 9.4L22.6 8L16 14.6L9.4 8L8 9.4l6.6 6.6L8 22.6L9.4 24l6.6-6.6l6.6 6.6l1.4-1.4l-6.6-6.6L24 9.4z" fill="currentColor"></path>
                            </svg>
                        </n-icon>
                    </n-button>
                </div>
            </div>

            <div class="body">
                <div class="editor">
                    <Editor
                        ref="editorRef"
                        :wails3Cfg="wails3Cfg"
                        :packageCfg="packageCfg"
                        :app-icon-version="iconVersion"
                        v-if="loaded"
                        @wails3-saved="handleWails3Saved"
                        @package-saved="setPackageCfg"
                    />
                </div>
            </div>

            <div class="foot">
                <div class="folder-card">
                    <div class="folder-icon">
                        <n-icon size="21">
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                <path d="M28 8H15.414l-2.707-2.707A1 1 0 0 0 12 5H4a2 2 0 0 0-2 2v18a2 2 0 0 0 2 2h24a2 2 0 0 0 2-2V10a2 2 0 0 0-2-2zM4 7h7.586l2.707 2.707A1 1 0 0 0 15 10h13v3H4zm0 18V15h24v10z" fill="currentColor"></path>
                            </svg>
                        </n-icon>
                    </div>

                    <div class="folder-content">
                        <div class="folder-label">{{ t("project.folder") }}</div>
                        <div class="folder-path">
                            <n-ellipsis>
                                {{ projectDir }}
                            </n-ellipsis>
                        </div>

                    </div>

                    <button
                        class="open-folder-button"
                        type="button"
                        :aria-label="t('project.openInFinder')"
                        :title="t('project.openInFinder')"
                        @click="openProjectFolder"
                    >
                        <n-icon size="18">
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                <path d="M26 28H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h9v2H6v20h20v-9h2v9a2 2 0 0 1-2 2z" fill="currentColor"></path>
                                <path d="M21 4v2h3.586L14.293 16.293l1.414 1.414L26 7.414V11h2V4z" fill="currentColor"></path>
                            </svg>
                        </n-icon>
                    </button>
                </div>

                <div class="build-card">
                    <n-button
                        class="build-button"
                        type="primary"
                        :loading="building"
                        :disabled="building"
                        @click="runBuild"
                        style="border-radius: 8px"
                    >
                        <div class="build-button-content">
                            <n-icon size="18">
                                <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 32 32"><path d="M7.288 23.292l7.997-7.997l1.415 1.414l-7.998 7.997z" fill="currentColor"></path><path d="M17 30a1 1 0 0 1-.37-.07a1 1 0 0 1-.62-.79l-1-7l2-.28l.75 5.27L21 24.52V17a1 1 0 0 1 .29-.71l4.07-4.07A8.94 8.94 0 0 0 28 5.86V4h-1.86a8.94 8.94 0 0 0-6.36 2.64l-4.07 4.07A1 1 0 0 1 15 11H7.48l-2.61 3.26l5.27.75l-.28 2l-7-1a1 1 0 0 1-.79-.62a1 1 0 0 1 .15-1l4-5A1 1 0 0 1 7 9h7.59l3.77-3.78A10.92 10.92 0 0 1 26.14 2H28a2 2 0 0 1 2 2v1.86a10.92 10.92 0 0 1-3.22 7.78L23 17.41V25a1 1 0 0 1-.38.78l-5 4A1 1 0 0 1 17 30z" fill="currentColor"></path></svg>
                            </n-icon>

                            <span>{{ t("project.runBuild") }}</span>
                        </div>

                    </n-button>
                </div>
            </div>
        </div>
        <Settings />

        <PackageRunCard
            v-model:show="packageModalVisible"
            :finished="packageFinished"
            :building="building"
            :progress="packageProgress"
            :error="packageError"
            :transaction-id="packageTransactionId"
            :log-entries="packageLogEntries"
            :result="packageResult"
            :directories="packageDirectories"
            @open-path="openPath"
        />
    </div>
</template>

<script setup>
import { Service as SettingsService } from "../../bindings/wails3-manager/core/settings"
import { Service as PackagingService } from "../../bindings/wails3-manager/core/packaging"
import { AppService } from "../../bindings/wails3-manager/desktop"
import { NButton, NEllipsis, NIcon, useMessage } from "naive-ui"
import { computed, inject, ref, watch } from "vue"
import router from "../router/index.js"
import { useRoute } from "vue-router"
import Environment from "../components/Environment.vue"
import Settings from "../components/Settings.vue";
import Editor from "../components/Editor/Editor.vue";
import PackageRunCard from "../components/PackageRunCard.vue"
import {useI18n} from "../i18n/index.js";
import { localFileUrl } from "../utils/files"
const { t } = useI18n()
const route = useRoute()
const formatTimestamp = inject("formatTimestamp")
const message = useMessage()
const store = inject("store")
const projectDir = computed(() => decodeURIComponent(route.query.projectDir || ""))
const wails3Cfg = ref({})
const packageCfg = ref({})
const editorRef = ref(null)
const loaded = ref(false)
const iconLoadError = ref(false)
const building = ref(false)
const packageModalVisible = ref(false)
const packageFinished = ref(false)
const packageProgress = ref(0)
const packageError = ref("")
const packageResult = ref(null)
const packageTransactionId = ref("")
const iconVersion = ref(0)
let loadProjectSeq = 0

const projectName = computed(() => {
    return wails3Cfg.value?.project?.wailsConfig?.info?.productName || t("project.unnamed")
})

const projectVersion = computed(() => {
    return wails3Cfg.value?.project?.wailsConfig?.info?.version || ""
})

const iconSrc = computed(() => {
    if (!wails3Cfg.value?.iconPath) return ""
    return localFileUrl(wails3Cfg.value.iconPath, iconVersion.value)
})

const packageLogEntries = computed(() => {
    if (!packageTransactionId.value) {
        return []
    }
    return store.packageTransactions[packageTransactionId.value]?.entries || []
})

const packageDirectories = computed(() => {
    if (!packageResult.value) {
        return []
    }
    const path = packageResult.value.packageOutputDir || packageResult.value.buildOutputDir
    if (!path) {
        return []
    }

    return [{
        label: packageResult.value.packageOutputDir
            ? t("project.finalPackageDir")
            : t("project.wailsOutputDir"),
        path
    }]
})

async function openProjectFolder() {
    await openPath(projectDir.value)
}

async function runBuild() {
    if (building.value) return

    building.value = true
    let packageStarted = false

    try {
        const saved = await editorRef.value?.saveAll?.()
        if (saved === false) {
            return
        }

        packageStarted = true
        packageTransactionId.value = `package-${Date.now()}`
        packageModalVisible.value = true
        packageFinished.value = false
        packageProgress.value = 8
        packageError.value = ""
        packageResult.value = null
        packageProgress.value = 32

        const result = await PackagingService.Package({
            projectDir: projectDir.value,
            platform: "auto",
            dryRun: false,
            runBuild: true,
            transactionId: packageTransactionId.value
        })

        packageResult.value = result
        packageProgress.value = 100
    } catch (error) {
        if (!packageStarted) {
            message.error(error?.message || String(error))
            return
        }
        packageError.value = error?.message || String(error)
        packageProgress.value = 100
    } finally {
        building.value = false
        if (packageStarted) {
            packageFinished.value = true
        }
    }
}

async function openPath(path) {
    try {
        await AppService.OpenPath(path)
    } catch (error) {
        message.error(error?.message || String(error))
    }
}

function setPackageCfg(nextCfg) {
    packageCfg.value = nextCfg
}

async function handleWails3Saved(payload) {
    const record = payload?.record
    if (!record) {
        return
    }

    const iconPath = await SettingsService.GetABSPath(
        record.projectDir,
        record.project.wailsConfig.icon
    )
    wails3Cfg.value = {
        ...record,
        iconPath
    }
    if (payload.iconChanged) {
        iconLoadError.value = false
        iconVersion.value = Date.now()
    }
}

async function loadProject() {
    const currentProjectDir = projectDir.value
    const seq = ++loadProjectSeq

    if (!currentProjectDir) {
        backToHome()
        return
    }

    loaded.value = false
    iconLoadError.value = false

    try {
        const wCfg = await SettingsService.OpenProject(currentProjectDir)
        const pCfg = await PackagingService.LoadPackagingConfig(currentProjectDir)

        const iconPath = await SettingsService.GetABSPath(
            wCfg.projectDir,
            wCfg.project.wailsConfig.icon
        )

        if (seq !== loadProjectSeq) {
            return
        }

        packageCfg.value = pCfg
        wails3Cfg.value = {
            ...wCfg,
            iconPath
        }

        loaded.value = true
    } catch (error) {
        if (seq !== loadProjectSeq) {
            return
        }
        message.error(error?.message || String(error))
        backToHome()
    }
}

watch(
    projectDir,
    () => {
        loadProject()
    },
    { immediate: true }
)


function backToHome() {
    router.push({
        name: "home"
    })
}

</script>

<style lang="scss" scoped>
.project-view {
    width: 100%;
    height: 100%;
    overflow: hidden;
}

.container {
    display: flex;
    flex-direction: column;
    gap: 16px;
    height: 100%;
    padding: 18px;
    box-sizing: border-box;
    overflow: hidden;
    color: var(--wm-text-secondary);
    background: var(--wm-bg-page-gradient);
}

.head {
    height: 78px;
    min-height: 78px;
    padding: 0 18px;
    box-sizing: border-box;

    display: flex;
    align-items: center;
    justify-content: space-between;

    border: 1px solid var(--wm-border-subtle);
    border-radius: 16px;
    background: var(--wm-surface-1);

    backdrop-filter: blur(18px);
}

.project-info {
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 14px;
}

.icon {
    width: 52px;
    height: 52px;
    min-width: 52px;
    border-radius: 12px;
    overflow: hidden;

    display: flex;
    align-items: center;
    justify-content: center;

    background:
        linear-gradient(180deg, var(--wm-surface-2), var(--wm-control-bg));
    border: 1px solid var(--wm-border-soft);

    img {
        width: 36px;
        height: 36px;
        object-fit: contain;
        display: block;
    }
}

.icon-placeholder {
    font-size: 18px;
    font-weight: 600;
    color: var(--wm-text-muted);
}

.info {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap:3px;
}

.status-line {
    height: 16px;
    display: flex;
    align-items: center;
    gap: 7px;
}

.status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--wm-color-primary);
    box-shadow: 0 0 0 4px var(--wm-color-primary-shadow);
}

.status-text {
    font-size: 11px;
    line-height: 1;
    color: var(--wm-text-muted);
    letter-spacing: 0.04em;
    text-transform: uppercase;
}

.version {
    height: 18px;
    padding: 0 8px;
    border-radius: 999px;

    font-size: 10px;
    line-height: 18px;
    color: var(--wm-color-primary-hover);
    background: var(--wm-color-primary-shadow);
    border: 1px solid var(--wm-color-primary-border);
}

.name {
    min-width: 0;
    max-width: 520px;
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 18px;
    font-weight: 500;
    line-height: 1.2;
    color: var(--wm-text-primary);

    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.path-line {
    min-width: 0;
    max-width: 720px;
    font-size: 12px;
}

.path {
    font-size: 12px;
    color: var(--wm-text-muted);
}

.head-actions {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 12px;
    margin-left: 20px;
}

.time {
    text-align: right;
    padding-right: 2px;
}

.time-label {
    margin-bottom: 3px;
    font-size: 11px;
    color: var(--wm-text-muted);
}

.time-value {
    font-size: 12px;
    color: var(--wm-text-secondary);
    white-space: nowrap;
}

.path-button {
    height: 32px;
    padding: 0 12px;
}

.close-button,.settings-button {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    color: var(--wm-text-secondary);
}

.close-button:hover {
    color: var(--wm-color-danger);
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

.body {
    flex: 1 1 0;
    min-height: 0;
    display: flex;
    overflow: hidden;
}

.editor {
    flex: 1;
    min-width: 0;
    min-height: 0;
    height: 100%;
}

.foot {
    height: 72px;
    min-height: 72px;
    padding: 0 14px;
    box-sizing: border-box;

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;

    border: 1px solid var(--wm-border-subtle);
    border-radius: 16px;
    background: var(--wm-surface-1);
    backdrop-filter: blur(18px);
}

.folder-card {
    flex: 1;
    min-width: 0;

    display: flex;
    align-items: center;
    gap: 12px;
}

.folder-icon {
    width: 40px;
    height: 40px;
    min-width: 40px;
    border-radius: 12px;

    display: flex;
    align-items: center;
    justify-content: center;

    color: var(--wm-color-primary-hover);
    background:
        radial-gradient(circle at 35% 20%, var(--wm-color-primary-shadow), transparent 58%),
        var(--wm-control-bg);
    border: 1px solid var(--wm-color-primary-border);
}

.folder-content {
    min-width: 0;
    flex: 1;
}

.folder-label {
    margin-bottom: 5px;
    font-size: 11px;
    color: var(--wm-text-muted);
}

.folder-path {
    max-width: 100%;
    font-size: 14px;
    color: var(--wm-text-secondary);
}

.open-folder-button {
    display: flex;
    background: var(--wm-control-bg);
    height: 36px;
    width: 36px;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    border: 1px solid var(--wm-border-soft);
    color: inherit;
    cursor: pointer;
    transition: 300ms;
    &:hover {
        color: var(--wm-color-primary-hover);
        border-color: var(--wm-color-primary-border);
    }
}

.build-card {
    flex: 0 0 auto;
}

.build-button {

    height: 36px;
    padding: 0 18px;
    display: flex;
    align-items: center;
    justify-content: center;

    gap: 8px;
}
.build-button-content {
    display: flex;
    align-items: center;
    gap: 8px;
    justify-content: center;
}
</style>
