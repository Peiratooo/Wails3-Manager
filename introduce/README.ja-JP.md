<div align="center">

<img src="../imgs/logo.png" alt="Wails3 Manager" width="96">

# Wails3 Manager

Wails 3 アプリを作成、設定、ビルド、パッケージ化するデスクトップ UI。

<img src="../imgs/badge.png" alt="Wails3 Manager badges" width="620">

[English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · **日本語** · [한국어](README.ko-KR.md) · [Français](README.fr-FR.md) · [Deutsch](README.de-DE.md) · [Русский](README.ru-RU.md)

</div>

> [!IMPORTANT]
> Wails3 Manager は Wails 3 beta.28 を対象としています。Windows インストーラーは Windows、macOS DMG は macOS で作成します。

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../imgs/english/dark/home.png">
  <source media="(prefers-color-scheme: light)" srcset="../imgs/english/light/home.png">
  <img alt="Wails3 Manager workspace" src="../imgs/english/light/home.png">
</picture>

## できること

Wails3 Manager は、よく使う Wails プロジェクト作業をひとつの画面にまとめます。プロジェクトの作成やインポート、アプリ情報とアイコンの編集、ビルド設定、ネイティブパッケージ作成、ローカル環境チェック、ビルドログ確認ができます。

便利な点は単純です。ターミナル、設定ファイル、アイコン変換、インストーラースクリプトを行き来する回数を減らします。関連ファイルを書き込み、必要なコマンドを実行し、パッケージ出力をまとめます。

## ダウンロード


1. [Releases](https://github.com/Peiratooo/Wails3-Manager/releases) から OS に合うビルドをダウンロードします。
2. Wails3 Manager を起動し、環境ページを確認します。
3. 新しい Wails 3 プロジェクトを作成するか、既存プロジェクトをインポートします。
4. ワークスペースで設定、ビルド、パッケージ化を行います。

ネイティブパッケージ化には、Windows の **Inno Setup**、macOS の Wails CLI とシステムのディスクイメージ作成・署名ツールが必要です。

DMG は公式の標準レイアウトと、アイコンを移動できるカスタムレイアウトに対応しています。新しい設定では公式モードを使用します。背景、ボリュームアイコン、DMG ファイルアイコンは個別に設定でき、背景を選択しない場合は内蔵のグリッド画像を使います。

アプリケーションバンドルは、パッケージ化の前に一時ディレクトリへコピーされます。新規または変更されたバンドルは ad-hoc 署名され、変更のないビルド済みバンドルは元の署名を保持します。Developer ID 署名と公証は別の配布手順です。既存の DMG は、空でない新しいファイルの生成に成功してから置き換えられます。

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

- Go **1.27.3**
- `github.com/wailsapp/wails/v3 v3.0.0-beta.28` と互換性のある Wails3 CLI
- Node.js **22.12+** と npm
- Go 1.27 でビルドしたアプリには macOS 13 Ventura 以降が必要です
- 利用 OS の Wails プラットフォームツールチェーン

```bash
git clone --branch 1.1.0 https://github.com/Peiratooo/Wails3-Manager.git wails3-manager
cd wails3-manager
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.28
cd frontend
npm ci
cd ..
wails3 task dev
```

便利なコマンド:

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

macOS では `wails3 task darwin:package:universal:dmg` で Intel と Apple Silicon の両方に対応する DMG を作成できます。一般配布では `darwin:package:universal`、`darwin:sign:notarize`、`darwin:create:dmg` の順に実行します。署名は既存のバンドルを使用し、再ビルドしません。ネイティブの単一アーキテクチャのみの場合は `darwin:package:universal` を `darwin:package` に置き換えてください。

## 管理対象プロジェクトのファイル

インポートしたプロジェクトには `builder/` ディレクトリが作成されます。

```text
builder/
├── project.json
├── packaging.json
├── backup/initial/
├── windows/inno.iss
├── macos/dmg-background.png
└── release/
```

## 現在の制限

- Wails 3 は Beta 段階で、API が変更される場合があります。
- ネイティブパッケージは対象 OS 上で作成します。
- Linux パッケージ化は未実装です。
- 管理画面は Developer ID 署名や公証を行いません。上記の CLI 配布手順を使用してください。新規または変更されたバンドルには ad-hoc 署名を行います。
- 自動公開、キャンセル、バージョン付きビルド履歴は未実装です。

## ライセンス

Licensed under the [Apache License 2.0](../LICENSE).
