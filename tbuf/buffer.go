package tbuf

import (
	"sync"
	"time"
)

type Buffer struct {
	m     sync.Mutex
	Lines []Line
	Pos   int
	Last  time.Time
}

func (this *Buffer) Size() int {
	this.m.Lock()
	defer this.m.Unlock()
	return len(this.Lines)
}

func (this *Buffer) At(pos int) Line {
	this.m.Lock()
	defer this.m.Unlock()
	return this.Lines[pos]
}

func (this *Buffer) Range(f func(i int, l *Line) bool) {
	this.m.Lock()
	defer this.m.Unlock()
	for i := 0; i < len(this.Lines); i++ {
		if !f(i, &this.Lines[i]) {
			return
		}
	}
}

func (this *Buffer) Mark() {
	this.m.Lock()
	defer this.m.Unlock()
	if len(this.Lines) == 0 {
		return
	}
	if this.Pos >= len(this.Lines) {
		return
	}
	if this.Pos < 0 {
		return
	}
	this.Lines[this.Pos].Mark = !this.Lines[this.Pos].Mark
}

func (this *Buffer) Get() (Line, bool) {
	this.m.Lock()
	defer this.m.Unlock()
	if len(this.Lines) == 0 {
		return Line{}, false
	}
	if this.Pos >= len(this.Lines) {
		return Line{}, false
	}
	if this.Pos < 0 {
		return Line{}, false
	}
	return this.Lines[this.Pos], true
}

func (this *Buffer) Down(f func(s Line) bool) (Line, bool) {
	cur := this.NewCursor()
	l, ok := cur.Down(f)
	this.m.Lock()
	defer this.m.Unlock()
	if ok {
		this.Pos = cur.Cur
	} else {
		this.Pos = len(this.Lines) // TAIL
	}
	return l, ok
}

func (this *Buffer) Up(f func(s Line) bool) (Line, bool) {
	cur := this.NewCursor()
	l, ok := cur.Up(f)
	this.m.Lock()
	defer this.m.Unlock()
	this.Pos = cur.Cur
	return l, ok
}
