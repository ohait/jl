package util

import (
	"bytes"
	"os/exec"
)

// very crude implementation using exec, as long as it works.
// Requires xclip to be installed
func ClipCopy(s string) error {
	cmd := exec.Command("xclip")
	cmd.Stdin = bytes.NewBufferString(s)
	return cmd.Run()
}

func ClipPaste() (string, error) {
	cmd := exec.Command("xclip", "-o")
	b := bytes.NewBuffer(nil)
	cmd.Stdout = b
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
