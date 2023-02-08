//go:build darwin || freebsd || openbsd || netbsd
// +build darwin freebsd openbsd netbsd

package screen

import (
	"golang.org/x/sys/unix"
)

const (
	GET_TERMIOS = unix.TIOCGETA
	SET_TERMIOS = unix.TIOCSETA
)
