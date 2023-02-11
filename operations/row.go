package operations


import (
	"bytes"
	"strings"

	"github.com/jobutterfly/gowrite/consts"
	"github.com/jobutterfly/gowrite/editor"
)

func CxToRx(row *editor.Row, cx int) int {
	var rx int = 0
	buf := row.Chars.Bytes()
	for i := 0; i < cx; i++ {
		if buf[i] == '\t' {
			rx += (consts.tabStopSize - 1)
		}
		rx++
	}
	return rx
}

func RxToCx(row *editor.Row, rx int) int {
	var cur_rx int = 0
	var cx int = 0
	buf := row.Chars.Bytes()

	for ; cx < row.Chars.Len(); cx++ {
		if buf[cx] == '\t' {
			cur_rx += consts.tabStopSize - 1
		}
		cur_rx++
		if cur_rx > rx {
			return cx
		}
	}

	return cx
}

func UpdateRow(row *editor.Row) {
	var newTab string = ""

	for i := 0; i < consts.tabStopSize; i++ {
		newTab = newTab + " "
	}
	chars := row.Chars.String()
	splitChars := strings.Split(chars, "\t")
	newChars := strings.Join(splitChars, newTab)
	row.Render = bytes.NewBufferString(newChars)
}

func InsertRow(s []byte, at int) error {
	n := &editor.Row{
		Chars: bytes.NewBuffer(s),
	}
	if editor.E.NumRows == at {
		editor.E.Rows = append(editor.E.Rows, n)
	} else {
		editor.E.Rows = append(editor.E.Rows[:at+1], editor.E.Rows[at:]...)
		editor.E.Rows[at] = n
	}
	UpdateRow(editor.E.Rows[at])
	editor.E.NumRows++
	editor.E.Dirty = true

	return nil
}

func DeleteRow(at int) {
	if at < 0 || at > editor.E.NumRows {
		return
	}
	editor.E.Rows = append(editor.E.Rows[:at], editor.E.Rows[at+1:]...)
	editor.E.NumRows--
	editor.E.Dirty = true
}

func RowInsertChar(row *editor.Row, at int, c byte) {
	var newC []byte
	if at < 0 || at > row.Chars.Len() {
		at = row.Chars.Len()
	}
	newC = append(newC, c)
	old := row.Chars.Bytes()
	joinBytes := [][]byte{old[:at], old[at:]}
	row.Chars = bytes.NewBuffer(bytes.Join(joinBytes, newC))
	UpdateRow(row)
	editor.E.Dirty = true
}

func RowAppendBytes(row *editor.Row, b []byte) {
	old := row.Chars.Bytes()
	row.Chars = bytes.NewBuffer(bytes.Join([][]byte{old, b}, []byte("")))
	UpdateRow(row)
}

func RowDeleteChar(row *editor.Row, at int) {
	if at < 0 || at > row.Chars.Len() {
		return
	}

	old := row.Chars.Bytes()
	joinBytes := [][]byte{old[:at-1], old[at:]}
	row.Chars = bytes.NewBuffer(bytes.Join(joinBytes, []byte("")))
	UpdateRow(row)
	editor.E.Dirty = true
}
