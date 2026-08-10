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

## Установка окружения

### macOS

1. **Homebrew** (если ещё не установлен):
   ```bash
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   ```

2. **Go**:
   ```bash
   brew install go
   ```

3. **Node.js**:
   ```bash
   brew install node
   ```

4. **Wails CLI**:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```
   Убедитесь, что `$HOME/go/bin` добавлен в `PATH`.

### Windows

1. **Go**: скачайте и установите с [go.dev/dl](https://go.dev/dl/).
2. **Node.js**: скачайте LTS с [nodejs.org](https://nodejs.org/).
3. **Git for Windows**: скачайте с [git-scm.com](https://git-scm.com/download/win).
4. **Wails CLI** (в PowerShell или cmd):
   ```powershell
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```
   Убедитесь, что `%USERPROFILE%\go\bin` добавлен в `PATH`.
5. **Компилятор C** (обязателен для Wails на Windows):
   - Установите **MSYS2**.
   - В MSYS2 UCRT64 терминале выполните:
     ```bash
     pacman -S mingw-w64-ucrt-x86_64-gcc
     ```
   - Добавьте `C:\msys64\ucrt64\bin` в `PATH`.

## Запуск в режиме разработки

```bash
cd pdfcreator-wails
wails dev
```

> `go run .` не работает для Wails-приложений — используйте только `wails dev` или `wails build`.

## Клонирование и сборка из исходников

```bash
git clone https://github.com/vizzyolog/pdfcreator-go.git
cd pdfcreator-go/pdfcreator-wails
```

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
