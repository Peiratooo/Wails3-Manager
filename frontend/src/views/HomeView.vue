<template>
	<div class="selector">
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
                        <n-button @click="importProject()">
                            <n-icon size="18" style="margin-right: 8px">
                                <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 48 48"><g fill="none"><path d="M40.75 24c.69 0 1.25-.56 1.25-1.25v-9.5A7.25 7.25 0 0 0 34.75 6h-21.5A7.25 7.25 0 0 0 6 13.25v21.5A7.25 7.25 0 0 0 13.25 42h9.5a1.25 1.25 0 1 0 0-2.5h-9.5a4.75 4.75 0 0 1-4.75-4.75v-21.5a4.75 4.75 0 0 1 4.75-4.75h21.5a4.75 4.75 0 0 1 4.75 4.75v9.5c0 .69.56 1.25 1.25 1.25zm-21.5-6c-.69 0-1.25.56-1.25 1.25v13.5a1.25 1.25 0 1 0 2.5 0V22.268l15.366 15.366a1.25 1.25 0 1 0 1.768-1.768L22.268 20.5H32.75a1.25 1.25 0 1 0 0-2.5h-13.5z" fill="currentColor"></path></g></svg>
                            </n-icon>
                            <label>
                                {{ t('home.importProject') }}
                            </label>
                        </n-button>
                    </div>

                </div>
            </div>
            <div class="project-table">
                <n-data-table :columns="columns" :data="filteredProjects" :rowProps="rowProps"/>
            </div>
        </div>
    </div>
</template>

<script setup>
import { useI18n } from '../i18n'
import {onMounted, ref, computed, h, inject} from "vue";
import {ProjectService} from "../../bindings/wails3-manager/core/project"
import {SettingsService} from "../../bindings/wails3-manager/core/settings"
import {AppService} from "../../bindings/wails3-manager/desktop"
import {NInput,NButton,NIcon,NImage,NDataTable} from "naive-ui"
const { t } = useI18n()

const keywords = ref("")

const projects = ref([])
const store = inject("store")
const rowProps = (row) => {
    return {
        style: {
            cursor: 'pointer'
        },
        onClick: (event) => {
            const delEl =  event.target.closest('[data-action="delete"]')
            if (delEl) {
                removeProject(row)
            } else {
                handleRowClick( row)
            }

        }
    }
}

const handleRowClick = (row) => {
    console.log(row)

}
async function importProject() {
    // ProjectService.ImportProject()
    const res = await AppService.ChooseFolder()
    if (res) {
       await ProjectService.ImportProject(res).then((resp)=>{
           if (resp) {
               initProjects()
           }
       }).catch(e=>{
           console.error(e)
       })
    }
}

function removeProject(project) {
    SettingsService.RemoveProject(project.projectDir,true).then((res)=>{

    }).finally(()=>{
        initProjects()
    })

}

const columns = computed(() => [
    {
        title: t('home.projectIcon'),
        key: 'projectIcon',
        render(row) {
            return h(NImage, {
                src: "/local/file?"+ row.iconPath,
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
                NIcon,
                {
                    class: 'delete-icon',
                    'data-action': 'delete'
                },
                {
                    default: () => h('span',{
                        innerHTML:`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 32 32"><path d="M12 12h2v12h-2z" fill="currentColor"></path><path d="M18 12h2v12h-2z" fill="currentColor"></path><path d="M4 6v2h2v20a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V8h2V6zm4 22V8h16v20z" fill="currentColor"></path><path d="M12 2h8v2h-8z" fill="currentColor"></path></svg>`
                    })
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

const formatTimestamp = (timestamp) => {
    if (!timestamp) return '-'

    const date = new Date(timestamp * 1000)

    const pad = (n) => String(n).padStart(2, '0')

    const year = date.getFullYear()
    const month = pad(date.getMonth() + 1)
    const day = pad(date.getDate())
    const hour = pad(date.getHours())
    const minute = pad(date.getMinutes())

    return `${year}-${month}-${day} ${hour}:${minute}`
}

function initProjects() {
    SettingsService.ListProjects().then(async (res)=>{
        projects.value = []
        for (let i of res) {
            console.log(i)
            i["iconPath"] = await SettingsService.GetABSPath(i.projectDir,i.project.wailsConfig.icon)
            projects.value.push(i)
        }
    })
}

onMounted(()=>{
    initProjects()
    SettingsService.GetSettings().then((res)=>{
        console.log(res)
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
    --n-color: linear-gradient(180deg, var(--wm-color-logo-start) 0%, var(--wm-color-primary) 100%) !important;
    --n-color-hover: linear-gradient(180deg, var(--wm-color-primary-hover) 0%, var(--wm-color-primary) 100%) !important;
    --n-color-pressed: linear-gradient(180deg, var(--wm-color-primary) 0%, var(--wm-color-primary-pressed) 100%) !important;
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

.delete-icon {
    cursor: pointer;
    color: var(--wm-text-muted);
    font-size: 17px;
    transition:
        color 0.18s ease,
        opacity 0.18s ease,
        transform 0.18s ease;
}

.delete-icon:hover {
    color: var(--wm-color-danger);
    transform: scale(1.06);
}
.import *{
    cursor: pointer !important;
}
</style>
