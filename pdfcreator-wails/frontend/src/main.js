import './style.css';

import {
  SelectTemplate,
  SelectExcel,
  SelectFont,
  GetFonts,
  GetColumns,
  GetPreviewRows,
  SaveFields,
  LoadFields,
  GeneratePreview,
  GenerateAll,
  GetOutputDir,
  GetDataDir,
  GetTemplatePath,
  GetExcelPath,
  GetTemplateBase64,
  GetSampleText,
  OpenFileDialog,
  OpenDirectoryDialog,
  OpenFile,
} from '../wailsjs/go/main/App';

const $ = (sel) => document.querySelector(sel);

const PAGE_W_MM = 297;
const PAGE_H_MM = 210;
let scale = 1;
let fields = [];
let columns = [];
let fonts = [];
let selectedIndex = -1;
let templatePath = '';
let excelPath = '';
let sampleTexts = {};

let dragState = null;
let dragRafId = null;
let resizeState = null;
let resizeRafId = null;

function mmToPx(mm) {
  return mm * scale;
}

function pxToMm(px) {
  return px / scale;
}

async function init() {
  renderLayout();
  bindEvents();
  await loadInitialState();
}

function renderLayout() {
  $('#app').innerHTML = `
    <header class="bg-teal-700 text-white px-4 py-3 flex items-center justify-between shadow">
      <h1 class="text-xl font-bold">PDF Creator Go</h1>
      <div class="text-sm text-teal-100">Desktop-редактор дипломов</div>
    </header>

    <div class="flex-1 flex overflow-hidden">
      <aside class="w-80 bg-white border-r overflow-y-auto p-4 flex flex-col gap-4">
        <section class="border rounded-lg p-3">
          <h2 class="font-semibold mb-2 text-gray-700">📁 Источники</h2>
          <button id="btn-template" class="w-full mb-2 px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded text-left text-sm">🖼️ Шаблон: <span id="lbl-template" class="text-gray-500">не выбран</span></button>
          <button id="btn-excel" class="w-full mb-2 px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded text-left text-sm">📊 Excel: <span id="lbl-excel" class="text-gray-500">не выбран</span></button>
          <button id="btn-font" class="w-full px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded text-left text-sm">🔤 Импорт шрифта</button>
        </section>

        <section class="border rounded-lg p-3">
          <div class="flex justify-between items-center mb-2">
            <h2 class="font-semibold text-gray-700">📝 Поля</h2>
            <button id="btn-add-field" class="text-xs px-2 py-1 bg-teal-600 text-white rounded hover:bg-teal-700">+ Добавить</button>
          </div>
          <div id="fields-list" class="space-y-1 max-h-48 overflow-y-auto"></div>
        </section>

        <section id="properties" class="border rounded-lg p-3 hidden">
          <h2 class="font-semibold mb-2 text-gray-700">⚙️ Свойства</h2>
          <div class="space-y-2 text-sm">
            <div>
              <label class="block text-gray-500">Столбец</label>
              <select id="prop-column" class="w-full border rounded px-2 py-1"></select>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div>
                <label class="block text-gray-500">X (мм)</label>
                <input id="prop-x" type="number" step="0.1" class="w-full border rounded px-2 py-1">
              </div>
              <div>
                <label class="block text-gray-500">Y (мм)</label>
                <input id="prop-y" type="number" step="0.1" class="w-full border rounded px-2 py-1">
              </div>
            </div>
            <div>
              <label class="block text-gray-500">Шрифт</label>
              <select id="prop-font" class="w-full border rounded px-2 py-1"></select>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div>
                <label class="block text-gray-500">Размер</label>
                <input id="prop-size" type="number" class="w-full border rounded px-2 py-1">
              </div>
              <div>
                <label class="block text-gray-500">Цвет</label>
                <input id="prop-color" type="color" class="w-full h-8 border rounded">
              </div>
            </div>
            <div>
              <label class="block text-gray-500">Выравнивание</label>
              <div id="prop-align-group" class="flex border rounded overflow-hidden mt-1">
                <button type="button" data-align="L" class="align-btn flex-1 px-2 py-1 text-xs hover:bg-gray-100 border-r">←</button>
                <button type="button" data-align="C" class="align-btn flex-1 px-2 py-1 text-xs hover:bg-gray-100 border-r">↔</button>
                <button type="button" data-align="R" class="align-btn flex-1 px-2 py-1 text-xs hover:bg-gray-100">→</button>
              </div>
            </div>
            <div>
              <label class="block text-gray-500">Ширина (мм)</label>
              <input id="prop-width" type="number" step="0.1" class="w-full border rounded px-2 py-1">
            </div>
            <div class="flex items-center gap-2">
              <input id="prop-autowrap" type="checkbox" class="w-4 h-4">
              <label for="prop-autowrap" class="text-sm text-gray-700">Автоперенос по шаблону превью</label>
            </div>
            <div class="text-xs text-gray-500 bg-gray-50 p-2 rounded">
              Область текста: <span id="prop-area-w">0</span> × <span id="prop-area-h">0</span> мм
            </div>
            <button id="btn-delete-field" class="w-full px-3 py-2 bg-red-100 text-red-700 rounded hover:bg-red-200">Удалить поле</button>
          </div>
        </section>
      </aside>

      <main class="flex-1 flex flex-col bg-gray-50 overflow-hidden">
        <div class="flex-1 flex items-center justify-center overflow-auto p-6" id="canvas-wrapper">
          <div id="canvas" class="editor-canvas bg-white"></div>
        </div>

        <footer class="bg-white border-t px-4 py-3 flex items-center justify-between">
          <div class="text-sm text-gray-500">
            Масштаб: <span id="scale-info">100%</span> | Кликните по полю, чтобы выбрать. Перетаскивайте мышкой.
          </div>
          <div class="flex gap-3">
            <button id="btn-preview" class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50">👁 Предпросмотр</button>
            <button id="btn-generate" class="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50">🚀 Генерировать все</button>
          </div>
        </footer>
      </main>
    </div>
  `;
}

function bindEvents() {
  $('#btn-template').addEventListener('click', chooseTemplate);
  $('#btn-excel').addEventListener('click', chooseExcel);
  $('#btn-font').addEventListener('click', chooseFont);
  $('#btn-add-field').addEventListener('click', addField);
  $('#btn-delete-field').addEventListener('click', deleteSelectedField);
  $('#btn-preview').addEventListener('click', doPreview);
  $('#btn-generate').addEventListener('click', doGenerate);

  $('#prop-column').addEventListener('change', updateSelectedField);
  $('#prop-x').addEventListener('input', updateSelectedField);
  $('#prop-y').addEventListener('input', updateSelectedField);
  $('#prop-font').addEventListener('change', updateSelectedField);
  $('#prop-size').addEventListener('input', updateSelectedField);
  $('#prop-color').addEventListener('input', updateSelectedField);
  $('#prop-width').addEventListener('input', updateSelectedField);
  $('#prop-autowrap').addEventListener('change', updateSelectedField);

  $('#prop-align-group').addEventListener('click', (e) => {
    const btn = e.target.closest('.align-btn');
    if (!btn) return;
    document.querySelectorAll('#prop-align-group .align-btn').forEach(b => {
      b.classList.remove('bg-teal-600', 'text-white');
      b.classList.add('hover:bg-gray-100');
    });
    btn.classList.add('bg-teal-600', 'text-white');
    btn.classList.remove('hover:bg-gray-100');
    updateSelectedField();
  });

  window.addEventListener('resize', async () => {
    await setupCanvas();
    renderFields();
  });
}

async function loadInitialState() {
  templatePath = await GetTemplatePath();
  excelPath = await GetExcelPath();
  fonts = await GetFonts();
  columns = await GetColumns();
  fields = await LoadFields();
  if (!Array.isArray(fields)) fields = [];

  await loadSampleTexts();
  updateSourceLabels();
  await setupCanvas();
  renderFieldsList();
  renderProperties();
  renderFields();
  updateButtons();
}

async function loadSampleTexts() {
  sampleTexts = {};
  if (!excelPath) return;
  for (const col of columns) {
    try {
      sampleTexts[col] = await GetSampleText(col);
    } catch (err) {
      sampleTexts[col] = col;
    }
  }
}

function updateSourceLabels() {
  $('#lbl-template').textContent = templatePath ? basename(templatePath) : 'не выбран';
  $('#lbl-excel').textContent = excelPath ? basename(excelPath) : 'не выбран';
}

function basename(path) {
  return path.split(/[\\/]/).pop();
}

async function chooseTemplate() {
  try {
    const path = await OpenFileDialog({
      Title: 'Выберите шаблон',
      Filters: [{DisplayName: 'Изображения', Pattern: '*.jpg;*.jpeg;*.png'}],
    });
    if (!path) return;
    await SelectTemplate(path);
    templatePath = path;
    updateSourceLabels();
    await setupCanvas();
    updateButtons();
    showToast('Шаблон выбран');
  } catch (err) {
    showToast('Ошибка: ' + err, true);
  }
}

async function chooseExcel() {
  try {
    const path = await OpenFileDialog({
      Title: 'Выберите Excel',
      Filters: [{DisplayName: 'Excel', Pattern: '*.xlsx;*.xls'}],
    });
    if (!path) return;
    await SelectExcel(path);
    excelPath = path;
    columns = await GetColumns();
    await loadSampleTexts();
    updateSourceLabels();
    if (fields.length === 0 && columns.length > 0) {
      addField();
    }
    renderFieldsList();
    renderProperties();
    renderFields();
    updateButtons();
    showToast('Excel выбран');
  } catch (err) {
    showToast('Ошибка: ' + err, true);
  }
}

async function chooseFont() {
  try {
    const path = await OpenFileDialog({
      Title: 'Выберите шрифт',
      Filters: [{DisplayName: 'Шрифты', Pattern: '*.ttf;*.otf'}],
    });
    if (!path) return;
    await SelectFont(path);
    fonts = await GetFonts();
    renderProperties();
    renderFields();
    showToast('Шрифт импортирован');
  } catch (err) {
    showToast('Ошибка: ' + err, true);
  }
}

async function setupCanvas() {
  const wrapper = $('#canvas-wrapper');
  const canvas = $('#canvas');
  if (!wrapper || !canvas) return;

  const availableW = wrapper.clientWidth - 48;
  const availableH = wrapper.clientHeight - 48;

  scale = Math.min(availableW / PAGE_W_MM, availableH / PAGE_H_MM);
  $('#scale-info').textContent = Math.round(scale * 100) + '%';

  canvas.style.width = mmToPx(PAGE_W_MM) + 'px';
  canvas.style.height = mmToPx(PAGE_H_MM) + 'px';

  if (templatePath) {
    try {
      const dataUrl = await GetTemplateBase64();
      canvas.style.backgroundImage = `url("${dataUrl}")`;
    } catch (err) {
      console.error('Failed to load template:', err);
      canvas.style.backgroundImage = '';
    }
  } else {
    canvas.style.backgroundImage = '';
  }
}

function addField() {
  if (columns.length === 0) {
    showToast('Сначала выберите Excel', true);
    return;
  }
  fields.push({
    column: columns[0],
    x: 95,
    y: 87,
    font: fonts[0] || 'Gilroy-Bold.ttf',
    fontSize: 38,
    color: '#007998',
    align: 'C',
    maxWidth: 100,
    autoWrap: false,
  });
  selectedIndex = fields.length - 1;
  saveFields();
  renderFieldsList();
  renderProperties();
  renderFields();
}

function deleteSelectedField() {
  if (selectedIndex < 0) return;
  fields.splice(selectedIndex, 1);
  selectedIndex = -1;
  saveFields();
  renderFieldsList();
  renderProperties();
  renderFields();
}

function renderFieldsList() {
  const list = $('#fields-list');
  list.innerHTML = '';
  fields.forEach((f, i) => {
    const div = document.createElement('div');
    div.className = `px-2 py-1 rounded text-sm cursor-pointer ${i === selectedIndex ? 'bg-teal-100 text-teal-800' : 'hover:bg-gray-100'}`;
    div.textContent = `${i + 1}. ${f.column || 'Без имени'}`;
    div.addEventListener('click', () => {
      selectedIndex = i;
      renderFieldsList();
      renderProperties();
      renderFields();
    });
    list.appendChild(div);
  });
}

function renderProperties() {
  const props = $('#properties');
  const colSel = $('#prop-column');
  const fontSel = $('#prop-font');

  colSel.innerHTML = columns.map(c => `<option value="${c}">${c}</option>`).join('');
  fontSel.innerHTML = fonts.map(f => `<option value="${f}">${f}</option>`).join('');

  if (selectedIndex < 0) {
    props.classList.add('hidden');
    return;
  }

  props.classList.remove('hidden');
  const f = fields[selectedIndex];
  colSel.value = f.column || '';
  $('#prop-x').value = round(f.x ?? 95);
  $('#prop-y').value = round(f.y ?? 87);
  fontSel.value = f.font || '';
  $('#prop-size').value = f.fontSize ?? 38;
  $('#prop-color').value = f.color || '#000000';
  setActiveAlignButton(f.align || 'L');
  $('#prop-width').value = round(f.maxWidth ?? 100);
  $('#prop-autowrap').checked = f.autoWrap || false;

  updateAreaInfo(f);
}

function updateAreaInfo(f) {
  // Estimate area using canvas text metrics scaled to mm
  const el = document.createElement('div');
  el.style.position = 'absolute';
  el.style.visibility = 'hidden';
  el.style.fontFamily = 'sans-serif';
  el.style.fontSize = (f.fontSize * scale * 0.35) + 'px';
  el.style.whiteSpace = 'pre-wrap';
  el.style.width = mmToPx(f.maxWidth || 100) + 'px';
  el.textContent = getDisplayText(f);
  document.body.appendChild(el);
  const w = pxToMm(el.scrollWidth);
  const h = pxToMm(el.scrollHeight);
  document.body.removeChild(el);

  $('#prop-area-w').textContent = round(w);
  $('#prop-area-h').textContent = round(h);
}

function getActiveAlign() {
  const active = document.querySelector('#prop-align-group .align-btn.bg-teal-600');
  return active ? active.dataset.align : 'L';
}

function setActiveAlignButton(align) {
  document.querySelectorAll('#prop-align-group .align-btn').forEach(btn => {
    const isActive = btn.dataset.align === align;
    if (isActive) {
      btn.classList.add('bg-teal-600', 'text-white');
      btn.classList.remove('hover:bg-gray-100');
    } else {
      btn.classList.remove('bg-teal-600', 'text-white');
      btn.classList.add('hover:bg-gray-100');
    }
  });
}

function updateSelectedField() {
  if (selectedIndex < 0) return;
  const f = fields[selectedIndex];
  f.column = $('#prop-column').value;
  f.x = parseFloat($('#prop-x').value);
  f.y = parseFloat($('#prop-y').value);
  f.font = $('#prop-font').value;
  f.fontSize = parseInt($('#prop-size').value);
  f.color = $('#prop-color').value;
  f.align = getActiveAlign();
  f.maxWidth = parseFloat($('#prop-width').value);
  f.autoWrap = $('#prop-autowrap').checked;

  updateAreaInfo(f);
  renderFieldsList();
  renderFields();
  saveFields();
}

function saveFields() {
  SaveFields(fields).catch(err => console.error(err));
}

function getDisplayText(field) {
  let text = sampleTexts[field.column] || field.column || 'Текст';
  if (field.autoWrap) {
    // Show a rough multi-line preview in the editor; exact mask is computed at generation time.
    text = text.split(/\s+/).join('\n');
  }
  return text;
}

function renderFields() {
  const canvas = $('#canvas');
  canvas.querySelectorAll('.text-element').forEach(el => el.remove());

  fields.forEach((f, i) => {
    const el = document.createElement('div');
    el.className = `text-element ${i === selectedIndex ? 'selected' : ''}`;
    el.dataset.index = i;
    el.style.left = mmToPx(f.x) + 'px';
    el.style.top = mmToPx(f.y) + 'px';
    el.style.width = mmToPx(f.maxWidth || 100) + 'px';
    el.style.fontFamily = 'sans-serif';
    el.style.fontSize = (f.fontSize * scale * 0.35) + 'px';
    el.style.color = f.color || '#000000';
    el.style.textAlign = f.align === 'C' ? 'center' : (f.align === 'R' ? 'right' : 'left');
    el.style.whiteSpace = 'pre-wrap';
    el.style.wordBreak = 'break-word';
    el.textContent = getDisplayText(f);

    el.addEventListener('mousedown', startDrag);

    const handleLeft = document.createElement('div');
    handleLeft.className = 'resize-handle left';
    handleLeft.dataset.index = i;
    handleLeft.addEventListener('mousedown', startResize);

    const handleRight = document.createElement('div');
    handleRight.className = 'resize-handle right';
    handleRight.dataset.index = i;
    handleRight.addEventListener('mousedown', startResize);

    el.appendChild(handleLeft);
    el.appendChild(handleRight);
    canvas.appendChild(el);
  });
}

function selectField(index) {
  selectedIndex = index;
  document.querySelectorAll('.text-element').forEach(el => el.classList.remove('selected'));
  const active = document.querySelector(`.text-element[data-index="${index}"]`);
  if (active) active.classList.add('selected');
  renderFieldsList();
  renderProperties();
}

function startDrag(e) {
  e.preventDefault();
  e.stopPropagation();

  const el = e.currentTarget;
  const index = parseInt(el.dataset.index);
  if (index !== selectedIndex) {
    selectField(index);
  }

  const startX = e.clientX;
  const startY = e.clientY;
  const startLeft = parseFloat(el.style.left) || 0;
  const startTop = parseFloat(el.style.top) || 0;

  dragState = { el, index, startX, startY, startLeft, startTop, dx: 0, dy: 0 };
  el.style.cursor = 'grabbing';

  document.addEventListener('mousemove', onDragMove);
  document.addEventListener('mouseup', onDragEnd);
}

function onDragMove(e) {
  if (!dragState) return;
  const { el, startX, startY } = dragState;

  const dx = e.clientX - startX;
  const dy = e.clientY - startY;
  dragState.dx = dx;
  dragState.dy = dy;

  el.style.transform = `translate3d(${dx}px, ${dy}px, 0)`;

  if (!dragRafId) {
    dragRafId = requestAnimationFrame(updateDragProps);
  }
}

function updateDragProps() {
  dragRafId = null;
  if (!dragState) return;
  const { index, startLeft, startTop, dx, dy } = dragState;
  const f = fields[index];
  f.x = pxToMm(startLeft + dx);
  f.y = pxToMm(startTop + dy);

  if (index === selectedIndex) {
    $('#prop-x').value = round(f.x);
    $('#prop-y').value = round(f.y);
  }
}

function onDragEnd() {
  if (!dragState) return;
  const { el, index, startLeft, startTop, dx, dy } = dragState;

  const newLeft = startLeft + dx;
  const newTop = startTop + dy;

  el.style.transform = '';
  el.style.left = newLeft + 'px';
  el.style.top = newTop + 'px';
  el.style.cursor = '';

  const f = fields[index];
  f.x = pxToMm(newLeft);
  f.y = pxToMm(newTop);

  if (index === selectedIndex) {
    $('#prop-x').value = round(f.x);
    $('#prop-y').value = round(f.y);
  }

  dragState = null;
  if (dragRafId) {
    cancelAnimationFrame(dragRafId);
    dragRafId = null;
  }
  document.removeEventListener('mousemove', onDragMove);
  document.removeEventListener('mouseup', onDragEnd);
  saveFields();
}

function startResize(e) {
  e.preventDefault();
  e.stopPropagation();

  const handle = e.currentTarget;
  const el = handle.parentElement;
  const index = parseInt(handle.dataset.index);
  if (index !== selectedIndex) {
    selectField(index);
  }

  const side = handle.classList.contains('left') ? 'left' : 'right';
  const startX = e.clientX;
  const startLeft = parseFloat(el.style.left) || 0;
  const startWidth = parseFloat(el.style.width) || el.offsetWidth;

  resizeState = { el, index, side, startX, startLeft, startWidth };

  document.addEventListener('mousemove', onResizeMove);
  document.addEventListener('mouseup', onResizeEnd);
}

function onResizeMove(e) {
  if (!resizeState) return;
  const { el } = resizeState;

  const dx = e.clientX - resizeState.startX;
  let newLeft = resizeState.startLeft;
  let newWidth = resizeState.startWidth;

  if (resizeState.side === 'right') {
    newWidth = Math.max(20, resizeState.startWidth + dx);
  } else {
    newWidth = Math.max(20, resizeState.startWidth - dx);
    newLeft = resizeState.startLeft + (resizeState.startWidth - newWidth);
  }

  el.style.left = newLeft + 'px';
  el.style.width = newWidth + 'px';

  if (!resizeRafId) {
    resizeRafId = requestAnimationFrame(updateResizeProps);
  }
}

function updateResizeProps() {
  resizeRafId = null;
  if (!resizeState) return;
  const f = fields[resizeState.index];
  f.x = pxToMm(parseFloat(resizeState.el.style.left) || 0);
  f.maxWidth = pxToMm(parseFloat(resizeState.el.style.width) || resizeState.el.offsetWidth);

  if (resizeState.index === selectedIndex) {
    $('#prop-x').value = round(f.x);
    $('#prop-width').value = round(f.maxWidth);
    updateAreaInfo(f);
  }
}

function onResizeEnd() {
  if (!resizeState) return;
  updateResizeProps();

  resizeState = null;
  if (resizeRafId) {
    cancelAnimationFrame(resizeRafId);
    resizeRafId = null;
  }
  document.removeEventListener('mousemove', onResizeMove);
  document.removeEventListener('mouseup', onResizeEnd);
  saveFields();
}

function updateButtons() {
  const ready = templatePath && excelPath;
  $('#btn-preview').disabled = !ready;
  $('#btn-generate').disabled = !ready;
}

async function doPreview() {
  if (!templatePath || !excelPath) {
    showToast('Сначала выберите шаблон и Excel', true);
    return;
  }
  $('#btn-preview').disabled = true;
  $('#btn-preview').textContent = 'Генерация...';

  try {
    const result = await GeneratePreview();
    fields = result.fields;
    await SaveFields(fields);
    renderFieldsList();
    renderProperties();
    renderFields();
    showWarnings(result.warnings);
    await OpenFile(result.path);
    showToast('Предпросмотр открыт');
  } catch (err) {
    showToast('Ошибка: ' + err, true);
  } finally {
    $('#btn-preview').disabled = false;
    $('#btn-preview').textContent = '👁 Предпросмотр';
    updateButtons();
  }
}

async function doGenerate() {
  if (!templatePath || !excelPath) {
    showToast('Сначала выберите шаблон и Excel', true);
    return;
  }
  $('#btn-generate').disabled = true;
  $('#btn-generate').textContent = 'Генерация...';

  try {
    const outputDir = await OpenDirectoryDialog({
      Title: 'Выберите папку для сохранения',
      DefaultDirectory: await GetOutputDir(),
    });
    if (!outputDir) return;

    const result = await GenerateAll(outputDir);
    fields = result.fields;
    await SaveFields(fields);
    renderFieldsList();
    renderProperties();
    renderFields();
    showWarnings(result.warnings);
    showToast(`Сгенерировано ${result.count} PDF`);
    await OpenFile(outputDir);
  } catch (err) {
    showToast('Ошибка: ' + err, true);
  } finally {
    $('#btn-generate').disabled = false;
    $('#btn-generate').textContent = '🚀 Генерировать все';
    updateButtons();
  }
}

function round(val) {
  return Math.round(val * 10) / 10;
}

function showToast(message, isError = false) {
  const toast = document.createElement('div');
  toast.className = `fixed bottom-4 right-4 px-4 py-2 rounded shadow text-white ${isError ? 'bg-red-600' : 'bg-teal-600'}`;
  toast.textContent = message;
  document.body.appendChild(toast);
  setTimeout(() => toast.remove(), 3000);
}

function showWarnings(warnings) {
  if (!warnings || warnings.length === 0) return;
  const list = warnings.slice(0, 5).join('\n');
  const more = warnings.length > 5 ? `\n…и ещё ${warnings.length - 5}` : '';
  showToast('Предупреждения:\n' + list + more, true);
}

init();
