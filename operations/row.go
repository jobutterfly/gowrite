package operations


import (
	"bytes"
	"strings"

	"github.com/jobutterfly/gowrite/editor"
)

func CxToRx(row *Row, cx int) int {
	var rx int = 0
	buf := row.Chars.Bytes()
	for i := 0; i < cx; i++ {
		if buf[i] == '\t' {
			rx += (tabStopSize - 1)
		}
		rx++
	}
	return rx
}

func RxToCx(row *Row, rx int) int {
	var cur_rx int = 0
	var cx int = 0
	buf := row.Chars.Bytes()

	for ; cx < row.Chars.Len(); cx++ {
		if buf[cx] == '\t' {
			cur_rx += tabStopSize - 1
		}
		cur_rx++
		if cur_rx > rx {
			return cx
		}
	}

	return cx
}

func UpdateRow(row *Row) {
	var newTab string = ""

	for i := 0; i < tabStopSize; i++ {
		newTab = newTab + " "
	}
	chars := row.Chars.String()
	splitChars := strings.Split(chars, "\t")
	newChars := strings.Join(splitChars, newTab)
	row.Render = bytes.NewBufferString(newChars)
}

func InsertRow(s []byte, at int) error {
	n := &row{
		Chars: bytes.NewBuffer(s),
	}
	if E.NumRows == at {
		E.Rows = append(E.Rows, n)
	} else {
		E.Rows = append(E.Rows[:at+1], E.Rows[at:]...)
		E.Rows[at] = n
	}
	UpdateRow(E.Rows[at])
	E.NumRows++
	E.Dirty = true

	return nil
}

func DeleteRow(at int) {
	if at < 0 || at > E.NumRows {
		return
	}
	E.Rows = append(E.Rows[:at], E.Rows[at+1:]...)
	E.NumRows--
	E.Dirty = true
}

func RowInsertChar(row *Row, at int, c byte) {
	var newC []byte
	if at < 0 || at > row.Chars.Len() {
		at = row.Chars.Len()
	}
	newC = append(newC, c)
	old := row.Chars.Bytes()
	joinBytes := [][]byte{old[:at], old[at:]}
	row.Chars = bytes.NewBuffer(bytes.Join(joinBytes, newC))
	UpdateRow(row)
	E.Dirty = true
}

func RowAppendBytes(row *Row, b []byte) {
	old := row.Chars.Bytes()
	row.Chars = bytes.NewBuffer(bytes.Join([][]byte{old, b}, []byte("")))
	UpdateRow(row)
}

func RowDeleteChar(row *Row, at int) {
	if at < 0 || at > row.Chars.Len() {
		return
	}

	old := row.Chars.Bytes()
	joinBytes := [][]byte{old[:at-1], old[at:]}
	row.Chars = bytes.NewBuffer(bytes.Join(joinBytes, []byte("")))
	UpdateRow(row)
	E.Dirty = true
}
