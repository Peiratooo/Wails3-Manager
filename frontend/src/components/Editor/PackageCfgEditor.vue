<template>
    <div class="package-config">
        <div class="config-head">
            <div>
                <div class="config-title">{{ t("packageEditor.title") }}</div>
                <div class="config-desc">{{ t("packageEditor.desc") }}</div>
            </div>

            <div class="head-actions">
                <div class="head-state">
                    <span class="state-dot" :class="{ active: isDirty }"></span>
                    <span>{{ isDirty ? t("editor.hasChanges") : t("editor.noChanges") }}</span>
                </div>

                <n-button
                    type="primary"
                    :disabled="!isDirty || saving"
                    :loading="saving"
                    @click="savePackageCfg"
                >
                    {{ t("editor.save") }}
                </n-button>
            </div>
        </div>

        <div class="config-scroll">
            <section class="section-card">
                <div class="section-head">
                    <div>
                        <div class="section-title">{{ t("packageEditor.buildTitle") }}</div>
                        <div class="section-desc">{{ t("packageEditor.buildDesc") }}</div>
                    </div>
                </div>

                <div class="build-content">
                    <div class="input-area full">
                        <div class="label">{{ t("packageEditor.appName") }}</div>
                        <n-input
                            v-model:value="cfg.build.appName"
                            :placeholder="t('packageEditor.appNamePlaceholder')"
                        />
                    </div>

                    <div class="build-icon-row">
                        <div class="setup-icon-stage">
                            <n-image
                                v-if="currentSetupIcon"
                                class="setup-icon"
                                object-fit="contain"
                                :src="currentSetupIcon"
                                preview-disabled
                            />

                            <div v-else class="icon-empty">
                                <n-icon size="28">
                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                        <path d="M26 4H6a2 2 0 0 0-2 2v20a2 2 0 0 0 2 2h20a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2zM6 6h20v13.17l-4.59-4.58a2 2 0 0 0-2.82 0L16 17.17l-5.59-5.58a2 2 0 0 0-2.82 0L6 13.17zm0 20v-9.99l3-3L19.99 24H6zm20 0h-3.17l-5.42-5.41L20 18l6 6z" fill="currentColor"></path>
                                    </svg>
                                </n-icon>
                            </div>

                            <div v-if="setupIcon.newPath" class="new-badge">
                                {{ t("editor.newIcon") }}
                            </div>
                        </div>

                        <div class="icon-info">
                            <div class="label">{{ t("packageEditor.setupIcon") }}</div>

                            <div class="icon-status">
                                <span class="icon-status-dot" :class="{ active: setupIcon.newPath }"></span>
                                <span>
                                    {{ setupIcon.newPath ? t("packageEditor.setupIconChanged") : t("packageEditor.usingCurrentSetupIcon") }}
                                </span>
                            </div>

                            <n-ellipsis class="icon-path">
                                {{ cfg.windows.setupIcon || t("packageEditor.noSetupIcon") }}
                            </n-ellipsis>
                        </div>

                        <div class="icon-actions">
                            <n-button
                                :disabled="saving"
                                @click="chooseSetupIcon"
                            >
                                {{ setupIcon.newPath ? t("editor.reselectIcon") : t("editor.selectIcon") }}
                            </n-button>

                            <n-button
                                v-if="setupIcon.newPath"
                                quaternary
                                :disabled="saving"
                                @click="cancelSetupIcon"
                            >
                                {{ t("editor.cancel") }}
                            </n-button>
                        </div>
                    </div>

                    <div class="switch-row">
                        <div class="switch-area">
                            <div>
                                <div class="label">{{ t("packageEditor.production") }}</div>
                                <div class="hint">{{ t("packageEditor.productionDesc") }}</div>
                            </div>

                            <n-switch v-model:value="cfg.build.production" />
                        </div>

                        <div class="switch-area">
                            <div>
                                <div class="label">{{ t("packageEditor.cgoEnabled") }}</div>
                                <div class="hint">{{ t("packageEditor.cgoEnabledDesc") }}</div>
                            </div>

                            <n-switch v-model:value="cfg.build.cgoEnabled" />
                        </div>
                    </div>
                </div>
            </section>

            <section class="section-card">
                <div class="section-head">
                    <div>
                        <div class="section-title">{{ t("packageEditor.entryTitle") }}</div>
                        <div class="section-desc">{{ t("packageEditor.entryDesc") }}</div>
                    </div>
                </div>

                <div class="path-picker">
                    <div class="path-icon">
                        <n-icon size="20">
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                <path d="M6 4h20a2 2 0 0 1 2 2v20a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2zm0 2v20h20V6z" fill="currentColor"></path>
                                <path d="M12 10l10 6l-10 6z" fill="currentColor"></path>
                            </svg>
                        </n-icon>
                    </div>

                    <div class="path-main">
                        <div class="label">{{ t("packageEditor.executablePath") }}</div>

                        <n-ellipsis class="path-value">
                            {{ executablePathText }}
                        </n-ellipsis>

                        <div class="path-hint" v-if="runtimeExecutableHint">
                            {{ runtimeExecutableHint }}
                        </div>
                    </div>

                    <n-button
                        :disabled="saving"
                        @click="chooseExecutablePath"
                    >
                        {{ t("packageEditor.selectExecutable") }}
                    </n-button>
                </div>
            </section>

            <section class="section-card">
                <div class="section-head">
                    <div>
                        <div class="section-title">{{ t("packageEditor.windowsTitle") }}</div>
                        <div class="section-desc">{{ t("packageEditor.windowsDesc") }}</div>
                    </div>

                    <n-switch v-model:value="cfg.windows.enabled" />
                </div>

                <div class="form-grid" :class="{ disabled: !cfg.windows.enabled }">
                    <div class="input-area">
                        <div class="label">{{ t("packageEditor.privilegesRequired") }}</div>

                        <n-select
                            v-model:value="cfg.windows.privilegesRequired"
                            :options="privilegeOptions"
                        />
                    </div>

                    <div class="input-area">
                        <div class="label">{{ t("packageEditor.appURL") }}</div>

                        <n-input
                            v-model:value="cfg.windows.appURL"
                            :placeholder="t('packageEditor.appURLPlaceholder')"
                        />
                    </div>

                    <div class="switch-area full">
                        <div>
                            <div class="label">{{ t("packageEditor.createDesktopShortcut") }}</div>
                            <div class="hint">{{ t("packageEditor.createDesktopShortcutDesc") }}</div>
                        </div>

                        <n-switch v-model:value="cfg.windows.createDesktopShortcut" />
                    </div>
                </div>
            </section>

            <section class="section-card">
                <div class="section-head">
                    <div>
                        <div class="section-title">{{ t("packageEditor.assetsTitle") }}</div>
                        <div class="section-desc">{{ t("packageEditor.assetsDesc") }}</div>
                    </div>

                    <div class="asset-add-actions">
                        <n-button
                            size="small"
                            :disabled="saving"
                            @click="chooseAssetFile"
                        >
                            {{ t("packageEditor.addFile") }}
                        </n-button>

                        <n-button
                            size="small"
                            :disabled="saving"
                            @click="chooseAssetDirectory"
                        >
                            {{ t("packageEditor.addDirectory") }}
                        </n-button>
                    </div>
                </div>

                <div class="asset-list" v-if="cfg.assets.length">
                    <div
                        class="asset-item"
                        v-for="(asset, index) in cfg.assets"
                        :key="index"
                    >
                        <div class="asset-icon">
                            <n-icon size="18">
                                <svg v-if="asset.type === 'directory'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                    <path d="M28 8H15.414l-2.707-2.707A1 1 0 0 0 12 5H4a2 2 0 0 0-2 2v18a2 2 0 0 0 2 2h24a2 2 0 0 0 2-2V10a2 2 0 0 0-2-2zM4 7h7.586l2.707 2.707A1 1 0 0 0 15 10h13v3H4zm0 18V15h24v10z" fill="currentColor"></path>
                                </svg>

                                <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                    <path d="M19 2H8a2 2 0 0 0-2 2v24a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9zm0 2.8L23.2 9H19zM8 28V4h9v7h7v17z" fill="currentColor"></path>
                                </svg>
                            </n-icon>
                        </div>

                        <div class="asset-main">
                            <n-input
                                v-model:value="asset.src"
                                :placeholder="t('packageEditor.assetSrcPlaceholder')"
                            />

                            <div class="asset-options">

                                <div class="asset-required">
                                    <span>{{ t("packageEditor.assetRequired") }}</span>
                                    <n-switch v-model:value="asset.required" />
                                </div>
                            </div>
                        </div>

                        <div
                            class="asset-remove"
                            @click="removeAsset(index)"
                        >
                            <n-icon size="17">
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                    <path d="M12 12h2v12h-2z" fill="currentColor"></path>
                                    <path d="M18 12h2v12h-2z" fill="currentColor"></path>
                                    <path d="M4 6v2h2v20a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V8h2V6zm4 22V8h16v20z" fill="currentColor"></path>
                                    <path d="M12 2h8v2h-8z" fill="currentColor"></path>
                                </svg>
                            </n-icon>
                        </div>
                    </div>
                </div>

                <div class="asset-empty" v-else>
                    <div class="asset-empty-icon">
                        <n-icon size="24">
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                <path d="M28 8H15.414l-2.707-2.707A1 1 0 0 0 12 5H4a2 2 0 0 0-2 2v18a2 2 0 0 0 2 2h24a2 2 0 0 0 2-2V10a2 2 0 0 0-2-2zM4 7h7.586l2.707 2.707A1 1 0 0 0 15 10h13v3H4zm0 18V15h24v10z" fill="currentColor"></path>
                            </svg>
                        </n-icon>
                    </div>

                    <div>{{ t("packageEditor.noAssets") }}</div>
                </div>
            </section>
        </div>
    </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue"
import {
    NButton,
    NEllipsis,
    NIcon,
    NImage,
    NInput,
    NSelect,
    NSwitch,
    useMessage
} from "naive-ui"
import { useI18n } from "../../i18n"
import { PackagingService } from "../../../bindings/wails3-manager/core/packaging"
import { AppService } from "../../../bindings/wails3-manager/desktop"

const props = defineProps({
    packageCfg: {
        type: Object,
        required: true
    },
    projectDir: {
        type: String,
        required: true
    }
})

const { t } = useI18n()
const message = useMessage()

const cfg = ref(cloneData(props.packageCfg))
const originalCfgJson = ref(JSON.stringify(cfg.value))
const saving = ref(false)
const runtimeInfo = ref({
    defaultExecutablePath: "",
    effectiveExecutablePath: "",
    usingDefaultExecutable: true
})

const setupIcon = ref({
    oldPath: "",
    newPath: ""
})

const privilegeOptions = computed(() => [
    {
        label: t("packageEditor.privileges.lowest"),
        value: "lowest"
    },
    {
        label: t("packageEditor.privileges.poweruser"),
        value: "poweruser"
    },
    {
        label: t("packageEditor.privileges.admin"),
        value: "admin"
    }
])

const assetTypeOptions = computed(() => [
    {
        label: t("packageEditor.assetFile"),
        value: "file"
    },
    {
        label: t("packageEditor.assetDirectory"),
        value: "directory"
    }
])

const currentSetupIcon = computed(() => {
    const path = setupIcon.value.newPath || setupIcon.value.oldPath
    if (setupIcon.value.newPath) {
        return "/local/file?"+path
    } else {
        return "/local/file?"+props.projectDir+'/'+path
    }

})

const isDirty = computed(() => {
    return JSON.stringify(cfg.value) !== originalCfgJson.value
})

const executablePathText = computed(() => {
    if (cfg.value.entry.executablePath) {
        return cfg.value.entry.executablePath
    }
    if (runtimeInfo.value.effectiveExecutablePath) {
        return t("packageEditor.autoExecutable", {
            path: runtimeInfo.value.effectiveExecutablePath
        })
    }
    return t("packageEditor.noExecutable")
})

const runtimeExecutableHint = computed(() => {
    if (!runtimeInfo.value.effectiveExecutablePath) {
        return ""
    }
    if (runtimeInfo.value.usingDefaultExecutable) {
        return ""
    }
    return t("packageEditor.effectiveExecutable", {
        path: runtimeInfo.value.effectiveExecutablePath
    })
})



onMounted(() => {
    initConfig(props.packageCfg)
    loadRuntimeInfo()
})

watch(
    () => props.packageCfg,
    (next) => {
        initConfig(next)
        loadRuntimeInfo()
    },
    { deep: true }
)

function initConfig(data) {
    cfg.value = cloneData(data)
    originalCfgJson.value = JSON.stringify(cfg.value)

    setupIcon.value.oldPath = cfg.value.windows.setupIcon
    setupIcon.value.newPath = ""
}

async function savePackageCfg() {
    if (!isDirty.value || saving.value) return

    try {
        saving.value = true

        const nextCfg = cloneData(cfg.value)

        const savedCfg = await PackagingService.SavePackagingConfig(props.projectDir, nextCfg)
        initConfig(savedCfg)
        await loadRuntimeInfo()

        message.success(t("editor.saveSuccess"))
    } catch (error) {
        message.error(error?.message || String(error))
    } finally {
        saving.value = false
    }
}

async function loadRuntimeInfo() {
    try {
        runtimeInfo.value = await PackagingService.GetPackagingRuntimeInfo(props.projectDir)
    } catch (error) {
        runtimeInfo.value = {
            defaultExecutablePath: "",
            effectiveExecutablePath: "",
            usingDefaultExecutable: true
        }
    }
}

async function chooseSetupIcon() {
    try {
        const path = await AppService.ChooseSetupIcon()
        setNewSetupIcon(path)
    } catch (error) {
        message.error(error?.message || String(error))
    }
}

function setNewSetupIcon(path) {
    if (!path) return

    setupIcon.value.newPath = path
    cfg.value.windows.setupIcon = path
}

function cancelSetupIcon() {
    setupIcon.value.newPath = ""
    cfg.value.windows.setupIcon = setupIcon.value.oldPath
}

async function chooseExecutablePath() {
    try {
        const path = await AppService.ChooseFile()
        setExecutablePath(path)
    } catch (error) {
        message.error(error?.message || String(error))
    }
}

function setExecutablePath(path) {
    cfg.value.entry.executablePath = path || ""
}

async function chooseAssetFile() {
    try {
        const path = await AppService.ChooseFile()
        addAsset(path, "file")
    } catch (error) {
        message.error(error?.message || String(error))
    }
}

async function chooseAssetDirectory() {
    try {
        const path = await AppService.ChooseFolder()
        addAsset(path, "directory")
    } catch (error) {
        message.error(error?.message || String(error))
    }
}

function addAsset(src, type = "file") {
    if (!src) return

    cfg.value.assets.push({
        src,
        type,
        required: true
    })
}

function removeAsset(index) {
    cfg.value.assets.splice(index, 1)
}

function cloneData(data) {
    return JSON.parse(JSON.stringify(data))
}
</script>

<style lang="scss" scoped>
.package-config {
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

.head-state {
    flex: 0 0 auto;

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

.config-scroll {
    flex: 1;
    min-height: 0;
    margin-top: 18px;
    overflow-y: auto;
    padding-right: 4px;
}

.config-scroll::-webkit-scrollbar {
    width: 6px;
}

.config-scroll::-webkit-scrollbar-thumb {
    border-radius: 999px;
    background: var(--wm-border-soft);
}

.section-card {
    padding: 18px;
    box-sizing: border-box;
    border-radius: 18px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-surface-2);
}

.section-card + .section-card {
    margin-top: 14px;
}

.section-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 18px;
}

.section-title {
    font-size: 15px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.section-desc {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.5;
    color: var(--wm-text-muted);
}

.build-content {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.form-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    align-content: start;
    gap: 16px;
}

.form-grid.disabled {
    opacity: 0.58;
    pointer-events: none;
}

.input-area {
    min-width: 0;
}

.input-area.full,
.switch-area.full {
    grid-column: 1 / -1;
}

.label {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--wm-text-muted);
}

.hint {
    margin-top: 5px;
    font-size: 12px;
    line-height: 1.5;
    color: var(--wm-text-muted);
}

.switch-row {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
}

.switch-area {
    min-width: 0;
    min-height: 58px;
    padding: 12px;
    box-sizing: border-box;
    border-radius: 14px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-control-bg);

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.build-icon-row {
    min-width: 0;
    padding: 12px;
    box-sizing: border-box;
    border-radius: 14px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-control-bg);

    display: flex;
    align-items: center;
    gap: 14px;
}

.setup-icon-stage {
    position: relative;
    width: 76px;
    height: 76px;
    min-width: 76px;
    border-radius: 18px;
    overflow: hidden;

    display: flex;
    align-items: center;
    justify-content: center;

    background:
        radial-gradient(circle at 30% 18%, var(--wm-color-primary-shadow), transparent 58%),
        var(--wm-control-bg);
    border: 1px solid var(--wm-border-soft);
}

.setup-icon {
    width: 52px;
    height: 52px;
}

:deep(.setup-icon img) {
    width: 52px;
    height: 52px;
    object-fit: contain;
}

.icon-empty {
    width: 44px;
    height: 44px;
    border-radius: 14px;

    display: flex;
    align-items: center;
    justify-content: center;

    color: var(--wm-text-muted);
    background: var(--wm-control-bg);
    border: 1px solid var(--wm-border-soft);
}

.new-badge {
    position: absolute;
    left: 6px;
    bottom: 6px;
    z-index: 2;

    height: 18px;
    padding: 0 7px;
    border-radius: 999px;

    font-size: 9px;
    line-height: 18px;

    color: var(--wm-text-inverse);
    background: linear-gradient(
        180deg,
        var(--wm-color-logo-start) 0%,
        var(--wm-color-primary) 100%
    );
    box-shadow: var(--wm-shadow-primary);
}

.icon-info {
    flex: 1;
    min-width: 0;
}

.icon-status {
    display: flex;
    align-items: center;
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

.icon-path {
    max-width: 100%;
    margin-top: 8px;
    font-size: 12px;
    color: var(--wm-text-secondary);
}

.icon-actions {
    flex: 0 0 auto;

    display: flex;
    align-items: center;
    gap: 10px;
}

.path-picker {
    min-width: 0;
    padding: 12px;
    box-sizing: border-box;
    border-radius: 14px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-control-bg);

    display: flex;
    align-items: center;
    gap: 12px;
}

.path-icon {
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

.path-main {
    flex: 1;
    min-width: 0;
}

.path-value {
    max-width: 100%;
    font-size: 12px;
    color: var(--wm-text-secondary);
}

.path-hint {
    margin-top: 6px;
    font-size: 11px;
    line-height: 1.45;
    color: var(--wm-text-muted);
}

.asset-add-actions {
    flex: 0 0 auto;

    display: flex;
    align-items: center;
    gap: 10px;
}

.asset-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.asset-item {
    min-width: 0;
    padding: 12px;
    box-sizing: border-box;
    border-radius: 14px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-control-bg);

    display: flex;
    align-items: center;
    gap: 12px;
}

.asset-icon {
    width: 34px;
    height: 34px;

    border-radius: 8px;

    display: flex;
    align-items: center;
    justify-content: center;

    color: var(--wm-text-muted);
    background: var(--wm-control-bg);
    border: 1px solid var(--wm-border-soft);
}

.asset-main {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 16px;
}

.asset-options {


    display: flex;
    align-items: center;
    gap: 12px;
}

.asset-type {
    width: 132px;
}

.asset-required {
    flex: 1;
    min-width: 0;

    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 10px;

    font-size: 12px;

    white-space:nowrap ;
    color: var(--wm-text-muted);
}

.asset-remove {

    display: flex;
    cursor: pointer;
    color: var(--wm-text-muted);
}

.asset-remove:hover {
    color: var(--wm-color-danger);
}

.asset-empty {
    min-height: 150px;
    border-radius: 16px;
    border: 1px dashed var(--wm-border-soft);
    color: var(--wm-text-muted);
    background: var(--wm-control-bg);

    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 10px;
    font-size: 13px;
}

.asset-empty-icon {
    width: 48px;
    height: 48px;
    border-radius: 16px;

    display: flex;
    align-items: center;
    justify-content: center;

    color: var(--wm-text-muted);
    background: var(--wm-control-bg);
    border: 1px solid var(--wm-border-soft);
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

:deep(.n-base-selection) {
    --n-color: var(--wm-control-bg) !important;
    --n-color-active: var(--wm-control-bg-hover) !important;
    --n-border: 1px solid var(--wm-border-soft) !important;
    --n-border-hover: 1px solid var(--wm-color-primary-border) !important;
    --n-border-active: 1px solid var(--wm-border-strong) !important;
    --n-box-shadow-active: 0 0 0 2px var(--wm-color-primary-shadow) !important;
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

:deep(.n-switch) {
    --n-rail-color-active: var(--wm-color-primary) !important;
    --n-loading-color: var(--wm-color-primary) !important;
    --n-box-shadow-focus: 0 0 0 2px var(--wm-color-primary-shadow) !important;
}

:deep(.n-button--disabled) {
    opacity: 0.5;
    box-shadow: none;
}
</style>
