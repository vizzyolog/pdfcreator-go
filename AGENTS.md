# AGENTS.md — инструкции для ассистентов

## Назначение проекта

PDF Creator Go — desktop-приложение для генерации PDF-дипломов/сертификатов по изображению-шаблону и таблице Excel. Позиционируется как замена старому Python-скрипту `PDFCreator.py`.

## Техстек

- **Backend:** Go 1.26.5, Wails v2.13.0
- **Frontend:** vanilla JavaScript, Tailwind CSS v3, Vite 7
- **PDF:** `github.com/go-pdf/fpdf`
- **Excel:** `github.com/xuri/excelize/v2`
- **Целевые платформы:** macOS arm64, Windows amd64

## Структура

```
PDFCreator/
├── README.md              # Человекочитаемая документация
├── PLAN.md                # План работ и статус
├── AGENTS.md              # Этот файл
├── PDFCreatorGo.app       # Релизный билд macOS
├── PDFCreatorGo.exe       # Релизный билд Windows
├── pdfcreator-wails/      # Основной код Wails-приложения
│   ├── app.go             # Bind-методы Wails
│   ├── internal/
│   │   ├── config/        # Модель полей (config.Field)
│   │   ├── excel/         # Чтение .xlsx
│   │   ├── pdfgen/        # Генерация PDF, авторазбиение текста
│   │   └── project/       # Пути проекта, шрифты, output
│   └── frontend/src/main.js
├── fonts/                 # Стартаповые шрифты (копируются/импортируются)
├── output/                # Сгенерированные PDF
└── data/                  # Конфиги и импортированные шрифты
```

## Как запускать и собирать

Разработка:

```bash
cd pdfcreator-wails
wails dev
```

> **Важно:** `go run .` не работает для Wails. Всегда используй `wails dev` или `wails build`.

Сборка:

```bash
# macOS arm64
cd pdfcreator-wails && wails build -platform darwin/arm64
cp -R build/bin/pdfcreator-wails.app ../PDFCreatorGo.app

# Windows amd64
cd pdfcreator-wails && wails build -platform windows/amd64
cp build/bin/pdfcreator-wails.exe ../PDFCreatorGo.exe
```

## Архитектура и ключевые файлы

- `pdfcreator-wails/internal/config/config.go` — модель `Field`. Содержит persisted поля + runtime `WrapPattern []int`.
- `pdfcreator-wails/internal/pdfgen/generator.go` — логика генерации PDF, авторазбиения по маске превью, warnings.
- `pdfcreator-wails/app.go` — Wails bind-методы, определение `projectHome`, открытие файлов.
- `pdfcreator-wails/frontend/src/main.js` — редактор: drag-and-drop, resize, свойства, вызовы Go.
- `pdfcreator-wails/AGENTS.md` — технические детали Wails-приложения.

## Особенности и ограничения

- `file://` URL в WebView на macOS не работают. Шаблон грузится как base64 data URL (`GetTemplateBase64`).
- Preview и папка результатов открываются через системный `open`/`start` (`OpenFile`), не через `BrowserOpenURL`.
- Корневая папка проекта (`projectHome`) определяется из пути бинарника. `output/` и `data/` располагаются в корне проекта, а не рядом с `.app` или `.exe`.
- Drag-and-drop оптимизирован: используется `transform: translate3d(...)`, `requestAnimationFrame`, throttle для input-полей.
- Ресайз ширины поля идёт через прямое изменение `width/left` (layout неизбежен), но свойства обновляются в `requestAnimationFrame`.
- Автоперенос текста строит маску по самой длинной записи столбца и применяет её ко всем записям. Ручной разделитель строк удалён.

## Правила при доработке

- Минимальные изменения. Не переписывай существующую логику без необходимости.
- Следуй стилю существующего кода (Go-форматирование, camelCase/snake_case по языку).
- Все изменения в frontend должны проходить через `wails build` (bind-методы генерируются автоматически).
- После изменений обновляй `README.md`, `PLAN.md` и `AGENTS.md`.
- Не коммить и не пушь в git без явного запроса пользователя.

## Текущий статус

См. `PLAN.md`. Основные фичи реализованы. Приоритетные возможные доработки:

- Тестирование на Windows.
- Улучшение авторазбиения для граничных случаев.
- Экспорт/импорт настроек полей.
- UI-улучшения: сетка, масштабирование, направляющие.
