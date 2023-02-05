package main

import (
	"bytes"
	"os"
	"errors"
	"fmt"
	"log"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

const (
	LEFT int = 1000 
	RIGHT int = 1001 
	UP int = 1002 
	DOWN int = 1003
	DELETE int = 1004
	HOME int = 1005
	END int = 1006
	PAGE_UP int = 1007
	PAGE_DOWN int = 1008
)

const gowriteVersion = "0.1"

type EditorConfig struct {
	cx	int
	cy	int
	termios *term.State
	winSize	*unix.Winsize
}

var E EditorConfig;


// terminal

func die(err error) {
	term.Restore(int(os.Stdin.Fd()), E.termios)
	if _, err := os.Stdout.Write([]byte("\x1b[2J")); err != nil {
	    log.Fatal("Could not clean screen")
	}
	fmt.Printf("%v\n", err)
	os.Exit(1)
}

func readKey() int {
	var b []byte = make([]byte, 1)

	nread, err := os.Stdin.Read(b)
	if err != nil {
	    die(err)
	}

	if nread != 1 {
	    die(errors.New(fmt.Sprintf("Wanted to read one character, got %d", nread)))
	}

	if b[0] == '\x1b' {
	    var seq []byte = make([]byte, 3)

	    _, err := os.Stdin.Read(seq)
	    if err != nil {
		die(err)
	    }

	    if seq[0] == '['{
		if seq[1] >= '0' && seq[1] <= '9' {
		    if seq[2] == '~' {
			switch seq[1] {
			case '1': return HOME
			case '3': return DELETE
			case '4': return END
			case '5': return PAGE_UP
			case '6': return PAGE_DOWN
			case '7': return HOME
			case '8': return END
			}
		    }
		} else {
		    switch seq[1] {
		    case 'A': return UP
		    case 'B': return DOWN
		    case 'C': return RIGHT
		    case 'D': return LEFT
		    case 'H': return HOME
		    case 'F': return END
		    }

		}
	    } else if seq[0] == 'O' {
		switch seq[1] {
		case 'H': return HOME
		case 'F': return END
		}
	    }

	    return '\x1b'
	} 

	return int(b[0])
}

func getCursorPosition() (*unix.Winsize, error) {
	var row, col int
	var buf []byte = make([]byte, 32)
	var i int = 0
	_, err := os.Stdout.Write([]byte("\x1b[6n"))
	if err != nil {
	    return nil, err
	}
	
	if _, err := os.Stdin.Read(buf); err != nil {
	    return nil, err
	}

	for ;i < len(buf); {
	    if buf[i] == 'R'{
		break
	    }
	    i++
	}

	if buf[0] != '\x1b' || buf[1] != '['  {
	    return nil, errors.New("Could not find escape sequence when getting cursor position")
	}

	if _, err := fmt.Sscanf(string(buf[2:i]), "%d;%d", &row, &col); err != nil {
	    return nil, err
	}

	winSize := unix.Winsize{
	    Row: uint16(row),
	    Col: uint16(col),
	}

	readKey()

	return &winSize, nil
}

func getWindowSize() (*unix.Winsize, error){
	winSize, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
	    return nil, err
	}

	if (winSize.Row < 1 || winSize.Col < 1) {
	    _, err := os.Stdout.Write([]byte("\x1b[999C\x1b[999B"))
	    if err != nil {
		return nil, err
	    }

	    return getCursorPosition()

	}
	return winSize, nil
}


// input

func moveCursor(key int) {
	switch (key){
	case LEFT: 
	    if (E.cx  != 0) {
		E.cx--
	    }
	case RIGHT:
	    if (E.cx != int(E.winSize.Col) - 1) {
		E.cx++
	    }
	case DOWN:
	    if (E.cy != int(E.winSize.Row) - 1) {
		E.cy++
	    }
	case UP:
	    if (E.cy != 0) {
		E.cy--
	    }
	}
}

func processKeyPress(oldState *term.State) bool{
	var char int = readKey()

	switch (char) {
	// see references in readme for ascii control codes
	// control q
	case 17:
	    if _, err := os.Stdout.Write([]byte("\x1b[2J")); err != nil {
		die(err)
	    }
	    if _, err := os.Stdout.Write([]byte("\x1b[H")); err != nil {
		die(err)
	    }
	    return true
	case UP:
	    moveCursor(char)
	case DOWN:
	    moveCursor(char)
	case RIGHT:
	    moveCursor(char)
	case LEFT:
	    moveCursor(char)
	case PAGE_UP: 
	for i:= E.winSize.Row ;i > 1; i-- {
	    moveCursor(UP)
	}
	case PAGE_DOWN:
	for i:= E.winSize.Row ;i > 1; i-- {
	    moveCursor(DOWN)
	}
	case HOME:
	    E.cx = 0;
	case END:
	    E.cx = int(E.winSize.Col) - 1
	}

	return false
}

// output

func drawRows(mainBuf *bytes.Buffer) error {
	for i := 0; i < int(E.winSize.Row) ; i++ {
	    if _, err := mainBuf.Write([]byte("~")); err != nil {
		return err
	    }

	    if i == int(E.winSize.Row) / 3 {
		welcome := fmt.Sprintf("gowrite version: %s", gowriteVersion)
		padding := (int(E.winSize.Col) - len(welcome)) / 2
		for ;padding > 1; padding-- {
		    mainBuf.Write([]byte(" "))
		}
		mainBuf.Write([]byte(welcome))
	    } 

	    if _, err := mainBuf.Write([]byte("\x1b[K")); err != nil {
		return err
	    }
	    if i < int(E.winSize.Row) - 1 {
		if _, err := mainBuf.Write([]byte("\r\n")); err != nil {
		    return err
		}
	    }
	}

	return nil
}

func refreshScreen() error{
	var mainBuf bytes.Buffer

	if _, err := mainBuf.Write([]byte("\x1b[?25l")); err != nil {
	    return err
	}
	if _, err := mainBuf.Write([]byte("\x1b[H")); err != nil {
	    return err
	}

	err := drawRows(&mainBuf)
	if err != nil {
	    return err
	}

	if _, err := mainBuf.Write([]byte(fmt.Sprintf("\x1b[%d;%dH", E.cy + 1, E.cx + 1))); err != nil {
	    return err
	}
	if _, err := mainBuf.Write([]byte("\x1b[?25h")); err != nil {
	    return err
	}

	if _, err := os.Stdout.Write(mainBuf.Bytes()); err != nil {
	    return err

	}
	mainBuf.Reset()

	return nil
}

// init

func initEditor() {
	E.cx = 0
	E.cy = 0
	winSize, err := getWindowSize()
	if err != nil {
	    die(err)
	}
	E.winSize = winSize
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
	    die(err)
	}
	E.termios = oldState
	defer term.Restore(int(os.Stdin.Fd()), E.termios)
	initEditor()

	for {
	    if err := refreshScreen(); err != nil {
		die(err)
	    }
	    if processKeyPress(oldState) {
		break
	    }
	}

}










