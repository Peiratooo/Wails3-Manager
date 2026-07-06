<template>
	<div class="selector">
        <div class="top-actions">
            <Environment />
            <n-button class="settings-button" circle quaternary @click="store.panels.settings = true">
                <n-icon size="20">
                    <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 32 32"><g fill="none"><path d="M16 11a5 5 0 1 0 0 10a5 5 0 0 0 0-10zm-3 5a3 3 0 1 1 6 0a3 3 0 0 1-6 0zm-.16 13.628c1.035.247 2.096.372 3.16.372a13.643 13.643 0 0 0 3.156-.375a1.478 1.478 0 0 0 1.13-1.276l.234-2.13a1.471 1.471 0 0 1 2.066-1.2l1.955.856a1.472 1.472 0 0 0 1.671-.345a14.245 14.245 0 0 0 3.156-5.443a1.478 1.478 0 0 0-.535-1.627l-1.729-1.275a1.481 1.481 0 0 1 .003-2.396l1.72-1.27a1.474 1.474 0 0 0 .537-1.63a14.199 14.199 0 0 0-3.157-5.443a1.48 1.48 0 0 0-1.674-.345l-1.946.856a1.483 1.483 0 0 1-2.067-1.2l-.236-2.12a1.476 1.476 0 0 0-1.147-1.283a15.123 15.123 0 0 0-3.127-.363a15.395 15.395 0 0 0-3.146.363a1.469 1.469 0 0 0-1.147 1.28l-.237 2.122a1.493 1.493 0 0 1-2.073 1.206l-1.946-.857a1.493 1.493 0 0 0-1.67.35a14.245 14.245 0 0 0-3.16 5.446a1.478 1.478 0 0 0 .536 1.625l1.725 1.272a1.488 1.488 0 0 1 0 2.397L3.167 18.47a1.477 1.477 0 0 0-.535 1.63a14.253 14.253 0 0 0 3.16 5.45a1.458 1.458 0 0 0 1.077.465c.203 0 .404-.042.591-.123l1.955-.859a1.485 1.485 0 0 1 2.065 1.2l.235 2.126a1.476 1.476 0 0 0 1.125 1.27zm5.501-1.866a11.638 11.638 0 0 1-4.677 0l-.195-1.74a3.48 3.48 0 0 0-1.14-2.208a3.534 3.534 0 0 0-3.718-.6l-1.606.7a12.237 12.237 0 0 1-2.348-4.05l1.424-1.052a3.488 3.488 0 0 0 0-5.616L4.66 12.147a12.243 12.243 0 0 1 2.348-4.046l1.6.7a3.45 3.45 0 0 0 1.4.294a3.5 3.5 0 0 0 3.467-3.108l.194-1.747c.774-.15 1.56-.23 2.347-.24c.782.01 1.562.09 2.33.24l.186 1.74a3.48 3.48 0 0 0 1.137 2.216a3.525 3.525 0 0 0 3.727.6l1.6-.7a12.212 12.212 0 0 1 2.35 4.047l-1.423 1.046a3.48 3.48 0 0 0 0 5.62l1.422 1.05A12.273 12.273 0 0 1 25 23.901l-1.6-.7a3.473 3.473 0 0 0-4.866 2.81l-.193 1.75z" fill="currentColor"></path></g></svg>
                </n-icon>
            </n-button>
        </div>
        <div class="title">
            <div class="welcome">{{ t('home.welcome') }}</div>
            <div class="desc">{{ t('home.desc') }}</div>
        </div>
        <div class="panel">
            <div class="operator">
                <div class="left">
                    {{ t('home.recently') }}
                </div>
                <div class="right">
                    <div class="search">
                        <n-input v-model:value="keywords" :placeholder="t('home.searchProject')" />
                    </div>
                    <div class="import">
                        <n-button
                            :loading="importingProject"
                            :disabled="importingProject"
                            @click="importProject()"
                        >
                            <n-icon size="18" style="margin-right: 8px">
                                <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 48 48"><g fill="none"><path d="M40.75 24c.69 0 1.25-.56 1.25-1.25v-9.5A7.25 7.25 0 0 0 34.75 6h-21.5A7.25 7.25 0 0 0 6 13.25v21.5A7.25 7.25 0 0 0 13.25 42h9.5a1.25 1.25 0 1 0 0-2.5h-9.5a4.75 4.75 0 0 1-4.75-4.75v-21.5a4.75 4.75 0 0 1 4.75-4.75h21.5a4.75 4.75 0 0 1 4.75 4.75v9.5c0 .69.56 1.25 1.25 1.25zm-21.5-6c-.69 0-1.25.56-1.25 1.25v13.5a1.25 1.25 0 1 0 2.5 0V22.268l15.366 15.366a1.25 1.25 0 1 0 1.768-1.768L22.268 20.5H32.75a1.25 1.25 0 1 0 0-2.5h-13.5z" fill="currentColor"></path></g></svg>
                            </n-icon>
                            <span>
                                {{ t('home.importProject') }}
                            </span>
                        </n-button>
                    </div>
                    <div class="create-project">
                        <n-button @click="store.panels.creator = true">
                            <n-icon size="18" style="margin-right: 8px">
                                <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 48 48"><g fill="none"><path d="M17.06 9c.833 0 1.64.277 2.295.784l.175.144l2.586 2.263c.19.166.425.27.673.3l.15.009H40.25a3.75 3.75 0 0 1 3.745 3.55l.005.2v7.806a12.024 12.024 0 0 0-2.5-1.724V16.25a1.25 1.25 0 0 0-1.122-1.243L40.25 15l-17.403-.001l-2.127 2.616a3.75 3.75 0 0 1-2.685 1.378L17.81 19L6.5 18.999V35.25c0 .647.492 1.18 1.122 1.243l.128.007h16.769c.268.881.635 1.719 1.087 2.501L7.75 39a3.75 3.75 0 0 1-3.745-3.55L4 35.25v-22.5a3.75 3.75 0 0 1 3.55-3.745L7.75 9h9.31zm0 2.5H7.75a1.25 1.25 0 0 0-1.244 1.122l-.006.128v3.749l11.31.001c.33 0 .643-.13.876-.358l.094-.104l1.635-2.013l-2.531-2.216a1.25 1.25 0 0 0-.673-.3l-.15-.009zM36 23c5.523 0 10 4.477 10 10s-4.477 10-10 10s-10-4.477-10-10s4.477-10 10-10zm0 4a1 1 0 0 0-.993.883L35 28v4h-4a1 1 0 0 0-.993.883L30 33a1 1 0 0 0 .883.993L31 34h4v4a1 1 0 0 0 .883.993L36 39a1 1 0 0 0 .993-.883L37 38v-4h4a1 1 0 0 0 .993-.883L42 33a1 1 0 0 0-.883-.993L41 32h-4v-4a1 1 0 0 0-.883-.993L36 27z" fill="currentColor"></path></g></svg>
                            </n-icon>
                            <span>
                                {{ t('home.createProject') }}
                            </span>
                        </n-button>
                    </div>
                </div>
            </div>
            <div class="project-table">
                <n-data-table :columns="columns" :data="filteredProjects" :rowProps="rowProps"/>
            </div>
        </div>
        <Settings />
        <Creator />
    </div>
</template>

<script setup>
import { useI18n } from '../i18n'
import {onMounted, ref, computed, h, inject} from "vue";
import { Service as ProjectService } from "../../bindings/wails3-manager/core/project"
import { Service as SettingsService } from "../../bindings/wails3-manager/core/settings"
import {AppService} from "../../bindings/wails3-manager/desktop"
import {NInput,NButton,NIcon,NImage,NDataTable,NPopover,useMessage} from "naive-ui"
import Settings from '../components/Settings.vue'
import Environment from "../components/Environment.vue"
import router from "../router/index.js";
import Creator from "../components/Creator/Creator.vue";
import { localFileUrl } from "../utils/files"
const { t } = useI18n()
const message = useMessage()
const store = inject("store")
const keywords = ref("")
const formatTimestamp = inject("formatTimestamp")
const projects = ref([])
const deletePopoverProjectDir = ref("")
const importingProject = ref(false)

const rowProps = (row) => {
    return {
        style: {
            cursor: 'pointer'
        },
        onClick: (event) => {
            if (event.target.closest('[data-action="delete-popover"]')) return

            router.push({
                name: "project",
                query: {
                    projectDir: encodeURIComponent(row.projectDir)
                }
            })
        }
    }
}

async function importProject() {
    if (importingProject.value) return

    importingProject.value = true
    try {
        const res = await AppService.ChooseFolder()
        if (res) {
            await ProjectService.ImportProject(res)
            await initProjects()
        }
    } catch (error) {
        message.error(error?.message || String(error))
    } finally {
        importingProject.value = false
    }
}

async function removeProject(project, restoreOriginal) {
    try {
        deletePopoverProjectDir.value = ""
        await SettingsService.RemoveProject(project.projectDir, restoreOriginal)
        await initProjects()
    } catch (error) {
        message.error(error?.message || String(error))
    }
}

const columns = computed(() => [
    {
        title: t('home.projectIcon'),
        key: 'projectIcon',
        render(row) {
            return h(NImage, {
                src: localFileUrl(row.iconPath),
                width: 32,
                height: 32,
                showToolbar: false
            })
        },
        width: 64
    },
    {
        title: t('home.projectName'),
        key: 'projectName',
        render(row) {
            return row.project.wailsConfig.info.productName || '-'
        },
        width: 150
    },
    {
        title: t('home.projectPath'),
        key: 'projectDir',
        width: 380,
    },
    {
        title: t('home.openTime'),
        key: 'lastOpenedAt',
        render(row) {
            return formatTimestamp(row.lastOpenedAt)
        },
        width: 150

    },
    {
        title: t('home.importTime'),
        key: 'importedAt',
        render(row) {
            return formatTimestamp(row.importedAt)
        },
        width: 150
    },
    {
        title: '',
        key: 'delete',
        render(row) {
            return h(
                NPopover,
                {
                    trigger: 'click',
                    placement: 'left',
                    show: deletePopoverProjectDir.value === row.projectDir,
                    'onUpdate:show': (show) => {
                        deletePopoverProjectDir.value = show ? row.projectDir : ""
                    }
                },
                {
                    trigger: () => h(
                        'div',
                        {
                            class: 'delete-trigger',
                            type: 'button',
                            title: t('home.deleteProject'),
                            'aria-label': t('home.deleteProject'),
                            'data-action': 'delete-popover',
                            onClick: (event) => event.stopPropagation()
                        },
                        [
                            h(NIcon, {
                                class: 'delete-icon',
                                innerHTML: `
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
      <path d="M12 12h2v12h-2z" fill="currentColor"></path>
      <path d="M18 12h2v12h-2z" fill="currentColor"></path>
      <path d="M4 6v2h2v20a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V8h2V6zm4 22V8h16v20z" fill="currentColor"></path>
      <path d="M12 2h8v2h-8z" fill="currentColor"></path>
    </svg>
  `
                            })
                        ]
                    ),
                    default: () => h(
                        'div',
                        {
                            class: 'delete-popover',
                            'data-action': 'delete-popover',
                            onClick: (event) => event.stopPropagation()
                        },
                        [
                            h('div', { class: 'delete-popover-title' }, t('home.deleteProject')),
                            h('div', { class: 'delete-popover-actions' }, [
                                h(
                                    NButton,
                                    {
                                        size: 'small',
                                        type: 'error',
                                        onClick: () => removeProject(row, true)
                                    },
                                    { default: () => t('home.deleteRestore') }
                                ),
                                h(
                                    NButton,
                                    {
                                        size: 'small',
                                        onClick: () => removeProject(row, false)
                                    },
                                    { default: () => t('home.deleteKeep') }
                                ),
                                h(
                                    NButton,
                                    {
                                        size: 'small',
                                        onClick: () => {
                                            deletePopoverProjectDir.value = ""
                                        }
                                    },
                                    { default: () => t('editor.cancel') }
                                )
                            ])
                        ]
                    )
                }
            )
        }
    }
])
const filteredProjects = computed(() => {
    const keyword = keywords.value.trim().toLowerCase()

    if (!keyword) {
        return projects.value
    }

    return projects.value.filter((item) => {
        return JSON.stringify(item).toLowerCase().includes(keyword)
    })
})



async function initProjects() {
    const res = await SettingsService.ListProjects()
    projects.value = []
    for (let i of res) {
        i["iconPath"] = await SettingsService.GetABSPath(i.projectDir,i.project.wailsConfig.icon)
        projects.value.push(i)
    }
}

onMounted(()=>{
    initProjects().catch((error) => {
        message.error(error?.message || String(error))
    })
})
</script>

<style lang="scss" scoped>


.selector {
    position: relative;
    width: 100%;
    min-height: 100vh;
    box-sizing: border-box;
    padding: 48px 38px 44px;
    overflow: hidden;
    color: var(--wm-text-secondary);
    background: var(--wm-bg-page-gradient);
}

.top-actions {
    position: absolute;
    top: 24px;
    right: 28px;
    display: flex;
    align-items: center;
    gap: 10px;
}

.settings-button {
    color: var(--wm-text-secondary);
    display: flex;
    align-items: center;
}

.settings-button:hover {
    color: var(--wm-color-primary-hover);
}

.title {
    text-align: center;
    margin-top: 2px;
    margin-bottom: 66px;

    .welcome {
        font-size: 32px;
        letter-spacing: 0.02em;
        color: var(--wm-text-primary);
    }

    .desc {
        margin-top: 10px;
        font-size: 16px;
        color: var(--wm-text-muted);
    }
}

.panel {
    width: min(80%, 100%);
    margin: 0 auto;
}

.operator {
    display: flex;
    align-items: center;
    justify-content: space-between;

    margin-bottom: 24px;

    .left {
        flex: 0 0 auto;
        font-size: 16px;
        font-weight: 500;
        color: var(--wm-text-primary);
    }

    .right {
        display: flex;
        align-items: center;
        gap: 12px;
    }
}

.search {
    display: flex;
    align-items: center;
    gap: 8px;


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

.operator > .right > :deep(.n-button:last-child) {
    --n-color: var(--wm-color-primary) !important;
    --n-color-hover: var(--wm-color-primary-hover) !important;
    --n-color-pressed: var(--wm-color-primary-pressed) !important;
    --n-border: 1px solid var(--wm-color-primary-border) !important;
    --n-border-hover: 1px solid var(--wm-border-strong) !important;
    --n-border-pressed: 1px solid var(--wm-color-primary-border) !important;
    --n-text-color: var(--wm-text-inverse) !important;
    --n-text-color-hover: var(--wm-text-inverse) !important;
    --n-text-color-pressed: var(--wm-text-inverse) !important;

    box-shadow: var(--wm-shadow-primary);
}

.project-table {
    width: 100%;
}

:deep(.n-data-table) {

    --n-th-color: transparent !important;
    --n-th-color-hover: transparent !important;
    --n-td-color: transparent !important;
    --n-td-color-hover: var(--wm-table-row-hover) !important;
    --n-border-color: var(--wm-border-subtle) !important;
    --n-th-text-color: var(--wm-table-header-text) !important;
    --n-td-text-color: var(--wm-table-row-text) !important;
    background: transparent;
}

:deep(.n-data-table-wrapper),
:deep(.n-data-table-base-table),
:deep(.n-data-table-table),
:deep(.n-data-table-thead),
:deep(.n-data-table-tbody) {
    background: transparent;
}

:deep(.n-data-table-th) {
    height: 52px;


    font-weight: 400;
    color: var(--wm-table-header-text);
    background: transparent;
    border-bottom: none;
}

:deep(.n-data-table-td) {
    font-size: 12px;
    color: var(--wm-table-row-text);
    background: transparent;
    border-bottom: 1px solid var(--wm-border-subtle);
    cursor: pointer;
}

:deep(.n-data-table-tr:hover .n-data-table-td) {
    background: var(--wm-table-row-hover);
}


.import *,.create-project *{
    cursor: pointer !important;
}

.delete-trigger {
    padding: 0;
    border: 0;
    display: inline-flex;
    color: inherit;
    background: transparent;
    cursor: pointer;
}

@media (max-width: 700px) {
    .selector {
        min-height: 100vh;
        padding: 24px 20px 34px;
        overflow-y: auto;
    }

    .top-actions {
        top: 24px;
        right: 28px;
    }

    .title {
        margin-top: 36px;
        margin-bottom: 38px;

        .welcome {
            font-size: 30px;
            line-height: 1.3;
        }

        .desc {
            font-size: 14px;
        }
    }

    .panel {
        width: 100%;
    }

    .operator {
        align-items: flex-start;
        flex-direction: column;
        gap: 12px;
        margin-bottom: 18px;

        .right {
            width: 100%;
            flex-wrap: wrap;
            align-items: stretch;
        }
    }

    .search {
        flex: 1 0 100%;
        width: 100%;
    }

    .search :deep(.n-input) {
        width: 100%;
    }

    .import,
    .create-project {
        flex: 1 1 0;
        min-width: 0;
    }

    .import :deep(.n-button),
    .create-project :deep(.n-button) {
        width: 100%;
    }

    .project-table {
        overflow-x: auto;
        -webkit-overflow-scrolling: touch;
    }

    .project-table :deep(.n-data-table) {
        min-width: 840px;
    }
}
</style>
