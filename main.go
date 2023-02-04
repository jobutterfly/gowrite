package main

import (
	"fmt"
	"os"
	"log"
	"unicode"

	"golang.org/x/term"
)

func main() {
	var b []byte = make([]byte, 1)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	for {
		_, err := os.Stdin.Read(b)
		if err != nil {
			log.Fatal(err)
		}

		if b[0] < unicode.MaxASCII {
		    fmt.Printf("%d ('%c') \r\n", b[0], b[0])
		} else {
		    fmt.Printf("%d \r\n", b[0])
		}

		if b[0] == 'q'{
		    break
		}
	}
}










