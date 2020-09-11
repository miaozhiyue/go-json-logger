// +build js nacl plan9

package logger

import (
	"io"
)

func checkIfTerminal(w io.Writer) bool {
	return false
}
