package terminal

import (
	"errors"
	"fmt"
	"log"
	"os"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)


func Die(err error) {
	term.Restore(int(os.Stdin.Fd()), E.Termios)
	if _, err := os.Stdout.Write([]byte("\x1b[2J")); err != nil {
		log.Fatal("Could not clean screen")
	}
	fmt.Printf("%v\n", err)
	os.Exit(1)
}

func ReadKey() int {
	var b []byte = make([]byte, 1)

	nread, err := os.Stdin.Read(b)
	if err != nil {
		Die(err)
	}

	if nread != 1 {
		Die(errors.New(fmt.Sprintf("Wanted to read one character, got %d", nread)))
	}

	if b[0] == '\x1b' {
		var seq []byte = make([]byte, 3)

		_, err := os.Stdin.Read(seq)
		if err != nil {
			Die(err)
		}

		if seq[0] == '[' {
			if seq[1] >= '0' && seq[1] <= '9' {
				if seq[2] == '~' {
					switch seq[1] {
					case '1':
						return HOME
					case '3':
						return DELETE
					case '4':
						return END
					case '5':
						return PAGE_UP
					case '6':
						return PAGE_DOWN
					case '7':
						return HOME
					case '8':
						return END
					}
				}
			} else {
				switch seq[1] {
				case 'A':
					return UP
				case 'B':
					return DOWN
				case 'C':
					return RIGHT
				case 'D':
					return LEFT
				case 'H':
					return HOME
				case 'F':
					return END
				}

			}
		} else if seq[0] == 'O' {
			switch seq[1] {
			case 'H':
				return HOME
			case 'F':
				return END
			}
		}

		return '\x1b'
	}

	return int(b[0])
}

func GetCursorPosition() (row int, col int, err error) {
	var buf []byte = make([]byte, 32)
	var i int = 0
	if _, err := os.Stdout.Write([]byte("\x1b[6n")); err != nil {
		return 0, 0, err
	}

	if _, err := os.Stdin.Read(buf); err != nil {
		return 0, 0, err
	}

	for i < len(buf) {
		if buf[i] == 'R' {
			break
		}
		i++
	}

	if buf[0] != '\x1b' || buf[1] != '[' {
		return 0, 0, errors.New("Could not find escape sequence when getting cursor position")
	}

	if _, err := fmt.Sscanf(string(buf[2:i]), "%d;%d", &row, &col); err != nil {
		return 0, 0, err
	}

	return row, col, nil
}

func GetWindowSize() (rows int, cols int, err error) {
	winSize, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}

	if winSize.Row < 1 || winSize.Col < 1 {
		_, err := os.Stdout.Write([]byte("\x1b[999C\x1b[999B"))
		if err != nil {
			return 0, 0, err
		}

		return GetCursorPosition()
	}
	return int(winSize.Row), int(winSize.Col), nil
}

