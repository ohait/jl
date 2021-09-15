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

func (this Cursor) Printf(f string, args ...interface{}) Cursor {
	return this.Print(fmt.Sprintf(f, args...))
}

var reKeywords = regexp.MustCompile(`\b(error|err|panic|close|invalid)`)

func (this Cursor) Print(s string) Cursor {
	if this.pattern == nil {
		return this.printHL(s, reKeywords, this.Style.Foreground(tcell.Color156), func(this Cursor, s string) Cursor {
			return this.print(s)
		})
	} else {
		return this.printHL(s, this.pattern, this.Style.Foreground(tcell.Color87).Bold(true), func(this Cursor, s string) Cursor {
			return this.printHL(s, reKeywords, this.Style.Foreground(tcell.Color156), func(this Cursor, s string) Cursor {
				return this.print(s)
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

func (this Cursor) print(s string) Cursor {
	for _, ch := range s {
		switch ch {
		case '"', '\'', '(', '{', '}', ')', ',', ':', '[', ']', '/':
			this.scr.scr.SetContent(this.X, this.Y, ch, nil, this.Style.Dim(true))
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

func (this Cursor) Col(col tcell.Color) Cursor {
	this.Style = this.Style.Foreground(col)
	return this
}

func (this Cursor) ColByLevel(l string) Cursor {
	switch strings.ToLower(l) {
	case "debug":
		return this.Col(tcell.ColorDarkCyan)
	case "info":
		return this.Col(tcell.ColorLightCyan)
	case "notice":
		return this.Col(tcell.ColorWhite)
	case "warn":
		return this.Col(tcell.ColorYellow)
	case "error":
		return this.Col(tcell.ColorRed)
	default:
		return this
	}
}

func (this Cursor) Level(l string) Cursor {
	fg, _, _ := this.Style.Decompose()
	switch strings.ToLower(l) {
	case "debug":
		this = this.Col(tcell.ColorDarkCyan).Print("dbg").Col(fg)
	case "info":
		this = this.Col(tcell.ColorLightCyan).Print("inf").Col(fg)
	case "notice":
		this = this.Col(tcell.ColorWhite).Print("---").Col(fg)
	case "warn":
		this = this.Col(tcell.ColorYellow).Print("WRN").Col(fg)
	case "error":
		this = this.Col(tcell.ColorRed).Print("ERR").Col(fg)
	default:
		this = this.Print(l)
	}
	return this
}

func (this Cursor) Line(l tbuf.Line, padding int) Cursor {
	if l.Mark {
		this.Style = this.Style.Background(tcell.Color52)
	}
	fg, _, _ := this.Style.Decompose()
	if !l.Time.IsZero() {
		this = this.Col(tcell.Color246).Time(l.Time)
		this = this.Print(" ").Col(fg)
	}
	this = this.Level(l.Level)
	this = this.Print(" ")
	if l.Short != "" {
		this = this.Print(l.Short)
	} else {
		this = this.Print(l.Str)
	}
	return this.Clear()
}
