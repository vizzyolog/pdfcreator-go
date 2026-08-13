# План разработки PDFCreatorGo (Wails)

## Описание проекта
Desktop-приложение на Go + Wails для заполнения бланков дипломов (PDF) по данным из Excel. Заменяет существующий Python-скрипт `PDFCreator.py`. Финальный артефакт — один EXE для Windows (`PDFCreatorGo.exe`). Python-файлы в корне проекта остаются для истории, но не участвуют в финальной сборке.

## Текущий статус
- ✅ Проект мигрирован на Wails v2.
- ✅ Реализован GUI-редактор с drag-and-drop.
- ✅ Реализованы загрузка шаблона, Excel, шрифтов.
- ✅ Реализованы предпросмотр и генерация PDF.
- ✅ Собраны macOS (.app) и Windows (.exe) артефакты.

## Целевой стек
- **Backend:** Go 1.26, Wails v2.
- **Desktop shell:** WebView2 (Windows), WebKit (macOS).
- **PDF:** `github.com/go-pdf/fpdf` (картинка-шаблон + текст).
- **Excel:** `github.com/xuri/excelize/v2`.
- **Frontend:** vanilla JS + Tailwind CSS (drag-and-drop и ресайз реализованы вручную).

## Структура проекта
```
PDFCreator/
├── PDFCreator.py          # исторический Python-скрипт
├── pdfcreator.spec        # историческая спецификация PyInstaller
├── fonts/                 # исходные шрифты
├── pdfcreator-wails/      # новый Wails-проект
│   ├── app.go             # Wails App + bind methods
│   ├── main.go            # точка входа
│   ├── wails.json         # конфиг Wails
│   ├── go.mod
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── excel/reader.go
│   │   ├── pdfgen/
│   │   │   ├── generator.go
│   │   │   └── color.go
│   │   └── project/project.go
│   ├── frontend/
│   │   ├── index.html
│   │   ├── src/main.js    # GUI редактор
│   │   ├── src/style.css
│   │   └── ...
│   └── build/bin/
│       ├── pdfcreator-wails.app   # macOS
│       └── pdfcreator-wails.exe   # Windows
├── PDFCreatorGo.exe       # финальный артефакт для Windows (копия)
└── PLAN.md
```

## Чекпоинты

### ✅ 1. Миграция на Wails
- [x] Установлен Wails CLI.
- [x] Инициализирован Wails-проект.
- [x] Перенесены internal-пакеты (config, excel, pdfgen).
- [x] Создан `app.go` с Wails App struct.
- [x] Настроен `wails.json`.
- [x] Проверены `wails dev` и `wails build`.

### ✅ 2. Backend API через Wails bind
- [x] `SelectTemplate(path)` — выбор шаблона.
- [x] `SelectExcel(path)` — выбор Excel.
- [x] `SelectFont(path)` — импорт шрифта.
- [x] `GetFonts()` — список шрифтов.
- [x] `GetColumns()` — список столбцов Excel.
- [x] `GetPreviewRows(limit)` — первые N строк.
- [x] `SaveFields(fields)` / `LoadFields()` — сохранение/загрузка полей.
- [x] `GeneratePreview()` — генерация preview PDF.
- [x] `GenerateAll(outputDir)` — генерация всех PDF.
- [x] `GetTemplateBase64()` — шаблон как data URL для WebView.
- [x] `OpenFileDialog` / `OpenDirectoryDialog` — диалоги выбора файлов.
- [x] `OpenFile(path)` — открытие файла/папки системным приложением.

### ✅ 3. Визуальный редактор обложки
- [x] Отображение обложки (JPG/PNG) на рабочей области через base64.
- [x] Drag-and-drop текстовых элементов по обложке (чистый JS, оптимизировано через `transform` и `requestAnimationFrame`).
- [x] Ресайз ширины текстового блока мышкой за левую/правую границу.
- [x] Панель свойств:
  - [x] Выбор столбца Excel.
  - [x] Шрифт, размер, цвет.
  - [x] Выравнивание (L/C/R) кнопками.
  - [x] Ширина текстового блока.
  - [x] Автоперенос по шаблону превью (вместо ручного разделителя строк).
- [x] Автоподбор размера шрифта, если текст не влезает в область.
- [x] Отображение размеров области текста (W × H).
- [x] Предпросмотр текста — самая длинная запись из Excel по выбранному столбцу.
- [x] Добавление/удаление текстовых элементов.
- [x] Реал-тайм обновление при перемещении/изменении.
- [x] Привязка к реальным размерам A4 landscape (297x210 мм).

### ✅ 4. Предпросмотр и генерация
- [x] Preview PDF открывается в системном PDF-viewer.
- [x] Генерация всех PDF с выбором папки сохранения.
- [x] Автоматическое создание ZIP.
- [x] Открытие папки с результатами.

### ✅ 5. Тестирование и сборка
- [x] `wails dev` работает на macOS.
- [x] `wails build -platform darwin/arm64` собирает `.app`.
- [x] `wails build -platform windows/amd64` собирает `.exe`.
- [x] Финальный `PDFCreatorGo.exe` скопирован в корень проекта.

### ✅ 6. Улучшения по обратной связи
- [x] Оптимизация drag-and-drop (плавное следование за курсором).
- [x] Ресайз ширины поля мышкой.
- [x] Имена сгенерированных PDF формируются по имени и фамилии ученика (issue #1).

## Критерии готовности
1. ✅ Приложение запускается в отдельном окне без адресной строки.
2. ✅ Можно загрузить шаблон (JPG/PNG) и Excel.
3. ✅ Можно мышкой перетаскивать надписи по обложке.
4. ✅ Можно менять шрифт, размер, цвет, выравнивание, многострочность.
5. ✅ Предпросмотр открывает PDF в системном viewer.
6. ✅ Генерация создаёт PDF для каждой строки Excel.
7. ✅ Собирается `PDFCreatorGo.exe` без ошибок.

## Примечания
- Шаблон — только картинка (JPG/PNG). PDF-шаблоны не поддерживаются.
- При первом запуске приложение копирует шрифты из `../fonts/` в `data/fonts/`.
- Wails на Windows требует WebView2 Runtime (обычно уже установлен на Windows 10/11).
- Preview и результаты открываются системным приложением, чтобы избежать ограничений WebView на `file://` URL.

## Как собрать
```bash
cd pdfcreator-wails

# Разработка
wails dev

# macOS
wails build -platform darwin/arm64

# Windows
wails build -platform windows/amd64
cp build/bin/pdfcreator-wails.exe ../PDFCreatorGo.exe
```

## Контекст для продолжения
- Пользователь: Олег, проект в `/Users/olegchernov/quatorium/pdf/PDFCreator`.
- Цель: заменить Python-скрипт PDFCreator.py на desktop-приложение с GUI-редактором.
- Финальный EXE: `PDFCreatorGo.exe` в корне проекта.
- Python-файлы оставляем без изменений, они не используются в новой сборке.
