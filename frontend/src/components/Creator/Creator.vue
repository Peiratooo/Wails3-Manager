<template>
    <n-modal
        v-model:show="store.panels.creator"
        class="creator-modal"
        :auto-focus="false"
        :trap-focus="false"
        :close-on-esc="false"
        :mask-closable="false"
    >
        <section
            class="creator-card"
            role="dialog"
            aria-modal="true"
            aria-labelledby="creator-title"
        >
            <header class="creator-head">
                <div class="creator-heading">
                    <div
                        id="creator-title"
                        class="creator-title"
                    >
                        {{ t("createProject.title") }}
                    </div>

                    <div class="creator-subtitle">
                        {{ t("createProject.description") }}
                    </div>
                </div>

                <div class="creator-head-actions">

                    <n-button
                        class="close-button"
                        quaternary
                        circle
                        size="small"
                        :disabled="creating"
                        :aria-label="t('editor.cancel')"
                        :title="t('editor.cancel')"
                        @click="requestClose"
                    >
                        <n-icon size="17">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                viewBox="0 0 32 32"
                                aria-hidden="true"
                            >
                                <path
                                    d="M17.4 16 24 9.4 22.6 8 16 14.6 9.4 8 8 9.4l6.6 6.6L8 22.6 9.4 24l6.6-6.6 6.6 6.6 1.4-1.4z"
                                    fill="currentColor"
                                />
                            </svg>
                        </n-icon>
                    </n-button>
                </div>
            </header>

            <transition name="close-confirm">
                <div
                    v-show="showCloseConfirm"
                    class="close-confirm"
                    role="alertdialog"
                    aria-live="assertive"
                >
                    <div
                        class="close-confirm-icon"
                        aria-hidden="true"
                    >
                        <n-icon size="17">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                viewBox="0 0 32 32"
                            >
                                <path
                                    d="M16 3 2.5 27h27zm0 5.1L25.2 25H6.8z"
                                    fill="currentColor"
                                />
                                <path
                                    d="M15 12h2v7h-2zm0 9h2v2h-2z"
                                    fill="currentColor"
                                />
                            </svg>
                        </n-icon>
                    </div>

                    <div class="close-confirm-text">
                        {{ t("createProject.messages.closeConfirm") }}
                    </div>

                    <div class="close-confirm-actions">
                        <n-button
                            size="small"
                            @click="showCloseConfirm = false"
                        >
                            {{ t("createProject.actions.keepEditing") }}
                        </n-button>

                        <n-button
                            size="small"
                            type="primary"
                            @click="closeCreator"
                        >
                            {{ t("createProject.actions.discard") }}
                        </n-button>
                    </div>
                </div>
            </transition>

            <div
                class="creator-body"
                @focusin="showCloseConfirm = false"
            >
                <aside class="creator-steps">
                    <div class="steps-caption">
                        {{ t("createProject.title") }}
                    </div>

                    <n-steps
                        vertical
                        size="small"
                        :current="activeStep + 1"
                    >
                        <n-step
                            v-for="step in steps"
                            :key="step.key"
                            :title="step.title"
                        />
                    </n-steps>
                </aside>

                <main class="creator-panel">
                    <transition
                        name="creator-step"
                        mode="out-in"
                    >
                        <div
                            :key="currentStep.key"
                            class="step-panel"
                        >
                            <div class="panel-head">
                                <div class="panel-kicker">
                                    {{ activeStep + 1 }}/{{ steps.length }}
                                </div>

                                <div class="panel-title">
                                    {{ currentStep.title }}
                                </div>

                                <div class="panel-desc">
                                    {{ currentStep.description }}
                                </div>
                            </div>

                            <n-scrollbar class="panel-scroll">
                                <!-- 基础信息 -->
                                <n-form
                                    v-if="currentStep.key === 'basic'"
                                    class="creator-form"
                                    :model="config"
                                    label-placement="top"
                                    :show-require-mark="false"
                                    size="large"
                                >
                                    <div class="form-grid">
                                        <n-form-item
                                            v-for="field in basicFields"
                                            :key="field.path"
                                            :path="field.path"
                                            :validation-status="
                                                visibleFieldError(field.path)
                                                    ? 'error'
                                                    : undefined
                                            "
                                            :feedback="
                                                visibleFieldError(field.path) ||
                                                undefined
                                            "
                                        >
                                            <template #label>
                                                <div class="field-label">
                                                    <span>{{ field.label }}</span>

                                                    <field-help
                                                        :content="field.description"
                                                    />
                                                </div>
                                            </template>

                                            <n-input
                                                v-model:value="config[field.path]"
                                                clearable
                                                :maxlength="field.maxlength"
                                                :disabled="creating"
                                                :placeholder="field.placeholder"
                                                @blur="touchField(field.path)"
                                            />
                                        </n-form-item>

                                        <n-form-item
                                            class="full"
                                            path="dir"
                                            :validation-status="
                                                visibleFieldError('dir')
                                                    ? 'error'
                                                    : undefined
                                            "
                                            :feedback="
                                                visibleFieldError('dir') ||
                                                undefined
                                            "
                                        >
                                            <template #label>
                                                <div class="field-label">
                                                    <span>
                                                        {{
                                                            t(
                                                                "createProject.fields.dir.label"
                                                            )
                                                        }}
                                                    </span>

                                                    <field-help
                                                        :content="
                                                            t(
                                                                'createProject.fields.dir.description'
                                                            )
                                                        "
                                                    />
                                                </div>
                                            </template>

                                            <n-input-group>
                                                <n-input
                                                    v-model:value="config.dir"
                                                    readonly
                                                    :disabled="creating"
                                                    :placeholder="
                                                        t(
                                                            'createProject.fields.dir.placeholder'
                                                        )
                                                    "
                                                    @click="chooseDirectory"
                                                    @blur="touchField('dir')"
                                                >
                                                    <template #prefix>
                                                        <n-icon size="18" style="margin-right: 6px;">
                                                            <svg
                                                                xmlns="http://www.w3.org/2000/svg"
                                                                viewBox="0 0 32 32"
                                                                aria-hidden="true"
                                                            >
                                                                <path
                                                                    d="M28 8H15.414l-2.707-2.707A1 1 0 0 0 12 5H4a2 2 0 0 0-2 2v18a2 2 0 0 0 2 2h24a2 2 0 0 0 2-2V10a2 2 0 0 0-2-2zM4 7h7.586l2.707 2.707A1 1 0 0 0 15 10h13v3H4zm0 18V15h24v10z"
                                                                    fill="currentColor"
                                                                />
                                                            </svg>
                                                        </n-icon>
                                                    </template>
                                                </n-input>

                                                <n-button
                                                    :disabled="creating"
                                                    @click="chooseDirectory"
                                                >
                                                    {{
                                                        t(
                                                            "createProject.actions.selectDirectory"
                                                        )
                                                    }}
                                                </n-button>
                                            </n-input-group>
                                        </n-form-item>
                                    </div>
                                </n-form>

                                <!-- 模板 -->
                                <n-form
                                    v-else-if="currentStep.key === 'template'"
                                    class="creator-form template-form"
                                    :model="config"
                                    label-placement="top"
                                    :show-require-mark="false"
                                    size="large"
                                >
                                    <div class="template-layout">
                                        <n-form-item
                                            class="template-select-item"
                                            path="template"
                                            :validation-status="
                                                visibleFieldError('template')
                                                    ? 'error'
                                                    : undefined
                                            "
                                            :feedback="
                                                visibleFieldError('template') ||
                                                undefined
                                            "
                                        >
                                            <template #label>
                                                <div class="field-label">
                                                    <span>
                                                        {{
                                                            t(
                                                                "createProject.fields.template.label"
                                                            )
                                                        }}
                                                    </span>

                                                    <field-help
                                                        :content="
                                                            t(
                                                                'createProject.fields.template.description'
                                                            )
                                                        "
                                                    />
                                                </div>
                                            </template>

                                            <div class="template-select-card">
                                                <div class="template-source-row">
                                                    <span class="template-source-caption">
                                                        {{
                                                            t(
                                                                "createProject.fields.template.label"
                                                            )
                                                        }}
                                                    </span>

                                                    <div
                                                        class="template-source-toggle"
                                                        role="group"
                                                        :aria-label="
                                                            t(
                                                                'createProject.fields.template.label'
                                                            )
                                                        "
                                                    >
                                                        <n-button
                                                            v-for="option in templateSourceOptions"
                                                            :key="option.value"
                                                            size="small"
                                                            :type="
                                                                templateSource === option.value
                                                                    ? 'primary'
                                                                    : 'default'
                                                            "
                                                            :disabled="creating"
                                                            @click="
                                                                setTemplateSource(
                                                                    option.value
                                                                )
                                                            "
                                                        >
                                                            {{ option.label }}
                                                        </n-button>
                                                    </div>
                                                </div>

                                                <div class="template-control">
                                                    <n-select
                                                        v-if="templateSource === 'official'"
                                                        v-model:value="config.template"
                                                        clearable
                                                        filterable
                                                        :consistent-menu-width="false"
                                                        :disabled="creating"
                                                        :loading="loadingTemplates"
                                                        :options="templateOptions"
                                                        :placeholder="
                                                            t(
                                                                'createProject.fields.template.placeholder'
                                                            )
                                                        "
                                                        @blur="touchField('template')"
                                                    />

                                                    <n-input
                                                        v-else
                                                        v-model:value="config.template"
                                                        clearable
                                                        :disabled="creating"
                                                        :placeholder="
                                                            t(
                                                                'createProject.fields.template.customPlaceholder'
                                                            )
                                                        "
                                                        @blur="touchField('template')"
                                                    />
                                                </div>
                                            </div>
                                        </n-form-item>

                                        <div class="option-card template-mode-card">
                                            <div class="option-copy">
                                                <div class="field-label">
                                                    <span>
                                                        {{
                                                            t(
                                                                "createProject.fields.useInterfaces.label"
                                                            )
                                                        }}
                                                    </span>

                                                    <field-help
                                                        :content="
                                                            t(
                                                                'createProject.fields.useInterfaces.description'
                                                            )
                                                        "
                                                    />
                                                </div>

                                                <span class="option-value">
                                                    {{
                                                        config.useInterfaces
                                                            ? t(
                                                                "createProject.fields.useInterfaces.interfaces"
                                                            )
                                                            : t(
                                                                "createProject.fields.useInterfaces.classes"
                                                            )
                                                    }}
                                                </span>
                                            </div>

                                            <n-switch
                                                v-model:value="config.useInterfaces"
                                                :disabled="creating"
                                            />
                                        </div>
                                    </div>
                                </n-form>

                                <!-- 产品信息 -->
                                <n-form
                                    v-else-if="currentStep.key === 'product'"
                                    class="creator-form"
                                    :model="config"
                                    label-placement="top"
                                    :show-require-mark="false"
                                    size="large"
                                >
                                    <div class="form-grid">
                                        <n-form-item
                                            v-for="field in productFields"
                                            :key="field.path"
                                            :class="{ full: field.full }"
                                            :path="field.path"
                                            :validation-status="
                                                visibleFieldError(field.path)
                                                    ? 'error'
                                                    : undefined
                                            "
                                            :feedback="
                                                visibleFieldError(field.path) ||
                                                undefined
                                            "
                                        >
                                            <template #label>
                                                <div class="field-label">
                                                    <span>{{ field.label }}</span>

                                                    <field-help
                                                        :content="field.description"
                                                    />
                                                </div>
                                            </template>

                                            <n-input
                                                v-model:value="config[field.path]"
                                                v-bind="field.props"
                                                clearable
                                                :disabled="creating"
                                                :placeholder="field.placeholder"
                                                @blur="touchField(field.path)"
                                            />
                                        </n-form-item>
                                    </div>
                                </n-form>

                                <!-- 确认 -->
                                <div
                                    v-else
                                    class="summary-list"
                                >
                                    <section
                                        v-for="section in summarySections"
                                        :key="section.key"
                                        class="summary-section"
                                    >
                                        <div class="summary-title">
                                            {{ section.title }}
                                        </div>

                                        <div
                                            v-for="item in section.items"
                                            :key="item.path"
                                            class="summary-row"
                                        >
                                            <span>{{ item.label }}</span>

                                            <n-ellipsis>
                                                {{ formatSummaryValue(item) }}
                                            </n-ellipsis>
                                        </div>
                                    </section>
                                </div>
                            </n-scrollbar>
                        </div>
                    </transition>
                </main>
            </div>

            <footer class="creator-foot">
                <div
                    v-if="progressMessage"
                    class="foot-status"
                >
                    {{ progressMessage }}
                </div>

                <div class="foot-actions">
                    <n-button
                        :disabled="activeStep === 0 || creating"
                        @click="previousStep"
                    >
                        {{ t("createProject.actions.previous") }}
                    </n-button>

                    <n-button
                        v-if="!isLastStep"
                        type="primary"
                        :disabled="creating"
                        @click="nextStep"
                    >
                        {{ t("createProject.actions.next") }}
                    </n-button>

                    <n-button
                        v-else
                        type="primary"
                        :disabled="!allStepsReady"
                        :loading="creating"
                        @click="createProject"
                    >
                        {{ t("createProject.actions.create") }}
                    </n-button>
                </div>
            </footer>
        </section>
    </n-modal>
</template>

<script setup>
import {
    NButton,
    NEllipsis,
    NForm,
    NFormItem,
    NIcon,
    NInput,
    NInputGroup,
    NModal,
    NScrollbar,
    NSelect,
    NStep,
    NSteps,
    NSwitch,
    useMessage
} from "naive-ui"

import {
    computed,
    inject,
    onMounted,
    reactive,
    ref,
    watch
} from "vue"

import router from "../../router/index.js"
import { useI18n } from "../../i18n"
import FieldHelp from "./FieldHelp.vue"

import { AppService } from "../../../bindings/wails3-manager/desktop"
import { Service as CreatorService } from "../../../bindings/wails3-manager/core/creator"
import { Service as ProjectService } from "../../../bindings/wails3-manager/core/project"
import { Service as SettingsService } from "../../../bindings/wails3-manager/core/settings"

const store = inject("store")
const message = useMessage()
const { t } = useI18n()

const activeStep = ref(0)
const creating = ref(false)
const loadingTemplates = ref(false)
const showCloseConfirm = ref(false)
const progressMessage = ref("")

const templateSource = ref("official")
const templateOptions = ref([])
const touchedFields = reactive({})
const config = reactive(defaultConfig())
const reservedProjectNamePattern = /^(con|prn|aux|nul|com[1-9]|lpt[1-9])(?:\.|$)/i

const steps = computed(() => [
    {
        key: "basic",
        title: t("createProject.steps.basic.title"),
        description: t("createProject.steps.basic.description")
    },
    {
        key: "template",
        title: t("createProject.steps.template.title"),
        description: t("createProject.steps.template.description")
    },
    {
        key: "product",
        title: t("createProject.steps.product.title"),
        description: t("createProject.steps.product.description")
    },
    {
        key: "confirm",
        title: t("createProject.steps.confirm.title"),
        description: t("createProject.steps.confirm.description")
    }
])

const basicFields = computed(() => [
    {
        path: "name",
        label: t("createProject.fields.name.label"),
        placeholder: t("createProject.fields.name.placeholder"),
        description: t("createProject.fields.name.description"),
        maxlength: 64
    },
    {
        path: "packageName",
        label: t("createProject.fields.packageName.label"),
        placeholder: t("createProject.fields.packageName.placeholder"),
        description: t("createProject.fields.packageName.description"),
        maxlength: 64
    }
])

const productFields = computed(() => [
    {
        path: "productName",
        label: t("createProject.fields.productName.label"),
        placeholder: t("createProject.fields.productName.placeholder"),
        description: t("createProject.fields.productName.description"),
        props: {
            maxlength: 100
        }
    },
    {
        path: "productVersion",
        label: t("createProject.fields.productVersion.label"),
        placeholder: t("createProject.fields.productVersion.placeholder"),
        description: t("createProject.fields.productVersion.description"),
        props: {
            maxlength: 40
        }
    },
    {
        path: "productCompany",
        label: t("createProject.fields.productCompany.label"),
        placeholder: t("createProject.fields.productCompany.placeholder"),
        description: t("createProject.fields.productCompany.description"),
        props: {
            maxlength: 100
        }
    },
    {
        path: "productIdentifier",
        label: t("createProject.fields.productIdentifier.label"),
        placeholder: t("createProject.fields.productIdentifier.placeholder"),
        description: t("createProject.fields.productIdentifier.description"),
        props: {
            maxlength: 160
        }
    },
    {
        path: "productDescription",
        label: t("createProject.fields.productDescription.label"),
        placeholder: t("createProject.fields.productDescription.placeholder"),
        description: t("createProject.fields.productDescription.description"),
        full: true,
        props: {
            type: "textarea",
            maxlength: 300,
            showCount: true,
            autosize: {
                minRows: 3,
                maxRows: 5
            }
        }
    },
    {
        path: "productCopyright",
        label: t("createProject.fields.productCopyright.label"),
        placeholder: t("createProject.fields.productCopyright.placeholder"),
        description: t("createProject.fields.productCopyright.description"),
        full: true,
        props: {
            maxlength: 200
        }
    },
    {
        path: "productComments",
        label: t("createProject.fields.productComments.label"),
        placeholder: t("createProject.fields.productComments.placeholder"),
        description: t("createProject.fields.productComments.description"),
        full: true,
        props: {
            type: "textarea",
            maxlength: 300,
            showCount: true,
            autosize: {
                minRows: 2,
                maxRows: 4
            }
        }
    }
])

const templateSourceOptions = computed(() => [
    {
        label: t("createProject.fields.template.official"),
        value: "official"
    },
    {
        label: t("createProject.fields.template.custom"),
        value: "custom"
    }
])

const currentStep = computed(() => {
    return steps.value[activeStep.value]
})

const isLastStep = computed(() => {
    return activeStep.value === steps.value.length - 1
})

const allStepsReady = computed(() => {
    return (
        isBasicReady() &&
        isTemplateReady() &&
        isProductReady()
    )
})

const summarySections = computed(() => [
    {
        key: "basic",
        title: t("createProject.summary.basic"),
        items: [
            {
                path: "name",
                label: t("createProject.fields.name.label")
            },
            {
                path: "dir",
                label: t("createProject.fields.dir.label")
            },
            {
                path: "packageName",
                label: t("createProject.fields.packageName.label")
            }
        ]
    },
    {
        key: "template",
        title: t("createProject.summary.template"),
        items: [
            {
                path: "template",
                label: t("createProject.fields.template.label")
            },
            {
                path: "useInterfaces",
                label: t("createProject.fields.useInterfaces.label"),
                format(value) {
                    return value
                        ? t("createProject.fields.useInterfaces.interfaces")
                        : t("createProject.fields.useInterfaces.classes")
                }
            }
        ]
    },
    {
        key: "product",
        title: t("createProject.summary.product"),
        items: [
            {
                path: "productName",
                label: t("createProject.fields.productName.label")
            },
            {
                path: "productVersion",
                label: t("createProject.fields.productVersion.label")
            },
            {
                path: "productCompany",
                label: t("createProject.fields.productCompany.label")
            },
            {
                path: "productIdentifier",
                label: t("createProject.fields.productIdentifier.label")
            },
            {
                path: "productDescription",
                label: t("createProject.fields.productDescription.label")
            },
            {
                path: "productCopyright",
                label: t("createProject.fields.productCopyright.label")
            },
            {
                path: "productComments",
                label: t("createProject.fields.productComments.label")
            }
        ]
    }
])

onMounted(loadTemplates)

watch(
    () => store.panels.creator,
    show => {
        if (show) {
            resetCreator()
            loadTemplates()
        } else {
            showCloseConfirm.value = false
        }
    }
)

function defaultConfig() {
    return {
        name: "my-wails-app",
        dir: "",
        template: "vanilla",
        packageName: "main",

        productName: "My Product",
        productDescription: "My Product Description",
        productVersion: "0.1.0",
        productCompany: "My Company",
        productCopyright: "Copyright (c) 2026",
        productComments: "This is a comment",
        productIdentifier: "com.mycompany.myproduct",

        useInterfaces: true
    }
}

function resetCreator() {
    Object.assign(config, defaultConfig())
    templateSource.value = "official"

    Object.keys(touchedFields).forEach(key => {
        delete touchedFields[key]
    })

    activeStep.value = 0
    creating.value = false
    showCloseConfirm.value = false
    progressMessage.value = ""
}

async function loadTemplates() {
    if (
        loadingTemplates.value ||
        templateOptions.value.length
    ) {
        return
    }

    loadingTemplates.value = true

    try {
        const templates =
            await CreatorService.ListTemplates()

        templateOptions.value = (templates || []).map(item => ({
            label: item.description
                ? `${item.name} · ${item.description}`
                : item.name,
            value: item.name
        }))
    } catch (error) {
        message.error(error?.message || String(error))
    } finally {
        loadingTemplates.value = false
    }
}


function setTemplateSource(value) {
    templateSource.value = value

    if (value === "custom") {
        if (isOfficialTemplate(config.template)) {
            config.template = ""
        }
        return
    }

    if (!isOfficialTemplate(config.template)) {
        config.template = "vanilla"
    }
}

function isOfficialTemplate(value) {
    const template = String(value || "").trim()

    return (
        template === "vanilla" ||
        templateOptions.value.some(item => item.value === template)
    )
}

async function chooseDirectory() {
    if (creating.value) {
        return
    }

    try {
        const dir = await AppService.ChooseFolder()

        if (dir) {
            config.dir = dir
            touchField("dir")
        }
    } catch (error) {
        message.error(error?.message || String(error))
    }
}

function requestClose() {
    if (!creating.value) {
        showCloseConfirm.value = true
    }
}

function closeCreator() {
    if (creating.value) {
        return
    }

    showCloseConfirm.value = false
    store.panels.creator = false
    resetCreator()
}

function previousStep() {
    if (
        activeStep.value > 0 &&
        !creating.value
    ) {
        showCloseConfirm.value = false
        activeStep.value -= 1
    }
}

function nextStep() {
    if (creating.value || isLastStep.value) {
        return
    }

    touchCurrentStep()

    if (!isCurrentStepReady()) {
        return
    }

    showCloseConfirm.value = false
    activeStep.value += 1
}

function touchCurrentStep() {
    if (currentStep.value.key === "basic") {
        touchField("name")
        touchField("dir")
        touchField("packageName")
    }

    if (currentStep.value.key === "template") {
        touchField("template")
    }

    if (currentStep.value.key === "product") {
        touchField("productVersion")
        touchField("productIdentifier")
    }
}

function isCurrentStepReady() {
    if (currentStep.value.key === "basic") {
        return isBasicReady()
    }

    if (currentStep.value.key === "template") {
        return isTemplateReady()
    }

    if (currentStep.value.key === "product") {
        return isProductReady()
    }

    return true
}

async function createProject() {
    if (
        !allStepsReady.value ||
        creating.value
    ) {
        return
    }

    creating.value = true
    showCloseConfirm.value = false
    progressMessage.value = ""
    let pendingMessage = null

    const runStep = async (pendingKey, failedKey, task) => {
        progressMessage.value = t(pendingKey)
        pendingMessage?.destroy()
        pendingMessage = message.loading(progressMessage.value, {
            duration: 0
        })

        try {
            return await task()
        } catch (error) {
            const text = `${t(failedKey)}: ${formatError(error)}`
            progressMessage.value = text
            message.error(text, {
                duration: 10000
            })
            throw error
        }
    }

    try {
        const projectDir =
            await runStep(
                "createProject.messages.creating",
                "createProject.messages.createFailed",
                () => CreatorService.CreateWailsProject({
                    ...config,
                    quiet: true,
                    skipRemoteTemplateWarning: true
                })
            )

        await runStep(
            "createProject.messages.importing",
            "createProject.messages.importFailed",
            () => ProjectService.ImportProject(projectDir)
        )

        await runStep(
            "createProject.messages.opening",
            "createProject.messages.openFailed",
            async () => {
                await SettingsService.OpenProject(projectDir)
                await router.push({
                    name: "project",
                    query: {
                        projectDir: encodeURIComponent(projectDir)
                    }
                })
            }
        )

        pendingMessage?.destroy()

        message.success(
            t("createProject.messages.createSuccess", {
                path: projectDir
            })
        )

        store.panels.creator = false
        resetCreator()
    } catch {
        // Error message is handled in runStep so the failed stage stays visible.
    } finally {
        pendingMessage?.destroy()
        creating.value = false
    }
}

function formatError(error) {
    return error?.message || String(error)
}

function touchField(path) {
    touchedFields[path] = true
}

function visibleFieldError(path) {
    const error = fieldError(path)

    if (!error) {
        return ""
    }

    const value = String(config[path] ?? "").trim()

    if (
        touchedFields[path] ||
        value
    ) {
        return error
    }

    return ""
}

function fieldError(path) {
    const value = String(config[path] ?? "").trim()

    if (path === "name") {
        if (!value) {
            return t(
                "createProject.validation.nameRequired"
            )
        }

        if (!isValidProjectDirectoryName(value)) {
            return t(
                "createProject.validation.nameInvalid"
            )
        }
    }

    if (
        path === "dir" &&
        !value
    ) {
        return t(
            "createProject.validation.dirRequired"
        )
    }

    if (path === "packageName") {
        if (!value) {
            return t(
                "createProject.validation.packageNameRequired"
            )
        }

        if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(value)) {
            return t(
                "createProject.validation.packageNameInvalid"
            )
        }
    }

    if (
        path === "template" &&
        !value
    ) {
        return t(
            "createProject.validation.templateRequired"
        )
    }

    if (
        path === "productVersion" &&
        !value
    ) {
        return t(
            "createProject.validation.versionInvalid"
        )
    }

    if (
        path === "productIdentifier" &&
        value &&
        !/^[A-Za-z][A-Za-z0-9]*(?:\.[A-Za-z0-9][A-Za-z0-9-]*)+$/.test(value)
    ) {
        return t(
            "createProject.validation.identifierInvalid"
        )
    }

    return ""
}

function isValidProjectDirectoryName(value) {
    return (
        /^[A-Za-z0-9_.-]+$/.test(value) &&
        value !== "." &&
        value !== ".." &&
        !value.endsWith(".") &&
        !reservedProjectNamePattern.test(value)
    )
}

function isBasicReady() {
    return (
        !fieldError("name") &&
        !fieldError("dir") &&
        !fieldError("packageName")
    )
}

function isTemplateReady() {
    return !fieldError("template")
}

function isProductReady() {
    return (
        !fieldError("productVersion") &&
        !fieldError("productIdentifier")
    )
}

function formatSummaryValue(item) {
    const value = config[item.path]

    if (item.format) {
        return item.format(value)
    }

    return (
        String(value || "").trim() ||
        t("createProject.common.empty")
    )
}
</script>

<style lang="scss" scoped>
.creator-card {
    position: relative;

    width: min(1020px, calc(100vw - 32px));
    height: min(700px, calc(100vh - 40px));

    box-sizing: border-box;

    display: flex;
    flex-direction: column;
    overflow: hidden;

    border: 1px solid var(--wm-border-soft);
    border-radius: 14px;

    background:
        linear-gradient(
            180deg,
            var(--wm-surface-1),
            transparent 120px
        ),
        var(--wm-bg-page);

    box-shadow: var(--wm-shadow-page);

    pointer-events: auto;
}


.creator-head {
    flex: 0 0 auto;

    min-height: 72px;
    padding: 16px 18px 16px 20px;

    box-sizing: border-box;

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 18px;

    border-bottom: 1px solid var(--wm-border-subtle);
}

.creator-heading {
    min-width: 0;
}

.creator-title {
    color: var(--wm-text-primary);
    font-size: 16px;
    font-weight: 600;
    line-height: 22px;
}

.creator-subtitle {
    margin-top: 4px;

    overflow: hidden;

    color: var(--wm-text-muted);
    font-size: 12px;
    line-height: 17px;

    text-overflow: ellipsis;
    white-space: nowrap;
}

.creator-head-actions {
    display: flex;
    align-items: center;
    gap: 8px;
}

:deep(.n-button.close-button) {
    color: var(--wm-text-muted) !important;
}

:deep(.n-button.close-button:hover) {
    color: var(--wm-text-primary) !important;
}

.close-confirm {
    position: absolute;
    top: 72px;
    left: 0;
    right: 0;
    z-index: 3;
    backdrop-filter: blur(12px);
    min-height: 52px;
    padding: 9px 14px 9px 18px;

    box-sizing: border-box;

    display: grid;
    grid-template-columns: 28px minmax(0, 1fr) auto;
    align-items: center;
    gap: 10px;

    border-bottom: 1px solid var(--wm-color-primary-border);

    color: var(--wm-text-primary);

    background:
        linear-gradient(
            90deg,
            rgba(var(--wm-color-primary-rgb), 0.08),
            rgba(var(--wm-color-primary-rgb), 0.025)
        );
}

.close-confirm-icon {
    width: 28px;
    height: 28px;

    display: flex;
    align-items: center;
    justify-content: center;

    border: 1px solid var(--wm-color-primary-border);
    border-radius: 8px;

    color: var(--wm-color-primary-hover);
    background: rgba(var(--wm-color-primary-rgb), 0.075);
}

.close-confirm-text {
    min-width: 0;

    font-size: 12px;
    line-height: 18px;
}

.close-confirm-actions {
    display: flex;
    align-items: center;
    gap: 8px;
}

.creator-body {
    flex: 1;
    min-height: 0;

    display: grid;
    grid-template-columns: 220px minmax(0, 1fr);
}

.creator-steps {
    min-height: 0;
    padding: 22px;

    box-sizing: border-box;

    border-right: 1px solid var(--wm-border-subtle);
    background: var(--wm-surface-1);
}

.steps-caption {
    margin-bottom: 16px;

    color: var(--wm-text-muted);

    font-size: 12px;
    letter-spacing: 0.08em;

    text-transform: uppercase;
}

.creator-panel {
    min-width: 0;
    min-height: 0;

    display: flex;
    flex-direction: column;

    overflow: hidden;
}

.step-panel {
    height: 100%;
    min-height: 0;
    padding: 24px 28px;

    box-sizing: border-box;

    display: flex;
    flex-direction: column;
}

.panel-head {
    flex: 0 0 auto;
    margin-bottom: 14px;
}

.panel-kicker {
    margin-bottom: 5px;

    color: var(--wm-color-primary-hover);

    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.06em;
}

.panel-title {
    color: var(--wm-text-primary);

    font-size: 18px;
    font-weight: 600;
    line-height: 24px;
}

.panel-desc {
    max-width: 640px;
    margin-top: 5px;

    color: var(--wm-text-muted);

    font-size: 11px;
    line-height: 17px;
}

.panel-scroll {
    flex: 1;
    min-height: 0;

    margin-top: 18px;
}

.creator-form {
    padding-right: 5px;
}

.form-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    align-content: start;
    padding-right: 16px;
    gap: 16px;
}

.form-grid > .full {
    grid-column: 1 / -1;
}

:deep(.n-form-item) {
    min-width: 0;
}



:deep(.n-form-item-feedback-wrapper) {
    min-height: 18px;
}

:deep(.n-form-item-feedback) {
    font-size: 10px;
    line-height: 16px;
}

.field-label {
    min-width: 0;

    display: inline-flex;
    align-items: center;
    gap: 6px;

    color: var(--wm-text-secondary);

    font-size: 12px;
    font-weight: 500;
}

.template-form {
    padding-right: 16px;
}

.template-layout {
    display: flex;
    flex-direction: column;
    gap: 14px;
}

.template-select-item {
    margin-bottom: 0;
}

:deep(.template-select-item > .n-form-item-blank) {
    display: block;
    width: 100%;
}

.template-select-card {
    width: 100%;
    padding: 16px;

    box-sizing: border-box;

    border: 1px solid var(--wm-border-subtle);
    border-radius: 12px;

    background:
        linear-gradient(
            135deg,
            rgba(var(--wm-color-primary-rgb), 0.045),
            transparent 58%
        ),
        var(--wm-surface-1);
}

.template-source-row {
    margin-bottom: 12px;

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.template-source-caption {
    color: var(--wm-text-muted);
    font-size: 11px;
    line-height: 16px;
}

.template-source-toggle {
    width: min(280px, 100%);

    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
}

.template-source-toggle :deep(.n-button) {
    min-width: 0;
}

.template-control {
    width: 100%;
}



.template-mode-card {
    min-height: 72px;
    padding: 12px;
}

.option-card {
    min-height: 66px;
    padding: 12px ;

    box-sizing: border-box;

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 18px;

    border: 1px solid var(--wm-border-subtle);
    border-radius: 10px;

    background: var(--wm-surface-1);
}

.option-copy {
    min-width: 0;
}

.option-value {
    display: block;
    margin-top: 5px;

    color: var(--wm-text-secondary);

    font-size: 14px;
    line-height: 16px;
}

.summary-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding-right: 16px;
}

.summary-section {
    min-width: 0;
    padding: 13px 14px;

    box-sizing: border-box;

    border: 1px solid var(--wm-border-subtle);
    border-radius: 10px;

    background: var(--wm-surface-1);
}

.summary-title {
    margin-bottom: 8px;

    color: var(--wm-text-primary);

    font-size: 12px;
    font-weight: 600;
}

.summary-row {
    min-height: 31px;

    display: grid;
    grid-template-columns: 145px minmax(0, 1fr);
    align-items: center;
    gap: 12px;

    color: var(--wm-text-secondary);

    font-size: 11px;
}

.summary-row + .summary-row {
    border-top: 1px solid var(--wm-border-subtle);
}

.summary-row > span:first-child {
    color: var(--wm-text-muted);
}

.creator-foot {
    flex: 0 0 auto;

    min-height: 64px;
    padding: 12px 18px;

    box-sizing: border-box;

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;

    border-top: 1px solid var(--wm-border-subtle);
    background: var(--wm-surface-1);
}

.foot-status {
    overflow: hidden;

    color: var(--wm-text-muted);

    font-size: 10px;

    text-overflow: ellipsis;
    white-space: nowrap;
}

.foot-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-left: auto;
}

.creator-step-enter-active,
.creator-step-leave-active {
    transition:
        opacity 150ms ease,
        transform 150ms ease;
}

.creator-step-enter-from {
    opacity: 0;
    transform: translateX(8px);
}

.creator-step-leave-to {
    opacity: 0;
    transform: translateX(-8px);
}

.close-confirm-enter-active,
.close-confirm-leave-active {
    transition:
        opacity 150ms ease,
        transform 150ms ease;
}

.close-confirm-enter-from,
.close-confirm-leave-to {
    opacity: 0;
    transform: translateY(-6px);
}

:deep(.n-steps .n-step-content-header) {
    color: var(--wm-text-secondary) !important;
    font-size: 14px !important;
}

:deep(.n-steps .n-step--process-status .n-step-content-header) {
    color: var(--wm-text-primary) !important;
}

:deep(.n-steps .n-step-indicator) {
    --n-indicator-border-color-process:
        var(--wm-color-primary-border) !important;

    --n-indicator-color-process:
        var(--wm-color-primary) !important;

    --n-indicator-color-finish:
        var(--wm-color-primary) !important;

    --n-indicator-border-color-finish:
        var(--wm-color-primary) !important;
}

:deep(.n-input),
:deep(.n-base-selection) {
    --n-color:
        var(--wm-control-bg) !important;

    --n-color-focus:
        var(--wm-control-bg-hover) !important;

    --n-color-active:
        var(--wm-control-bg-hover) !important;

    --n-border:
        1px solid var(--wm-border-soft) !important;

    --n-border-hover:
        1px solid var(--wm-color-primary-border) !important;

    --n-border-focus:
        1px solid var(--wm-border-strong) !important;

    --n-border-active:
        1px solid var(--wm-border-strong) !important;

    --n-box-shadow-focus:
        0 0 0 2px var(--wm-color-primary-shadow) !important;

    --n-box-shadow-active:
        0 0 0 2px var(--wm-color-primary-shadow) !important;

    --n-text-color:
        var(--wm-text-secondary) !important;

    --n-placeholder-color:
        var(--wm-placeholder) !important;

    --n-border-radius:
        8px !important;
}

:deep(.n-input input),
:deep(.n-input textarea) {
    user-select: text !important;
    -webkit-user-select: text !important;

    cursor: text;
}

:deep(.n-input-group > .n-button) {
    min-width: 94px;
}

:deep(.n-button) {
    --n-border-radius:
        7px !important;

    --n-text-color:
        var(--wm-text-secondary) !important;

    --n-text-color-hover:
        var(--wm-color-primary-hover) !important;

    --n-text-color-pressed:
        var(--wm-color-primary-pressed) !important;

    --n-color:
        var(--wm-control-bg) !important;

    --n-color-hover:
        var(--wm-control-bg-hover) !important;

    --n-color-pressed:
        var(--wm-control-bg-active) !important;

    --n-border:
        1px solid var(--wm-border-soft) !important;

    --n-border-hover:
        1px solid var(--wm-color-primary-border) !important;

    --n-border-pressed:
        1px solid var(--wm-border-strong) !important;
}

:deep(.n-button--primary-type) {
    --n-color:
        var(--wm-color-primary) !important;

    --n-color-hover:
        var(--wm-color-primary-hover) !important;

    --n-color-pressed:
        var(--wm-color-primary-pressed) !important;

    --n-border:
        1px solid var(--wm-color-primary-border) !important;

    --n-border-hover:
        1px solid var(--wm-border-strong) !important;

    --n-border-pressed:
        1px solid var(--wm-color-primary-border) !important;

    --n-text-color:
        var(--wm-text-inverse) !important;

    --n-text-color-hover:
        var(--wm-text-inverse) !important;

    --n-text-color-pressed:
        var(--wm-text-inverse) !important;

    box-shadow: var(--wm-shadow-primary);
}

:deep(.n-switch) {
    --n-rail-color-active:
        var(--wm-color-primary) !important;

    --n-loading-color:
        var(--wm-color-primary) !important;

    --n-box-shadow-focus:
        0 0 0 2px var(--wm-color-primary-shadow) !important;
}

:deep(.n-button--disabled) {
    opacity: 0.48;
    box-shadow: none;
}

:deep(.n-button--primary-type.n-button--disabled) {
    --n-color-disabled:
        var(--wm-control-bg) !important;

    --n-border-disabled:
        1px solid var(--wm-border-soft) !important;

    --n-text-color-disabled:
        var(--wm-text-muted) !important;
}

@media (max-width: 760px) {
    .creator-card {
        width: calc(100vw - 20px);
        height: calc(100vh - 20px);

        border-radius: 12px;
    }

    .creator-body {
        grid-template-columns: 1fr;
    }

    .creator-steps {
        padding: 12px 16px;

        border-right: 0;
        border-bottom: 1px solid var(--wm-border-subtle);
    }

    .steps-caption {
        display: none;
    }

    :deep(.n-steps) {
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
    }

    :deep(.n-steps .n-step-splitor) {
        display: none;
    }

    :deep(.n-steps .n-step-content-header) {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .step-panel {
        padding: 18px;
    }

    .form-grid {
        grid-template-columns: 1fr;
    }

    .template-form {
        padding-right: 0;
    }

    .form-grid > .full {
        grid-column: auto;
    }

    .close-confirm {
        grid-template-columns: 28px minmax(0, 1fr);
    }

    .close-confirm-actions {
        grid-column: 1 / -1;
        justify-content: flex-end;
    }
}

@media (max-width: 500px) {
    .creator-subtitle,
    .foot-status {
        display: none;
    }

    .creator-head {
        min-height: 62px;
        padding: 13px 14px;
    }

    .close-confirm {
        top: 62px;
    }

    .creator-steps {
        display: none;
    }

    .template-select-card {
        padding: 13px;
    }

    .template-source-row {
        align-items: stretch;
        flex-direction: column;
        gap: 8px;
    }

    .template-source-toggle {
        width: 100%;
    }

    .creator-foot {
        justify-content: flex-end;
    }

    .summary-row {
        grid-template-columns: 110px minmax(0, 1fr);
    }
}

@media (prefers-reduced-motion: reduce) {
    .creator-step-enter-active,
    .creator-step-leave-active,
    .close-confirm-enter-active,
    .close-confirm-leave-active {
        transition: none;
    }
}
</style>
