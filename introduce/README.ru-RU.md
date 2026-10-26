<div align="center">

<img src="../imgs/logo.png" alt="Wails3 Manager" width="72">

# Wails3 Manager

Десктопное приложение для создания, настройки, сборки и упаковки приложений Wails 3.

<img src="../imgs/badge.png" alt="Значки Wails3 Manager" width="512">

[English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md) · [Français](README.fr-FR.md) · [Deutsch](README.de-DE.md) · **Русский**

</div>

> [!IMPORTANT]
> Wails3 Manager рассчитан на Wails 3 beta.28. Установщики Windows собираются в Windows, а образы DMG для macOS — в macOS.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../imgs/english/dark/home.png">
  <source media="(prefers-color-scheme: light)" srcset="../imgs/english/light/home.png">
  <img alt="Рабочая область Wails3 Manager" src="../imgs/english/dark/home.png">
</picture>

## Для чего нужен Wails3 Manager

Wails3 Manager объединяет типовые операции с проектами Wails в одном окне. Можно создавать и импортировать проекты, редактировать сведения о приложении и значки, настраивать сборку, создавать установочные пакеты для операционной системы, проверять инструменты разработки и просматривать журналы сборки.

Больше не нужно постоянно переключаться между терминалом, конфигурационными файлами, редакторами значков и скриптами установщиков. Менеджер изменяет необходимые файлы проекта, выполняет команды и собирает результаты упаковки в одном месте.

## Загрузка

1. Скачайте сборку для своей операционной системы.
2. Запустите Wails3 Manager и проверьте страницу системного окружения.
3. Создайте новый проект Wails 3 или импортируйте существующий.
4. Настройте, соберите и упакуйте приложение в рабочей области.

Для создания нативных пакетов необходимы **Inno Setup** в Windows, а в macOS — Wails CLI и системные инструменты создания образов дисков и подписи.

Для DMG доступны официальное стандартное оформление и своё оформление с перемещаемыми значками. Новые конфигурации используют официальный режим. Фон и значки тома и файла DMG настраиваются отдельно; без выбранного фона используется встроенное изображение с сеткой.

Перед упаковкой приложение копируется во временный каталог. Новые и изменённые приложения получают подпись ad-hoc, а неизменённые готовые приложения сохраняют исходную подпись. Подпись Developer ID и нотариальное заверение выполняются отдельно. Предыдущий DMG заменяется только после создания непустого нового файла.

## Возможности

Основные функции:

- Создание проектов из официальных, локальных или удалённых шаблонов Wails с последующим открытием в менеджере.
- Импорт существующих проектов Wails 3 и ведение списка последних проектов.
- Редактирование метаданных приложения и замена значка; PNG преобразуются в ресурсы Wails для целевых платформ.
- Настройка имени сборки, производственного режима, CGO, запускаемого файла и включаемых в пакет ресурсов.
- Создание установщиков Windows с помощью Inno Setup и пакетов DMG для macOS в соответствующей ОС.
- Проверка Go, Node.js, npm, Wails3 CLI, Git и инструментов упаковки перед сборкой.
- Отображение журналов сборки в реальном времени и сохранение результатов в `builder/release`.

## Скриншоты

<table>
<tr>
<td width="50%" valign="top"><strong>Создание проекта</strong><br><br><img src="../imgs/english/light/creator.png" alt="Создание проекта Wails 3"></td>
<td width="50%" valign="top"><strong>Метаданные проекта</strong><br><br><img src="../imgs/english/light/wails3-config.png" alt="Редактирование метаданных проекта Wails"></td>
</tr>
<tr>
<td width="50%" valign="top"><strong>Сборка и упаковка</strong><br><br><img src="../imgs/english/dark/package-config.png" alt="Настройка сборок и установщиков"></td>
<td width="50%" valign="top"><strong>Проверка окружения</strong><br><br><img src="../imgs/english/dark/enviroment.png" alt="Отчёт о системном окружении"></td>
</tr>
</table>

## Разработка

Технологический стек: Wails 3, Go, Vue 3, Vite, Naive UI, Pinia, Vue Router, Vue I18n, vue3-moveable.

Требования:

- Go **1.27.3** (версия указана в go.mod)
- Wails3 CLI **v3.0.0-beta.28**, соответствующий модулю Go и среде выполнения фронтенда
- Node.js и npm
- Инструментарий Wails для вашей ОС

```bash
git clone <repository-url>
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.28
cd wails3-manager/frontend
npm install
cd ..
wails3 task dev
```

Полезные команды:

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

## Файлы под управлением менеджера

В импортированных проектах менеджер создаёт и обслуживает каталог `builder/`:

```text
builder/
├── project.json
├── packaging.json
├── backup/initial/
├── windows/inno.iss
├── macos/dmg.sh
└── release/
```

## Текущие ограничения

- Wails 3 пока находится на стадии бета-разработки.
- Нативные пакеты необходимо создавать в целевой операционной системе.
- Упаковка для Linux пока не реализована.
- Подпись кода, нотариальное заверение, автоматическая публикация, отмена операций и история сборок с версиями пока не реализованы.

## Лицензия

Проект распространяется на условиях [Apache License 2.0](../LICENSE).

---

*Перевёл: [Xenon007](https://github.com/xenon007).*
