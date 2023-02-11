package io

import (
	"fmt"
	"os"
	"unicode"

	"golang.org/x/term"
	"github.com/jobutterfly/gowrite/editor"
	"github.com/jobutterfly/gowrite/operations"
	"github.com/jobutterfly/gowrite/terminal"
	"github.com/jobutterfly/gowrite/consts"
)


func Prompt(prompt string, callBack func([]byte, int)) string {
	var buf []byte

	for ;; {
		SetStatusMsg(fmt.Sprintf("%s%s", prompt, buf))
		RefreshScreen()

		var c int = terminal.ReadKey()
		// 8 is ctrl h
		if c == consts.DELETE || c == 8 || c == consts.BACKSPACE {
			if len(buf) != 0 {
				buf = buf[:len(buf) - 1]
			}
		} else if c == '\x1b' {
			// needing to press escape twice to exit
			SetStatusMsg("")
			if callBack != nil {
				callBack(buf, c)
			}
			return ""
		} else if c == '\r' {
			if len(buf) != 0 {
				SetStatusMsg("")
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
	var row *editor.Row
	var lastRow bool = false
	if editor.E.Cy >= editor.E.NumRows {
		lastRow = true
	} else {
		row = editor.E.Rows[editor.E.Cy]
	}

	switch key {
	case consts.LEFT:
		if editor.E.Cx != 0 {
			editor.E.Cx--
		} else if editor.E.Cy > 0 {
			editor.E.Cy--
			editor.E.Cx = editor.E.Rows[editor.E.Cy].Chars.Len()
		}
	case consts.RIGHT:
		if !lastRow {
			// we are going past the below condition
			if editor.E.Cx < row.Chars.Len() {
				editor.E.Cx++
			} else if editor.E.Cx == row.Chars.Len() {
				editor.E.Cy++
				editor.E.Cx = 0
			}
		}
	case consts.DOWN:
		if editor.E.Cy < editor.E.NumRows {
			editor.E.Cy++
		}
	case consts.UP:
		if editor.E.Cy != 0 {
			editor.E.Cy--
		}
	}

	var rowLen int = 0
	if editor.E.Cy >= editor.E.NumRows {
		lastRow = true
	} else {
		row = editor.E.Rows[editor.E.Cy]
		rowLen = editor.E.Rows[editor.E.Cy].Chars.Len()
	}
	if editor.E.Cx > rowLen {
		editor.E.Cx = rowLen
	}

}

func ProcessKeyPress(oldState *term.State) bool {
	var char int = terminal.ReadKey()

	switch char {
	// see references in readme for ascii control codes
	case '\r':
		operations.InsertNewLine()
	// ctrl q
	case 17:
		if editor.E.Dirty && editor.E.QuitTimes > 0 {
			SetStatusMsg(fmt.Sprintf("Warning! File has unsaved changes. Press Ctrl-Q %d times to quit anyway.", editor.E.QuitTimes))
			editor.E.QuitTimes--
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
		EditorSave()
	case consts.UP:
		MoveCursor(char)
	case consts.DOWN:
		MoveCursor(char)
	case consts.RIGHT:
		MoveCursor(char)
	case consts.LEFT:
		MoveCursor(char)
	case consts.PAGE_UP:
		editor.E.Cy = editor.E.RowOff
		for i := editor.E.ScreenRows; i > 1; i-- {
			MoveCursor(consts.UP)
		}
	case consts.PAGE_DOWN:
		editor.E.Cy = editor.E.RowOff + editor.E.ScreenRows - 1
		if editor.E.Cy > editor.E.NumRows {
			editor.E.Cy = editor.E.NumRows
		}
		for i := editor.E.ScreenRows; i > 1; i-- {
			MoveCursor(consts.DOWN)
		}
	case consts.BACKSPACE:
		operations.DeleteChar()
	// ctrl h
	case 8:
		operations.DeleteChar()
	case consts.DELETE:
		MoveCursor(consts.RIGHT)
		operations.DeleteChar()
	case consts.HOME:
		editor.E.Cx = 0
	case consts.END:
		if editor.E.Cy < editor.E.NumRows {
			editor.E.Cx = editor.E.Rows[editor.E.Cy].Chars.Len()
		}
	// ctrl f
	case 6:
		Find()
	// ctrl l
	case 12:
		break
	case '\x1b':
		break
	default:
		operations.InsertChar(byte(char))
	}

	editor.E.QuitTimes = consts.QUIT_TIMES 
	return false
}

