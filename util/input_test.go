package util

import "testing"

func TestInput(t *testing.T) {
	h := &HistoryInput{}
	i := h.NewUnlessEmpty()
	i.Left()
	i.Right()
	i.Append("1")
	i.Left()
	i.Append("0")
	if i.String() != "01" {
		t.Fatalf("expected 01, got %q", i)
	}
	i = h.Up()
	if i.String() != "01" {
		t.Fatalf("expected 01, got %q", i)
	}

	i = h.NewUnlessEmpty()
	if i.String() != "" {
		t.Fatalf("expected empty, got %q", i)
	}
	i = h.Up()
	if i.String() != "01" {
		t.Fatalf("expected 01, got %q", i)
	}

}
