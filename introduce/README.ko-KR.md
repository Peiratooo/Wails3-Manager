<div align="center">

<img src="../imgs/logo.png" alt="Wails3 Manager" width="96">

# Wails3 Manager

Wails 3 앱을 생성, 설정, 빌드, 패키징하는 데스크톱 UI입니다.

<img src="../imgs/badge.png" alt="Wails3 Manager badges" width="620">

[English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · **한국어** · [Français](README.fr-FR.md) · [Deutsch](README.de-DE.md) · [Русский](README.ru-RU.md)

</div>

> [!IMPORTANT]
> Wails3 Manager는 Wails 3 beta.28을 대상으로 합니다. Windows 설치 파일은 Windows에서, macOS DMG는 macOS에서 만듭니다.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../imgs/english/dark/home.png">
  <source media="(prefers-color-scheme: light)" srcset="../imgs/english/light/home.png">
  <img alt="Wails3 Manager workspace" src="../imgs/english/light/home.png">
</picture>

## 도움이 되는 부분

Wails3 Manager는 자주 하는 Wails 프로젝트 작업을 한 화면에 모읍니다. 프로젝트 생성 또는 가져오기, 앱 정보와 아이콘 편집, 빌드 설정, 네이티브 패키지 생성, 로컬 도구 확인, 빌드 로그 확인을 할 수 있습니다.

편한 점은 단순합니다. 터미널, 설정 파일, 아이콘 도구, 설치 스크립트 사이를 오가는 일을 줄입니다. 관련 프로젝트 파일을 쓰고, 필요한 명령을 실행하며, 패키지 결과물을 한곳에 모읍니다.

## 다운로드

1. [Releases](https://github.com/Peiratooo/Wails3-Manager/releases)에서 운영체제에 맞는 빌드를 다운로드합니다.
2. Wails3 Manager를 실행하고 환경 페이지를 확인합니다.
3. 새 Wails 3 프로젝트를 만들거나 기존 프로젝트를 가져옵니다.
4. 작업 공간에서 설정, 빌드, 패키징을 진행합니다.

네이티브 패키징에는 Windows의 **Inno Setup**, macOS의 Wails CLI와 시스템 디스크 이미지 및 서명 도구가 필요합니다.

DMG는 공식 표준 레이아웃과 아이콘을 드래그할 수 있는 사용자 지정 레이아웃을 지원합니다. 새 설정은 공식 모드를 사용합니다. 배경, 볼륨 아이콘, DMG 파일 아이콘은 각각 설정하며, 배경을 선택하지 않으면 내장 격자 이미지를 사용합니다.

패키징 전에 앱 번들을 임시 디렉터리로 복사합니다. 새 번들이나 수정된 번들은 ad-hoc 서명을 적용하고, 변경되지 않은 사전 빌드 번들은 원래 서명을 유지합니다. Developer ID 서명과 공증은 별도의 배포 단계입니다. 비어 있지 않은 새 DMG 파일을 성공적으로 만든 후에만 기존 파일을 교체합니다.

## 기능

주요 기능:

- 공식, 로컬, 원격 Wails 템플릿으로 프로젝트를 만들고 작업 공간에서 엽니다.
- 기존 Wails 3 프로젝트를 가져오고 최근 프로젝트 목록을 유지합니다.
- 제품 정보를 편집하고 앱 아이콘을 교체합니다. PNG는 Wails 플랫폼 자산으로 변환됩니다.
- 빌드 이름, 프로덕션 모드, CGO, 실행 프로그램, 포함 파일을 설정합니다.
- 맞는 OS에서 Windows Inno Setup 설치 파일과 macOS DMG를 만듭니다.
- 빌드 전에 Go, Node.js, npm, Wails3 CLI, Git, 패키징 도구를 확인합니다.
- 빌드 로그를 표시하고 최종 결과물을 `builder/release`에 모읍니다.

## 스크린샷

<table>
<tr>
<td width="50%" valign="top"><strong>Project creation</strong><br><br><img src="../imgs/english/light/creator.png" alt="Create a Wails 3 project"></td>
<td width="50%" valign="top"><strong>Project metadata</strong><br><br><img src="../imgs/english/light/wails3-config.png" alt="Edit Wails project metadata"></td>
</tr>
<tr>
<td width="50%" valign="top"><strong>Build and packaging</strong><br><br><img src="../imgs/english/dark/package-config.png" alt="Configure builds and installers"></td>
<td width="50%" valign="top"><strong>Environment check</strong><br><br><img src="../imgs/english/dark/enviroment.png" alt="System environment report"></td>
</tr>
</table>

## 개발

기술 스택: Wails 3, Go, Vue 3, Vite, Naive UI, Pinia, Vue Router, Vue I18n, vue3-moveable.

요구 사항:

- Go **1.27.3**
- `github.com/wailsapp/wails/v3 v3.0.0-beta.28`와 호환되는 Wails3 CLI
- Node.js **22.12+** 및 npm
- Go 1.27로 빌드한 앱은 macOS 13 Ventura 이상 필요
- 현재 OS의 Wails 플랫폼 툴체인

```bash
git clone --branch 1.1.0 https://github.com/Peiratooo/Wails3-Manager.git wails3-manager
cd wails3-manager
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.28
cd frontend
npm ci
cd ..
wails3 task dev
```

자주 쓰는 명령:

```bash
cd frontend
npm run i18n:check
npm run build
cd ..
go vet ./...
wails3 task build
wails3 task release
wails3 task package
```

macOS에서 `wails3 task darwin:package:universal:dmg`는 Intel과 Apple Silicon을 모두 지원하는 DMG를 만듭니다. 공개 배포 시 `darwin:package:universal`, `darwin:sign:notarize`, `darwin:create:dmg`를 순서대로 실행하세요. 서명 단계는 기존 번들을 사용하며 다시 빌드하지 않습니다. 네이티브 아키텍처만 필요하면 `darwin:package:universal` 대신 `darwin:package`를 사용하세요.

## 관리되는 프로젝트 파일

가져온 프로젝트에는 `builder/` 디렉터리가 만들어집니다.

```text
builder/
├── project.json
├── packaging.json
├── backup/initial/
├── windows/inno.iss
├── macos/dmg-background.png
└── release/
```

## 현재 제한

- Wails 3는 Beta 단계이며 API가 변경될 수 있습니다.
- 네이티브 패키지는 대상 OS에서 만들어야 합니다.
- Linux 패키징은 구현되지 않았습니다.
- 관리 화면은 Developer ID 서명과 공증을 처리하지 않습니다. 위 CLI 배포 단계를 사용하세요. 새 번들이나 수정된 번들에는 ad-hoc 서명을 적용합니다.
- 자동 배포, 취소, 버전별 빌드 기록은 구현되지 않았습니다.

## 라이선스

Licensed under the [Apache License 2.0](../LICENSE).
