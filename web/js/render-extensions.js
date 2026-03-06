// render-extensions.js - KaTeX 수학 렌더링 및 Mermaid 다이어그램 초기화
// SPEC-RENDER-001: 확장 렌더링 모듈

(function() {
    'use strict';

    // 최대 Mermaid 다이어그램 수 (성능 보호)
    var MAX_MERMAID_DIAGRAMS = 50;

    // 대형 수식 타임아웃 (500자 이상)
    var LARGE_FORMULA_TIMEOUT_MS = 5000;
    var LARGE_FORMULA_THRESHOLD = 500;

    /**
     * 수학 수식을 렌더링한다.
     * $...$ 는 인라인 수식, $$...$$ 는 블록 수식으로 처리한다.
     * <code>, <pre> 내부의 달러 기호는 무시한다.
     */
    function renderMath() {
        if (typeof katex === 'undefined') {
            return;
        }

        var content = document.getElementById('content');
        if (!content) return;

        // HTML을 직접 치환하여 수식을 렌더링한다
        var html = content.innerHTML;

        // 이스케이프된 달러 기호를 임시 토큰으로 치환
        var ESCAPED_DOLLAR = '\u0000ESCAPED_DOLLAR\u0000';
        html = html.replace(/\\\$/g, ESCAPED_DOLLAR);

        // <code>, <pre> 블록을 임시로 보호한다
        var codeBlocks = [];
        html = html.replace(/<(pre|code)[^>]*>[\s\S]*?<\/\1>/gi, function(match) {
            codeBlocks.push(match);
            return '\u0000CODEBLOCK_' + (codeBlocks.length - 1) + '\u0000';
        });

        // $$...$$ 블록 수식 (먼저 처리하여 $...$와 충돌 방지)
        html = html.replace(/\$\$([\s\S]+?)\$\$/g, function(match, tex) {
            return renderKaTeX(tex.trim(), true);
        });

        // $...$ 인라인 수식 (통화 $100 등을 피하기 위해 내용이 공백으로 시작하지 않아야 한다)
        html = html.replace(/\$([^\s$](?:[^$]*[^\s$])?)\$/g, function(match, tex) {
            return renderKaTeX(tex.trim(), false);
        });

        // 보호된 코드 블록을 복원한다
        html = html.replace(/\u0000CODEBLOCK_(\d+)\u0000/g, function(match, idx) {
            return codeBlocks[parseInt(idx, 10)];
        });

        // 이스케이프된 달러 기호를 복원한다
        html = html.replace(new RegExp(ESCAPED_DOLLAR.replace(/\u0000/g, '\\u0000'), 'g'), '$');
        // 간단한 문자열 치환으로 복원
        while (html.indexOf(ESCAPED_DOLLAR) !== -1) {
            html = html.replace(ESCAPED_DOLLAR, '$');
        }

        content.innerHTML = html;
    }

    /**
     * KaTeX로 TeX 문자열을 렌더링하여 HTML 문자열을 반환한다.
     * @param {string} tex - TeX 수식 문자열
     * @param {boolean} displayMode - true면 블록 수식, false면 인라인 수식
     * @returns {string} 렌더링된 HTML 문자열
     */
    function renderKaTeX(tex, displayMode) {
        try {
            // 대형 수식 경고 (타임아웃 대신 단순 길이 제한)
            if (tex.length > LARGE_FORMULA_THRESHOLD) {
                // 렌더링 시도하되, 너무 복잡하면 KaTeX 자체가 에러를 발생시킨다
            }

            return katex.renderToString(tex, {
                displayMode: displayMode,
                throwOnError: false,
                errorColor: '#cc0000'
            });
        } catch (e) {
            // 렌더링 실패 시 에러 메시지를 인라인으로 표시
            var tag = displayMode ? 'div' : 'span';
            return '<' + tag + ' class="katex-error" title="' +
                escapeHtml(e.message || 'KaTeX 렌더링 오류') +
                '" style="color:#cc0000;">' +
                escapeHtml(tex) + '</' + tag + '>';
        }
    }

    /**
     * Mermaid 다이어그램을 렌더링한다.
     * <pre><code class="language-mermaid"> 요소를 찾아 Mermaid로 변환한다.
     */
    function renderMermaid() {
        if (typeof mermaid === 'undefined') {
            return;
        }

        var content = document.getElementById('content');
        if (!content) return;

        // language-mermaid 클래스를 가진 코드 블록을 찾는다
        var codeBlocks = content.querySelectorAll('pre code.language-mermaid');
        var count = Math.min(codeBlocks.length, MAX_MERMAID_DIAGRAMS);

        if (count === 0) return;

        // mermaid 초기화
        mermaid.initialize({
            startOnLoad: false,
            theme: 'default',
            securityLevel: 'strict'
        });

        for (var i = 0; i < count; i++) {
            var codeEl = codeBlocks[i];
            var preEl = codeEl.parentElement;
            var diagramText = codeEl.textContent;

            // <pre> 를 <div class="mermaid">로 교체
            var mermaidDiv = document.createElement('div');
            mermaidDiv.className = 'mermaid';
            mermaidDiv.textContent = diagramText;
            preEl.parentNode.replaceChild(mermaidDiv, preEl);
        }

        // Mermaid 렌더링 실행
        try {
            mermaid.run({ querySelector: '.mermaid' }).catch(function(err) {
                // 개별 다이어그램 에러는 Mermaid가 자체 처리한다
                console.warn('Mermaid 렌더링 경고:', err);
            });
        } catch (e) {
            console.warn('Mermaid 초기화 오류:', e);
        }
    }

    /**
     * HTML 특수 문자를 이스케이프한다.
     * @param {string} str - 이스케이프할 문자열
     * @returns {string} 이스케이프된 문자열
     */
    function escapeHtml(str) {
        var div = document.createElement('div');
        div.appendChild(document.createTextNode(str));
        return div.innerHTML;
    }

    /**
     * 확장 렌더링 진입점 - 수학 수식과 다이어그램을 렌더링한다.
     */
    function renderExtensions() {
        renderMath();
        renderMermaid();
    }

    /**
     * KaTeX가 로딩될 때까지 대기한 후 수식을 렌더링한다.
     * defer 스크립트 실행 순서가 보장되지 않는 환경(WebView2 등)을 위한 폴백이다.
     */
    function renderExtensionsWithRetry() {
        if (typeof katex !== 'undefined') {
            renderExtensions();
            return;
        }
        // KaTeX가 아직 로딩되지 않은 경우 재시도 (최대 50회, 100ms 간격 = 5초)
        var retries = 0;
        var maxRetries = 50;
        var timer = setInterval(function() {
            retries++;
            if (typeof katex !== 'undefined' || retries >= maxRetries) {
                clearInterval(timer);
                renderExtensions();
            }
        }, 100);
    }

    // 전역에 노출하여 WebSocket 업데이트 후 호출 가능하게 한다
    window.renderExtensions = renderExtensions;

    // 모든 리소스 로드 완료 후 실행 (defer 스크립트 타이밍 문제 방지)
    if (document.readyState === 'complete') {
        renderExtensionsWithRetry();
    } else {
        window.addEventListener('load', renderExtensionsWithRetry);
    }
})();
