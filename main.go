package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/unix"
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

type editorConfig struct {
	cx            int
	cy            int
	rx            int
	rowOff        int
	colOff        int
	termios       *term.State
	screenRows    int
	screenCols    int
	numRows       int
	rows          []*row
	dirty         bool
	fileName      string
	statusMsg     string
	statusMsgTime time.Time
	quitTimes     int
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
			rx += (tabStopSize - 1)
		}
		rx++
	}
	return rx
}

func updateRow(row *row) {
	var newTab string = ""

	for i := 0; i < tabStopSize; i++ {
		newTab = newTab + " "
	}
	chars := row.chars.String()
	splitChars := strings.Split(chars, "\t")
	newChars := strings.Join(splitChars, newTab)
	row.render = bytes.NewBufferString(newChars)
}

func appendRow(s []byte) error {
	n := &row{
		chars: bytes.NewBuffer(s),
	}
	E.rows = append(E.rows, n)
	updateRow(E.rows[len(E.rows)-1])
	E.numRows++
	E.dirty = true

	return nil
}

func deleteRow(at int) {
	var newRows []*row = E.rows
	if at < 0 || at > E.numRows {
		return
	}
	E.rows = append(newRows[:at], newRows[at+1:]...)
	E.numRows--
	E.dirty = true
}

func rowInsertChar(row *row, at int, c byte) {
	var newC []byte
	if at < 0 || at > row.chars.Len() {
		at = row.chars.Len()
	}
	newC = append(newC, c)
	old := row.chars.Bytes()
	joinBytes := [][]byte{old[:at], old[at:]}
	row.chars = bytes.NewBuffer(bytes.Join(joinBytes, newC))
	updateRow(row)
	E.dirty = true
}

func rowAppendBytes(row *row, b []byte) {
	old := row.chars.Bytes()
	row.chars = bytes.NewBuffer(bytes.Join([][]byte{old, b}, []byte("")))
	updateRow(row)
}

func rowDeleteChar(row *row, at int) {
	if at < 0 || at > row.chars.Len() {
		return
	}

	old := row.chars.Bytes()
	joinBytes := [][]byte{old[:at-1], old[at:]}
	row.chars = bytes.NewBuffer(bytes.Join(joinBytes, []byte("")))
	updateRow(row)
	E.dirty = true
}

// editor operations

func insertChar(c byte) {
	if E.cy == E.numRows {
		appendRow([]byte(""))
	}
	rowInsertChar(E.rows[E.cy], E.cx, c)
	E.cx++
}

func deleteChar() {
	if E.cy == E.numRows {
		return
	}
	if E.cx == 0 && E.cy == 0 {
		return
	}
	if E.cx > 0 {
		rowDeleteChar(E.rows[E.cy], E.cx)
		E.cx--
	} else {
		E.cx = E.rows[E.cy - 1].chars.Len()
		rowAppendBytes(E.rows[E.cy - 1], E.rows[E.cy].chars.Bytes())
		deleteRow(E.cy)
		E.cy--
	}
}

// file i/o

func rowsToString() string {
	var rowsArr [][]byte

	for _, r := range E.rows {
		rowsArr = append(rowsArr, r.chars.Bytes())
	}
	final := bytes.Join(rowsArr, []byte("\n"))
	final = append(final, '\n')

	return string(final)
}

func editorOpen(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	E.fileName = fileName

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for i := 0; scanner.Scan(); i++ {
		if scanner.Err() != nil {
			die(err)
		}
		if err := appendRow([]byte(scanner.Text())); err != nil {
			return err
		}
	}
	E.dirty = false

	return nil
}

func editorSave() {
	if E.fileName == "" {
	}

	buf := rowsToString()
	if err := os.WriteFile(E.fileName, []byte(buf), 0644); err != nil {
		setStatusMsg(fmt.Sprintf("Can't save! i/o error: %v", err))
	}
	E.dirty = false

	setStatusMsg("bytes written to disk")
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
		if _, err := buf.Write([]byte("\r\n")); err != nil {
			return err
		}
	}

	return nil
}

func drawStatusBar(buf *bytes.Buffer) error {
	var fileName string = E.fileName
	var dirtyText string = ""
	if _, err := buf.Write([]byte("\x1b[7m")); err != nil {
		return err
	}
	if fileName == "" {
		fileName = "[No Name]"
	}
	if E.dirty {
		dirtyText = "(modified)"
	} 

	status := fmt.Sprintf("%.20s - %d lines %s", fileName, E.numRows, dirtyText)
	rowStatus := fmt.Sprintf("%d/%d", E.cy+1, E.numRows)
	length := len(status)
	rLength := len(rowStatus)
	if length > E.screenCols {
		length = E.screenCols
		status = status[:length]
	}

	if _, err := buf.Write([]byte(status)); err != nil {
		return err
	}
	// die(errors.New(fmt.Sprintf("legnth: %d\nrLength: %d\nE.screenCols: %d\nrowStatus: %s\n", length, rLength, E.screenCols, rowStatus)))

	for i := length; i < E.screenCols; {
		if E.screenCols-i == rLength {
			if _, err := buf.Write([]byte(rowStatus)); err != nil {
				return err
			}
			break
		} else {
			if _, err := buf.Write([]byte(" ")); err != nil {
				return err
			}
			i++
		}
	}

	if _, err := buf.Write([]byte("\x1b[m")); err != nil {
		return err
	}
	if _, err := buf.Write([]byte("\r\n")); err != nil {
		return err
	}

	return nil
}

func drawMessageBar(buf *bytes.Buffer) error {
	if _, err := buf.Write([]byte("\x1b[K")); err != nil {
		return err
	}

	if E.statusMsg != "" {
		msgLen := len(E.statusMsg)

		if msgLen > E.screenCols {
			msgLen = E.screenCols
		}
		if (time.Now()).Unix() - E.statusMsgTime.Unix() < 5 {
			if _, err := buf.Write([]byte(E.statusMsg)); err != nil {
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

	if err := drawRows(&mainBuf); err != nil {
		return err
	}
	if err := drawStatusBar(&mainBuf); err != nil {
		return err
	}
	if err := drawMessageBar(&mainBuf); err != nil {
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

func setStatusMsg(msg string) {
	E.statusMsg = msg
	E.statusMsgTime = time.Now()
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
	case '\r':
	// todo
	// ctrl q
	case 17:
		if E.dirty && E.quitTimes > 0 {
			setStatusMsg(fmt.Sprintf("Warning! File has unsaved changes. Press Ctrl-Q %d times to quit anyway.", E.quitTimes))
			E.quitTimes--
			return false
		}
		if _, err := os.Stdout.Write([]byte("\x1b[2J")); err != nil {
			die(err)
		}
		if _, err := os.Stdout.Write([]byte("\x1b[H")); err != nil {
			die(err)
		}
		return true
	case 19: 
		editorSave()
	case UP:
		moveCursor(char)
	case DOWN:
		moveCursor(char)
	case RIGHT:
		moveCursor(char)
	case LEFT:
		moveCursor(char)
	case PAGE_UP:
		E.cy = E.rowOff
		for i := E.screenRows; i > 1; i-- {
			moveCursor(UP)
		}
	case PAGE_DOWN:
		E.cy = E.rowOff + E.screenRows - 1
		if E.cy > E.numRows {
			E.cy = E.numRows
		}
		for i := E.screenRows; i > 1; i-- {
			moveCursor(DOWN)
		}
	case BACKSPACE:
		deleteChar()
	// ctrl h
	case 8:
		deleteChar()
	case DELETE:
		moveCursor(RIGHT)
		deleteChar()
	case HOME:
		E.cx = 0
	case END:
		if E.cy < E.numRows {
			E.cx = E.rows[E.cy].chars.Len()
		}
	// ctrl l
	case 12:
		break
	case '\x1b':
		break
	default:
		insertChar(byte(char))
	}

	E.quitTimes = QUIT_TIMES 
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
	E.dirty = false
	E.fileName = ""
	E.statusMsg = ""
	E.statusMsgTime = time.Now()

	E.screenRows -= 2
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

	setStatusMsg("HELP: Ctrl-Q = quit | Ctrl-S = save")

	for {
		if err := refreshScreen(); err != nil {
			die(err)
		}
		if processKeyPress(oldState) {
			break
		}
	}

}
