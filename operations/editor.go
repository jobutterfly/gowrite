package operations

import (
	"bytes"
)

func InsertChar(c byte) {
	if E.Cy == E.NumRows {
		InsertRow([]byte(""), E.NumRows)
	}
	RowInsertChar(E.Rows[E.Cy], E.Cx, c)
	E.Cx++
}

func InsertNewLine() {
	if E.Cx == 0 {
		InsertRow([]byte(""), E.Cy)
	} else {
		InsertRow((E.Rows[E.Cy].Chars.Bytes())[E.Cx:], E.cy + 1)
		E.Rows[E.Cy].Chars = bytes.NewBuffer((E.Rows[E.cy].Chars.Bytes())[:E.Cx])
		UpdateRow(E.Rows[E.Cy])
	}
	E.Cy++
	E.Cx = 0
}

func DeleteChar() {
	if E.Cy == E.NumRows {
		return
	}
	if E.Cx == 0 && E.Cy == 0 {
		return
	}
	if E.Cx > 0 {
		RowDeleteChar(E.Rows[E.Cy], E.Cx)
		E.Cx--
	} else {
		E.Cx = E.Rows[E.Cy - 1].Chars.Len()
		RowAppendBytes(E.Rows[E.Cy - 1], E.Rows[E.cy].Chars.Bytes())
		DeleteRow(E.Cy)
		E.Cy--
	}
}

