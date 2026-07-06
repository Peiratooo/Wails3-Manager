<div align="center">

<img src="../imgs/logo.png" alt="Wails3 Manager" width="96">

# Wails3 Manager

用桌面介面建立、設定、建置和打包 Wails 3 應用程式。

<img src="../imgs/badge.png" alt="Wails3 Manager badges" width="620">

[English](../README.md) · [简体中文](README.zh-CN.md) · **繁體中文** · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md) · [Français](README.fr-FR.md) · [Deutsch](README.de-DE.md)

</div>

> [!IMPORTANT]
> Wails3 Manager 仍在開發中，並跟隨 Wails 3 Alpha 版本變化。Windows 安裝套件需要在 Windows 上編譯，macOS DMG 需要在 macOS 上建立。

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../imgs/chinese/dark/home.png">
  <source media="(prefers-color-scheme: light)" srcset="../imgs/chinese/light/home.png">
  <img alt="Wails3 Manager 工作區" src="../imgs/chinese/light/home.png">
</picture>

## 它能幫什麼忙

Wails3 Manager 把常見的 Wails 專案操作放到一個視窗裡。你可以建立或匯入專案，編輯應用程式資訊和圖示，準備建置設定，產生原生安裝套件，檢查本機工具鏈，並查看建置日誌。

它帶來的便利很直接：少切換終端機，少手改設定檔，少處理圖示和安裝腳本。管理器會寫入相關專案檔案、執行需要的命令，並把安裝套件產物放到一起。

## 下載

一般使用直接下載發行版即可：

**[下載最新版本](../../../releases/latest)** · [查看全部版本](../../../releases)

1. 下載符合目前系統的版本。
2. 啟動 Wails3 Manager，並檢查開發環境。
3. 建立新專案，或匯入既有 Wails 3 專案。
4. 在工作區完成設定、建置和打包。

使用原生打包功能時，需要安裝對應工具：Windows 使用 **Inno Setup**，macOS 使用 **create-dmg**。

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

- Go **1.25.0**
- 與 `github.com/wailsapp/wails/v3 v3.0.0-alpha.96` 相容的 Wails3 CLI
- Node.js 和 npm
- 目前系統所需的 Wails 平台工具鏈

```bash
git clone <你的倉庫地址>
cd wails3-manager/frontend
npm install
cd ..
wails3 task dev
```

常用命令：

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

## 被管理專案檔案

匯入專案後，應用會在目標專案中建立 `builder/` 目錄：

```text
builder/
├── project.json
├── packaging.json
├── backup/initial/
├── windows/inno.iss
├── macos/dmg.sh
└── release/
```

## 目前限制

- Wails 3 仍處於 Alpha 階段。
- 原生安裝套件必須在目標系統上建立。
- Linux 安裝套件尚未實作。
- 程式碼簽章、公證、自動發布、任務取消和版本化建置歷史尚未實作。
- 專案還原只使用匯入時的單一快照，不是完整備份歷史。

## 授權

Wails3 Manager 採用 [Apache License 2.0](../LICENSE)。
