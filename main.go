package main

import (
	"os"

	"golang.org/x/term"
	"github.com/jobutterlfy/gowrite/editor"
)


func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		terminal.Die(err)
	}
	E.Termios = oldState
	defer term.Restore(int(os.Stdin.Fd()), E.Termios)
	editor.InitEditor()

	args := os.Args
	if len(args) >= 2 {
		if err := fileio.EditorOpen(args[1]); err != nil {
			terminal.Die(err)
		}
	}

	output.SetStatusMsg("HELP: Ctrl-Q = quit | Ctrl-S = save | Ctrl-F = find")

	for {
		if err := RefreshScreen(); err != nil {
			terminal.Die(err)
		}
		if processKeyPress(oldState) {
			break
		}
	}

}
