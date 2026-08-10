# AGENTS.md — инструкции для ассистентов (pdfcreator-wails)

## Что это

Wails v2 приложение — GUI-редактор для генерации PDF-дипломов по шаблону и Excel.

## Запуск

```bash
cd pdfcreator-wails
wails dev
```

Никогда `go run .` — Wails требует build tags.

## Сборка

```bash
# macOS arm64
wails build -platform darwin/arm64

# Windows amd64
wails build -platform windows/amd64
```

Билды появляются в `build/bin/`.

## Структура

```
pdfcreator-wails/
├── app.go                    # Wails bind-методы, projectHome, диалоги
├── internal/
│   ├── config/config.go      # Field, Config, Load/Save
│   ├── excel/                # Чтение .xlsx
│   ├── pdfgen/generator.go   # Генерация PDF, авторазбиение, warnings
│   └── project/              # Проектные директории, шрифты
├── frontend/
│   ├── src/main.js           # Логика редактора
│   ├── src/style.css         # Стили элементов + Tailwind
│   └── wailsjs/              # Сгенерированные bindings (не править руками)
└── build/bin/                # Результат сборки
```

## Важные детали реализации

### config.Field

```go
type Field struct {
    Column      string  `json:"column"`
    X           float64 `json:"x"`
    Y           float64 `json:"y"`
    Font        string  `json:"font"`
    FontSize    int     `json:"fontSize"`
    Color       string  `json:"color"`
    Align       Align   `json:"align"`
    MaxWidth    float64 `json:"maxWidth"`
    AutoWrap    bool    `json:"autoWrap"`
    WrapPattern []int   `json:"-"` // runtime-only
}
```

`WrapPattern` заполняется в `pdfgen.fitFieldsForRows` и используется в `drawText`.

### pdfgen/generator.go

- `fitFieldsForRows` — подбирает `FontSize` для каждого поля.
- `fitAutoWrap` — строит маску разбиения для самой длинной записи и подбирает шрифт.
- `buildWrapPattern` — жадный word-wrap по ширине.
- `applyWrapPattern` — применяет маску к конкретной записи.
- `checkWrapWarnings` — ищет записи, в которых слов меньше, чем строк в маске.
- `GenerateResult` содержит пути, поля и warnings.

### frontend/src/main.js

- `mmToPx(px)` / `pxToMm(px)` — координаты в редакторе.
- `renderFields()` — создаёт `.text-element` с drag- и resize-захватами.
- `startDrag` / `onDragMove` / `onDragEnd` — оптимизированный drag через `transform`.
- `startResize` / `onResizeMove` / `onResizeEnd` — ресайз ширины.
- `selectField` — обновляет selectedIndex, список полей и свойства без пересоздания DOM.
- `doPreview` / `doGenerate` — вызывают Go, обновляют fields, показывают warnings.

### projectHome

В `app.go` функция `detectProjectHome` поднимается от бинарника к корню проекта (`PDFCreator/`). Там располагаются `data/` и `output/`.

### Безопасность WebView

- Шаблон передаётся как base64 data URL (`GetTemplateBase64`).
- PDF и папки открываются через системные команды (`OpenFile`), не через браузер.

## Рекомендации при изменениях

1. **Изменил struct Field?** Пересобери bindings: `wails build`.
2. **Изменил frontend?** Проверь в `wails dev`, затем собери production.
3. **Добавил runtime-поле?** Пометь его `json:"-"`, чтобы не попадало в config.
4. **Warnings?** Возвращай их из `GeneratePreview` / `GenerateImageTemplate` и показывай в `showWarnings`.
5. **Минимальные правки** — не переписывай существующие механизмы без согласования.

## Частые ошибки

- `go run .` → `Wails applications will not build without the correct build tags`.
- Забыть `WrapPattern` runtime → маска сохранится в JSON или потеряется.
- Изменить сигнатуру bind-метода и не пересобрать frontend → JS упадёт.
