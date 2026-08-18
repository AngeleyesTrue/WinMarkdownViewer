// Package assets 는 애플리케이션 리소스(아이콘 등)를 go:embed로 포함한다.
package assets

import _ "embed"

// IconData 는 시스템 트레이 및 윈도우 아이콘에 사용되는 ICO 파일 데이터이다.
//
//go:embed icon.ico
var IconData []byte
