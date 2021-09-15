package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ohait/jl/screen"
	"github.com/ohait/jl/tbuf"
	"golang.org/x/sys/unix"
)

var buffer = &tbuf.Buffer{}

var scr *screen.Screen

func main() {
	log("INIT #######################################################")
	defer log("EXIT")

	scr = screen.NewScreen(log, buffer)
	defer scr.Close()

	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	readInit()

	var last time.Time
	t := time.NewTimer(30 * time.Millisecond)
LOOP:
	for {
		select {
		case <-scr.Done:
			break LOOP

		case <-t.C:
			t.Reset(50 * time.Millisecond)
			if buffer.Last != last {
				last = buffer.Last
				scr.Repaint()
			} else if scr.Refresh {
				scr.Repaint()
			}

		case sig := <-sigchan:
			log("SIG: %v", sig)
			break LOOP
		}
	}
}

func log(args ...interface{}) {
	now := time.Now()
	_, file, line, _ := runtime.Caller(1)
	prefix := fmt.Sprintf("%s %s:%d ", now.Format(`15:04:05.000`), file, line)
	var ln string
	switch len(args) {
	case 0:
		return
	case 1:
		ln = fmt.Sprintf("%+v", args[0])
	default:
		switch f := args[0].(type) {
		case string:
			ln = fmt.Sprintf(f, args[1:]...)
		default:
			ln = fmt.Sprintf("%+v", args...)
		}
	}
	f, err := os.OpenFile("/tmp/less.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	if err == nil {
		fmt.Fprintln(f, prefix+ln)
	} else {
		panic(err)
	}
}

func IsTerminal(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd), unix.TCGETS)
	return err == nil
}

func readInit() {
	if len(os.Args) == 1 {
		if !IsTerminal(os.Stdin.Fd()) {
			// only slurp stdin if not a terminal
			go read(os.Stdin, "STDIN")
			//defer os.Stdin.Close()
		} else {
			buffer.Append("terminal and nothing to read", log)
			return
		}
	} else {
		f, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}
		filename := f.Name()
		go read(f, filename)
	}
}

func read(f io.Reader, fname string) {
	r := bufio.NewReader(f)
	log("reading... %q", fname)
	defer log("done reading")
	for {
		l, err := r.ReadString('\n')
		if err != nil {
			return
		}
		buffer.Append(l, log)
	}
}
