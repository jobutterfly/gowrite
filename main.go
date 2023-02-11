package main

import (
	"os"

	"golang.org/x/term"
	"github.com/jobutterfly/gowrite/editor"
	"github.com/jobutterfly/gowrite/terminal"
	"github.com/jobutterfly/gowrite/io"
)


func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		terminal.Die(err)
	}
	editor.E.Termios = oldState
	defer term.Restore(int(os.Stdin.Fd()), editor.E.Termios)
	terminal.InitEditor()

	args := os.Args
	if len(args) >= 2 {
		if err := io.EditorOpen(args[1]); err != nil {
			terminal.Die(err)
		}
	}

	io.SetStatusMsg("HELP: Ctrl-Q = quit | Ctrl-S = save | Ctrl-F = find")

	for {
		if err := io.RefreshScreen(); err != nil {
			terminal.Die(err)
		}
		if io.ProcessKeyPress(oldState) {
			break
		}
	}

}
