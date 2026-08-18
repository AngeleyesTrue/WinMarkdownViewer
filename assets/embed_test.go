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

// TestIconDataMultiSize 멀티사이즈 ICO 파일이 6개 이미지 엔트리를 포함하는지 검증한다.
// 기대 크기: 16x16, 32x32, 48x48, 64x64, 128x128, 256x256
func TestIconDataMultiSize(t *testing.T) {
	if len(IconData) < 6 {
		t.Fatal("IconData가 너무 작다: ICONDIR 헤더를 읽을 수 없다")
	}

	const expectedCount = 6
	count := int(binary.LittleEndian.Uint16(IconData[4:6]))
	if count != expectedCount {
		t.Errorf("ICO 이미지 엔트리 수: want %d, got %d", expectedCount, count)
	}
}

// TestIconDataMultiSizeReasonableFileSize 멀티사이즈 ICO 파일이 충분한 크기인지 검증한다.
// 6개 사이즈를 포함하는 ICO는 최소 10KB 이상이어야 한다.
func TestIconDataMultiSizeReasonableFileSize(t *testing.T) {
	const minBytes = 10 * 1024 // 10KB
	if len(IconData) < minBytes {
		t.Errorf("멀티사이즈 ICO 파일 크기가 너무 작다: want >= %d bytes, got %d bytes",
			minBytes, len(IconData))
	}
}

// TestIconDataICONDIRENTRY 각 ICONDIRENTRY가 유효한 크기 정보를 가지는지 검증한다.
// ICONDIRENTRY 구조체: Width(1) + Height(1) + ColorCount(1) + Reserved(1)
//
//	+ Planes(2) + BitCount(2) + SizeInBytes(4) + Offset(4) = 16바이트
func TestIconDataICONDIRENTRY(t *testing.T) {
	if len(IconData) < 6 {
		t.Fatal("IconData가 너무 작다: ICONDIR 헤더를 읽을 수 없다")
	}

	count := int(binary.LittleEndian.Uint16(IconData[4:6]))

	// ICONDIR(6) + count * ICONDIRENTRY(16) 이상이어야 한다
	requiredSize := 6 + count*16
	if len(IconData) < requiredSize {
		t.Fatalf("IconData가 너무 작다: %d개 엔트리에 최소 %d바이트 필요, got %d",
			count, requiredSize, len(IconData))
	}

	// 기대하는 크기 목록 (0은 256을 의미)
	expectedDims := map[byte]bool{
		16:  true, // 16x16
		32:  true, // 32x32
		48:  true, // 48x48
		64:  true, // 64x64
		128: true, // 128x128
		0:   true, // 256x256 (ICO 포맷에서 0은 256)
	}

	foundDims := make(map[byte]bool)

	for i := 0; i < count; i++ {
		offset := 6 + i*16
		width := IconData[offset]
		height := IconData[offset+1]

		// 너비와 높이가 같아야 한다 (정사각형 아이콘)
		if width != height {
			t.Errorf("엔트리 %d: 정사각형이 아니다 (width=%d, height=%d)", i, width, height)
		}

		// SizeInBytes가 0보다 커야 한다
		sizeInBytes := binary.LittleEndian.Uint32(IconData[offset+8 : offset+12])
		if sizeInBytes == 0 {
			t.Errorf("엔트리 %d (dim=%d): SizeInBytes가 0이다", i, width)
		}

		// Offset이 유효한 범위 안에 있어야 한다
		dataOffset := binary.LittleEndian.Uint32(IconData[offset+12 : offset+16])
		if int(dataOffset+sizeInBytes) > len(IconData) {
			t.Errorf("엔트리 %d (dim=%d): 데이터 범위 초과 (offset=%d, size=%d, total=%d)",
				i, width, dataOffset, sizeInBytes, len(IconData))
		}

		foundDims[width] = true
	}

	// 기대하는 모든 크기가 존재하는지 검증
	for dim := range expectedDims {
		if !foundDims[dim] {
			displayDim := int(dim)
			if dim == 0 {
				displayDim = 256
			}
			t.Errorf("기대하는 아이콘 크기 %dx%d이 없다", displayDim, displayDim)
		}
	}
}
