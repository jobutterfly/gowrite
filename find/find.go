package find

import (
	"bytes"
)

func Find() {
	savedCx := E.Cx
	savedCy := E.Cy
	savedColOff := E.ColOff
	savedRowOff := E.RowOff

	lastMatch := -1
	direction := 1

	findCallback := func (query []byte, key int) {
		if key == '\r' || key == '\x1b' {
			lastMatch = -1
			return
		} else if key == RIGHT || key == DOWN {
			direction = 1
		} else if key == LEFT || key == UP {
			direction = -1
		} else {
			lastMatch = -1
			direction = 1
		}

		if lastMatch == -1 {
			direction = 1
		}
		current := lastMatch
		for i := 0; i < E.NumRows; i++ {
			current += direction
			if current == -1 {
				current = E.NumRows - 1
			} else if current == E.NumRows {
				current = 0
			}
			rowBytes := E.Rows[current].Render.Bytes()
			index := bytes.Index(rowBytes, query)
			if index >= 0 {
				lastMatch = current
				E.Cy = current
				E.Cx = operations.RxToCx(E.Rows[current], index)
				E.RowOff = E.NumRows
				break
			}
		}
	}

	query := input.Prompt("Search: ", findCallback)
	if query != "" {
		return
	} else {
		E.Cx = savedCx
		E.Cy = savedCy
		E.ColOff = savedColOff
		E.RowOff = savedRowOff
	}
}
