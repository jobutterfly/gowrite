package output

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

func Scroll() {
	E.Rx = 0
	if E.Cy < E.NumRows {
		E.Rx = operations.CxToRx(E.Rows[E.Cy], E.Cx)
	}

	if E.Cy < E.RowOff {
		E.RowOff = E.Cy
	}
	if E.Cy >= E.RowOff+E.ScreenRows {
		E.RowOff = E.Cy - E.ScreenRows + 1
	}
	if E.Rx < E.ColOff {
		E.ColOff = E.Rx
	}
	if E.Rx >= E.ColOff+E.ScreenCols {
		E.ColOff = E.Rx - E.ScreenCols + 1
	}
}

func DrawRows(buf *bytes.Buffer) error {
	for i := 0; i < E.ScreenRows; i++ {
		fileRow := i + E.RowOff
		if fileRow >= E.NumRows {
			if E.NumRows == 0 && i == E.ScreenRows/3 {
				if _, err := buf.Write([]byte("~")); err != nil {
					return err
				}
				welcome := fmt.Sprintf("gowrite version: %s", gowriteVersion)
				padding := (E.ScreenCols - len(welcome)) / 2
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
			length := E.ScreenCols + E.ColOff
			from := E.ColOff
			if length < 0 {
				length = 0
			}
			if length >= E.Rows[fileRow].Render.Len() {
				length = E.Rows[fileRow].Render.Len()
			}
			if from >= length {
				from = length
			}
			if _, err := buf.Write(E.Rows[fileRow].Render.Bytes()[from:length]); err != nil {
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

func DrawStatusBar(buf *bytes.Buffer) error {
	var FileName string = E.fileName
	var DirtyText string = ""
	if _, err := buf.Write([]byte("\x1b[7m")); err != nil {
		return err
	}
	if FileName == "" {
		FileName = "[No Name]"
	}
	if E.Dirty {
		DirtyText = "(modified)"
	} 

	status := fmt.Sprintf("%.20s - %d lines %s", FileName, E.NumRows, DirtyText)
	rowStatus := fmt.Sprintf("%d/%d", E.Cy+1, E.NumRows)
	length := len(status)
	rLength := len(rowStatus)
	if length > E.ScreenCols {
		length = E.ScreenCols
		status = status[:length]
	}

	if _, err := buf.Write([]byte(status)); err != nil {
		return err
	}

	for i := length; i < E.ScreenCols; {
		if E.ScreenCols-i == rLength {
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

func DrawMessageBar(buf *bytes.Buffer) error {
	if _, err := buf.Write([]byte("\x1b[K")); err != nil {
		return err
	}

	if E.StatusMsg != "" {
		msgLen := len(E.StatusMsg)

		if msgLen > E.ScreenCols {
			msgLen = E.ScreenCols
		}
		if (time.Now()).Unix() - E.StatusMsgTime.Unix() < 5 {
			if _, err := buf.Write([]byte(E.StatusMsg)); err != nil {
				return err
			}
		}
	}
	return nil
}

func RefreshScreen() error {
	Scroll()

	var mainBuf bytes.Buffer

	if _, err := mainBuf.Write([]byte("\x1b[?25l")); err != nil {
		return err
	}
	if _, err := mainBuf.Write([]byte("\x1b[H")); err != nil {
		return err
	}

	if err := DrawRows(&mainBuf); err != nil {
		return err
	}
	if err := DrawStatusBar(&mainBuf); err != nil {
		return err
	}
	if err := DrawMessageBar(&mainBuf); err != nil {
		return err
	}

	if _, err := mainBuf.Write([]byte(fmt.Sprintf("\x1b[%d;%dH", E.Cy-E.RowOff+1, E.Rx-E.ColOff+1))); err != nil {
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

func SetStatusMsg(msg string) {
	E.StatusMsg = msg
	E.StatusMsgTime = time.Now()
}

