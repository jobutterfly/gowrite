package io

import (
	"bytes"

	"github.com/jobutterlfy/gowrite/editor"
	"github.com/jobutterlfy/gowrite/operations"
	"github.com/jobutterlfy/gowrite/consts"
)

func Find() {
	savedCx := editor.E.Cx
	savedCy := editor.E.Cy
	savedColOff := editor.E.ColOff
	savedRowOff := editor.E.RowOff

	lastMatch := -1
	direction := 1

	findCallback := func (query []byte, key int) {
		if key == '\r' || key == '\x1b' {
			lastMatch = -1
			return
		} else if key == consts.RIGHT || key == consts.DOWN {
			direction = 1
		} else if key == consts.LEFT || key == consts.UP {
			direction = -1
		} else {
			lastMatch = -1
			direction = 1
		}

		if lastMatch == -1 {
			direction = 1
		}
		current := lastMatch
		for i := 0; i < editor.E.NumRows; i++ {
			current += direction
			if current == -1 {
				current = editor.E.NumRows - 1
			} else if current == editor.E.NumRows {
				current = 0
			}
			rowBytes := editor.E.Rows[current].Render.Bytes()
			index := bytes.Index(rowBytes, query)
			if index >= 0 {
				lastMatch = current
				editor.E.Cy = current
				editor.E.Cx = operations.RxToCx(editor.E.Rows[current], index)
				editor.E.RowOff = editor.E.NumRows
				break
			}
		}
	}

	query := Prompt("Search: ", findCallback)
	if query != "" {
		return
	} else {
		editor.E.Cx = savedCx
		editor.E.Cy = savedCy
		editor.E.ColOff = savedColOff
		editor.E.RowOff = savedRowOff
	}
}
