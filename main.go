package main

import (
	"os"

	"golang.org/x/term"
)

const (
	BACKSPACE int = 127
	LEFT      int = 1000
	RIGHT     int = 1001
	UP        int = 1002
	DOWN      int = 1003
	DELETE    int = 1004
	HOME      int = 1005
	END       int = 1006
	PAGE_UP   int = 1007
	PAGE_DOWN int = 1008
)

const gowriteVersion = "0.1"
const tabStopSize = 8
const QUIT_TIMES = 2

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		terminal.Die(err)
	}
	E.Termios = oldState
	defer term.Restore(int(os.Stdin.Fd()), E.Termios)
	initEditor()

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
