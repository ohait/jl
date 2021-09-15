package util

import "testing"

func TestClip(t *testing.T) {
	err := ClipCopy(t.Name())
	if err != nil {
		t.Logf("ignoring... missing xclip? %v", err)
		t.SkipNow()
	}
	s, err := ClipPaste()
	if err != nil {
		t.Fatalf("failed to paste back: %v", err)
	}
	if s != t.Name() {
		t.Fatalf("got back the wrong clipboard")
	}
}
