# Frontend Contract

前端当前只提供工程壳，不做业务页面。统一从 `src/services/appService.js` 调后端：

```js
import { appService } from '@/services/appService'
```

后端不暴露 `PickDirectory`、`PickFile`、`OpenPath`。

## Services

`appService` 同时提供分组服务和常用方法别名：

```js
appService.project
appService.creator
appService.settings
appService.environment
appService.packaging
```

### CreatorService

```js
await appService.CreateWailsProject(request)
```

`CreateWailsProject` 是 `wails3 init` 的薄封装，成功后返回新项目绝对路径，不会自动导入项目列表。`request.name` 必填；`request.dir` 是目标项目目录，省略时使用当前工作目录下的项目名目录。可选参数映射 Wails3 CLI：`template`、`packageName`、`goModule`、`git`、`productName`、`productDescription`、`productVersion`、`productCompany`、`productCopyright`、`productComments`、`productIdentifier`、`quiet`、`skipRemoteTemplateWarning`、`skipGoModTidy`。

### ProjectService

```js
await appService.ScanProject(projectDir)
await appService.ImportProject(projectDir)
await appService.SaveProject(projectRecord)
await appService.ReplaceProjectIcon(projectDir, pngBase64)
```

`ImportProject` 成功时不返回项目记录；前端需要重新调用 `ListProjects` 刷新列表。它会创建 `builder/project.json`、`builder/backup/initial` 和 `builder/packaging.json`，并按当前系统生成默认模板：Windows 为 Inno，macOS 为 DMG。

`ReplaceProjectIcon` 接收 PNG 的 base64 字符串，可以是纯 base64，也可以是 `data:image/png;base64,...`。后端会覆盖 `build/appicon.png`，然后执行 `wails3 task common:update:build-assets`。

### SettingsService

```js
await appService.ListProjects()
await appService.OpenProject(projectDir)
await appService.RemoveProject(projectDir, restoreOriginal)
await appService.GetSettings()
await appService.SaveSettings(settings)
await appService.ClearLogs()
```

`ManagerSettings` 使用布尔主题字段：

```js
{
  isDark: true,
  language: 'zh-CN',
  recordLogs: true
}
```

运行日志不再通过 `LogsSince` 轮询读取。后端每写入一条日志时会推送 Wails 事件 `manager:log-line`，payload 为 `{ line: string }`；前端使用 `Events.On('manager:log-line', handler)` 订阅。

`RemoveProject(projectDir, true)` 会恢复唯一的导入前快照、删除 `builder/`，并从本地项目列表移除。快照缺失时会报错并保留项目记录。

项目管理模块不会维护备份列表，也不会在保存项目或替换图标时创建额外备份。恢复来源固定为 `builder/backup/initial`。

### EnvironmentService

```js
await appService.CheckEnvironment()
```

返回当前 OS/Arch 和 Go、Node、包管理器、Wails3、Inno Setup、create-dmg 等检测结果。

### PackagingService

```js
await appService.InitPackaging(projectDir)
await appService.LoadPackagingConfig(projectDir)
await appService.SavePackagingConfig(projectDir, packagingConfig)
await appService.GetPackagingRuntimeInfo(projectDir)
await appService.Package({ projectDir, platform, dryRun, runBuild })
await appService.Artifacts(projectDir)
```

打包配置固定写入 `builder/packaging.json`。`platform` 可用 `auto`、`windows`、`darwin`、`all`。

打包阶段不支持 before/after 脚本字段，前端不要提交 `scripts` 配置；需要自定义构建时设置 `packagingConfig.build.command`。

当前 `PackagingConfig` 只保留一个入口程序和一个资产列表：

```js
{
  schemaVersion: 1,
  build: {
    taskfile: 'Taskfile.yml',
    task: 'release',
    command: null,
    production: true,
    cgoEnabled: true,
    appName: 'demo'
  },
  entry: {
    executablePath: ''
  },
  assets: [
    { src: 'README.md', type: 'file', required: false },
    { src: 'assets', type: 'directory', required: true }
  ],
  windows: {
    appURL: 'https://example.com/demo',
    createDesktopShortcut: true
  }
}
```

`packaging.json` 不保存产品名、版本、bundleId、描述、版权或图标；打包时这些基础信息从 `builder/project.json` 读取。初始化只写当前系统的平台配置：Windows 写 `windows`，macOS 写 `macos`。`windows.appURL` 是 Inno 专属 URL，会写入 `MyAppURL`；`MyAppExeName` 由 `entry.executablePath` 或默认的 `bin/${build.appName}.exe` 推导。`build.appName` 是打包构建的应用名来源；默认构建命令会把 `APP_NAME`、`PRODUCTION`、`CGO_ENABLED` 传给 Task。资产不再区分平台；文件和目录都会放到安装根目录，目录会保留目录本身。

`GetPackagingRuntimeInfo(projectDir)` 返回只读运行信息：

```js
{
  defaultExecutablePath: 'bin/demo.exe',
  effectiveExecutablePath: 'bin/demo.exe',
  usingDefaultExecutable: true
}
```

当前入口为空时，后端不会把默认值写回 `packaging.json`，但会在打包和前端提示里统一使用 Wails3 构建产物：Windows 为 `bin/${build.appName}.exe`，macOS 为 `bin/${build.appName}.app`。

## ProjectRecord

```js
{
  projectDir: 'C:/path/to/app',
  importedAt: 1770000000,
  lastOpenedAt: 1770000300,
  project: {
    projectDir: 'C:/path/to/app',
    currentPlatform: 'windows',
    wailsConfig: {
      info: {
        companyName: 'Acme',
        productName: 'Demo',
        productIdentifier: 'com.acme.demo',
        description: 'Demo app',
        copyright: 'Acme',
        comments: '',
        version: '1.0.0'
      },
      icon: 'build/appicon.png',
      fileAssociations: []
    }
  }
}
```

`importedAt` 和 `lastOpenedAt` 是 Unix 秒。
