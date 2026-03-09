package assets

import (
	"encoding/binary"
	"testing"
)

// TestIconDataNotEmpty 임베디드 아이콘 데이터가 비어있지 않은지 검증한다.
func TestIconDataNotEmpty(t *testing.T) {
	if len(IconData) == 0 {
		t.Fatal("IconData가 비어있다: go:embed가 실패했을 수 있다")
	}
}

// TestIconDataIsValidICO ICO 파일 헤더가 유효한지 검증한다.
func TestIconDataIsValidICO(t *testing.T) {
	if len(IconData) < 6 {
		t.Fatal("IconData가 너무 작다: 최소 6바이트 (ICONDIR 헤더) 필요")
	}

	// ICONDIR 헤더 검증
	reserved := binary.LittleEndian.Uint16(IconData[0:2])
	if reserved != 0 {
		t.Errorf("ICO Reserved 필드: want 0, got %d", reserved)
	}

	imageType := binary.LittleEndian.Uint16(IconData[2:4])
	if imageType != 1 {
		t.Errorf("ICO Type 필드: want 1 (icon), got %d", imageType)
	}

	count := binary.LittleEndian.Uint16(IconData[4:6])
	if count < 1 {
		t.Errorf("ICO Count 필드: want >= 1, got %d", count)
	}
}

// TestIconDataMinimumSize ICO 파일이 최소 크기 이상인지 검증한다.
// ICONDIR(6) + ICONDIRENTRY(16) + BITMAPINFOHEADER(40) = 최소 62바이트
func TestIconDataMinimumSize(t *testing.T) {
	minSize := 62
	if len(IconData) < minSize {
		t.Errorf("IconData 크기가 너무 작다: want >= %d, got %d", minSize, len(IconData))
	}
}
