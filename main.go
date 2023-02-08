package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

const (
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
const tabStop = "        "
const tabStopSize = 8

type editorConfig struct {
	cx         int
	cy         int
	rx         int
	rowOff     int
	colOff     int
	termios    *term.State
	screenRows int
	screenCols int
	numRows    int
	rows       []*row
}

type row struct {
	chars  *bytes.Buffer
	render *bytes.Buffer
}

var E editorConfig

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

		if seq[0] == '[' {
			if seq[1] >= '0' && seq[1] <= '9' {
				if seq[2] == '~' {
					switch seq[1] {
					case '1':
						return HOME
					case '3':
						return DELETE
					case '4':
						return END
					case '5':
						return PAGE_UP
					case '6':
						return PAGE_DOWN
					case '7':
						return HOME
					case '8':
						return END
					}
				}
			} else {
				switch seq[1] {
				case 'A':
					return UP
				case 'B':
					return DOWN
				case 'C':
					return RIGHT
				case 'D':
					return LEFT
				case 'H':
					return HOME
				case 'F':
					return END
				}

			}
		} else if seq[0] == 'O' {
			switch seq[1] {
			case 'H':
				return HOME
			case 'F':
				return END
			}
		}

		return '\x1b'
	}

	return int(b[0])
}

func getCursorPosition() (row int, col int, err error) {
	var buf []byte = make([]byte, 32)
	var i int = 0
	if _, err := os.Stdout.Write([]byte("\x1b[6n")); err != nil {
		return 0, 0, err
	}

	if _, err := os.Stdin.Read(buf); err != nil {
		return 0, 0, err
	}

	for i < len(buf) {
		if buf[i] == 'R' {
			break
		}
		i++
	}

	if buf[0] != '\x1b' || buf[1] != '[' {
		return 0, 0, errors.New("Could not find escape sequence when getting cursor position")
	}

	if _, err := fmt.Sscanf(string(buf[2:i]), "%d;%d", &row, &col); err != nil {
		return 0, 0, err
	}

	return row, col, nil
}

func getWindowSize() (rows int, cols int, err error) {
	winSize, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}

	if winSize.Row < 1 || winSize.Col < 1 {
		_, err := os.Stdout.Write([]byte("\x1b[999C\x1b[999B"))
		if err != nil {
			return 0, 0, err
		}

		return getCursorPosition()
	}
	return int(winSize.Row), int(winSize.Col), nil
}

// row operations

func cxToRx(row *row, cx int) int {
	var rx int = 0
	buf := row.chars.Bytes()
	for i := 0; i < cx; i++ {
		if buf[i] == '\t' {
			rx += tabStopSize - 1 - (rx % tabStopSize)
		}
		rx++
	}
	return rx
}

func updateRow(row *row) {
	chars := row.chars.String()
	splitChars := strings.Split(chars, "\t")
	newChars := strings.Join(splitChars, tabStop)
	row.render = bytes.NewBufferString(newChars)
}

func appendRow(s []byte) error {
	n := &row{
		chars: bytes.NewBuffer(s),
	}
	E.rows = append(E.rows, n)
	updateRow(E.rows[len(E.rows)-1])
	E.numRows++

	return nil
}

// file i/o

func editorOpen(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	i := 0

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for ; scanner.Scan(); i++ {
		if scanner.Err() != nil {
			die(err)
		}
		/*
		   Don't know why, but if scanner.Bytes() is called instead
		   what is show below, part of the first line of E.row becomes
		   overwritten by the last line of the file, with the black
		   magic below it doesn't happen. probably something to do
		   with memory
		*/
		if err := appendRow([]byte(scanner.Text())); err != nil {
			return err
		}
	}

	return nil
}

// output

func scroll() {
	E.rx = 0
	if E.cy < E.numRows {
		E.rx = cxToRx(E.rows[E.cy], E.cx)
	}

	if E.cy < E.rowOff {
		E.rowOff = E.cy
	}
	if E.cy >= E.rowOff+E.screenRows {
		E.rowOff = E.cy - E.screenRows + 1
	}
	if E.rx < E.colOff {
		E.colOff = E.rx
	}
	if E.rx >= E.colOff+E.screenCols {
		E.colOff = E.rx - E.screenCols + 1
	}
}

func drawRows(buf *bytes.Buffer) error {
	for i := 0; i < E.screenRows; i++ {
		fileRow := i + E.rowOff
		if fileRow >= E.numRows {
			if E.numRows == 0 && i == E.screenRows/3 {
				if _, err := buf.Write([]byte("~")); err != nil {
					return err
				}
				welcome := fmt.Sprintf("gowrite version: %s", gowriteVersion)
				padding := (E.screenCols - len(welcome)) / 2
				for ; padding > 1; padding-- {
					buf.Write([]byte(" "))
				}
				buf.Write([]byte(welcome))
			} else {
				if _, err := buf.Write([]byte("~")); err != nil {
					return err
				}
			}

		} else {
			length := E.screenCols + E.colOff
			from := E.colOff
			if length < 0 {
				length = 0
			}
			if length >= E.rows[fileRow].render.Len() {
				length = E.rows[fileRow].render.Len()
			}
			if from >= length {
				from = length
			}
			if _, err := buf.Write(E.rows[fileRow].render.Bytes()[from:length]); err != nil {
				return err
			}
		}

		if _, err := buf.Write([]byte("\x1b[K")); err != nil {
			return err
		}
		if i < E.screenRows-1 {
			if _, err := buf.Write([]byte("\r\n")); err != nil {
				return err
			}
		}
	}

	return nil
}

func refreshScreen() error {
	scroll()

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

	if _, err := mainBuf.Write([]byte(fmt.Sprintf("\x1b[%d;%dH", E.cy-E.rowOff+1, E.rx-E.colOff+1))); err != nil {
		return err
	}
	if _, err := mainBuf.Write([]byte("\x1b[?25h")); err != nil {
		return err
	}

	if _, err := os.Stdout.Write(mainBuf.Bytes()); err != nil {
		return err

	}

	return nil
}

// input

func moveCursor(key int) {
	var row *row
	var lastRow bool = false
	if E.cy >= E.numRows {
		lastRow = true
	} else {
		row = E.rows[E.cy]
	}

	switch key {
	case LEFT:
		if E.cx != 0 {
			E.cx--
		} else if E.cy > 0 {
			E.cy--
			E.cx = E.rows[E.cy].chars.Len()
		}
	case RIGHT:
		if !lastRow {
			// we are going past the below condition
			if E.cx < row.chars.Len() {
				E.cx++
			} else if E.cx == row.chars.Len() {
				E.cy++
				E.cx = 0
			}
		}
	case DOWN:
		if E.cy < E.numRows {
			E.cy++
		}
	case UP:
		if E.cy != 0 {
			E.cy--
		}
	}

	var rowLen int = 0
	if E.cy >= E.numRows {
		lastRow = true
	} else {
		row = E.rows[E.cy]
		rowLen = E.rows[E.cy].chars.Len()
	}
	if E.cx > rowLen {
		E.cx = rowLen
	}

}

func processKeyPress(oldState *term.State) bool {
	var char int = readKey()

	switch char {
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
		for i := E.screenRows; i > 1; i-- {
			moveCursor(UP)
		}
	case PAGE_DOWN:
		for i := E.screenRows; i > 1; i-- {
			moveCursor(DOWN)
		}
	case HOME:
		E.cx = 0
	case END:
		E.cx = E.screenCols - 1
	}

	return false
}

// init

func initEditor() {
	E.cx = 0
	E.cy = 0
	E.rx = 0
	rows, cols, err := getWindowSize()
	if err != nil {
		die(err)
	}
	E.screenRows = rows
	E.screenCols = cols
	E.numRows = 0
	E.rowOff = 0
	E.colOff = 0
	E.rows = []*row{}
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		die(err)
	}
	E.termios = oldState
	defer term.Restore(int(os.Stdin.Fd()), E.termios)
	initEditor()

	args := os.Args
	if len(args) >= 2 {
		if err := editorOpen(args[1]); err != nil {
			die(err)
		}
	}

	for {
		if err := refreshScreen(); err != nil {
			die(err)
		}
		if processKeyPress(oldState) {
			break
		}
	}

}
