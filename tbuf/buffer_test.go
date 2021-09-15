package tbuf

import (
	"strings"
	"testing"
)

func TestBuffer(t *testing.T) {
	b := &Buffer{}
	b.Down(nil)
	b.Up(nil)
	b.Down(nil)
	b.Up(nil)
	if b.Pos != 0 {
		t.Fatalf("non zero pos")
	}

	b.Append("one", t.Log)
	if b.Pos != 1 {
		t.Fatal()
	}
	b.Down(nil)
	if b.Pos != 1 {
		t.Fatal()
	}
	b.Up(nil)
	if b.Pos != 0 {
		t.Fatal()
	}
	b.Up(nil)
	if b.Pos != 0 {
		t.Fatal()
	}

	b.Append("two", t.Log)
	if b.Pos != 0 {
		t.Fatal() // won't move, not in tail mode
	}
	b.Append("three foo", t.Log)
	b.Append("four foo", t.Log)
	b.Append("five bar", t.Log)
	b.Append("six", t.Log)

	b.Down(func(s Line) bool { return strings.Contains(s.Str, "foo") })
	if b.Pos != 2 {
		t.Fatal(b.Pos)
	}

	b.Down(func(s Line) bool { return strings.Contains(s.Str, "foo") })
	if b.Pos != 3 {
		t.Fatal(b.Pos)
	}

	b.Down(func(s Line) bool { return strings.Contains(s.Str, "bar") })
	if b.Pos != 4 {
		t.Fatal(b.Pos)
	}

	b.Up(func(s Line) bool { return strings.Contains(s.Str, "foo") })
	if b.Pos != 3 {
		t.Fatal(b.Pos)
	}

	b.Up(func(s Line) bool { return strings.Contains(s.Str, "none") })
	if b.Pos != 0 {
		t.Fatal(b.Pos)
	}
}
