package widget

import "testing"

func TestPickBoardSizeSelectsLargestFittingGeometry(t *testing.T) {
	cases := []struct {
		name          string
		width, height int
		want          BoardSize
	}{
		{"tiny terminal falls back to small", 30, 12, BoardSmall},
		{"medium fits but large does not", 50, 28, BoardMedium},
		{"large fits", 100, 40, BoardLarge},
		{"height too short for medium", 50, 18, BoardSmall},
		{"width too narrow for large", 55, 40, BoardMedium},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := PickBoardSize(tc.width, tc.height)
			if got != tc.want {
				t.Fatalf("PickBoardSize(%d,%d) = %v, want %v", tc.width, tc.height, got, tc.want)
			}
		})
	}
}

func TestBoardSizeFootprintMatchesCellGeometry(t *testing.T) {
	for _, sz := range []BoardSize{BoardSmall, BoardMedium, BoardLarge} {
		cw, ch := sz.Cell()
		fw, fh := sz.Footprint()
		if fw < cw*8 || fh < ch*8 {
			t.Errorf("size %v footprint %dx%d smaller than 8x8 cells at %dx%d", sz, fw, fh, cw, ch)
		}
	}
}
