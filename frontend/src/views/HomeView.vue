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
                        <n-input :placeholder="t('home.searchProject')" v-model:value="keywords" />
                        <n-button>{{ t('home.search') }}</n-button>
                    </div>
                    <n-button>
                        <n-icon></n-icon>
                        <label>
                            {{ t('home.importProject') }}
                        </label>
                    </n-button>
                </div>
            </div>
            <div class="project-table">
                <n-data-table :columns="columns" :data="projects" />
            </div>
        </div>
    </div>
</template>

<script setup>
import { useI18n } from '../i18n'
import {onMounted, ref, computed, h, inject} from "vue";
import {ProjectService} from "../../bindings/wails3-manager/core/project"
import {SettingsService} from "../../bindings/wails3-manager/core/settings"
import {NInput,NButton,NIcon,NImage,NDataTable} from "naive-ui"
const { t } = useI18n()

const keywords = ref("")

const projects = ref([])
const store = inject("store")
const columns = computed(() => [
    {
        title: t('home.projectIcon'),
        key: 'projectIcon',
        render(row) {
            return h(NImage, {
                src: "/local/file?"+ row.iconPath
            })
        }
    },
    {
        title: t('home.projectName'),
        key: 'projectName',
        render(row) {
            return row.project.wailsConfig.info.productName || '-'
        }
    },
    {
        title: t('home.projectPath'),
        key: 'projectDir'
    },
    {
        title: t('home.openTime'),
        key: 'lastOpenedAt',
        render(row) {
            return formatTimestamp(row.lastOpenedAt)
        }
    },
    {
        title: t('home.importTime'),
        key: 'importedAt',
        render(row) {
            return formatTimestamp(row.importedAt)
        }
    },
    {
        title: '',
        key: 'delete',
        render(row) {
            return h(
                NIcon,
                {
                    class: 'delete-icon',
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
onMounted(()=>{
   SettingsService.ListProjects().then(async (res)=>{
       for (let i of res) {
           i["iconPath"] = await SettingsService.GetABSPath(i.projectDir,i.project.wailsConfig.icon)
           projects.value.push(i)

       }
   })
})
</script>

<style lang="scss" scoped>


.selector {
    background: #0f1115;
    position: relative;
    width: 100%;
    min-height: 100vh;
    box-sizing: border-box;
    padding: 48px 38px 44px;
    overflow: hidden;

    color: rgba(235, 239, 245, 0.86);
    background:
        radial-gradient(circle at 50% 0%, rgba(80, 90, 110, 0.18), transparent 34%),
        radial-gradient(circle at 18% 92%, rgba(80, 90, 110, 0.10), transparent 36%),
        linear-gradient(135deg, #15181d 0%, #101318 48%, #15181d 100%);

    border: 1px solid rgba(255, 255, 255, 0.07);

    box-shadow:
        inset 0 1px 0 rgba(255, 255, 255, 0.04),
        0 24px 80px rgba(0, 0, 0, 0.35);
}

.title {
    text-align: center;
    margin-top: 2px;
    margin-bottom: 66px;

    .welcome {
        font-size: 21px;
        font-weight: 500;
        letter-spacing: 0.02em;
        color: rgba(245, 247, 250, 0.88);
    }

    .desc {
        margin-top: 10px;
        font-size: 12px;
        line-height: 1.4;
        color: rgba(210, 216, 225, 0.48);
    }
}

.panel {
    width: min(760px, 100%);
    margin: 0 auto;
}

.operator {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    margin-bottom: 24px;

    .left {
        flex: 0 0 auto;
        font-size: 13px;
        font-weight: 500;
        color: rgba(245, 247, 250, 0.86);
    }

    .right {
        display: flex;
        align-items: center;
        gap: 10px;
    }
}

.search {
    display: flex;
    align-items: center;
    gap: 8px;

    :deep(.n-input) {
        width: 130px;
    }
}

:deep(.n-input) {
    --n-height: 28px !important;
    --n-color: rgba(255, 255, 255, 0.035) !important;
    --n-color-focus: rgba(255, 255, 255, 0.05) !important;
    --n-border: 1px solid rgba(255, 255, 255, 0.08) !important;
    --n-border-hover: 1px solid rgba(255, 255, 255, 0.14) !important;
    --n-border-focus: 1px solid rgba(70, 130, 255, 0.52) !important;
    --n-box-shadow-focus: 0 0 0 2px rgba(70, 130, 255, 0.12) !important;
    --n-text-color: rgba(240, 244, 250, 0.84) !important;
    --n-placeholder-color: rgba(210, 216, 225, 0.34) !important;
    --n-border-radius: 4px !important;

    font-size: 11px;
}

:deep(.n-button) {
    --n-height: 28px !important;
    --n-padding: 0 12px !important;
    --n-font-size: 11px !important;
    --n-border-radius: 4px !important;
    --n-text-color: rgba(235, 239, 245, 0.86) !important;
    --n-text-color-hover: rgba(255, 255, 255, 0.96) !important;
    --n-text-color-pressed: rgba(255, 255, 255, 0.96) !important;
    --n-color: rgba(255, 255, 255, 0.045) !important;
    --n-color-hover: rgba(255, 255, 255, 0.075) !important;
    --n-color-pressed: rgba(255, 255, 255, 0.055) !important;
    --n-border: 1px solid rgba(255, 255, 255, 0.08) !important;
    --n-border-hover: 1px solid rgba(255, 255, 255, 0.14) !important;
    --n-border-pressed: 1px solid rgba(255, 255, 255, 0.12) !important;

    backdrop-filter: blur(12px);
}

.operator > .right > :deep(.n-button:last-child) {
    --n-color: linear-gradient(180deg, #3f78ff 0%, #2861df 100%) !important;
    --n-color-hover: linear-gradient(180deg, #4a82ff 0%, #306beb 100%) !important;
    --n-color-pressed: linear-gradient(180deg, #2e64e6 0%, #2356c8 100%) !important;
    --n-border: 1px solid rgba(105, 150, 255, 0.36) !important;
    --n-border-hover: 1px solid rgba(132, 170, 255, 0.48) !important;
    --n-border-pressed: 1px solid rgba(105, 150, 255, 0.32) !important;
    --n-text-color: #ffffff !important;
    --n-text-color-hover: #ffffff !important;
    --n-text-color-pressed: #ffffff !important;

    box-shadow: 0 8px 18px rgba(43, 98, 220, 0.22);
}

.project-table {
    width: 100%;
}

:deep(.n-data-table) {
    --n-font-size: 11px !important;
    --n-th-color: transparent !important;
    --n-th-color-hover: transparent !important;
    --n-td-color: transparent !important;
    --n-td-color-hover: rgba(255, 255, 255, 0.035) !important;
    --n-border-color: rgba(255, 255, 255, 0.055) !important;
    --n-th-text-color: rgba(210, 216, 225, 0.48) !important;
    --n-td-text-color: rgba(225, 230, 238, 0.70) !important;

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
    height: 34px;
    padding: 0 12px;
    font-size: 11px;
    font-weight: 400;
    color: rgba(210, 216, 225, 0.48);
    background: transparent;
    border-bottom: none;
}

:deep(.n-data-table-td) {
    height: 46px;
    padding: 0 12px;
    font-size: 11px;
    color: rgba(225, 230, 238, 0.70);
    background: transparent;
    border-bottom: 1px solid rgba(255, 255, 255, 0.052);
}

:deep(.n-data-table-tr:hover .n-data-table-td) {
    background: rgba(255, 255, 255, 0.035);
}

:deep(.n-data-table-th:first-child),
:deep(.n-data-table-td:first-child) {
    width: 68px;
    padding-left: 12px;
}

:deep(.n-image) {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
}

:deep(.n-image img) {
    width: 24px;
    height: 24px;
    object-fit: contain;
    border-radius: 6px;
}

.delete-icon {
    cursor: pointer;
    color: rgba(210, 216, 225, 0.32);
    font-size: 17px;
    transition:
        color 0.18s ease,
        opacity 0.18s ease,
        transform 0.18s ease;
}

.delete-icon:hover {
    color: #ff5f57;
    transform: scale(1.06);
}

@media (max-width: 720px) {
    .selector {
        padding: 42px 22px 32px;
    }

    .title {
        margin-bottom: 42px;
    }

    .operator {
        align-items: flex-start;
        flex-direction: column;
        gap: 14px;

        .right {
            width: 100%;
            justify-content: space-between;
        }
    }

    .search {
        flex: 1;

        :deep(.n-input) {
            width: 100%;
        }
    }
}
</style>
