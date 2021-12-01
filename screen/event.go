package screen

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/ohait/jl/tbuf"
	"github.com/ohait/jl/util"
)

var Exit = errors.New("EXIT")

func (this *Screen) event(ev tcell.Event) error {
	this.help = false
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if this.query != "" { // user query
			return this.eventQuery(ev)
		} else {
			return this.eventNav(ev)
		}
	case *tcell.EventResize:
		this.Refresh = true
		return nil
	default:
		return fmt.Errorf("unexpected event: %#v", ev)
	}
}

func (this *Screen) eventQuery(ev *tcell.EventKey) error {
	this.log("eventQuery(%+v)", ev)
	switch ev.Key() {
	case tcell.KeyLeft:
		this.input.Get().Left()
		this.Repaint()
	case tcell.KeyRight:
		this.input.Get().Right()
		this.Repaint()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		this.input.Get().Backspace()
		this.onChange()
		this.Repaint()
	case tcell.KeyUp:
		this.input.Up()
		this.onChange()
		this.Repaint()
	case tcell.KeyDown:
		this.input.Down()
		this.onChange()
		this.Repaint()
	case tcell.KeyPgDn:
		this.input.Last()
		this.onChange()
		this.Repaint()
	case tcell.KeyEnter:
		this.onEnter()
		this.Repaint()
	case tcell.KeyRune:
		this.log("query char: %q", ev.Rune())
		switch ev.Rune() {
		default:
			this.input.Get().Append(string(ev.Rune()))
			this.onChange()
			this.Repaint()
		}
	default:
		this.log("query key: %#v", ev)
	}
	return nil
}

func (this *Screen) eventNav(ev *tcell.EventKey) error {
	this.log("eventNav(%+v)", ev)
	_, h := this.scr.Size()
	switch ev.Key() {
	case tcell.KeyF1:
		this.help = true
		this.Repaint()
	case tcell.KeyUp:
		this.detoffset = 0
		this.buffer.Up(nil)
		if this.row > 0 {
			this.row--
		}
		this.Repaint()
	case tcell.KeyDown:
		this.detoffset = 0
		this.buffer.Down(nil)
		if this.row < h-2 {
			this.row++
		}
		this.Repaint()
	case tcell.KeyPgUp:
		if this.details > 0 {
			this.detoffset -= 5
			if this.detoffset < 0 {
				this.detoffset = 0
			}
		} else {
			for i := 0; i < 25; i++ {
				this.buffer.Up(nil)
			}
			this.row -= 25
			if this.row < 0 {
				this.row = 0
			}
		}
		this.Refresh = true
	case tcell.KeyPgDn:
		if this.details > 0 {
			this.detoffset += 5
			if this.detoffset < 0 {
				this.detoffset = 0
			}
		} else {
			for i := 0; i < 25; i++ {
				this.buffer.Down(nil)
			}
			this.row += 25
			if this.row > h-2 {
				this.row = h - 2
			}
		}
		this.Refresh = true

	case tcell.KeyLeft:
		if this.col >= 20 {
			this.col -= 20
		} else {
			this.col = 0
		}
		this.Repaint()
	case tcell.KeyRight:
		this.col += 20
		this.Repaint()

	default:
		switch ev.Rune() {
		case 'h':
			this.help = true
			this.Repaint()
		case '0':
			this.buffer.Pos = 0
			this.row = 0
			this.Repaint()
		case 'F': // Follow
			this.buffer.Pos = len(this.buffer.Lines)
			this.row = h - 2
			if this.row > this.buffer.Pos {
				this.row = this.buffer.Pos
			}
			this.Repaint()
		case '/':
			// search
			this.query = "SEARCH"
			this.input.NewUnlessEmpty()
			this.onEnter = func() {
				this.query = ""
			}
			this.onChange = func() {
				this.pattern = nil
				p := this.input.Get().String()
				this.log("SEARCH: %q", p)
				if len(p) == 0 {
					return
				}
				var err error
				this.pattern, err = regexp.Compile(p)
				if err != nil {
					this.pattern, _ = regexp.Compile(regexp.QuoteMeta(p))
				}
			}
			this.onChange()
			this.Repaint()

		case 'n': //scan next
			this.detoffset = 0
			if this.pattern != nil {
				this.buffer.Down(func(l tbuf.Line) bool {
					if this.row < h-2 {
						this.row++
					}
					return l.Mark || this.pattern.MatchString(l.Str)
				})
				this.Repaint()
			} else {
				this.buffer.Down(func(l tbuf.Line) bool {
					if this.row < h-2 {
						this.row++
					}
					return l.Mark
				})
				this.Repaint()
			}
		case 'N': //scan prev
			this.detoffset = 0
			if this.pattern != nil {
				this.buffer.Up(func(l tbuf.Line) bool {
					if this.row > 0 {
						this.row--
					}
					return l.Mark || this.pattern.MatchString(l.Str)
				})
				this.Repaint()
			} else {
				this.buffer.Up(func(l tbuf.Line) bool {
					if this.row > 0 {
						this.row--
					}
					return l.Mark
				})
				this.Repaint()
			}

		case ' ':
			if _, ok := this.buffer.Get(); ok {
				this.buffer.Mark()
				this.buffer.Down(nil)
				if this.row < h-2 {
					this.row++
				}
			}
			this.Repaint()

		case 'M': // unmark all
			this.buffer.Range(func(i int, l *tbuf.Line) bool {
				l.Mark = false
				return true
			})
			this.Repaint()

		case 'm': // mark searches
			if this.pattern != nil {
				this.buffer.Range(func(i int, l *tbuf.Line) bool {
					if this.pattern.MatchString(l.Str) {
						l.Mark = true
					}
					return true
				})
				this.pattern = nil // usually makes sense
			}
			this.Repaint()

		case 'g': // grep mode (only show marked)
			b := &tbuf.Buffer{}
			this.buffer.Range(func(i int, l *tbuf.Line) bool {
				if l.Mark {
					l := *l
					l.Mark = false
					b.AppendLine(l)
				}
				if i == this.buffer.Pos {
					b.Pos = len(b.Lines)
				}
				return true
			})
			if len(b.Lines) > 0 {
				this.buffer = b
				this.Repaint()
			}
		case 'G': // grep mode inverted (only unmarked)
			b := &tbuf.Buffer{}
			this.buffer.Range(func(i int, l *tbuf.Line) bool {
				if !l.Mark {
					l := *l
					l.Mark = false
					b.AppendLine(l)
				}
				if i == this.buffer.Pos {
					b.Pos = len(b.Lines)
				}
				return true
			})
			if len(b.Lines) > 0 {
				this.buffer = b
				this.Repaint()
			}

		case 'O':
			this.detoffset = 0
			this.buffer = this.origBuf
			this.Repaint()

		//case '?': // search backward ?
		case 'd':
			this.detoffset = 0
			this.details++
			if this.details > 2 {
				this.details = 0
			}
			this.Repaint()
		case 'q':
			return Exit

		case 'c': // copy line
			if line, ok := this.buffer.Get(); ok {
				this.log("copying %s", line.Str)
				err := util.ClipCopy(line.Str)
				if err != nil {
					this.log("copy error: %v", err)
				}
			}
		case 'C': // copy marked
			out := []string{}
			this.buffer.Range(func(i int, l *tbuf.Line) bool {
				if l.Mark {
					out = append(out, l.Str+"\n")
				}
				return true
			})
			this.log("copying %d lines", len(out))
			err := util.ClipCopy(strings.Join(out, ""))
			if err != nil {
				this.log("copy error: %v", err)
			}
		}
	}
	return nil
}
