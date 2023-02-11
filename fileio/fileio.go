package fileio

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)


func RowsToString() string {
	var rowsArr [][]byte

	for _, r := range E.Rows {
		rowsArr = append(rowsArr, r.Chars.Bytes())
	}
	final := bytes.Join(rowsArr, []byte("\n"))
	final = append(final, '\n')

	return string(final)
}

func EditorOpen(FileName string) error {
	file, err := os.Open(FileName)
	if err != nil {
		return err
	}
	defer file.Close()
	E.FileName = fileName

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for i := 0; scanner.Scan(); i++ {
		if scanner.Err() != nil {
			terminal.Die(err)
		}
		if err := operations.InsertRow([]byte(scanner.Text()), E.NumRows); err != nil {
			return err
		}
	}
	E.Dirty = false

	return nil
}

func EditorSave() {
	if E.FileName == "" {
		E.FileName = input.Prompt("Save as: ", nil)
		if E.FileName == "" {
			output.SetStatusMsg("Save aborted")
			return
		}
	}

	buf := RowsToString()
	if err := os.WriteFile(E.FileName, []byte(buf), 0644); err != nil {
		output.SetStatusMsg(fmt.Sprintf("Can't save! i/o error: %v", err))
	}
	E.Dirty = false

	output.SetStatusMsg("bytes written to disk")
}

