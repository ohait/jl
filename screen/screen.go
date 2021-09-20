package screen

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/ohait/jl/tbuf"
	"github.com/ohait/jl/util"
	"golang.org/x/sys/unix"
)

type Screen struct {
	Done     chan struct{}
	m        sync.Mutex
	scr      tcell.Screen
	origBuf  *tbuf.Buffer
	buffer   *tbuf.Buffer
	row      int
	col      int
	log      func(...interface{})
	details  int
	help     bool
	pattern  *regexp.Regexp
	Refresh  bool
	input    *util.HistoryInput
	onEnter  func()
	onChange func()
	query    string
	queryEnd string
}

func NewScreen(log func(...interface{}), buffer *tbuf.Buffer) *Screen {
	scr, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	scr.Init()
	scr.HideCursor()
	undoTcellSig()

	this := &Screen{
		scr:     scr,
		log:     log,
		origBuf: buffer,
		buffer:  buffer,
		input:   &util.HistoryInput{},
		Done:    make(chan struct{}),
	}
	//this.pattern = regexp.MustCompile(`lighthouse`)
	go util.Recover(func() {
		defer close(this.Done)
		for {
			ev := scr.PollEvent()
			err := this.handleEvent(ev)
			if err != nil {
				this.log("EXIT: %v", err)
				return
			}
		}
	}, func(r interface{}, stack string) {
		scr.Clear()
		scr.Fini()
		fmt.Printf("panic: %v\n%s\n", r, stack)
		os.Exit(-1)
	})
	scr.Fill(' ', tcell.StyleDefault)
	w, h := scr.Size()
	this.log("size: %d/%d", w, h)
	this.row = h - 5
	return this
}

func (this *Screen) handleEvent(ev tcell.Event) error {
	this.m.Lock()
	defer this.m.Unlock()
	err := this.event(ev)
	if err == nil {
		return nil
	}
	this.scr.Sync()
	switch err {
	case Exit:
		return err
	default:
		this.log("on poll: %v", err)
		_, h := this.scr.Size()
		this.NewCursor(0, h-1).Fg(tcell.ColorPink).Printf("ERROR: %v", err).Clear()
		time.Sleep(time.Second)
		return nil
	}
}

func (this *Screen) Close() {
	this.scr.Fini()
}

func (this *Screen) Repaint() {
	this.Refresh = false
	w, h := this.scr.Size()
	this.log("repaint (size: %d/%d, this.row: %d, buff: %d)", w, h, this.row, this.buffer.Pos)

	nomatch := tcell.StyleDefault.Foreground(tcell.Color246).Background(tcell.Color232)

	if this.details > 0 {
		if this.row < 5 {
			this.row++
			this.Refresh = true
		} else if this.row > 5 {
			this.row--
			this.Refresh = true
		}
	}

	// before
	bc := this.buffer.NewCursor()
	for y := 1; y < 100; y++ {
		if this.row-y < 0 {
			break
		}
		if line, ok := bc.Up(nil); ok {
			cur := this.NewCursor(-this.col, this.row-y)
			if this.pattern != nil && !this.pattern.MatchString(line.Str) {
				cur.Style = nomatch
			}
			cur.Line(line, this.col)
		} else {
			this.NewCursor(0, this.row-y).Clear()
			this.row--
			this.log("REPAINT for better framing")
			this.Refresh = true
		}
	}

	// after
	bc = this.buffer.NewCursor()
	for y := 1; y < 100; y++ {
		if y+this.row >= h {
			break
		}
		if line, ok := bc.Down(nil); ok {
			cur := this.NewCursor(-this.col, this.row+y)
			if this.pattern != nil && !this.pattern.MatchString(line.Str) {
				cur.Style = nomatch
			}
			cur.Line(line, this.col)
		} else {
			this.NewCursor(0, y+this.row).Clear()
		}
	}

	// cursor
	cur := this.NewCursor(-this.col, this.row)
	cur.Style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	if line, ok := this.buffer.Get(); ok {
		//this.log("cur: %+v", line)

		cur.pattern = this.pattern
		cur = cur.Line(line, this.col)
		cur.Style = tcell.StyleDefault.Background(tcell.Color236)

		switch this.details {
		case 0: // none
		case 1: // tags
			cur.X = 24 - this.col
			cur = cur.Fg(tcell.ColorTeal).Printf(" time: %v, level: %s", line.Time.UTC(), line.Level).Clear()
			for _, tag := range line.SortedTags() {
				v := line.Tags[tag]
				if len(v) == 0 {
					continue
				}
				cur.X = 24 - this.col
				cur = cur.Fg(tcell.ColorOrange).Printf(" %s: ", tag).Fg(tcell.ColorWhite)
				if len(v) > 40 && v[0] == '{' {
					var x interface{}
					err := json.Unmarshal(v, &x)
					if err == nil {
						j, _ := json.MarshalIndent(x, "", "  ")
						lines := strings.Split(string(j), "\n")
						this.log("subtag: %q", lines[1])
						for _, l := range lines[1 : len(lines)-1] {
							this.log("subtag: %q", l)
							cur = cur.Clear()
							cur.X = 24 - this.col
							cur = cur.PrintfHL("    %s", l)
						}
					} else {
						cur = cur.PrintHL(string(v))
					}
				} else {
					cur = cur.PrintHL(string(v))
				}
				cur = cur.Clear()
			}
			if this.row > 0 && cur.Y > h {
				this.row--
				this.log("REPAINT for better framing")
				this.Refresh = true
				return
			}

		case 2: // expand and prettify
			//cur.X = 24
			//cur = cur.Col(tcell.ColorTeal).Printf(" level: ").Level(line.Level).Printf(", time: %v", line.Time).Clear()
			str := line.Short
			if str == "" {
				str = line.Str
			}
			for _, s := range util.Prettify(str) {
				cur.X = 24 - this.col
				cur = cur.PrintfHL(" %s", s).Clear()
			}
			cur.X = 24 - this.col
			cur = cur.Fg(tcell.ColorTeal).Printf(" time: %v, level: %s", line.Time.UTC(), line.Level).Clear()
			//cur.X = 24
			//cur = cur.Clear()
		}

	} else {
		cur.Clear()
	}

	// status bar
	{
		cur := this.NewCursor(0, h-1)
		cur.Style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
		cur = cur.Printf("file: %d/%d ", this.buffer.Pos, len(this.buffer.Lines))
		if this.col != 0 {
			cur = cur.Printf("col: %d ", this.col)
		}
		if this.buffer != this.origBuf {
			cur = cur.Printf(" (orig: %d lines)", len(this.origBuf.Lines))
		}
		this.scr.HideCursor()
		switch this.query {
		case "":
			if this.pattern != nil {
				cur = cur.Printf(" /%+v/", this.pattern)
			}
		case "SEARCH":
			cur = cur.Print(" /")
			l, r := this.input.Get().Details()
			cur.Style = cur.Style.Bold(true)
			cur = cur.Print(l)
			this.scr.ShowCursor(cur.X, cur.Y)
			cur = cur.Print(r)
			cur.Style = cur.Style.Bold(false)
			cur = cur.Print("/")
		default:
			if this.pattern != nil {
				cur = cur.Printf(" /%+v/", this.pattern)
			}
			cur = cur.Print(" ")
			cur = cur.Print(this.query)
			l, r := this.input.Get().Details()
			cur.Style = cur.Style.Bold(true)
			cur = cur.Print(l)
			this.scr.ShowCursor(cur.X, cur.Y)
			cur = cur.Print(r)
			cur.Style = cur.Style.Bold(false)
			cur = cur.Print(this.queryEnd)
		}
		cur.Clear()
	}

	// HELP

	if this.help {
		cur := this.NewCursor(20, 2)
		cur.Style = tcell.StyleDefault.Background(tcell.Color237).Foreground(tcell.Color222)
		cur = cur.Printf("   [h] Help                  ").CR(20)
		cur = cur.Printf("   [/] Search                ").CR(20)
		cur = cur.Printf("   [M] Mark matches          ").CR(20)
		cur = cur.Printf(" [⇧+M] unmark all            ").CR(20)
		cur = cur.Printf("   [ ] Mark current line     ").CR(20)
		cur = cur.Printf("   [G] grep marked           ").CR(20)
		cur = cur.Printf(" [⇧+G] grep unmarked         ").CR(20)
		cur = cur.Printf(" [⇧+O] original buffer       ").CR(20)
		cur = cur.Printf("   [C] copy current line     ").CR(20)
		cur = cur.Printf(" [⇧+C] copy marked lines     ").CR(20)
		cur = cur.Printf("   [0] first line            ").CR(20)
		cur = cur.Printf(" [⇧+F] tail mode             ").CR(20)
		cur = cur.Printf("   [D] show details          ").CR(20)
		cur = cur.Printf("   [Q] quit                  ").CR(20)
	}

	this.scr.Show()
}

func undoTcellSig() error {
	tio, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TCGETS)
	if err != nil {
		return err
	}
	raw := &unix.Termios{
		Cflag: tio.Cflag,
		Oflag: tio.Oflag,
		Iflag: tio.Iflag,
		Lflag: tio.Lflag,
		Cc:    tio.Cc,
	}
	raw.Lflag |= unix.ISIG
	return unix.IoctlSetTermios(int(os.Stdout.Fd()), unix.TCSETS, raw)
}
