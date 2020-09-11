package logger

import (
	"golang.org/x/sys/unix"
)

func isTerminal(fd int) bool {
	_, err := unix.IoctlGetTermio(fd, unix.TCGETA)
	return err == nil
}
