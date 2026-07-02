<template>
    <n-modal
        :show="show"
        class="dmg-layout-modal"
        :mask-closable="false"
        :close-on-esc="false"
    >
        <div class="dmg-layout-card">
            <div class="dmg-layout-head">
                <div>
                    <div class="dmg-layout-title">{{ t("packageEditor.dmgLayoutTitle") }}</div>
                    <div class="dmg-layout-subtitle">
                        {{ local.windowWidth }} x {{ local.windowHeight }}
                    </div>
                </div>

                <div class="head-actions">
                    <n-button quaternary @click="closeEditor">
                        {{ t("editor.cancel") }}
                    </n-button>

                    <n-button type="primary" @click="saveEditor">
                        {{ t("editor.save") }}
                    </n-button>
                </div>
            </div>

            <div class="dmg-layout-body">
                <div class="dmg-stage-pane">
                    <div class="pane-title">{{ t("packageEditor.dmgCanvas") }}</div>

                    <n-scrollbar class="stage-scroll">
                        <div class="stage-wrap" @pointerdown="onBlankPointerDown">
                            <div
                                ref="stageRef"
                                class="dmg-stage"
                                :style="stageStyle"
                            >
                                <div
                                    ref="appRef"
                                    class="dmg-item app-item"
                                    :class="{ active: selectedElement === 'app' }"
                                    :style="appStyle"
                                    @pointerdown.stop="selectElement('app')"
                                >
                                    <img
                                        class="app-preview"
                                        :src="appIconURL"
                                        alt=""
                                    >
                                    <div class="item-label">{{ t("packageEditor.appIcon") }}</div>
                                </div>

                                <div
                                    ref="applicationsRef"
                                    class="dmg-item folder-item"
                                    :class="{ active: selectedElement === 'applications' }"
                                    :style="applicationsStyle"
                                    @pointerdown.stop="selectElement('applications')"
                                >
                                    <img
                                        class="applications-preview"
                                        :src="applicationsIcon"
                                        alt=""
                                    >
                                    <div class="item-label">{{ t("packageEditor.applicationsFolder") }}</div>
                                </div>

                                <Moveable
                                    ref="moveableRef"
                                    v-if="moveableTarget"
                                    :key="selectedElement"
                                    :target="moveableTarget"
                                    :container="stageRef"
                                    :draggable="true"
                                    :resizable="true"
                                    :snappable="true"
                                    :keep-ratio="true"
                                    :origin="false"
                                    :use-accurate-position="true"
                                    :vertical-guidelines="verticalGuidelines"
                                    :horizontal-guidelines="horizontalGuidelines"
                                    :snap-directions="snapDirections"
                                    :snap-gap="false"
                                    :element-guidelines="[]"
                                    :element-snap-directions="false"
                                    :throttle-drag="0"
                                    :throttle-resize="1"
                                    :snap-threshold="8"
                                    @dragStart="onDragStart"
                                    @drag="onDrag"
                                    @dragEnd="onDragEnd"
                                    @resizeStart="onResizeStart"
                                    @resize="onResize"
                                    @resizeEnd="onResizeEnd"
                                />
                            </div>
                        </div>
                    </n-scrollbar>
                </div>

                <div class="dmg-params-pane">
                    <div class="pane-title">{{ t("packageEditor.dmgParameters") }}</div>

                    <n-scrollbar class="params-scroll">
                        <div class="param-list">
                            <div class="param-group">
                                <div class="param-title">{{ t("packageEditor.dmgBackgroundSection") }}</div>

                                <div class="path-row">
                                    <label class="param-field">
                                        <span class="label">{{ t("packageEditor.background") }}</span>
                                        <n-input v-model:value="local.background" />
                                    </label>

                                    <n-button @click="chooseBackground">
                                        {{ t("packageEditor.selectBackground") }}
                                    </n-button>
                                </div>
                            </div>

                            <div class="param-group">
                                <div class="param-title">{{ t("packageEditor.dmgWindowSection") }}</div>

                                <div class="param-row">
                                    <label class="param-field">
                                        <span class="label">{{ t("packageEditor.windowWidth") }}</span>
                                        <n-input-number v-model:value="local.windowWidth" :min="1" />
                                    </label>

                                    <label class="param-field">
                                        <span class="label">{{ t("packageEditor.windowHeight") }}</span>
                                        <n-input-number v-model:value="local.windowHeight" :min="1" />
                                    </label>
                                </div>
                            </div>

                            <div class="param-group">
                                <div class="param-title">{{ t("packageEditor.dmgIconSection") }}</div>

                                <div class="param-row">
                                    <label class="param-field">
                                        <span class="label">{{ t("packageEditor.iconSize") }}</span>
                                        <n-input-number v-model:value="local.iconSize" :min="1" />
                                    </label>
                                </div>

                                <div class="param-row">
                                    <label class="param-field">
                                        <span class="label">{{ t("packageEditor.appX") }}</span>
                                        <n-input-number v-model:value="local.appX" />
                                    </label>

                                    <label class="param-field">
                                        <span class="label">{{ t("packageEditor.appY") }}</span>
                                        <n-input-number v-model:value="local.appY" />
                                    </label>
                                </div>

                                <div class="param-row">
                                    <label class="param-field">
                                        <span class="label">{{ t("packageEditor.applicationsX") }}</span>
                                        <n-input-number v-model:value="local.applicationsX" />
                                    </label>

                                    <label class="param-field">
                                        <span class="label">{{ t("packageEditor.applicationsY") }}</span>
                                        <n-input-number v-model:value="local.applicationsY" />
                                    </label>
                                </div>
                            </div>
                        </div>
                    </n-scrollbar>
                </div>
            </div>
        </div>
    </n-modal>
</template>

<script setup>
import { computed, nextTick, ref, watch } from "vue"
import Moveable from "vue3-moveable"
import { NButton, NInput, NInputNumber, NModal, NScrollbar, useMessage } from "naive-ui"
import { AppService } from "../../../bindings/wails3-manager/desktop"
import { useI18n } from "../../i18n"
import applicationsIcon from "../../../../assets/applications.png"
import { localFileUrl } from "../../utils/files"

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    macos: {
        type: Object,
        required: true
    },
    projectDir: {
        type: String,
        required: true
    }
})

const emit = defineEmits(["update:show", "save"])

const { t } = useI18n()
const message = useMessage()

const local = ref(cloneData(props.macos))
const selectedElement = ref(null)
const stageRef = ref(null)
const appRef = ref(null)
const applicationsRef = ref(null)
const moveableRef = ref(null)
const backgroundVersion = ref(0)
const interactionBase = ref(null)
const snapDirections = {
    center: true,
    middle: true,
}

watch(
    () => props.macos,
    (next) => {
        local.value = cloneData(next)
        backgroundVersion.value += 1
    },
    { deep: true }
)

watch(
    () => props.show,
    (visible) => {
        if (!visible) return

        local.value = cloneData(props.macos)
        selectedElement.value = null
        interactionBase.value = null
        backgroundVersion.value += 1
    }
)

watch(
    () => [
        local.value.windowWidth,
        local.value.windowHeight,
        local.value.iconSize,
        local.value.appX,
        local.value.appY,
        local.value.applicationsX,
        local.value.applicationsY,
    ],
    () => {
        if (interactionBase.value) return

        nextTick(() => moveableRef.value?.updateRect())
    }
)

const moveableTarget = computed(() => {
    if (!selectedElement.value) {
        return null
    }
    return selectedElement.value === "app" ? appRef.value : applicationsRef.value
})

const stageStyle = computed(() => ({
    width: `${local.value.windowWidth}px`,
    height: `${local.value.windowHeight}px`,
    backgroundImage: backgroundURL.value ? `url("${backgroundURL.value}")` : "none",
}))

const verticalGuidelines = computed(() => {
    const peerX = selectedElement.value === "app" ? local.value.applicationsX : local.value.appX
    return snapLines(
        local.value.windowWidth,
        local.value.windowWidth - peerX,
    )
})

const horizontalGuidelines = computed(() => {
    const peerY = selectedElement.value === "app" ? local.value.applicationsY : local.value.appY
    return snapLines(
        local.value.windowHeight,
        local.value.windowHeight - peerY,
    )
})

const backgroundURL = computed(() => {
    if (!local.value.background) {
        return ""
    }
    return localFileURL(local.value.background, backgroundVersion.value)
})

const appIconURL = computed(() => {
    return localFileURL("build/appicon.png", 0)
})

const appStyle = computed(() => itemStyle("app", local.value.appX, local.value.appY))
const applicationsStyle = computed(() => itemStyle("applications", local.value.applicationsX, local.value.applicationsY))

function itemStyle(element, x, y) {
    const base = interactionBase.value?.element === element ? interactionBase.value : null
    const size = base ? base.size : local.value.iconSize
    const centerX = base ? base.x : x
    const centerY = base ? base.y : y
    return {
        width: `${size}px`,
        height: `${size}px`,
        left: `${Math.round(centerX - size / 2)}px`,
        top: `${Math.round(centerY - size / 2)}px`,
    }
}

function snapLines(length, symmetryLine) {
    const values = [length / 2, symmetryLine]
    return [...new Set(values.map((value) => Math.round(value)).filter((value) => value >= 0 && value <= length))]
        .sort((a, b) => a - b)
}

function projectFilePath(path) {
    if (/^[A-Za-z]:[\\/]/.test(path) || path.startsWith("/") || path.startsWith("\\\\")) {
        return path
    }
    return props.projectDir + "/" + path
}

function localFileURL(path, version) {
    return localFileUrl(projectFilePath(path), version)
}

async function chooseBackground() {
    try {
        const path = await AppService.ChooseFile()
        if (!path) return

        local.value.background = path
        backgroundVersion.value += 1
    } catch (error) {
        message.error(error?.message || String(error))
    }
}

function selectElement(element) {
    selectedElement.value = element
    nextTick(() => moveableRef.value?.updateRect())
}

function onBlankPointerDown(event) {
    if (event.button !== 0) return
    if (!(event.target instanceof Element)) return
    if (event.target.closest(".dmg-item, .moveable-control-box, .moveable-control, .moveable-line, .moveable-area")) return

    selectedElement.value = null
}

function onDragStart(event) {
    setInteractionBase()
    event.target.style.transform = "translate(0px, 0px)"
}

function onDrag(event) {
    const next = getDragState(event)

    event.target.style.transform = event.transform
    if (!next) return

    setElementCenter(next.element, next.centerX, next.centerY)
    interactionBase.value.last = next
}

function onDragEnd(event) {
    commitInteraction(event.target, Boolean(event.isDrag), false)
}

function onResizeStart(event) {
    setInteractionBase()
    event.target.style.transform = ""
}

function onResize(event) {
    const next = getResizeState(event)

    if (!next) return

    applyTargetBox(event.target, next)
    interactionBase.value.last = next
}

function onResizeEnd(event) {
    commitInteraction(event.target, Boolean(event.isDrag), true)
}

function setInteractionBase() {
    if (selectedElement.value === "app") {
        interactionBase.value = {
            element: "app",
            x: local.value.appX,
            y: local.value.appY,
            size: local.value.iconSize,
            last: null,
        }
        return
    }

    if (selectedElement.value === "applications") {
        interactionBase.value = {
            element: "applications",
            x: local.value.applicationsX,
            y: local.value.applicationsY,
            size: local.value.iconSize,
            last: null,
        }
        return
    }

    interactionBase.value = null
}

function getDragState(event) {
    const base = interactionBase.value
    if (!base) return null

    const [dx, dy] = getEventPair(event.beforeTranslate || event.beforeDist || event.dist)
    return {
        element: base.element,
        centerX: Math.round(base.x + dx),
        centerY: Math.round(base.y + dy),
        size: base.size,
    }
}

function getResizeState(event) {
    const base = interactionBase.value
    if (!base) return null

    const dragDist = getEventPair(event.drag?.beforeDist || event.drag?.dist)
    const eventSize = Math.max(
        Number.isFinite(event.width) ? event.width : base.size,
        Number.isFinite(event.height) ? event.height : base.size,
    )
    const size = Math.max(1, Math.round(eventSize))
    const left = base.x - base.size / 2 + dragDist[0]
    const top = base.y - base.size / 2 + dragDist[1]

    return {
        element: base.element,
        centerX: Math.round(left + size / 2),
        centerY: Math.round(top + size / 2),
        size,
    }
}

function getEventPair(value, fallback = 0) {
    if (!Array.isArray(value)) {
        return [fallback, fallback]
    }

    return [
        Number.isFinite(value[0]) ? value[0] : fallback,
        Number.isFinite(value[1]) ? value[1] : fallback,
    ]
}

function commitInteraction(target, changed, commitSize) {
    const base = interactionBase.value
    const next = changed ? base?.last : null

    target.style.transform = ""

    if (next) {
        applyTargetBox(target, next)
    } else if (base) {
        applyTargetBox(target, {
            size: base.size,
            centerX: base.x,
            centerY: base.y,
        })
        local.value.iconSize = base.size
        setElementCenter(base.element, base.x, base.y)
    }

    if (commitSize && next) {
        local.value.iconSize = next.size
    }
    if (next) {
        setElementCenter(next.element, next.centerX, next.centerY)
    }
    interactionBase.value = null
    nextTick(() => moveableRef.value?.updateRect())
}

function applyTargetBox(target, box) {
    target.style.width = `${box.size}px`
    target.style.height = `${box.size}px`
    target.style.left = `${Math.round(box.centerX - box.size / 2)}px`
    target.style.top = `${Math.round(box.centerY - box.size / 2)}px`
}

function setElementCenter(element, x, y) {
    if (element === "app") {
        local.value.appX = x
        local.value.appY = y
        return
    }
    if (element === "applications") {
        local.value.applicationsX = x
        local.value.applicationsY = y
    }
}

function closeEditor() {
    emit("update:show", false)
}

function saveEditor() {
    emit("save", cloneData(local.value))
    emit("update:show", false)
}

function cloneData(data) {
    return JSON.parse(JSON.stringify(data))
}
</script>

<style lang="scss" scoped>
.dmg-layout-card {
    width: min(1060px, calc(100vw - 36px));
    height: min(720px, calc(100vh - 48px));
    border-radius: 8px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-bg-page);
    overflow: hidden;

    display: flex;
    flex-direction: column;
}

.dmg-layout-head {
    height: 64px;
    min-height: 64px;
    padding: 0 18px;
    box-sizing: border-box;
    border-bottom: 1px solid var(--wm-border-subtle);

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.dmg-layout-title {
    font-size: 16px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.dmg-layout-subtitle {
    margin-top: 4px;
    max-width: 420px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 12px;
    color: var(--wm-text-muted);
}

.head-actions {
    display: flex;
    align-items: center;
    gap: 10px;
}

.dmg-layout-body {
    flex: 1;
    min-height: 0;
    padding: 16px;
    box-sizing: border-box;

    display: grid;
    grid-template-columns: minmax(0, 1fr) 320px;
    gap: 16px;
}

.dmg-stage-pane,
.dmg-params-pane {
    min-width: 0;
    min-height: 0;
    border-radius: 8px;
    border: 1px solid var(--wm-border-subtle);
    background: var(--wm-surface-1);
    overflow: hidden;

    display: flex;
    flex-direction: column;
}

.pane-title {
    height: 40px;
    padding: 0 14px;
    box-sizing: border-box;
    border-bottom: 1px solid var(--wm-border-subtle);

    display: flex;
    align-items: center;

    font-size: 13px;
    color: var(--wm-text-primary);
}

.stage-scroll,
.params-scroll {
    flex: 1;
    min-height: 0;
}

.stage-wrap {
    min-width: 100%;
    min-height: 100%;
    padding: 24px;
    box-sizing: border-box;

    display: flex;
    align-items: center;
    justify-content: center;
}

.dmg-stage {
    position: relative;
    flex: 0 0 auto;
    overflow: hidden;
    border-radius: 8px;
    border: 1px solid var(--wm-border-soft);
    background-color: var(--wm-control-bg);
    background-position: center;
    background-size: cover;
    background-repeat: no-repeat;
    box-shadow: 0 18px 48px rgba(0, 0, 0, 0.22);
}

.dmg-item {
    position: absolute;
    border-radius: 8px;
    cursor: move;
    user-select: none;
    overflow: visible;
}

.dmg-item.active {
    background: color-mix(in srgb, var(--wm-color-primary) 9%, transparent);
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--wm-color-primary) 24%, transparent);
}

.app-preview,
.applications-preview {
    position: absolute;
    left: 50%;
    top: 50%;
    width: 86%;
    height: 86%;
    object-fit: contain;
    display: block;
    pointer-events: none;
    transform: translate(-50%, -50%);
}

.item-label {
    position: absolute;
    left: 50%;
    top: calc(100% + 6px);
    width: max(112px, 120%);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: center;
    font-size: 11px;
    line-height: 1.25;
    color: rgba(0, 0, 0, 0.88);
    text-shadow: 0 1px 2px rgba(255, 255, 255, 0.72);
    pointer-events: none;
    transform: translateX(-50%);
}

.param-list {
    padding: 14px;
    display: flex;
    flex-direction: column;
    gap: 14px;
}

.param-group {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.param-title {
    font-size: 12px;
    font-weight: 500;
    color: var(--wm-text-primary);
}

.param-row {
    min-width: 0;
    display: flex;
    gap: 10px;
}

.param-field {
    min-width: 0;
    flex: 1;
}

.label {
    display: block;
    margin-bottom: 7px;
    font-size: 12px;
    color: var(--wm-text-muted);
}

.path-row {
    display: flex;
    align-items: flex-end;
    gap: 8px;
}

.path-row .n-button {
    flex: 0 0 auto;
}

:deep(.n-input),
:deep(.n-input-number) {
    --n-color: var(--wm-control-bg) !important;
    --n-color-focus: var(--wm-control-bg-hover) !important;
    --n-border: 1px solid var(--wm-border-soft) !important;
    --n-border-hover: 1px solid var(--wm-color-primary-border) !important;
    --n-border-focus: 1px solid var(--wm-border-strong) !important;
    --n-box-shadow-focus: 0 0 0 2px var(--wm-color-primary-shadow) !important;
    --n-text-color: var(--wm-text-secondary) !important;
}

:deep(.n-button) {
    --n-border-radius: 4px !important;
}
</style>
