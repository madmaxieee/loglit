//go:build development

package utils

import (
	"io"
	"os"

	"golang.org/x/term"
)

func HandleError(err error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		_, _ = io.ReadAll(os.Stdin)
	}
	panic(err)
}
