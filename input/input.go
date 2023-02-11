package input

import (
	"fmt"
	"os"
	"unicode"

	"golang.org/x/term"
)


func Prompt(prompt string, callBack func([]byte, int)) string {
	var buf []byte

	for ;; {
		output.SetStatusMsg(fmt.Sprintf("%s%s", prompt, buf))
		output.RefreshScreen()

		var c int = terminal.ReadKey()
		// 8 is ctrl h
		if c == DELETE || c == 8 || c == BACKSPACE {
			if len(buf) != 0 {
				buf = buf[:len(buf) - 1]
			}
		} else if c == '\x1b' {
			// needing to press escape twice to exit
			output.SetStatusMsg("")
			if callBack != nil {
				callBack(buf, c)
			}
			return ""
		} else if c == '\r' {
			if len(buf) != 0 {
				output.SetStatusMsg("")
				if callBack != nil {
					callBack(buf, c)
					return string(buf)
				}
				return string(buf)
			}
		} else if !unicode.IsControl(rune(c)) && c < 128 {
			buf = append(buf, byte(c))
		}
		if callBack != nil {
			callBack(buf, c)
		}
	}
}

func MoveCursor(key int) {
	var row *Row
	var lastRow bool = false
	if E.Cy >= E.NumRows {
		lastRow = true
	} else {
		row = E.Rows[E.Cy]
	}

	switch key {
	case LEFT:
		if E.Cx != 0 {
			E.Cx--
		} else if E.Cy > 0 {
			E.Cy--
			E.Cx = E.Rows[E.Cy].Chars.Len()
		}
	case RIGHT:
		if !lastRow {
			// we are going past the below condition
			if E.Cx < row.Chars.Len() {
				E.Cx++
			} else if E.Cx == row.Chars.Len() {
				E.Cy++
				E.Cx = 0
			}
		}
	case DOWN:
		if E.Cy < E.NumRows {
			E.Cy++
		}
	case UP:
		if E.Cy != 0 {
			E.Cy--
		}
	}

	var rowLen int = 0
	if E.Cy >= E.NumRows {
		lastRow = true
	} else {
		row = E.Rows[E.Cy]
		rowLen = E.Rows[E.Cy].Chars.Len()
	}
	if E.Cx > rowLen {
		E.Cx = rowLen
	}

}

func processKeyPress(oldState *term.State) bool {
	var char int = terminal.ReadKey()

	switch char {
	// see references in readme for ascii control codes
	case '\r':
		operations.InsertNewLine()
	// ctrl q
	case 17:
		if E.Dirty && E.QuitTimes > 0 {
			output.SetStatusMsg(fmt.Sprintf("Warning! File has unsaved changes. Press Ctrl-Q %d times to quit anyway.", E.QuitTimes))
			E.QuitTimes--
			return false
		}
		if _, err := os.Stdout.Write([]byte("\x1b[2J")); err != nil {
			terminal.Die(err)
		}
		if _, err := os.Stdout.Write([]byte("\x1b[H")); err != nil {
			terminal.Die(err)
		}
		return true
	// ctrl s
	case 19: 
		fileio.EditorSave()
	case UP:
		MoveCursor(char)
	case DOWN:
		MoveCursor(char)
	case RIGHT:
		MoveCursor(char)
	case LEFT:
		MoveCursor(char)
	case PAGE_UP:
		E.Cy = E.RowOff
		for i := E.ScreenRows; i > 1; i-- {
			MoveCursor(UP)
		}
	case PAGE_DOWN:
		E.Cy = E.RowOff + E.ScreenRows - 1
		if E.Cy > E.NumRows {
			E.Cy = E.NumRows
		}
		for i := E.ScreenRows; i > 1; i-- {
			MoveCursor(DOWN)
		}
	case BACKSPACE:
		operations.DeleteChar()
	// ctrl h
	case 8:
		operations.DeleteChar()
	case DELETE:
		MoveCursor(RIGHT)
		operations.DeleteChar()
	case HOME:
		E.Cx = 0
	case END:
		if E.Cy < E.NumRows {
			E.Cx = E.Rows[E.Cy].Chars.Len()
		}
	// ctrl f
	case 6:
		find.Find()
	// ctrl l
	case 12:
		break
	case '\x1b':
		break
	default:
		operations.InsertChar(byte(char))
	}

	E.QuitTimes = QUIT_TIMES 
	return false
}

