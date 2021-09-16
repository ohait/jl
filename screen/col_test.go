package screen

import (
	"testing"

	"github.com/gdamore/tcell"
)

func TestColor(t *testing.T) {
	s := tcell.StyleDefault
	fg, bg, attr := s.Decompose()
	t.Logf("fg: %v, bg: %v, attr: %v", fg, bg, attr)
	{
		r, g, b := bg.RGB()
		t.Logf("bg: RGB: %d, %d, %d", r, g, b)
	}
	{
		r, g, b := fg.RGB()
		t.Logf("fg: RGB: %d, %d, %d", r, g, b)
	}
	t.Logf("test: \x1b[38;5;226mXXXX\x1b[0m")
	t.Logf("test: \x1b[38;2;150;250;50mRGB\x1b[0m")
}
