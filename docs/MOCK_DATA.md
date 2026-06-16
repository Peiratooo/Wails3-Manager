# Mock JSON Data

这份文件覆盖当前稳定 JSON DTO：

- `core/contracts` 中的前端/配置数据结构。
- `core/project/backup.go` 中落盘的初始快照 manifest。

运行期服务结构体，例如 `ProjectService`、`SettingsService`、`Runner`、`Logger`，不是 JSON 数据合约，不提供模拟数据。

## Scalar Values

`Platform`：

```json
"windows"
```

可选值：

```json
["auto", "windows", "darwin", "all"]
```

`UnixTime`：

```json
1770000000
```

`ValidationSeverity`：

```json
"warning"
```

可选值：

```json
["ok", "warning", "error"]
```

## PackagingConfig

`PackagingConfig` 是唯一的打包配置结构，固定保存为 `builder/packaging.json`。初始化只写当前系统的平台配置；下面示例为 Windows。

```json
{
  "schemaVersion": 1,
  "build": {
    "taskfile": "Taskfile.yml",
    "task": "builder:release",
    "command": ["wails3", "task", "builder:release"],
    "production": true,
    "cgoEnabled": true,
    "appName": "demo-desktop"
  },
  "entry": {
    "executablePath": ""
  },
  "assets": [
    {
      "src": "README.md",
      "type": "file",
      "required": false
    },
    {
      "src": "assets",
      "type": "directory",
      "required": true
    }
  ],
  "windows": {
    "enabled": true,
    "innoScript": "builder/windows/inno.iss",
    "isccPath": "C:/Program Files (x86)/Inno Setup 6/ISCC.exe",
    "defaultDirName": "{autopf}/${project.name}",
    "privilegesRequired": "lowest",
    "setupIcon": "build/windows/icon.ico",
    "outputBaseName": "${build.appName}-${project.version}-windows-setup",
    "appURL": "https://example.com/demo",
    "createDesktopShortcut": true
  },
  "artifacts": {
    "outputRoot": "builder/release"
  }
}
```

## BuildSettings

`appName` 是打包构建的应用名来源，用于安装包文件名占位符 `${build.appName}`；默认构建会把它作为 `APP_NAME` Task 变量和环境变量传入。

```json
{
  "taskfile": "Taskfile.yml",
  "task": "builder:release",
  "command": ["wails3", "task", "builder:release"],
  "production": true,
  "cgoEnabled": true,
  "appName": "demo-desktop"
}
```

## ProgramEntry

```json
{
  "executablePath": ""
}
```

为空时 Windows 默认使用 `bin/${build.appName}.exe`。

## PackagingAsset

`type` 使用 `"file"` 或 `"directory"`。资产统一放到安装根目录；目录会以目录本身作为根目录子项，不会只展开目录内容。

```json
{
  "src": "assets",
  "type": "directory",
  "required": true
}
```

## WindowsConfig

```json
{
  "enabled": true,
  "innoScript": "builder/windows/inno.iss",
  "isccPath": "C:/Program Files (x86)/Inno Setup 6/ISCC.exe",
  "defaultDirName": "{autopf}/${project.name}",
  "privilegesRequired": "lowest",
  "setupIcon": "build/windows/icon.ico",
  "outputBaseName": "${build.appName}-${project.version}-windows-setup",
  "appURL": "https://example.com/demo",
  "createDesktopShortcut": true
}
```

## MacOSConfig

```json
{
  "enabled": true,
  "appBundle": "bin/${build.appName}.app",
  "dmgScript": "builder/macos/dmg.sh",
  "background": "builder/macos/background.png",
  "outputName": "${build.appName}-${project.version}",
  "createDmgPath": "create-dmg",
  "windowWidth": 640,
  "windowHeight": 420,
  "iconSize": 96,
  "appX": 180,
  "appY": 210,
  "applicationsX": 460,
  "applicationsY": 210
}
```

## ArtifactConfig

```json
{
  "outputRoot": "builder/release"
}
```

## ScanResult

```json
{
  "projectDir": "C:/work/demo-desktop",
  "score": 90,
  "isWailsProject": true
}
```

## Artifact

```json
{
  "kind": "installer",
  "name": "demo-desktop-setup.exe",
  "path": "C:/work/demo-desktop/builder/release/windows/demo-desktop-setup.exe",
  "size": 52428800,
  "createdAt": 1770000300
}
```

## CommandResult

```json
{
  "ok": true,
  "message": "Package completed",
  "logs": ["running wails3 task builder:release", "generated installer"],
  "artifacts": [
    {
      "kind": "installer",
      "name": "demo-desktop-setup.exe",
      "path": "C:/work/demo-desktop/builder/release/windows/demo-desktop-setup.exe",
      "size": 52428800,
      "createdAt": 1770000300
    }
  ],
  "artifactChecks": [
    {
      "platform": "windows",
      "kind": "installer",
      "path": "C:/work/demo-desktop/builder/release/windows/demo-desktop-setup.exe",
      "found": true,
      "message": "found"
    }
  ],
  "warnings": [],
  "runId": "20260529-160000",
  "cacheDir": "C:/work/demo-desktop/builder/cache/20260529-160000",
  "releaseDir": "C:/work/demo-desktop/builder/release",
  "buildArtifacts": [],
  "finalArtifacts": []
}
```

## ArtifactCheck

```json
{
  "platform": "windows",
  "kind": "installer",
  "path": "C:/work/demo-desktop/builder/release/windows/demo-desktop-setup.exe",
  "found": true,
  "message": "found"
}
```

## ToolRequirement

```json
{
  "id": "inno",
  "name": "Inno Setup",
  "platform": "windows",
  "command": "ISCC.exe",
  "required": false,
  "found": true,
  "path": "C:/Program Files (x86)/Inno Setup 6/ISCC.exe",
  "configuredPath": "C:/Program Files (x86)/Inno Setup 6/ISCC.exe",
  "version": "6.3.3",
  "message": "Inno Setup is available",
  "installHint": "Install Inno Setup 6",
  "downloadUrl": "https://jrsoftware.org/isdl.php",
  "installCommand": ["winget", "install", "JRSoftware.InnoSetup"],
  "canAutoInstall": false,
  "canChoosePath": true
}
```

## ToolCheck

```json
{
  "id": "node",
  "name": "Node.js",
  "command": "node",
  "found": true,
  "version": "v22.12.0",
  "path": "C:/Program Files/nodejs/node.exe",
  "required": true,
  "platform": "all",
  "message": "Node.js is available"
}
```

## EnvironmentReport

```json
{
  "os": "windows",
  "arch": "amd64",
  "ok": true,
  "checks": [
    {
      "id": "go",
      "name": "Go",
      "command": "go",
      "found": true,
      "version": "go1.25.4",
      "path": "C:/Program Files/Go/bin/go.exe",
      "required": true,
      "platform": "all",
      "message": "Go is available"
    }
  ]
}
```

## ValidationCheck

```json
{
  "key": "windows.innoScript",
  "title": "Inno script",
  "severity": "warning",
  "message": "Script will be generated before packaging",
  "path": "builder/windows/inno.iss"
}
```

## ValidationReport

```json
{
  "ok": true,
  "checks": [
    {
      "key": "project.name",
      "title": "Project name",
      "severity": "ok",
      "message": "Project name is set"
    }
  ],
  "warnings": 0,
  "errors": 0
}
```

## WailsProjectConfig

```json
{
  "info": {
    "companyName": "Acme Inc.",
    "productName": "Demo Desktop",
    "productIdentifier": "com.acme.demo",
    "description": "Demo Wails desktop app",
    "copyright": "Copyright 2026 Acme Inc.",
    "comments": "Internal demo build",
    "version": "1.2.3"
  },
  "icon": "build/appicon.png",
  "fileAssociations": [
    {
      "ext": ".demo",
      "name": "Demo Document",
      "description": "Demo project document",
      "iconName": "demo-doc",
      "role": "Editor",
      "mimeType": "application/x-demo"
    }
  ]
}
```

## WailsAppInfo

```json
{
  "companyName": "Acme Inc.",
  "productName": "Demo Desktop",
  "productIdentifier": "com.acme.demo",
  "description": "Demo Wails desktop app",
  "copyright": "Copyright 2026 Acme Inc.",
  "comments": "Internal demo build",
  "version": "1.2.3"
}
```

## WailsFileAssociation

```json
{
  "ext": ".demo",
  "name": "Demo Document",
  "description": "Demo project document",
  "iconName": "demo-doc",
  "role": "Editor",
  "mimeType": "application/x-demo"
}
```

## WailsProjectManager

```json
{
  "projectDir": "C:/work/demo-desktop",
  "currentPlatform": "windows",
  "wailsConfig": {
    "info": {
      "companyName": "Acme Inc.",
      "productName": "Demo Desktop",
      "productIdentifier": "com.acme.demo",
      "description": "Demo Wails desktop app",
      "copyright": "Copyright 2026 Acme Inc.",
      "comments": "Internal demo build",
      "version": "1.2.3"
    },
    "icon": "build/appicon.png",
    "fileAssociations": []
  }
}
```

## ProjectRecord

```json
{
  "projectDir": "C:/work/demo-desktop",
  "project": {
    "projectDir": "C:/work/demo-desktop",
    "currentPlatform": "windows",
    "wailsConfig": {
      "info": {
        "companyName": "Acme Inc.",
        "productName": "Demo Desktop",
        "productIdentifier": "com.acme.demo",
        "description": "Demo Wails desktop app",
        "copyright": "Copyright 2026 Acme Inc.",
        "comments": "",
        "version": "1.2.3"
      },
      "icon": "build/appicon.png",
      "fileAssociations": []
    }
  },
  "importedAt": 1770000000,
  "lastOpenedAt": 1770000300
}
```

## UserState

```json
{
  "projects": [
    {
      "projectDir": "C:/work/demo-desktop",
      "project": {
        "projectDir": "C:/work/demo-desktop",
        "currentPlatform": "windows",
        "wailsConfig": {
          "info": {
            "companyName": "Acme Inc.",
            "productName": "Demo Desktop",
            "productIdentifier": "com.acme.demo",
            "description": "Demo Wails desktop app",
            "copyright": "Copyright 2026 Acme Inc.",
            "comments": "",
            "version": "1.2.3"
          },
          "icon": "build/appicon.png",
          "fileAssociations": []
        }
      },
      "importedAt": 1770000000,
      "lastOpenedAt": 1770000300
    }
  ]
}
```

## ManagerSettings

```json
{
  "isDark": true,
  "language": "zh-CN",
  "recordLogs": true
}
```

## PackageRequest

```json
{
  "projectDir": "C:/work/demo-desktop",
  "platform": "windows",
  "dryRun": false,
  "runBuild": true
}
```

## PackageResult

```json
{
  "ok": true,
  "message": "打包流程完成",
  "runId": "20260529-160000",
  "artifacts": [
    {
      "kind": "installer",
      "name": "demo-desktop-setup.exe",
      "path": "C:/work/demo-desktop/builder/release/windows/demo-desktop-setup.exe",
      "size": 52428800,
      "createdAt": 1770000300
    }
  ],
  "warnings": []
}
```

## LogEntry

```json
{
  "line": "wails3 task builder:release completed",
  "createdAt": 1770000300
}
```

## LogLineEvent

```json
{
  "line": "[12:00:00] running wails3 task builder:release"
}
```

## SnapshotFile

`SnapshotFile` 只用于 `builder/backup/initial/manifest.json`。

```json
{
  "source": "build",
  "backup": "initial/build"
}
```

## SnapshotManifest

`SnapshotManifest` 是导入项目时唯一创建的初始快照 manifest。保存项目、替换图标不会新增其他快照。

```json
{
  "id": "initial",
  "reason": "initial-import",
  "createdAt": 1770000000,
  "files": [
    {
      "source": "build",
      "backup": "initial/build"
    },
    {
      "source": "Taskfile.yml",
      "backup": "initial/Taskfile.yml"
    }
  ]
}
```
