package screen

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/ohait/jl/tbuf"
)

func (this *Screen) NewCursor(x, y int) Cursor {
	return Cursor{
		scr:     this,
		X:       x,
		Y:       y,
		pattern: this.pattern,
		Style:   tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.Color234),
	}
}

type Cursor struct {
	scr     *Screen
	X       int
	Y       int
	Style   tcell.Style
	pattern *regexp.Regexp
}

func (this Cursor) CR(x int) Cursor {
	this.X = x
	this.Y++
	return this
}

func (this Cursor) PrintfHL(f string, args ...interface{}) Cursor {
	return this.PrintHL(fmt.Sprintf(f, args...))
}

func (this Cursor) Printf(f string, args ...interface{}) Cursor {
	return this.Print(fmt.Sprintf(f, args...))
}

var reKeywords = regexp.MustCompile(`\b((?i)error|err|panic|closed?|invalid)\b`)

func (this Cursor) PrintHL(s string) Cursor {
	if this.pattern == nil {
		return this.printHL(s, reKeywords, this.Style.Foreground(tcell.Color173), func(this Cursor, s string) Cursor {
			return this.print(s, true)
		})
	} else {
		return this.printHL(s, this.pattern, this.Style.Foreground(tcell.Color87).Bold(true), func(this Cursor, s string) Cursor {
			return this.printHL(s, reKeywords, this.Style.Foreground(tcell.Color173), func(this Cursor, s string) Cursor {
				return this.print(s, true)
			})
		})
	}
}

func (this Cursor) printHL(s string, p *regexp.Regexp, st tcell.Style, pfunc func(Cursor, string) Cursor) Cursor {
	orig := this.Style
	for {
		pairs := p.FindStringIndex(s)
		if pairs == nil {
			break
		}
		this = pfunc(this, s[0:pairs[0]])
		this.Style = st
		this = pfunc(this, s[pairs[0]:pairs[1]])
		this.Style = orig
		s = s[pairs[1]:]
	}
	this = pfunc(this, s)
	return this
}

func (this Cursor) Print(s string) Cursor {
	return this.print(s, false)
}

func (this Cursor) print(s string, dim bool) Cursor {
	for _, ch := range s {
		switch ch {
		case '"', '\'', '(', '{', '}', ')', ',', ':', '[', ']', '/':
			this.scr.scr.SetContent(this.X, this.Y, ch, nil, this.Style.Dim(dim))
		case '\t':
			this.scr.scr.SetContent(this.X, this.Y, '⇥', nil, this.Style.Bold(true).Foreground(tcell.Color220))
		case '\n':
			this.scr.scr.SetContent(this.X, this.Y, '␍', nil, this.Style.Bold(true).Foreground(tcell.Color220))
		default:
			this.scr.scr.SetContent(this.X, this.Y, ch, nil, this.Style)
		}
		this.X++
	}
	return this
}

func (this Cursor) Time(t time.Time) Cursor {
	d := time.Since(t)
	if d < 0 {
		d -= d
	}
	var s string
	if d < 70*time.Hour {
		s = t.UTC().Format("15:04:05.000")
	} else if d < 10*24*time.Hour {
		s = t.UTC().Format("01-02 15:04:")
	} else {
		s = t.UTC().Format("06-01-02 15:")
	}
	return this.Print(s)
}

func (this Cursor) Fill(r rune) Cursor {
	w, _ := this.scr.scr.Size()
	for ; this.X < w; this.X++ {
		this.scr.scr.SetContent(this.X, this.Y, r, nil, this.Style)
	}
	this.X = 0
	this.Y++
	return this
}

func (this Cursor) Clear() Cursor {
	return this.Fill(' ')
}

func (this Cursor) Reverse(r bool) Cursor {
	this.Style = this.Style.Reverse(r)
	return this
}

func (this Cursor) Col(fg, bg tcell.Color) Cursor {
	this.Style = this.Style.Foreground(fg).Background(bg)
	return this
}

func (this Cursor) Fg(col tcell.Color) Cursor {
	this.Style = this.Style.Foreground(col)
	return this
}

func (this Cursor) Bg(col tcell.Color) Cursor {
	this.Style = this.Style.Background(col)
	return this
}

func (this Cursor) ColByLevel(l string) Cursor {
	_, bg, _ := this.Style.Decompose()
	r, g, b := bg.RGB()
	invert := r+g+b > 200
	if bg == -1 {
		invert = false
	}
	var col tcell.Color
	switch strings.ToLower(l) {
	case "debug":
		col = tcell.Color66
	case "info":
		col = tcell.Color116
	case "notice":
		col = tcell.ColorWhite
	case "warn":
		col = tcell.ColorYellow
	case "error":
		col = tcell.ColorRed
	default:
		return this
	}
	if invert {
		return this.Bg(col)
	} else {
		return this.Fg(col)
	}
}

func (this Cursor) Level(l string) Cursor {
	fg, bg, _ := this.Style.Decompose()
	var label string
	switch strings.ToLower(l) {
	case "debug":
		label = "dbg"
	case "info":
		label = "inf"
	case "notice":
		label = "ntc"
	case "warn":
		label = "WRN"
	case "error":
		label = "ERR"
	default:
		label = l
	}
	this = this.ColByLevel(l).print(label, false).Col(fg, bg)
	return this
}

func (this Cursor) Line(l tbuf.Line, padding int) Cursor {
	if l.Mark {
		this.Style = this.Style.Background(tcell.Color52)
	}
	fg, _, _ := this.Style.Decompose()
	if !l.Time.IsZero() {
		this = this.Fg(tcell.Color246).Time(l.Time)
		this = this.Print(" ").Fg(fg)
	}
	this = this.Level(l.Level)
	this = this.Print(" ")
	if l.Short != "" {
		this = this.PrintHL(l.Short)
	} else {
		this = this.PrintHL(l.Str)
	}
	return this.Clear()
}
