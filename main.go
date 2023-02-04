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
	fmt.Printf("%v\n", err)
	if _, err := os.Stdout.Write([]byte("\x1b[2J")); err != nil {
	    log.Fatal("Could not clean screen")
	    
	}
	os.Exit(1)
}

func readKey() byte {
	var b []byte = make([]byte, 1)

	nread, err := os.Stdin.Read(b)
	if err != nil {
	    die(err)
	}

	if nread != 1 {
	    die(errors.New(fmt.Sprintf("Wanted to read one character, got %d", nread)))
	}

	return b[0]
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

func processKeyPress(oldState *term.State) bool{
	var char byte = readKey()

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
	}

	return false
}

// output

func drawRows(mainBuf bytes.Buffer) error {
	for i := 0; i < int(E.winSize.Row) ; i++ {
	    if i == int(E.winSize.Row) / 3 {
		welcome := fmt.Sprintf("gowrite version: %s", gowriteVersion)
		mainBuf.Write([]byte(welcome))
	    } else {
		if _, err := mainBuf.Write([]byte("~")); err != nil {
		    return err
		}
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

	err := drawRows(mainBuf)
	if err != nil {
	    return err
	}
	if _, err := mainBuf.Write([]byte("\x1b[H")); err != nil {
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










