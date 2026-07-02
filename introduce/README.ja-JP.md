<div align="center">

<img src="../imgs/logo.png" alt="Wails3 Manager" width="96">

# Wails3 Manager

Wails 3 アプリを作成、設定、ビルド、パッケージ化するデスクトップ UI。

<img src="../imgs/badge.png" alt="Wails3 Manager badges" width="620">

[English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · **日本語** · [한국어](README.ko-KR.md) · [Français](README.fr-FR.md) · [Deutsch](README.de-DE.md)

</div>

> [!IMPORTANT]
> Wails3 Manager は開発中で、Wails 3 Alpha の変更に追従しています。Windows インストーラーは Windows、macOS DMG は macOS で作成します。

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../imgs/english/dark/home.png">
  <source media="(prefers-color-scheme: light)" srcset="../imgs/english/light/home.png">
  <img alt="Wails3 Manager workspace" src="../imgs/english/light/home.png">
</picture>

## できること

Wails3 Manager は、よく使う Wails プロジェクト作業をひとつの画面にまとめます。プロジェクトの作成やインポート、アプリ情報とアイコンの編集、ビルド設定、ネイティブパッケージ作成、ローカル環境チェック、ビルドログ確認ができます。

便利な点は単純です。ターミナル、設定ファイル、アイコン変換、インストーラースクリプトを行き来する回数を減らします。関連ファイルを書き込み、必要なコマンドを実行し、パッケージ出力をまとめます。

## ダウンロード

通常利用ではリリース版を使えます。

**[最新版をダウンロード](../../../releases/latest)** · [すべてのリリース](../../../releases)

1. OS に合うビルドをダウンロードします。
2. Wails3 Manager を起動し、環境ページを確認します。
3. 新しい Wails 3 プロジェクトを作成するか、既存プロジェクトをインポートします。
4. ワークスペースで設定、ビルド、パッケージ化を行います。

ネイティブパッケージ化には、Windows では **Inno Setup**、macOS では **create-dmg** が必要です。

## 機能

主な機能:

- 公式、ローカル、リモートの Wails テンプレートから作成し、ワークスペースで開く。
- 既存の Wails 3 プロジェクトをインポートし、最近使ったプロジェクトを保持。
- 製品情報を編集し、アプリアイコンを差し替え。PNG は Wails のプラットフォーム資産へ変換。
- ビルド名、本番モード、CGO、起動プログラム、同梱ファイルを設定。
- 対応 OS 上で Windows Inno Setup インストーラーと macOS DMG を作成。
- ビルド前に Go、Node.js、npm、Wails3 CLI、Git、パッケージングツールを確認。
- ビルドログを表示し、最終成果物を `builder/release` に集約。

## スクリーンショット

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

## 開発

技術スタック: Wails 3、Go、Vue 3、Vite、Naive UI、Pinia、Vue Router、Vue I18n、vue3-moveable。

要件:

- Go **1.25.0**
- `github.com/wailsapp/wails/v3 v3.0.0-alpha.96` と互換性のある Wails3 CLI
- Node.js と npm
- 利用 OS の Wails プラットフォームツールチェーン

```bash
git clone <repository-url>
cd wails3-manager/frontend
npm install
cd ..
wails3 task dev
```

便利なコマンド:

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

## 管理対象プロジェクトのファイル

インポートしたプロジェクトには `builder/` ディレクトリが作成されます。

```text
builder/
├── project.json
├── packaging.json
├── backup/initial/
├── windows/inno.iss
├── macos/dmg.sh
└── release/
```

## 現在の制限

- Wails 3 はまだ Alpha です。
- ネイティブパッケージは対象 OS 上で作成します。
- Linux パッケージ化は未実装です。
- 署名、公証、自動公開、キャンセル、バージョン付きビルド履歴は未実装です。
- 復元はインポート時の単一スナップショットを使います。

## ライセンス

Licensed under the [Apache License 2.0](../LICENSE).
