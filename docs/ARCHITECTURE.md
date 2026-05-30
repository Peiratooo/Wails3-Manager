# Architecture

## Module Boundary

后端拆成四个 Wails 服务，不再通过单个 `Core` 聚合：

```text
desktop.Run
  -> ProjectService
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
  -> project.UpsertProjectRecord writes local state.json
```

`OpenProject` 属于 `SettingsService`，只打开已经导入的项目。它读取 `builder/project.json`，再从真实项目文件刷新 `WailsProjectManager`，更新 `lastOpenedAt`。

`SaveProject` 属于 `ProjectService`，直接写回项目文件：

- `WailsConfig.Info` -> `build/config.yml`
- `TaskVars.AppName` / `Production` / `CGOEnabled` -> `Taskfile.yml`
- 保存后执行 `wails3 task common:update:build-assets`

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
settings.json   theme / language / log settings
```

## Packaging

`PackagingService` 只处理打包数据和打包运行，不参与项目导入/保存。

- `InitPackaging` 创建 `builder/packaging.json`，并生成 Inno / DMG 模板。
- Windows 打包生成 Inno 脚本；非 Windows 或 dry-run 时不执行 ISCC。
- macOS 打包生成 DMG 脚本；非 macOS 或 dry-run 时不执行 create-dmg。
- 打包阶段不再支持 before/after 脚本。需要自定义构建时，只配置 `packaging.json` 里的显式 build command。

## Import Rules

架构测试会检查：

- 不能出现循环 import。
- `project` 不依赖 `settings` 或 `packaging`。
- `project` 负责目标项目内文件和本地 ProjectRecord 索引。
- `settings` 负责管理器设置、日志读取，并通过 project 包打开/删除已导入项目。
- `contracts` 只保存 DTO，不包含业务行为。
