package tbuf

type Cursor struct {
	Buffer *Buffer
	Cur    int
}

func (this *Buffer) NewCursor() *Cursor {
	return &Cursor{
		Buffer: this,
		Cur:    this.Pos,
	}
}

func (this *Cursor) Down(f func(s Line) bool) (Line, bool) {
	size := this.Buffer.Size()
	if this.Cur >= size { // already to the tail
		return Line{}, false
	}
	if size == 0 {
		return Line{}, false
	}

	p := this.Cur
	for {
		p++
		if p >= size {
			this.Cur = p
			return this.Buffer.At(size - 1), false
		}
		if f == nil || f(this.Buffer.At(p)) {

			this.Cur = p
			return this.Buffer.At(p), true
		}
	}
}

func (this *Cursor) Up(f func(s Line) bool) (Line, bool) {
	if this.Cur == 0 {
		return Line{}, false
	}

	p := this.Cur
	for {
		p--
		if p <= 0 {
			this.Cur = 0
			return this.Buffer.At(0), true
		}
		if f == nil || f(this.Buffer.At(p)) {
			this.Cur = p
			return this.Buffer.At(p), true
		}
	}
}

func (this *Cursor) Commit() {
	this.Buffer.m.Lock()
	defer this.Buffer.m.Unlock()
	this.Buffer.Pos = this.Cur
}
