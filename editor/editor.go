package editor


import (
	"bytes"
	"time"

	"golang.org/x/term"
)


type editorConfig struct {
	Cx            int
	Cy            int
	Rx            int
	RowOff        int
	ColOff        int
	Termios       *term.State
	ScreenRows    int
	ScreenCols    int
	NumRows       int
	Rows          []*Row
	Dirty         bool
	FileName      string
	StatusMsg     string
	StatusMsgTime time.Time
	QuitTimes     int
}

type Row struct {
	Chars  *bytes.Buffer
	Render *bytes.Buffer
}

var E editorConfig
