package editor


import (
	"bytes"
	"time"

	"golang.org/x/term"
	"github.com/jobutterlfy/gowrite/terminal"
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

func InitEditor() {
	E.Cx = 0
	E.Cy = 0
	E.Rx = 0
	rows, cols, err := terminal.GetWindowSize()
	if err != nil {
		terminal.Die(err)
	}
	E.ScreenRows = rows
	E.ScreenCols = cols
	E.NumRows = 0
	E.RowOff = 0
	E.ColOff = 0
	E.Rows = []*Row{}
	E.Dirty = false
	E.FileName = ""
	E.StatusMsg = ""
	E.StatusMsgTime = time.Now()

	E.ScreenRows -= 2
}
