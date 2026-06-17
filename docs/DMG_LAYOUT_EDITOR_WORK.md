# DMG 编辑器与打包配置工作文档

日期：2026-06-17

## 目标

优化 macOS DMG 安装页编辑体验，收敛吸附规则，修复缩放抖动，并调整打包配置页的平台切换和安装包开关位置。

## 本次范围

- `frontend/src/components/Editor/DmgLayoutEditor.vue`
- `frontend/src/components/Editor/PackageCfgEditor.vue`
- `frontend/src/i18n/messages/*`

## 主要改动

### 1. 吸附规则收敛

- 只保留画布垂直居中线、水平居中线。
- 只保留两个图标之间基于画布中心的镜像对称线。
- 关闭 Moveable 的 gap、元素边缘和网格类吸附。

### 2. 缩放交互改为 transform 驱动

- 将 Moveable 的 `resizable` 改为 `scalable`。
- 拖动和缩放过程中只写元素 `transform`，右侧参数实时写当前坐标和大小。
- 交互结束时一次性把最终 `iconSize`、`appX/appY`、`applicationsX/applicationsY` 写回配置结构。

这样处理的原因：

- DMG 图标坐标是内部可信结构，编辑器不做额外兜底或字段补全。
- 抖动来自交互中同时改元素尺寸、位置和 Moveable transform。
- 使用交互基准值锁住元素起始样式，可以让参数面板实时变化，同时避免 Vue 样式重绘和 Moveable 控制框互相影响。

### 3. 参数面板分组

- 背景图、窗口尺寸、图标布局分成独立模块。
- 参数输入仍直接绑定 `local.macos` 对应字段。
- 保存按钮只回写外层打包配置，是否持久化仍由打包配置页保存按钮决定。

### 4. 平台切换与安装包开关换位

- 安装包生成开关放到平台配置标题右侧。
- Windows/macOS 切换独立放到标题下方，避免和开关高度不一致。

## 验证

已执行：

```powershell
npm run i18n:check
npm run build:dev
go test ./core/... ./desktop/...
wails3 task build
```

结果：全部通过。

## 2026-06-17 追加修复

- 修复 DMG 元素松开鼠标后向右下偏移 1px：坐标读取改为基于舞台内容区，扣除舞台 1px border。
- 简化 DMG 参数面板层级：去掉内层卡片式包裹，按背景、窗口、图标布局直接分组。
- 图标布局参数按数据类型排列：尺寸单独一行，App 坐标一行，Applications 坐标一行。
- 项目页主体布局改为明确的两列 grid，环境面板固定为 300px，避免环境数据加载前后宽度跳变。
