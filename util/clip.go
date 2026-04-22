package util

import (
	"bytes"
	"errors"
	"os/exec"
)

// Clipboard support across common environments.
// macOS: pbcopy/pbpaste
// Linux/X11: xclip
func ClipCopy(s string) error {
	if p, err := exec.LookPath("pbcopy"); err == nil {
		cmd := exec.Command(p)
		cmd.Stdin = bytes.NewBufferString(s)
		return cmd.Run()
	}
	if p, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command(p)
		cmd.Stdin = bytes.NewBufferString(s)
		return cmd.Run()
	}
	return errors.New("no clipboard tool found (tried pbcopy and xclip)")
}

func ClipPaste() (string, error) {
	var cmd *exec.Cmd
	if p, err := exec.LookPath("pbpaste"); err == nil {
		cmd = exec.Command(p)
	} else if p, err := exec.LookPath("xclip"); err == nil {
		cmd = exec.Command(p, "-o")
	} else {
		return "", errors.New("no clipboard tool found (tried pbpaste and xclip)")
	}

	b := bytes.NewBuffer(nil)
	cmd.Stdout = b
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
