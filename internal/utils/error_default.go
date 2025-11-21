//go:build !development

package utils

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

func HandleError(err error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		_, _ = io.ReadAll(os.Stdin)
	}
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	os.Exit(1)
}
