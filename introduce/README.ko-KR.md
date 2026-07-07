<div align="center">

<img src="../imgs/logo.png" alt="Wails3 Manager" width="96">

# Wails3 Manager

Wails 3 앱을 생성, 설정, 빌드, 패키징하는 데스크톱 UI입니다.

<img src="../imgs/badge.png" alt="Wails3 Manager badges" width="620">

[English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · **한국어** · [Français](README.fr-FR.md) · [Deutsch](README.de-DE.md)

</div>

> [!IMPORTANT]
> Wails3 Manager는 개발 중이며 Wails 3 Alpha 변경 사항을 따릅니다. Windows 설치 파일은 Windows에서, macOS DMG는 macOS에서 만듭니다.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../imgs/english/dark/home.png">
  <source media="(prefers-color-scheme: light)" srcset="../imgs/english/light/home.png">
  <img alt="Wails3 Manager workspace" src="../imgs/english/light/home.png">
</picture>

## 도움이 되는 부분

Wails3 Manager는 자주 하는 Wails 프로젝트 작업을 한 화면에 모읍니다. 프로젝트 생성 또는 가져오기, 앱 정보와 아이콘 편집, 빌드 설정, 네이티브 패키지 생성, 로컬 도구 확인, 빌드 로그 확인을 할 수 있습니다.

편한 점은 단순합니다. 터미널, 설정 파일, 아이콘 도구, 설치 스크립트 사이를 오가는 일을 줄입니다. 관련 프로젝트 파일을 쓰고, 필요한 명령을 실행하며, 패키지 결과물을 한곳에 모읍니다.

## 다운로드

1. 운영체제에 맞는 빌드를 다운로드합니다.
2. Wails3 Manager를 실행하고 환경 페이지를 확인합니다.
3. 새 Wails 3 프로젝트를 만들거나 기존 프로젝트를 가져옵니다.
4. 작업 공간에서 설정, 빌드, 패키징을 진행합니다.

네이티브 패키징에는 Windows의 **Inno Setup**, macOS의 **create-dmg**가 필요합니다.

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

- Go **1.25.0**
- `github.com/wailsapp/wails/v3 v3.0.0-alpha.96`와 호환되는 Wails3 CLI
- Node.js 및 npm
- 현재 OS의 Wails 플랫폼 툴체인

```bash
git clone <repository-url>
cd wails3-manager/frontend
npm install
cd ..
wails3 task dev
```

자주 쓰는 명령:

```bash
cd frontend
npm run i18n:check
npm run build
cd ..
go test ./core/... ./desktop/...
wails3 task build
wails3 task release
wails3 task package
```

## 관리되는 프로젝트 파일

가져온 프로젝트에는 `builder/` 디렉터리가 만들어집니다.

```text
builder/
├── project.json
├── packaging.json
├── backup/initial/
├── windows/inno.iss
├── macos/dmg.sh
└── release/
```

## 현재 제한

- Wails 3는 아직 Alpha 단계입니다.
- 네이티브 패키지는 대상 OS에서 만들어야 합니다.
- Linux 패키징은 구현되지 않았습니다.
- 코드 서명, 공증, 자동 배포, 취소, 버전별 빌드 기록은 구현되지 않았습니다.

## 라이선스

Licensed under the [Apache License 2.0](../LICENSE).
