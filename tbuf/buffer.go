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
	this.m.Lock()
	defer this.m.Unlock()
	if this.Pos >= len(this.Lines) { // already to the tail
		return Line{}, false
	}
	if len(this.Lines) == 0 {
		return Line{}, false
	}

	p := this.Pos
	for {
		p++
		if p >= len(this.Lines) {
			this.Pos = p
			return this.Lines[len(this.Lines)-1], false
		}
		if f == nil || f(this.Lines[p]) {

			this.Pos = p
			return this.Lines[p], true
		}
	}
}

func (this *Buffer) Up(f func(s Line) bool) (Line, bool) {
	this.m.Lock()
	defer this.m.Unlock()
	if this.Pos == 0 {
		return Line{}, false
	}

	p := this.Pos
	for {
		p--
		if p <= 0 {
			this.Pos = 0
			return this.Lines[0], true
		}
		if f == nil || f(this.Lines[p]) {
			this.Pos = p
			return this.Lines[p], true
		}
	}
}
