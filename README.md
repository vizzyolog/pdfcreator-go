# PDF Creator Go

Desktop-приложение для генерации PDF-дипломов и сертификатов по шаблону и таблице Excel.

## Технологии

- **Backend:** Go 1.26, Wails v2
- **Frontend:** Vanilla JS, Tailwind CSS v3, Vite 7
- **PDF:** `github.com/go-pdf/fpdf`
- **Excel:** `github.com/xuri/excelize/v2`

## Структура проекта

```
PDFCreator/
├── pdfcreator-wails/          # Wails v2 приложение
│   ├── app.go                 # Wails bind-методы
│   ├── internal/
│   │   ├── config/            # Модель полей и конфигурация
│   │   ├── excel/             # Чтение Excel
│   │   ├── pdfgen/            # Генерация PDF и авторазбиение текста
│   │   └── project/           # Проектные пути и директории
│   ├── frontend/              # UI (Vite + Tailwind)
│   │   └── src/main.js        # Логика редактора
│   └── build/bin/             # Собранные бинарники
├── PDFCreatorGo.app           # Готовый билд для macOS (arm64)
└── PDFCreatorGo.exe           # Готовый билд для Windows (amd64)
```

## Запуск в режиме разработки

```bash
cd pdfcreator-wails
wails dev
```

> `go run .` не работает для Wails-приложений — используйте только `wails dev` или `wails build`.

## Сборка релизов

macOS (arm64):

```bash
cd pdfcreator-wails
wails build -platform darwin/arm64
cp build/bin/pdfcreator-wails.app ../PDFCreatorGo.app
```

Windows (amd64):

```bash
cd pdfcreator-wails
wails build -platform windows/amd64
cp build/bin/pdfcreator-wails.exe ../PDFCreatorGo.exe
```

## Функциональность

- Выбор шаблона-изображения (JPG/PNG).
- Выбор таблицы Excel (`.xlsx`).
- Импорт TTF/OTF шрифтов.
- Визуальный редактор текстовых полей на превью шаблона.
- **Drag-and-drop** полей с оптимизацией через `requestAnimationFrame` и `transform`.
- **Изменение ширины поля** мышкой за левую/правую границу.
- Настройки поля: столбец Excel, координаты, шрифт, размер, цвет, выравнивание, ширина.
- **Автоперенос по шаблону превью:** если самая длинная запись столбца не влезает, программа строит маску разбиения и применяет её ко всем записям.
- Предупреждения, если запись не подходит под маску автопереноса.
- Предпросмотр одного PDF и генерация всех PDF с упаковкой в ZIP.

## Данные и выходные файлы

- Папка `output/` — сгенерированные PDF и ZIP.
- Папка `data/fonts/` — импортированные шрифты.
- Файл `data/fields.json` — сохранённые настройки полей.
