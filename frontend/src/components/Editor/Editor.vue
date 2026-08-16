<template>
    <div class="editor-card">
        <n-tabs
            v-model:value="activeTab"
            class="editor-tabs"
            type="line"
            animated
        >
            <n-tab-pane
                v-for="item in tabs"
                :key="item.name"
                :name="item.name"
                display-directive="show"
            >
                <template #tab>
                    <div class="tab-label">
                        <n-icon size="17">
                            <span v-html="item.icon"></span>
                        </n-icon>
                        <span>{{ item.label }}</span>
                    </div>
                </template>

                <div class="tab-content">
                    <component
                        :is="item.component"
                        :ref="setEditorRef(item.name)"
                        v-bind="item.props"
                    />
                </div>
            </n-tab-pane>
        </n-tabs>
    </div>
</template>

<script setup>
import { computed, markRaw, onBeforeUnmount, onMounted, ref } from "vue"
import { NTabs, NTabPane, NIcon, useMessage } from "naive-ui"
import { useI18n } from "../../i18n"

import InfoCfgEditor from "./InfoCfgEditor.vue"
import PackageCfgEditor from "./PackageCfgEditor.vue"

const props = defineProps({
    wails3Cfg: {
        type: Object,
        required: true
    },
    packageCfg: {
        type: Object,
        required: true
    },
    appIconVersion: {
        type: Number,
        default: 0
    },
})
const emit = defineEmits(["package-saved", "wails3-saved"])

const { t } = useI18n()
const message = useMessage()

const activeTab = ref("wails3")
const editorRefs = ref({})
const savingAll = ref(false)
const editorNames = ["wails3", "package"]

const tabs = computed(() => [
    {
        name: "wails3",
        label: t("editor.wails3"),
        icon: `
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                <path d="M6 4h20a2 2 0 0 1 2 2v20a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2zm0 2v20h20V6z" fill="currentColor"></path>
                <path d="M10 10h12v2H10z" fill="currentColor"></path>
                <path d="M10 15h8v2h-8z" fill="currentColor"></path>
                <path d="M10 20h10v2H10z" fill="currentColor"></path>
            </svg>
        `,
        component: markRaw(InfoCfgEditor),
        props: {
            wails3Cfg: props.wails3Cfg,
            savingAll: savingAll.value,
            onSaveRequest: saveAllConfigs,
            onSaved: (payload) => emit("wails3-saved", payload)
        }
    },
    {
        name: "package",
        label: t("editor.package"),
        icon: `
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
                <path d="M16 3L4 9v14l12 6l12-6V9zm0 2.24L24.76 9L16 12.76L7.24 9zM6 10.52l9 3.86v11.38l-9-4.5zm11 15.24V14.38l9-3.86v10.74z" fill="currentColor"></path>
            </svg>
        `,
        component: markRaw(PackageCfgEditor),
        props: {
            packageCfg: props.packageCfg,
            projectDir: props.wails3Cfg.projectDir,
            currentPlatform: props.wails3Cfg.project?.currentPlatform,
            appIconVersion: props.appIconVersion,
            savingAll: savingAll.value,
            onSaveRequest: saveAllConfigs,
            onSaved: (cfg) => emit("package-saved", cfg)
        }
    }
])

function setEditorRef(name) {
    return (el) => {
        if (el) {
            editorRefs.value[name] = el
        } else {
            delete editorRefs.value[name]
        }
    }
}

async function saveAllConfigs() {
    if (savingAll.value) {
        return false
    }

    savingAll.value = true
    try {
        for (const name of editorNames) {
            const editor = editorRefs.value[name]
            if (editor?.validate?.() === false) {
                activeTab.value = name
                return false
            }
        }

        const hasChanges = editorNames.some((name) => {
            return editorRefs.value[name]?.isDirty?.() === true
        })
        if (!hasChanges) {
            return true
        }

        for (const name of editorNames) {
            const editor = editorRefs.value[name]
            if (!editor?.save) {
                continue
            }

            const saved = await editor.save()
            if (saved === false) {
                activeTab.value = name
                return false
            }
        }

        message.success(t("editor.saveSuccess"))
        return true
    } finally {
        savingAll.value = false
    }
}

function onKeydown(event) {
    if (
        (event.ctrlKey || event.metaKey) &&
        event.key.toLowerCase() === "s"
    ) {
        event.preventDefault()
        void saveAllConfigs()
    }
}

onMounted(() => {
    window.addEventListener("keydown", onKeydown)
})

onBeforeUnmount(() => {
    window.removeEventListener("keydown", onKeydown)
})

defineExpose({
    saveAll: saveAllConfigs
})
</script>

<style lang="scss" scoped>
.editor-card {
    width: 100%;
    height: 100%;
    min-height: 0;
    box-sizing: border-box;
    padding: 8px 16px 18px 16px;
    border-radius: 18px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-surface-1);
    backdrop-filter: blur(18px);

    display: flex;
    flex-direction: column;
    overflow: hidden;
}


.editor-tabs {
    flex: 1;
    min-height: 0;

}

.tab-label {
    display: flex;
    align-items: center;
    gap: 6px;

}

.tab-content {
    height: 100%;
    min-height: 0;
    box-sizing: border-box;

    overflow: hidden;
}

:deep(.n-tabs) {
    display: flex;
    flex-direction: column;
    height: 100%;
}

:deep(.n-tabs-nav) {
    flex: 0 0 auto;
}

:deep(.n-tabs-tab) {
    color: var(--wm-text-muted);
}

:deep(.n-tabs-tab:hover) {
    color: var(--wm-color-primary-hover);
}

:deep(.n-tabs-tab.n-tabs-tab--active) {
    color: var(--wm-color-primary-hover);
}

:deep(.n-tabs-bar) {
    background: var(--wm-color-primary);
}

:deep(.n-tab-pane) {
    height: 100%;
}

:deep(.n-tabs-pane-wrapper) {
    flex: 1;
    min-height: 0;
}

:deep(.n-tabs-pane-wrapper > .n-tab-pane) {
    height: 100%;
}
</style>
