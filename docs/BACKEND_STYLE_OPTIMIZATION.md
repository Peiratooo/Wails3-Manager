# 后端代码风格优化工作文档

日期：2026-06-16

## 目标

按“输入校验 -> 明确赋值 -> 明确执行 -> 明确返回”的风格优化后端，减少内部可信数据流里的重复 normalize、默认值补全、猜测式路径和字段推断。

## 本次优化范围

- 项目记录链路：`core/project`
- 打包配置与打包执行链路：`core/packaging`
- 创建项目入口：`core/creator`
- 文件系统通用工具：`core/fsx`

## 主要改动

### 1. 项目记录不再从嵌套字段推断 projectDir

涉及文件：

- `core/project/project.go`
- `core/project/record.go`
- `core/project/state.go`

改动：

- 删除 `NormalizeProjectRecord`、`NormalizeProjectDir`、`applyRecordDefaults`。
- `SaveProject` 只接受顶层 `record.ProjectDir` 作为项目路径来源，不再从 `record.Project.ProjectDir` 兜底推断。
- `LoadProjectRecord` 读取 `builder/project.json` 后，以调用方传入的项目目录作为权威路径，直接赋值到 `record.ProjectDir` 和 `record.Project.ProjectDir`。
- `UpsertProjectRecord` 要求传入记录已经有明确 `ProjectDir`，缺失时直接报错。

原因：

- `ProjectRecord.ProjectDir` 是服务层已经确认的内部结构字段。
- 嵌套路径兜底会掩盖调用方没有正确构造记录的问题。
- 删除重复 normalize/default 后，导入、保存、打开项目的执行链路更直接。

### 2. packaging.json 从“补默认值”改为“显式校验”

涉及文件：

- `core/packaging/config/config.go`
- `core/packaging/service.go`
- `core/packaging/config/output.go`
- `core/packaging/config/render.go`

改动：

- 删除保存/读取时对 `packaging.json` 的默认值补全逻辑。
- 新增配置校验：缺少 `schemaVersion`、`build.taskfile`、`build.appName`、`artifacts.outputRoot`、资源 `src/type` 等必要字段时直接返回错误。
- `InitPackaging` 只有在 `packaging.json` 不存在时才创建默认配置；文件存在但格式错误或缺必要字段时不再静默重建。
- `build.command` 为空时必须明确提供 `build.task`。
- `build.appName` 渲染占位符时直接使用配置值，不再兜成 `app`。

原因：

- `packaging.json` 是外部持久化配置，读取/保存入口需要校验。
- 配置存在但不完整时继续补默认值，会掩盖配置错误并导致后续打包结果不可预期。

### 3. 打包执行不再接受未知平台

涉及文件：

- `core/packaging/service.go`

改动：

- 删除平台解析 helper。
- `Package` 中直接处理 `auto/空值`，并只接受 `windows`、`darwin`、`all`。
- 其他平台值直接报错，不再跳过所有平台后返回“Packaging completed.”。

原因：

- 平台来自外部调用参数，需要校验。
- 未知平台继续执行会给出错误成功态。

### 4. 安装包生成不再猜配置

涉及文件：

- `core/packaging/inno/generator.go`
- `core/packaging/dmg/generator.go`

改动：

- Windows 生成器不再为空 `outputBaseName`、`privilegesRequired`、`appId` 补默认值。
- 资源类型只看 `asset.type`，不再根据文件系统猜测目录。
- macOS 生成器直接使用 `macos.createDmgPath`。
- 项目名、版本、Windows `productIdentifier` 缺失时直接报错。

原因：

- 生成器收到的是已校验配置和项目信息，主流程应直接使用结构字段。
- 资源类型和安装包元数据缺失属于配置错误，不应在模板生成时猜测。

### 5. 创建项目入口保持外部输入处理，但不再单独封装默认补全

涉及文件：

- `core/creator/service.go`

改动：

- 删除 `applyRequestDefaults`。
- 在 `CreateWailsProject` 中直接 trim 用户输入，并明确设置 `template=vanilla`、`packageName=main`。

原因：

- 创建项目请求来自用户输入，入口清洗和产品默认值是必要的。
- 这些赋值很简单，放在主流程中更清晰。

### 6. 删除未使用的通用兜底 helper

涉及文件：

- `core/fsx/fs.go`

改动：

- 删除 `FirstNonEmpty`。

原因：

- 后端已不再使用该通用兜底函数。
- 保留未使用的万能 helper 会鼓励后续继续用兜底掩盖数据结构问题。

## 保留的校验或兜底及原因

### `fsx.NormalizePath`

- 数据来源：用户选择目录、接口参数、系统路径。
- 解决问题：去除空白和引号，转换为绝对 clean 路径，避免空路径和相对路径进入文件操作。
- 不保留后果：可能在错误目录读写文件或产生不可比较的路径。
- 更直接写法：各入口分别 `TrimSpace + Abs + Clean`，但会重复且更容易不一致；当前集中在低层文件系统工具中是合理边界。

### 图标 base64 处理

涉及 `core/project/icon.go` 的 `normalizePNGBase64`。

- 数据来源：前端传入的用户图片数据。
- 解决问题：支持标准 PNG data URL 和带换行的 base64，并明确拒绝非 PNG。
- 不保留后果：用户选择的合法 PNG data URL 可能无法写入；非 PNG 可能进入 Wails 图标流程。
- 更直接写法：只接受裸 base64；但前端图片上传常见格式就是 data URL，所以保留入口处理。

### 命令版本输出清洗

涉及 `core/environment/service.go` 的 `normalizeCommandVersion`。

- 数据来源：`go/node/npm/wails3/git` 等第三方命令输出。
- 解决问题：去掉 ANSI 和多余前缀，给 UI 稳定展示版本号。
- 不保留后果：环境检测 UI 会显示不可读或不稳定的命令输出。
- 更直接写法：直接展示首行；但不同命令输出格式差异明显，保留针对命令输出的清洗是合理的。

### 设置文件默认值

涉及 `core/settings/store.go`。

- 数据来源：用户目录下的 settings.json，属于外部持久化文件。
- 解决问题：首次启动或设置文件损坏时，应用仍能进入可用状态。
- 不保留后果：设置文件缺失或损坏会阻断应用启动体验。
- 更直接写法：启动时报错并要求用户修复 settings.json；对普通桌面应用不友好，因此本次不改。

### Wails config 写入辅助

涉及 `core/project/wailsconfig.go` 的 `ensureInfoSection`、`ensureVersionPrefix`。

- 数据来源：项目内 `build/config.yml`，属于外部项目文件。
- 解决问题：保存用户编辑的项目信息时，保证 Wails 需要的 `info` 段和 `v` 版本格式存在。
- 不保留后果：保存项目信息可能无法落到正确 YAML 段，或版本格式不符合 Wails 配置预期。
- 更直接写法：引入完整 YAML 结构化读写；当前代码是局部文本编辑，本次仅保留必要边界处理。

## 测试

已执行：

```powershell
go test ./core/... ./desktop/...
```

结果：全部通过。

新增/调整测试重点：

- `packaging.json` 缺少必填字段时保存失败。
- `Package` 收到未知平台时直接报错。
- 原有运行时可执行文件路径测试改为使用完整打包配置。
