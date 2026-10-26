<div align="center">

<img src="../imgs/logo.png" alt="Wails3 Manager" width="96">

# Wails3 Manager

用桌面介面建立、設定、建置和打包 Wails 3 應用程式。

<img src="../imgs/badge.png" alt="Wails3 Manager badges" width="620">

[English](../README.md) · [简体中文](README.zh-CN.md) · **繁體中文** · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md) · [Français](README.fr-FR.md) · [Deutsch](README.de-DE.md) · [Русский](README.ru-RU.md)

</div>

> [!IMPORTANT]
> Wails3 Manager 以 Wails 3 beta.28 為目標。Windows 安裝套件需要在 Windows 上編譯，macOS DMG 需要在 macOS 上建立。

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../imgs/chinese/dark/home.png">
  <source media="(prefers-color-scheme: light)" srcset="../imgs/chinese/light/home.png">
  <img alt="Wails3 Manager 工作區" src="../imgs/chinese/light/home.png">
</picture>

## 它能幫什麼忙

Wails3 Manager 把常見的 Wails 專案操作放到一個視窗裡。你可以建立或匯入專案，編輯應用程式資訊和圖示，準備建置設定，產生原生安裝套件，檢查本機工具鏈，並查看建置日誌。

它帶來的便利很直接：少切換終端機，少手改設定檔，少處理圖示和安裝腳本。管理器會寫入相關專案檔案、執行需要的命令，並把安裝套件產物放到一起。

## 下載

1. 從 [Releases](https://github.com/Peiratooo/Wails3-Manager/releases) 下載符合目前系統的版本。
2. 啟動 Wails3 Manager，並檢查開發環境。
3. 建立新專案，或匯入既有 Wails 3 專案。
4. 在工作區完成設定、建置和打包。

原生打包需要 Windows 上的 **Inno Setup**，以及 macOS 上的 Wails CLI、系統磁碟映像與簽章工具。

DMG 提供官方標準配置和可拖曳圖示的自訂配置，新建設定預設使用官方模式。背景、卷宗圖示和 DMG 檔案圖示分別設定；未選擇背景時使用內建網格圖。

應用程式套件會先複製到暫存目錄。新建或修改過的套件使用 ad-hoc 簽章，未修改的預先建置套件保留原簽章。Developer ID 簽章和公證是獨立的發佈步驟。只有成功產生非空的新檔案後才會替換舊 DMG。

## 功能

主要功能：

- 從官方、本機或遠端 Wails 範本建立專案，並直接開啟到工作區。
- 匯入既有 Wails 3 專案，並保留最近專案列表。
- 編輯產品資訊並替換應用程式圖示；PNG 圖示會轉換成 Wails 平台資源。
- 設定建置名稱、生產模式、CGO、啟動程式和附加檔案。
- 在對應系統上產生 Windows Inno Setup 安裝套件和 macOS DMG。
- 建置前檢查 Go、Node.js、npm、Wails3 CLI、Git 和打包工具。
- 顯示即時建置日誌，並把最終產物收集到 `builder/release`。

## 介面截圖

<table>
<tr>
<td width="50%" valign="top"><strong>專案建立</strong><br><br><img src="../imgs/chinese/light/creator.png" alt="建立 Wails 3 專案"></td>
<td width="50%" valign="top"><strong>專案資訊</strong><br><br><img src="../imgs/chinese/light/wails3-config.png" alt="編輯 Wails 專案資訊"></td>
</tr>
<tr>
<td width="50%" valign="top"><strong>建置與打包</strong><br><br><img src="../imgs/chinese/dark/package-config.png" alt="設定建置與打包"></td>
<td width="50%" valign="top"><strong>環境檢查</strong><br><br><img src="../imgs/chinese/dark/enviroment.png" alt="檢查開發環境"></td>
</tr>
</table>

## 開發

技術棧：Wails 3、Go、Vue 3、Vite、Naive UI、Pinia、Vue Router、Vue I18n、vue3-moveable。

開發需求：

- Go **1.27.3**
- 與 `github.com/wailsapp/wails/v3 v3.0.0-beta.28` 相容的 Wails3 CLI
- Node.js **22.12+** 和 npm
- 使用 Go 1.27 建置的 macOS 應用程式需要 macOS 13 Ventura 或更新版本
- 目前系統所需的 Wails 平台工具鏈

```bash
git clone --branch 1.1.0 https://github.com/Peiratooo/Wails3-Manager.git wails3-manager
cd wails3-manager
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.28
cd frontend
npm ci
cd ..
wails3 task dev
```

常用命令：

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

macOS 雙架構 DMG 使用 `wails3 task darwin:package:universal:dmg` 建置，同時支援 Intel 與 Apple Silicon。公開發佈時，依序執行 `darwin:package:universal`、`darwin:sign:notarize` 和 `darwin:create:dmg`。簽章操作使用既有套件，不會重新建置。只建置本機架構時，將 `darwin:package:universal` 換成 `darwin:package`。

## 被管理專案檔案

匯入專案後，應用會在目標專案中建立 `builder/` 目錄：

```text
builder/
├── project.json
├── packaging.json
├── backup/initial/
├── windows/inno.iss
├── macos/dmg-background.png
└── release/
```

## 目前限制

- Wails 3 仍處於 Beta 階段，API 可能變更。
- 原生安裝套件必須在目標系統上建立。
- Linux 安裝套件尚未實作。
- 管理器介面不處理 Developer ID 簽章和公證，請使用上述 CLI 發佈步驟；新建或修改過的套件使用 ad-hoc 簽章。
- 自動發佈、任務取消和版本化建置歷史尚未實作。

## 授權

Wails3 Manager 採用 [Apache License 2.0](../LICENSE)。
