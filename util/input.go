package util

type Input struct {
	left  string
	right string
}

func (this *Input) String() string {
	return this.left + this.right
}

func (this *Input) Details() (string, string) {
	return this.left, this.right
}

func (this *Input) Append(s string) {
	this.left = this.left + s
}

func (this *Input) Backspace() {
	Chop(&this.left)
}

func (this *Input) Left() {
	ch := Chop(&this.left)
	this.right = ch + this.right
}

func (this *Input) Right() {
	ch := Shift(&this.right)
	this.left += ch
}

type HistoryInput struct {
	Pos     int
	History []*Input
}

func (this *HistoryInput) Get() *Input {
	return this.History[this.Pos]
}

func (this *HistoryInput) NewUnlessEmpty() *Input {
	if len(this.History) > 0 {
		i := this.Last()
		if i.String() == "" {
			return i
		}
	}
	i := &Input{}
	this.History = append(this.History, i)
	this.Pos = len(this.History) - 1
	return i
}

func (this *HistoryInput) Up() *Input {
	if this.Pos > 0 {
		this.Pos--
	}
	return this.History[this.Pos]
}

func (this *HistoryInput) Down() *Input {
	if this.Pos < len(this.History)-1 {
		this.Pos++
	}
	return this.History[this.Pos]
}

func (this *HistoryInput) Last() *Input {
	this.Pos = len(this.History) - 1
	if this.Pos < 0 {
		return nil
	}
	return this.History[this.Pos]
}
