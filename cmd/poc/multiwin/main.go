// Package main 은 go-webview2 다중 인스턴스 지원 여부를 검증하는 PoC이다.
//
// 두 개의 WebView2 인스턴스를 별도 고루틴에서 실행하여
// 동일 프로세스 내 다중 윈도우 지원 가능 여부를 확인한다.
//
// 사용법: go run ./cmd/poc/multiwin
//
// 예상 결과:
//   - Path A: 두 윈도우가 모두 표시되고 독립적으로 동작 (고루틴별 윈도우 방식 작동)
//   - Path B: 하나 또는 둘 다 실패 (단일 스레드 접근 필요)
//   - Path C: 프로세스 크래시 (멀티 프로세스 모델 필요)
package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/jchv/go-webview2"
)

// htmlPage 는 PoC 테스트용 간단한 HTML 페이지를 생성한다.
func htmlPage(windowNum int, bgColor string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="display:flex;align-items:center;justify-content:center;height:100vh;margin:0;background:%s;font-family:sans-serif;">
  <div style="text-align:center;">
    <h1>Window %d - PoC Test</h1>
    <p>go-webview2 다중 인스턴스 검증</p>
    <p style="color:#666;">이 윈도우를 닫으면 테스트가 완료됩니다.</p>
  </div>
</body>
</html>`, bgColor, windowNum)
}

func main() {
	fmt.Println("[PoC] go-webview2 다중 인스턴스 테스트 시작")
	fmt.Printf("[PoC] Go version: %s, GOOS: %s, GOARCH: %s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	var wg sync.WaitGroup
	errors := make(chan error, 2)

	// 윈도우 1: 왼쪽 배치, 파란 배경
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()

		fmt.Println("[Window 1] 생성 시도...")
		w := webview2.NewWithOptions(webview2.WebViewOptions{
			Debug:     true,
			AutoFocus: true,
			WindowOptions: webview2.WindowOptions{
				Title:  "Window 1 - PoC Test",
				Width:  600,
				Height: 400,
			},
		})
		if w == nil {
			errors <- fmt.Errorf("[Window 1] WebView2 생성 실패 (nil 반환)")
			return
		}
		defer w.Destroy()

		fmt.Println("[Window 1] 생성 성공, HTML 로딩...")
		w.SetHtml(htmlPage(1, "#e3f2fd"))
		fmt.Println("[Window 1] 이벤트 루프 시작 (Run)")
		w.Run()
		fmt.Println("[Window 1] 종료됨")
	}()

	// 윈도우 2: 오른쪽 배치, 녹색 배경
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()

		fmt.Println("[Window 2] 생성 시도...")
		w := webview2.NewWithOptions(webview2.WebViewOptions{
			Debug:     true,
			AutoFocus: true,
			WindowOptions: webview2.WindowOptions{
				Title:  "Window 2 - PoC Test",
				Width:  600,
				Height: 400,
			},
		})
		if w == nil {
			errors <- fmt.Errorf("[Window 2] WebView2 생성 실패 (nil 반환)")
			return
		}
		defer w.Destroy()

		fmt.Println("[Window 2] 생성 성공, HTML 로딩...")
		w.SetHtml(htmlPage(2, "#e8f5e9"))
		fmt.Println("[Window 2] 이벤트 루프 시작 (Run)")
		w.Run()
		fmt.Println("[Window 2] 종료됨")
	}()

	// 모든 윈도우가 닫힐 때까지 대기
	wg.Wait()
	close(errors)

	// 결과 출력
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	fmt.Println()
	fmt.Println("=== PoC 결과 ===")
	if len(errs) == 0 {
		fmt.Println("[Path A] 성공: 두 윈도우가 모두 정상 동작함")
		fmt.Println("  -> 고루틴별 윈도우 방식 (runtime.LockOSThread + 별도 고루틴) 사용 가능")
	} else {
		fmt.Printf("[Path B] 부분 실패: %d개 윈도우에서 오류 발생\n", len(errs))
		for _, err := range errs {
			fmt.Printf("  -> %v\n", err)
		}
		fmt.Println("  -> 단일 스레드 또는 멀티 프로세스 접근 필요")
		os.Exit(1)
	}
}
