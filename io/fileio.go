package io

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/jobutterfly/gowrite/editor"
	"github.com/jobutterfly/gowrite/operations"
	"github.com/jobutterfly/gowrite/terminal"
)


func RowsToString() string {
	var rowsArr [][]byte

	for _, r := range editor.E.Rows {
		rowsArr = append(rowsArr, r.Chars.Bytes())
	}
	final := bytes.Join(rowsArr, []byte("\n"))
	final = append(final, '\n')

	return string(final)
}

func EditorOpen(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	editor.E.FileName = fileName

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for i := 0; scanner.Scan(); i++ {
		if scanner.Err() != nil {
			terminal.Die(err)
		}
		if err := operations.InsertRow([]byte(scanner.Text()), editor.E.NumRows); err != nil {
			return err
		}
	}
	editor.E.Dirty = false

	return nil
}

func EditorSave() {
	if editor.E.FileName == "" {
		editor.E.FileName = Prompt("Save as: ", nil)
		if editor.E.FileName == "" {
			SetStatusMsg("Save aborted")
			return
		}
	}

	buf := RowsToString()
	if err := os.WriteFile(editor.E.FileName, []byte(buf), 0644); err != nil {
		SetStatusMsg(fmt.Sprintf("Can't save! i/o error: %v", err))
	}
	editor.E.Dirty = false

	SetStatusMsg("bytes written to disk")
}

