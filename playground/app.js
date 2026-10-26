(() => {
  'use strict';

  const MONACO_BASE = 'https://cdn.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs';
  const FONT = '"SFMono-Regular", "JetBrains Mono", Menlo, Consolas, monospace';

  const state = {
    target: 'cdk',
    view: 'code',
    monaco: null,
    input: null,
    output: null,
    debounce: 0,
  };

  const $ = (id) => document.getElementById(id);

  const samples = {
    web: () => $('sample-web').textContent.trim(),
    storage: () => $('sample-storage').textContent.trim(),
  };

  /* ---------------- loaders ---------------- */

  function loadMonaco() {
    return new Promise((resolve, reject) => {
      if (typeof window.require !== 'function') {
        reject(new Error('Monaco loader failed to load. Check your network connection.'));
        return;
      }
      window.require.config({ paths: { vs: MONACO_BASE } });
      window.require(['vs/editor/editor.main'], () => resolve(window.monaco), reject);
    });
  }

  async function loadWasm() {
    if (typeof window.Go !== 'function') {
      throw new Error('wasm_exec.js is missing. Run `make wasm` before serving the playground.');
    }
    const go = new window.Go();
    let instance;
    try {
      const streamed = await WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject);
      instance = streamed.instance;
    } catch (_) {
      const response = await fetch('main.wasm');
      const bytes = await response.arrayBuffer();
      const compiled = await WebAssembly.instantiate(bytes, go.importObject);
      instance = compiled.instance;
    }
    go.run(instance);

    const deadline = Date.now() + 30000;
    while (typeof window.openiacTranspile !== 'function') {
      if (Date.now() > deadline) throw new Error('WebAssembly runtime did not start.');
      await new Promise((resolve) => setTimeout(resolve, 25));
    }
    // The converter and the IR toggle read separate exports, so a build with
    // only the first would leave the pane showing the other view's result.
    if (typeof window.openiacIR !== 'function') {
      throw new Error('This WebAssembly build predates the IR view. Run `make wasm` and reload the page.');
    }
  }

  /* ---------------- monaco setup ---------------- */

  function registerHCL(monaco) {
    monaco.languages.register({ id: 'hcl' });
    monaco.languages.setLanguageConfiguration('hcl', {
      comments: { lineComment: '#' },
      brackets: [['{', '}'], ['[', ']'], ['(', ')']],
      autoClosingPairs: [
        { open: '{', close: '}' },
        { open: '[', close: ']' },
        { open: '(', close: ')' },
        { open: '"', close: '"' },
      ],
    });
    monaco.languages.setMonarchTokensProvider('hcl', {
      keywords: [
        'resource', 'data', 'variable', 'output', 'module', 'provider', 'terraform',
        'locals', 'depends_on', 'for_each', 'count', 'lifecycle', 'dynamic', 'content',
        'true', 'false', 'null',
      ],
      tokenizer: {
        root: [
          [/[a-zA-Z_]\w*/, { cases: { '@keywords': 'keyword', '@default': 'identifier' } }],
          { include: '@whitespace' },
          [/"([^"\\]|\\.)*$/, 'string.invalid'],
          [/"/, { token: 'string.quote', bracket: '@open', next: '@string' }],
          [/\d+(\.\d+)?/, 'number'],
          [/[{}()[\]]/, '@brackets'],
          [/[=:,.]/, 'delimiter'],
        ],
        string: [
          [/[^\\"$]+/, 'string'],
          [/\$\{/, { token: 'delimiter.interpolation', next: '@interpolation' }],
          [/\\./, 'string.escape'],
          [/"/, { token: 'string.quote', bracket: '@close', next: '@pop' }],
        ],
        interpolation: [
          [/[^}]+/, 'variable'],
          [/\}/, { token: 'delimiter.interpolation', next: '@pop' }],
        ],
        whitespace: [
          [/[ \t\r\n]+/, 'white'],
          [/#.*$/, 'comment'],
          [/\/\/.*$/, 'comment'],
          [/\/\*/, 'comment', '@comment'],
        ],
        comment: [
          [/[^/*]+/, 'comment'],
          [/\*\//, 'comment', '@pop'],
          [/[/*]/, 'comment'],
        ],
      },
    });
  }

  function defineTheme(monaco) {
    monaco.editor.defineTheme('openiac', {
      base: 'vs-dark',
      inherit: true,
      rules: [
        { token: 'keyword', foreground: '7dd3fc', fontStyle: 'bold' },
        { token: 'identifier', foreground: 'dbeafe' },
        { token: 'string', foreground: 'a7f3d0' },
        { token: 'number', foreground: 'fbbf24' },
        { token: 'comment', foreground: '5b6472', fontStyle: 'italic' },
        { token: 'variable', foreground: 'c4b5fd' },
        { token: 'delimiter', foreground: '94a3b8' },
      ],
      colors: {
        'editor.background': '#0d121c',
        'editor.foreground': '#e6edf7',
        'editor.lineHighlightBackground': '#131a26',
        'editorLineNumber.foreground': '#38414f',
        'editorLineNumber.activeForeground': '#8b98aa',
        'editor.selectionBackground': '#1d4ed84d',
        'editorCursor.foreground': '#22d3ee',
        'editorIndentGuide.background1': '#1a2230',
        'editorIndentGuide.activeBackground1': '#273244',
        'scrollbarSlider.background': '#33415555',
        'scrollbarSlider.hoverBackground': '#47556977',
      },
    });
  }

  function createEditors(monaco) {
    state.monaco = monaco;
    const common = {
      theme: 'openiac',
      automaticLayout: true,
      minimap: { enabled: false },
      fontSize: 13,
      lineHeight: 21,
      fontFamily: FONT,
      scrollBeyondLastLine: false,
      renderLineHighlight: 'all',
      padding: { top: 12, bottom: 12 },
      tabSize: 2,
      smoothScrolling: true,
      cursorSmoothCaretAnimation: 'on',
      scrollbar: { verticalScrollbarSize: 10, horizontalScrollbarSize: 10 },
    };

    state.input = monaco.editor.create($('input-editor'), {
      ...common,
      value: samples.web(),
      language: 'hcl',
      wordWrap: 'on',
    });

    state.output = monaco.editor.create($('output-editor'), {
      ...common,
      value: '// TypeScript output appears here.\n',
      language: 'typescript',
      readOnly: true,
      domReadOnly: true,
      wordWrap: 'off',
    });

    state.input.onDidChangeModelContent(() => {
      clearTimeout(state.debounce);
      state.debounce = setTimeout(convert, 350);
    });

    state.input.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, convert);
  }

  /* ---------------- conversion ---------------- */

  function targetLabel() {
    return state.target === 'cdk' ? 'AWS CDK' : 'Pulumi';
  }

  // Shared by both views so the numbers describe the same graph. Data sources
  // only appear when the input has them, which keeps the common case short.
  function statsLine(result) {
    const parts = [`${result.resources || 0} resources`];
    if (result.data) parts.push(`${result.data} data`);
    parts.push(`${result.variables || 0} vars`, `${result.outputs || 0} outputs`, `${result.edges || 0} edges`);
    return parts.join(' · ');
  }

  // The output pane shows either generated TypeScript or the IR, so the editor
  // language, file name and download name follow the active view.
  function updateOutputLanguage() {
    const model = state.output && state.output.getModel();
    if (!model) return;
    state.monaco.editor.setModelLanguage(model, state.view === 'ir' ? 'json' : 'typescript');
  }

  function setView(view) {
    state.view = view;
    const irButton = $('ir-button');
    irButton.classList.toggle('is-active', view === 'ir');
    irButton.setAttribute('aria-pressed', view === 'ir' ? 'true' : 'false');
    irButton.title = view === 'ir' ? 'Back to the generated TypeScript' : 'Show the universal IR (JSON) both emitters share';
    $('ir-button-label').textContent = view === 'ir' ? 'Hide IR' : 'View IR';
    $('output-file').textContent = view === 'ir' ? 'ir.json' : 'index.ts';
    $('target-badge').textContent = view === 'ir' ? 'IR' : targetLabel();
    updateOutputLanguage();
    convert();
  }

  function convert() {
    if (!state.input) return;

    const source = state.input.getValue();
    if (!source.trim()) {
      state.output.setValue('');
      renderWarnings([]);
      setStatus('ok', 'Waiting for Terraform input.', '');
      return;
    }

    if (state.view === 'ir') {
      convertIR(source);
      return;
    }

    if (typeof window.openiacTranspile !== 'function') return;

    let result;
    try {
      result = JSON.parse(window.openiacTranspile(source, state.target));
    } catch (error) {
      fail(error);
      return;
    }

    if (!result.ok) {
      fail(new Error(result.error || 'Conversion failed.'));
      return;
    }

    const files = result.files || [];
    const file = files.find((entry) => entry.target === state.target) || files[0];
    state.output.setValue(file ? file.content : '');

    const warnings = result.warnings || [];
    renderWarnings(warnings);

    const stats = statsLine(result);
    setStatus('ok', `Converted to ${targetLabel()}.`, `${stats} · ${warnings.length} warnings`);
  }

  function convertIR(source) {
    if (typeof window.openiacIR !== 'function') {
      fail(new Error('This WebAssembly build has no IR export. Run `make wasm` and reload the page.'));
      return;
    }

    let result;
    try {
      result = JSON.parse(window.openiacIR(source));
    } catch (error) {
      fail(error);
      return;
    }
    if (!result.ok) {
      fail(new Error(result.error || 'IR generation failed.'));
      return;
    }

    let schema = '';
    try {
      schema = JSON.parse(result.ir).version || '';
    } catch (_) {
      schema = '';
    }

    state.output.setValue(result.ir || '');
    renderWarnings([]);

    const stats = statsLine(result);
    setStatus('ok', 'Generated the universal IR.', schema ? `${stats} · schema ${schema}` : stats);
  }

  function fail(error) {
    console.error(error);
    const message = String(error.message || error);
    state.output.setValue(state.view === 'ir'
      ? JSON.stringify({ error: message }, null, 2)
      : `// OpenIaC could not convert this input.\n// ${message.replace(/\n/g, '\n// ')}`);
    renderWarnings([]);
    setStatus('error', message, '');
  }

  function setStatus(kind, text, stats) {
    const dot = $('status-dot');
    dot.className = `status-dot is-${kind === 'loading' ? 'loading' : kind === 'error' ? 'error' : 'ok'}`;
    $('status-text').textContent = text;
    $('status-stats').textContent = stats || '';
  }

  function renderWarnings(warnings) {
    const panel = $('warning-panel');
    const list = $('warning-list');
    if (!warnings.length) {
      panel.hidden = true;
      list.replaceChildren();
      return;
    }
    panel.hidden = false;
    $('warning-count').textContent = String(warnings.length);
    list.replaceChildren(...warnings.map((warning) => {
      const item = document.createElement('li');
      item.textContent = warning;
      return item;
    }));
  }

  /* ---------------- controls ---------------- */

  function setupControls() {
    document.querySelectorAll('.segment').forEach((button) => {
      button.addEventListener('click', () => {
        state.target = button.dataset.target;
        document.querySelectorAll('.segment').forEach((other) => {
          other.classList.toggle('is-active', other === button);
        });
        setView('code');
      });
    });

    // The IR stays hidden until this button is pressed, then toggles back to
    // the generated TypeScript.
    $('ir-button').addEventListener('click', () => {
      setView(state.view === 'ir' ? 'code' : 'ir');
    });

    $('run-button').addEventListener('click', convert);

    $('copy-button').addEventListener('click', async () => {
      const text = state.output.getValue();
      if (!text) return;
      try {
        await navigator.clipboard.writeText(text);
        flash('copy-button', 'Copied');
      } catch (_) {
        flash('copy-button', 'Failed');
      }
    });

    $('download-button').addEventListener('click', () => {
      const text = state.output.getValue();
      if (!text) return;
      const ir = state.view === 'ir';
      const blob = new Blob([text], { type: ir ? 'application/json' : 'text/typescript' });
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = ir ? 'openiac.ir.json' : `openiac-${state.target}.ts`;
      anchor.click();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    });

    $('example-select').addEventListener('change', (event) => {
      const sample = samples[event.target.value];
      if (sample && state.input) {
        state.input.setValue(sample());
        convert();
      }
    });
  }

  function flash(id, text) {
    const button = $(id);
    const original = button.textContent;
    button.textContent = text;
    setTimeout(() => { button.textContent = original; }, 1100);
  }

  /* ---------------- boot ---------------- */

  async function boot() {
    try {
      setStatus('loading', 'Loading Monaco and WebAssembly…', '');
      const [monaco] = await Promise.all([loadMonaco(), loadWasm()]);
      registerHCL(monaco);
      defineTheme(monaco);
      createEditors(monaco);
      setupControls();
      setStatus('ok', 'Ready. Edit the Terraform on the left.', '');
      convert();
    } catch (error) {
      console.error(error);
      setStatus('error', String(error.message || error), '');
    }
  }

  document.addEventListener('DOMContentLoaded', boot);
})();
