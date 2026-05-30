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
appService.settings
appService.environment
appService.packaging
```

### ProjectService

```js
await appService.ScanProject(projectDir)
await appService.ImportProject(projectDir)
await appService.SaveProject(projectRecord)
await appService.ReplaceProjectIcon(projectDir, sourcePath)
```

`ImportProject` 会创建 `builder/project.json` 和 `builder/backup/initial`，不会创建 `builder/packaging.json`。

### SettingsService

```js
await appService.ListProjects()
await appService.OpenProject(projectDir)
await appService.RemoveProject(projectDir, restoreOriginal)
await appService.GetSettings()
await appService.SaveSettings(settings)
await appService.LogsSince(cursor)
await appService.ClearLogs()
```

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
    task: 'builder:release',
    command: ['wails3', 'task', 'builder:release'],
    production: true,
    cgoEnabled: true,
    appName: 'demo'
  },
  entry: {
    executablePath: 'bin/demo.exe'
  },
  assets: [
    { src: 'README.md', type: 'file', required: false },
    { src: 'assets', type: 'directory', required: true }
  ],
  windows: {
    createDesktopShortcut: true
  }
}
```

`build.appName` 来自项目 Taskfile 的 `APP_NAME`，用于 `${build.appName}` 占位符。资产不再区分平台；文件和目录都会放到安装根目录，目录会保留目录本身。

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
    },
    taskVars: {
      appName: 'demo',
      production: false,
      cgoEnabled: '1'
    }
  }
}
```

`importedAt` 和 `lastOpenedAt` 是 Unix 秒。
