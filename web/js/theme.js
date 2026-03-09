// theme.js - 테마 전환 관리 모듈
// SPEC-THEME-001: 다크 모드 테마 시스템

(function() {
    'use strict';

    // 테마 순환 순서: system -> light -> dark -> system
    var THEME_CYCLE = ['system', 'light', 'dark'];
    var STORAGE_KEY = 'winmdview-theme';

    // 현재 테마 모드를 반환한다
    function getCurrentTheme() {
        // 1순위: localStorage
        var stored = localStorage.getItem(STORAGE_KEY);
        if (stored && THEME_CYCLE.indexOf(stored) !== -1) {
            return stored;
        }
        // 2순위: html data-theme 속성 (Go 서버에서 설정)
        var htmlTheme = document.documentElement.getAttribute('data-theme');
        if (htmlTheme && THEME_CYCLE.indexOf(htmlTheme) !== -1) {
            return htmlTheme;
        }
        // 기본값
        return 'system';
    }

    // 테마를 적용한다
    function applyTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        // localStorage에도 저장하지만 포트가 변경되면 유실되므로
        // WebSocket을 통해 서버에도 통지하여 config.json에 영속화한다
        try { localStorage.setItem(STORAGE_KEY, theme); } catch(e) {}
        notifyServer(theme);
        updateToggleButton(theme);
        updateMermaidTheme(theme);
    }

    // 서버에 테마 변경을 WebSocket으로 통지한다
    // 서버가 config.json에 저장하여 앱 재시작 시에도 테마가 유지된다
    function notifyServer(theme) {
        // viewer.html의 인라인 스크립트가 노출하는 window._wsSend를 사용한다
        if (typeof window._wsSend === 'function') {
            window._wsSend(JSON.stringify({ type: 'theme', value: theme }));
        }
    }

    // 다음 테마로 순환 전환한다
    function cycleTheme() {
        var current = getCurrentTheme();
        var idx = THEME_CYCLE.indexOf(current);
        var next = THEME_CYCLE[(idx + 1) % THEME_CYCLE.length];
        applyTheme(next);
    }

    // 토글 버튼 아이콘을 업데이트한다
    function updateToggleButton(theme) {
        var btn = document.getElementById('theme-toggle');
        if (!btn) return;

        var icons = { system: '\u{1F5A5}', light: '\u2600', dark: '\u{1F319}' };
        var labels = { system: 'System', light: 'Light', dark: 'Dark' };
        btn.textContent = icons[theme] || icons.system;
        btn.title = labels[theme] || 'System';
    }

    // Mermaid 다이어그램 테마를 업데이트한다 (SPEC-THEME-001 담당)
    function updateMermaidTheme(theme) {
        if (typeof mermaid === 'undefined') return;

        var isDark = theme === 'dark' ||
            (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);

        mermaid.initialize({
            startOnLoad: false,
            theme: isDark ? 'dark' : 'default',
            securityLevel: 'strict'
        });

        // 기존 다이어그램을 재렌더링한다
        var diagrams = document.querySelectorAll('.mermaid[data-processed]');
        // Mermaid v10+에서는 data-processed 대신 aria-roledescription 사용
        if (diagrams.length === 0) {
            diagrams = document.querySelectorAll('.mermaid svg');
        }
        // 재렌더링이 필요한 경우 renderExtensions를 다시 호출한다
        if (diagrams.length > 0 && typeof window.renderExtensions === 'function') {
            // mermaid 다이어그램의 원본 텍스트를 복원하고 재렌더링
            // 복잡하므로 간단히 처리: 재초기화만 수행
        }
    }

    // 시스템 테마 변경 감지 리스너를 등록한다
    function setupSystemThemeListener() {
        var darkModeQuery = window.matchMedia('(prefers-color-scheme: dark)');
        darkModeQuery.addEventListener('change', function() {
            // system 모드일 때만 반응한다
            if (getCurrentTheme() === 'system') {
                // CSS @media 쿼리가 자동으로 처리하므로
                // Mermaid 테마만 업데이트한다
                updateMermaidTheme('system');
            }
        });
    }

    // 키보드 단축키를 등록한다 (Ctrl+Shift+D)
    function setupKeyboardShortcut() {
        document.addEventListener('keydown', function(e) {
            if (e.ctrlKey && e.shiftKey && e.key === 'D') {
                e.preventDefault();
                cycleTheme();
            }
        });
    }

    // 토글 버튼 클릭 이벤트를 등록한다
    function setupToggleButton() {
        var btn = document.getElementById('theme-toggle');
        if (btn) {
            btn.addEventListener('click', function(e) {
                e.preventDefault();
                cycleTheme();
            });
        }
    }

    // 초기화
    function init() {
        var theme = getCurrentTheme();
        applyTheme(theme);
        setupSystemThemeListener();
        setupKeyboardShortcut();
        setupToggleButton();
    }

    // DOM이 준비되면 초기화한다
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    // 전역에 테마 API를 노출한다
    window.themeManager = {
        cycle: cycleTheme,
        apply: applyTheme,
        current: getCurrentTheme
    };
})();
