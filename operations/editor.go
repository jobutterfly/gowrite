package operations

import (
	"bytes"

	"github.com/jobutterlfy/gowrite/editor"
)

func InsertChar(c byte) {
	if editor.E.Cy == editor.E.NumRows {
		InsertRow([]byte(""), editor.E.NumRows)
	}
	RowInsertChar(editor.E.Rows[editor.E.Cy], editor.E.Cx, c)
	editor.E.Cx++
}

func InsertNewLine() {
	if editor.E.Cx == 0 {
		InsertRow([]byte(""), editor.E.Cy)
	} else {
		InsertRow((editor.E.Rows[editor.E.Cy].Chars.Bytes())[editor.E.Cx:], editor.E.Cy + 1)
		editor.E.Rows[editor.E.Cy].Chars = bytes.NewBuffer((editor.E.Rows[editor.E.Cy].Chars.Bytes())[:editor.E.Cx])
		UpdateRow(editor.E.Rows[editor.E.Cy])
	}
	editor.E.Cy++
	editor.E.Cx = 0
}

func DeleteChar() {
	if editor.E.Cy == editor.E.NumRows {
		return
	}
	if editor.E.Cx == 0 && editor.E.Cy == 0 {
		return
	}
	if editor.E.Cx > 0 {
		RowDeleteChar(editor.E.Rows[editor.E.Cy], editor.E.Cx)
		editor.E.Cx--
	} else {
		editor.E.Cx = editor.E.Rows[editor.E.Cy - 1].Chars.Len()
		RowAppendBytes(editor.E.Rows[editor.E.Cy - 1], editor.E.Rows[editor.E.Cy].Chars.Bytes())
		DeleteRow(editor.E.Cy)
		editor.E.Cy--
	}
}

