# Architecture

## Module Boundary

后端拆成四个 Wails 服务，不再通过单个 `Core` 聚合：

```text
desktop.Run
  -> ProjectService
  -> CreatorService
  -> EnvironmentService
  -> SettingsService
  -> PackagingService
```

跨模块数据放在 `core/contracts`；底层文件和命令能力只保留 `core/fsx` 与 `core/runlog`。业务包保持单向依赖：`project` 不 import `settings/packaging`，`settings` 可以读取 `project` 的项目文件工具，`packaging` 可以读取 `project` 记录和 `environment` 工具检测。

## Data Flow

```text
ProjectService.ImportProject(path)
  -> project/scanner score >= 60
  -> project reads build/config.yml and Taskfile.yml
  -> project creates builder/backup/initial
  -> project writes builder/project.json
  -> injected PackagingService.InitPackaging writes builder/packaging.json and templates
  -> project.UpsertProjectRecord writes local state.json
```

`ImportProject` 成功时只返回 nil error，不向前端返回 ProjectRecord；前端导入后重新读取项目列表。

`OpenProject` 属于 `SettingsService`，只打开已经导入的项目。它读取 `builder/project.json`，再从真实项目文件刷新 `WailsProjectManager`，更新 `lastOpenedAt`。

`SaveProject` 属于 `ProjectService`，直接写回项目文件：

- `WailsConfig.Info` -> `build/config.yml`
- 保存后执行 `wails3 task common:update:build-assets`

`APP_NAME` / `PRODUCTION` / `CGO_ENABLED` 不再属于 `project.json`。首次初始化打包配置时会从 Taskfile 读取旧默认值，之后以 `builder/packaging.json` 的 `build` 字段为准。

项目管理只在导入时创建一次 `builder/backup/initial`。保存项目、替换图标等后续编辑不会追加备份，也不会维护备份列表。

## Stored Files

项目内：

```text
builder/
├── project.json
├── packaging.json
└── backup/
    └── initial/
        ├── build/
        ├── Taskfile.yml
        └── manifest.json
```

本地管理器状态：

```text
state.json      ProjectRecord[]
settings.json   isDark / language / recordLogs
```

运行日志保存在内存中；`recordLogs=true` 时每条日志通过 Wails 事件 `manager:log-line` 推送给前端，`recordLogs=false` 时不缓存也不推送新日志。

## Packaging

`PackagingService` 处理打包数据和打包运行。`desktop.Run` 会把 `InitPackaging` 注入给 `ProjectService`，让导入成功时同时初始化打包配置；`core/project` 仍然不直接 import `core/packaging`。

- `InitPackaging` 创建或读取 `builder/packaging.json`，只初始化当前运行系统的配置和模板：Windows 生成 Inno，macOS 生成 DMG。
- `packaging.json` 不保存 name/version/bundleId/icon 等项目基础信息；Inno/DMG 生成时直接读取 `builder/project.json`。
- `windows.appURL` 保存 Inno 专属 URL；`MyAppExeName` 从 `entry.executablePath` 或 `build.appName` 推导，不额外存一份 exe 名。
- 默认构建会把 `build.appName`、`build.production`、`build.cgoEnabled` 作为 Task 变量和环境变量传入；未配置 `entry.executablePath` 时，Windows 默认启动程序为 `bin/${build.appName}.exe`。
- 前端通过 `GetPackagingRuntimeInfo` 读取只读运行信息；默认入口不会写回 `packaging.json`，避免配置和派生值混在一起。
- Windows 打包生成 Inno 脚本；非 Windows 或 dry-run 时不执行 ISCC。
- macOS 打包生成 DMG 脚本；非 macOS 或 dry-run 时不执行 create-dmg。
- 打包阶段不再支持 before/after 脚本。需要自定义构建时，只配置 `packaging.json` 里的显式 build command。

## Project Creation

`CreatorService` 是独立服务，只负责把前端参数映射成 `wails3 init` 命令并返回新项目绝对路径。它不导入项目、不写管理器状态，也不依赖 `project` 或 `packaging` 包。

## Import Rules

架构测试会检查：

- 不能出现循环 import。
- `project` 不依赖 `settings` 或 `packaging`。
- `project` 负责目标项目内文件和本地 ProjectRecord 索引。
- `settings` 负责管理器设置、日志开关，并通过 project 包打开/删除已导入项目。
- `contracts` 只保存 DTO，不包含业务行为。
