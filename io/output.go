package io

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/jobutterfly/gowrite/editor"
	"github.com/jobutterfly/gowrite/operations"
	"github.com/jobutterfly/gowrite/consts"
)

func Scroll() {
	editor.E.Rx = 0
	if editor.E.Cy < editor.E.NumRows {
		editor.E.Rx = operations.CxToRx(editor.E.Rows[editor.E.Cy], editor.E.Cx)
	}

	if editor.E.Cy < editor.E.RowOff {
		editor.E.RowOff = editor.E.Cy
	}
	if editor.E.Cy >= editor.E.RowOff+editor.E.ScreenRows {
		editor.E.RowOff = editor.E.Cy - editor.E.ScreenRows + 1
	}
	if editor.E.Rx < editor.E.ColOff {
		editor.E.ColOff = editor.E.Rx
	}
	if editor.E.Rx >= editor.E.ColOff+editor.E.ScreenCols {
		editor.E.ColOff = editor.E.Rx - editor.E.ScreenCols + 1
	}
}

func DrawRows(buf *bytes.Buffer) error {
	for i := 0; i < editor.E.ScreenRows; i++ {
		fileRow := i + editor.E.RowOff
		if fileRow >= editor.E.NumRows {
			if editor.E.NumRows == 0 && i == editor.E.ScreenRows/3 {
				if _, err := buf.Write([]byte("~")); err != nil {
					return err
				}
				welcome := fmt.Sprintf("gowrite version: %s", consts.VERSION)
				padding := (editor.E.ScreenCols - len(welcome)) / 2
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
			length := editor.E.ScreenCols + editor.E.ColOff
			from := editor.E.ColOff
			if length < 0 {
				length = 0
			}
			if length >= editor.E.Rows[fileRow].Render.Len() {
				length = editor.E.Rows[fileRow].Render.Len()
			}
			if from >= length {
				from = length
			}
			if _, err := buf.Write(editor.E.Rows[fileRow].Render.Bytes()[from:length]); err != nil {
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
	var FileName string = editor.E.FileName
	var DirtyText string = ""
	if _, err := buf.Write([]byte("\x1b[7m")); err != nil {
		return err
	}
	if FileName == "" {
		FileName = "[No Name]"
	}
	if editor.E.Dirty {
		DirtyText = "(modified)"
	} 

	status := fmt.Sprintf("%.20s - %d lines %s", FileName, editor.E.NumRows, DirtyText)
	rowStatus := fmt.Sprintf("%d/%d", editor.E.Cy+1, editor.E.NumRows)
	length := len(status)
	rLength := len(rowStatus)
	if length > editor.E.ScreenCols {
		length = editor.E.ScreenCols
		status = status[:length]
	}

	if _, err := buf.Write([]byte(status)); err != nil {
		return err
	}

	for i := length; i < editor.E.ScreenCols; {
		if editor.E.ScreenCols-i == rLength {
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

	if editor.E.StatusMsg != "" {
		msgLen := len(editor.E.StatusMsg)

		if msgLen > editor.E.ScreenCols {
			msgLen = editor.E.ScreenCols
		}
		if (time.Now()).Unix() - editor.E.StatusMsgTime.Unix() < 5 {
			if _, err := buf.Write([]byte(editor.E.StatusMsg)); err != nil {
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

	if _, err := mainBuf.Write([]byte(fmt.Sprintf("\x1b[%d;%dH", editor.E.Cy-editor.E.RowOff+1, editor.E.Rx-editor.E.ColOff+1))); err != nil {
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
	editor.E.StatusMsg = msg
	editor.E.StatusMsgTime = time.Now()
}

