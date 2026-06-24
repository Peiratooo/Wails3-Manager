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
                    :disabled="!isDirty || saving || hasAssetTargetErrors"
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
<!--                        <div class="section-desc">{{ t("packageEditor.buildDesc") }}</div>-->
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

                    <div v-if="isWindows" class="build-icon-row">
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
<!--                        <div class="section-desc">{{ t("packageEditor.entryDesc") }}</div>-->
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
                        <div class="section-title">
                            {{ activePlatform === "windows" ? t("packageEditor.windowsTitle") : t("packageEditor.macosTitle") }}
                        </div>
<!--                        <div class="section-desc">-->
<!--                            {{ activePlatform === "windows" ? t("packageEditor.windowsDesc") : t("packageEditor.macosDesc") }}-->
<!--                        </div>-->
                    </div>

                    <div class="platform-actions">
                        <div class="package-toggle compact">
                            <span>
                                {{ activePlatform === "windows" ? t("packageEditor.windowsEnabled") : t("packageEditor.macosEnabled") }}
                            </span>

                            <n-switch
                                v-if="isWindows"
                                v-model:value="cfg.windows.enabled"
                            />

                            <n-switch
                                v-else-if="isMacOS"
                                v-model:value="cfg.macos.enabled"
                            />
                        </div>
                    </div>
                </div>

                <div
                    v-if="isWindows"
                    class="form-grid"
                    :class="{ disabled: !cfg.windows.enabled }"
                >
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

                            <div>{{ t("packageEditor.createDesktopShortcut") }}</div>
<!--                            <div class="hint">{{ t("packageEditor.createDesktopShortcutDesc") }}</div>-->


                        <n-switch v-model:value="cfg.windows.createDesktopShortcut" />
                    </div>
                </div>

                <div
                    v-else-if="isMacOS"
                    class="form-grid"
                    :class="{ disabled: !cfg.macos.enabled }"
                >
                    <div class="path-picker full">
                        <div class="path-icon">
                            <n-icon size="20">
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                    <path d="M6 4h20a2 2 0 0 1 2 2v20a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2zm0 2v20h20V6z" fill="currentColor"></path>
                                    <path d="M10 11h12v2H10zm0 4h8v2h-8zm0 4h10v2H10z" fill="currentColor"></path>
                                </svg>
                            </n-icon>
                        </div>

                        <div class="path-main">
                            <div class="label">{{ t("packageEditor.dmgLayoutTitle") }}</div>
                            <n-ellipsis class="path-value">
                                {{ cfg.macos.windowWidth }} x {{ cfg.macos.windowHeight }} · {{ t("packageEditor.iconSize") }} {{ cfg.macos.iconSize }}
                            </n-ellipsis>
                            <div class="path-hint">
                                App {{ cfg.macos.appX }}, {{ cfg.macos.appY }} · Applications {{ cfg.macos.applicationsX }}, {{ cfg.macos.applicationsY }}
                            </div>
                        </div>

                        <n-button
                            :disabled="saving"
                            @click="dmgEditorVisible = true"
                        >
                            {{ t("packageEditor.editDmgLayout") }}
                        </n-button>
                    </div>
                </div>
            </section>

            <section class="section-card">
                <div class="section-head">
                    <div>
                        <div class="section-title">{{ t("packageEditor.assetsTitle") }}</div>
<!--                        <div class="section-desc">{{ t("packageEditor.assetsDesc") }}</div>-->
                    </div>

                    <div v-if="isWindows" class="asset-add-actions">
                        <n-button
                            size="small"
                            :disabled="saving"
                            @click="chooseAssetFile()"
                        >
                            {{ t("packageEditor.addFile") }}
                        </n-button>

                        <n-button
                            size="small"
                            :disabled="saving"
                            @click="chooseAssetDirectory()"
                        >
                            {{ t("packageEditor.addDirectory") }}
                        </n-button>
                    </div>
                </div>

                <div v-if="isMacOS" class="mac-asset-board">
                    <div class="asset-area-toolbar">
                        <n-button
                            v-for="area in availableFixedMacAreas"
                            :key="area"
                            size="small"
                            :disabled="saving"
                            @click="addMacArea(area)"
                        >
                            + {{ area }}
                        </n-button>

                        <div class="custom-area-adder">
                            <n-input
                                v-model:value="customMacArea"
                                size="small"
                                placeholder="自定义区域"
                            />
                            <n-button
                                size="small"
                                :disabled="saving || !canAddCustomMacArea"
                                @click="addCustomMacArea"
                            >
                                添加区域
                            </n-button>
                        </div>
                    </div>

                    <div v-if="customMacAreaError" class="asset-target-error">
                        {{ customMacAreaError }}
                    </div>

                    <div
                        class="asset-region"
                        v-for="area in macAssetAreas"
                        :key="area"
                    >
                        <div class="asset-region-head">
                            <div>
                                <div class="asset-region-title">{{ area }}</div>
                                <div class="path-hint">Contents/{{ area }}</div>
                            </div>

                            <div class="asset-add-actions">
                                <n-button
                                    size="small"
                                    :disabled="saving"
                                    @click="chooseAssetFile(area)"
                                >
                                    {{ t("packageEditor.addFile") }}
                                </n-button>

                                <n-button
                                    size="small"
                                    :disabled="saving"
                                    @click="chooseAssetDirectory(area)"
                                >
                                    {{ t("packageEditor.addDirectory") }}
                                </n-button>

                                <n-button
                                    v-if="canRemoveMacArea(area)"
                                    size="small"
                                    quaternary
                                    :disabled="saving"
                                    @click="removeMacArea(area)"
                                >
                                    删除区域
                                </n-button>
                            </div>
                        </div>

                        <div v-if="macAreaTargetError(area)" class="asset-target-error region-error">
                            {{ macAreaTargetError(area) }}
                        </div>

                        <div class="asset-list compact">
                            <div
                                v-if="area === 'MacOS'"
                                class="asset-item asset-implicit"
                            >
                                <div class="asset-icon locked">
                                    <n-icon size="18">
                                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                            <path d="M6 4h20a2 2 0 0 1 2 2v20a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2zm0 2v20h20V6z" fill="currentColor"></path>
                                            <path d="M12 10l10 6l-10 6z" fill="currentColor"></path>
                                        </svg>
                                    </n-icon>
                                </div>

                                <div class="asset-main vertical">
                                    <div class="asset-label">主程序</div>
                                    <n-ellipsis class="path-value">{{ macMainProgramPath }}</n-ellipsis>
                                </div>
                            </div>

                            <div
                                v-if="area === 'MacOS' && macStartupProgramPath"
                                class="asset-item asset-implicit"
                            >
                                <div class="asset-icon locked">
                                    <n-icon size="18">
                                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                                            <path d="M6 4h20a2 2 0 0 1 2 2v20a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2zm0 2v20h20V6z" fill="currentColor"></path>
                                            <path d="M10 11h12v2H10zm0 4h8v2h-8zm0 4h10v2H10z" fill="currentColor"></path>
                                        </svg>
                                    </n-icon>
                                </div>

                                <div class="asset-main vertical">
                                    <div class="asset-label">启动程序</div>
                                    <n-ellipsis class="path-value">{{ macStartupProgramPath }}</n-ellipsis>
                                </div>
                            </div>

                            <div
                                class="asset-item"
                                v-for="{ asset, index } in assetsForMacArea(area)"
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

                            <div
                                v-if="area !== 'MacOS' && !assetsForMacArea(area).length"
                                class="asset-empty compact"
                            >
                                {{ t("packageEditor.noAssets") }}
                            </div>
                        </div>
                    </div>
                </div>

                <div v-else-if="isWindows">
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

                            <div class="asset-target-summary">
                                <button
                                    class="asset-target-toggle"
                                    type="button"
                                    :class="{ invalid: !!windowsAssetTargetError(asset) }"
                                    @click="toggleWindowsTarget(index)"
                                >
                                    目标目录 {{ normalizedWindowsTarget(asset.target) }}
                                </button>

                                <div
                                    v-if="expandedWindowsTargets[index]"
                                    class="asset-target-panel"
                                >
                                    <n-input
                                        v-model:value="asset.target"
                                        size="small"
                                        placeholder="/"
                                    />

                                    <div
                                        v-if="windowsAssetTargetError(asset)"
                                        class="asset-target-error"
                                    >
                                        {{ windowsAssetTargetError(asset) }}
                                    </div>
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
                </div>
            </section>
        </div>

        <DmgLayoutEditor
            v-if="isMacOS"
            v-model:show="dmgEditorVisible"
            :macos="cfg.macos"
            :project-dir="projectDir"
            @save="setMacOSLayout"
        />
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
import DmgLayoutEditor from "./DmgLayoutEditor.vue"

const props = defineProps({
    packageCfg: {
        type: Object,
        required: true
    },
    projectDir: {
        type: String,
        required: true
    },
    currentPlatform: {
        type: String,
        default: "windows"
    }
})

const emit = defineEmits(["saved"])
const { t } = useI18n()
const message = useMessage()

const cfg = ref(cloneData(props.packageCfg))
const originalCfgJson = ref(JSON.stringify(cfg.value))
const saving = ref(false)
const dmgEditorVisible = ref(false)
const runtimeInfo = ref({
    defaultExecutablePath: "",
    effectiveExecutablePath: "",
    usingDefaultExecutable: true
})

const setupIcon = ref({
    oldPath: "",
    newPath: ""
})
const customMacArea = ref("")
const addedMacAreas = ref([])
const expandedWindowsTargets = ref({})

const defaultMacAssetAreas = ["MacOS", "Resources"]
const fixedMacAssetAreas = ["Frameworks", "PlugIns", "SharedSupport"]

const activePlatform = computed(() => props.currentPlatform === "darwin" ? "macos" : "windows")
const isWindows = computed(() => activePlatform.value === "windows")
const isMacOS = computed(() => activePlatform.value === "macos")

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

const currentSetupIcon = computed(() => {
    const path = setupIcon.value.newPath || setupIcon.value.oldPath
    if (!path) {
        return ""
    }
    return "/local/file?" + projectFilePath(path)
})

const isDirty = computed(() => {
    return JSON.stringify(cfg.value) !== originalCfgJson.value
})

const hasAssetTargetErrors = computed(() => {
    if (!cfg.value.assets?.length) {
        return false
    }
    if (isWindows.value) {
        return cfg.value.assets.some((asset) => !!windowsAssetTargetError(asset))
    }
    if (isMacOS.value) {
        return cfg.value.assets.some((asset) => !!macAssetTargetError(asset))
    }
    return false
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

const macAssetAreas = computed(() => {
    const areas = new Set(defaultMacAssetAreas)
    for (const area of addedMacAreas.value) {
        areas.add(area)
    }
    for (const asset of cfg.value.assets || []) {
        areas.add(normalizedMacTarget(asset.target))
    }
    return Array.from(areas)
})

const availableFixedMacAreas = computed(() => {
    return fixedMacAssetAreas.filter((area) => !macAssetAreas.value.includes(area))
})

const customMacAreaError = computed(() => {
    const area = customMacArea.value.trim()
    if (!area) {
        return ""
    }
    if (macAssetAreas.value.includes(area)) {
        return "区域已存在"
    }
    return macTargetError(area, false)
})

const canAddCustomMacArea = computed(() => {
    return customMacArea.value.trim() !== "" && !customMacAreaError.value
})

const macMainProgramPath = computed(() => {
    const name = cfg.value.build?.appName || "app"
    return `Contents/MacOS/${name}`
})

const macStartupProgramPath = computed(() => {
    const entry = String(cfg.value.entry?.executablePath || "").trim()
    if (!entry) {
        return ""
    }
    return `Contents/MacOS/${baseName(entry)}`
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
    if (!Array.isArray(cfg.value.assets)) {
        cfg.value.assets = []
    }
    originalCfgJson.value = JSON.stringify(cfg.value)

    setupIcon.value.oldPath = cfg.value.windows?.setupIcon || ""
    setupIcon.value.newPath = ""
    customMacArea.value = ""
    addedMacAreas.value = []
    expandedWindowsTargets.value = {}
}

async function savePackageCfg() {
    if (!isDirty.value || saving.value || hasAssetTargetErrors.value) return

    try {
        saving.value = true

        const nextCfg = cloneData(cfg.value)
        normalizeAssetTargetsForSave(nextCfg)

        const savedCfg = await PackagingService.SavePackagingConfig(props.projectDir, nextCfg)
        initConfig(savedCfg)
        emit("saved", cloneData(savedCfg))
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

function setMacOSLayout(nextMacOS) {
    cfg.value.macos = cloneData(nextMacOS)
}

function projectFilePath(path) {
    if (/^[A-Za-z]:[\\/]/.test(path) || path.startsWith("/") || path.startsWith("\\\\")) {
        return path
    }
    return props.projectDir + "/" + path
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

async function chooseAssetFile(target = defaultAssetTarget()) {
    try {
        const path = await AppService.ChooseFile()
        addAsset(path, "file", target)
    } catch (error) {
        message.error(error?.message || String(error))
    }
}

async function chooseAssetDirectory(target = defaultAssetTarget()) {
    try {
        const path = await AppService.ChooseFolder()
        addAsset(path, "directory", target)
    } catch (error) {
        message.error(error?.message || String(error))
    }
}

function addAsset(src, type = "file", target = defaultAssetTarget()) {
    if (!src) return

    cfg.value.assets.push({
        src,
        type,
        required: true,
        target
    })
}

function removeAsset(index) {
    const asset = cfg.value.assets[index]
    if (isMacOS.value && asset) {
        const area = normalizedMacTarget(asset.target)
        if (!defaultMacAssetAreas.includes(area) && !addedMacAreas.value.includes(area)) {
            addedMacAreas.value.push(area)
        }
    }
    cfg.value.assets.splice(index, 1)
}

function assetsForMacArea(area) {
    return (cfg.value.assets || [])
        .map((asset, index) => ({ asset, index }))
        .filter(({ asset }) => normalizedMacTarget(asset.target) === area)
}

function addMacArea(area) {
    const next = String(area || "").trim()
    if (!next || macAssetAreas.value.includes(next) || macTargetError(next, false)) {
        return
    }
    addedMacAreas.value.push(next)
}

function addCustomMacArea() {
    if (!canAddCustomMacArea.value) {
        return
    }
    addMacArea(customMacArea.value.trim())
    customMacArea.value = ""
}

function canRemoveMacArea(area) {
    return !defaultMacAssetAreas.includes(area) && assetsForMacArea(area).length === 0
}

function removeMacArea(area) {
    if (!canRemoveMacArea(area)) {
        return
    }
    addedMacAreas.value = addedMacAreas.value.filter((item) => item !== area)
}

function toggleWindowsTarget(index) {
    expandedWindowsTargets.value[index] = !expandedWindowsTargets.value[index]
}

function defaultAssetTarget() {
    return isMacOS.value ? "MacOS" : "/"
}

function normalizeAssetTargetsForSave(nextCfg) {
    nextCfg.assets = (nextCfg.assets || []).map((asset) => ({
        ...asset,
        target: isMacOS.value ? normalizedMacTarget(asset.target) : normalizedWindowsTarget(asset.target)
    }))
}

function normalizedWindowsTarget(target) {
    const value = String(target || "").trim()
    return value || "/"
}

function normalizedMacTarget(target) {
    const value = String(target || "").trim()
    return value || "MacOS"
}

function windowsAssetTargetError(asset) {
    const target = normalizedWindowsTarget(asset?.target)
    if (!target.startsWith("/")) {
        return "目标目录必须以 / 开始"
    }
    if (target.includes("\\")) {
        return "目标目录只能使用 / 分隔"
    }
    return targetPartsError(target.slice(1), target)
}

function macAssetTargetError(asset) {
    return macTargetError(normalizedMacTarget(asset?.target), true)
}

function macAreaTargetError(area) {
    return macTargetError(area, false)
}

function macTargetError(target, allowEmpty) {
    const value = String(target || "").trim()
    if (!value) {
        return allowEmpty ? "" : "请输入区域名称"
    }
    if (value.startsWith("/") || value.startsWith("\\")) {
        return "区域必须是 Contents 下的相对路径"
    }
    if (value.includes("\\")) {
        return "区域只能使用 / 分隔"
    }
    return targetPartsError(value, value)
}

function targetPartsError(raw, original) {
    if (/[:*?"<>|]/.test(raw)) {
        return "路径包含非法字符"
    }
    if (!raw) {
        return ""
    }
    const parts = raw.split("/")
    if (parts.some((part) => !part || part === "." || part === "..")) {
        return `非法路径：${original}`
    }
    return ""
}

function baseName(path) {
    const clean = String(path || "").replace(/\\/g, "/").replace(/\/+$/, "")
    const parts = clean.split("/")
    return parts[parts.length - 1] || clean || "app"
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
    margin-bottom: 12px;
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

.platform-actions {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
}

.platform-switch-row {
    margin-bottom: 12px;
    display: flex;
    align-items: center;
    gap: 12px;
}

.package-toggle {
    min-height: 52px;
    padding: 9px 12px;
    box-sizing: border-box;
    border-radius: 12px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-control-bg);

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
}

.package-toggle.compact {
    min-height: 34px;
    padding: 0 8px 0 10px;
    border-radius: 6px;
}

.package-toggle.compact span {
    font-size: 12px;
    color: var(--wm-text-secondary);
}

.package-toggle-copy {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 5px;
}

.package-toggle-copy span {
    font-size: 12px;
    line-height: 1.2;
    color: var(--wm-text-secondary);
}

.package-toggle-copy small {
    font-size: 10px;
    line-height: 1.2;
    color: var(--wm-text-muted);
}

.platform-tabs {
    height: 34px;
    padding: 3px;
    box-sizing: border-box;
    border-radius: 6px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-control-bg);

    display: flex;
    align-items: center;
    gap: 3px;
}

.platform-switch-hint {
    min-width: 0;
    font-size: 11px;
    line-height: 1.4;
    color: var(--wm-text-muted);
}

.platform-tab {
    height: 26px;
    padding: 0 10px;
    border: 0;
    border-radius: 4px;
    background: transparent;
    color: var(--wm-text-muted);
    font-size: 12px;
    cursor: pointer;
}

.platform-tab.active {
    color: var(--wm-text-inverse);
    background: var(--wm-color-primary);
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

.path-picker.full {
    grid-column: 1 / -1;
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

.mac-asset-board {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.asset-area-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
}

.custom-area-adder {
    min-width: 260px;
    display: flex;
    align-items: center;
    gap: 8px;
}

.asset-region {
    padding: 12px;
    box-sizing: border-box;
    border-radius: 14px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-control-bg);
}

.asset-region-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 10px;
}

.asset-region-title {
    font-size: 13px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.asset-list.compact {
    gap: 8px;
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

.asset-icon.locked {
    color: var(--wm-color-primary-hover);
    border-color: var(--wm-color-primary-border);
}

.asset-main {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 16px;
}

.asset-main.vertical {
    align-items: flex-start;
    flex-direction: column;
    gap: 3px;
}

.asset-main :deep(.n-input) {
    flex: 1;
    min-width: 220px;
}

.asset-label {
    font-size: 12px;
    color: var(--wm-text-muted);
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

.asset-target-summary {
    flex: 0 0 100%;
    min-width: 0;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
}

.asset-target-toggle {
    border: 0;
    padding: 0;
    background: transparent;
    color: var(--wm-text-muted);
    font-size: 12px;
    cursor: pointer;
}

.asset-target-toggle:hover {
    color: var(--wm-color-primary-hover);
}

.asset-target-toggle.invalid,
.asset-target-error {
    color: var(--wm-color-danger);
}

.asset-target-panel {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.asset-target-error {
    font-size: 12px;
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

.asset-empty.compact {
    min-height: 58px;
    border-radius: 10px;
    font-size: 12px;
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
