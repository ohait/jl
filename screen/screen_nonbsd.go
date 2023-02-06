//go:build !darwin && !freebsd && !netbsd && !openbsd && !windows
// +build !darwin,!freebsd,!netbsd,!openbsd,!windows

package screen

import (
	"golang.org/x/sys/unix"
)

const (
	GET_TERMIOS = unix.TCGETS
	SET_TERMIOS = unix.TCSETS
)
