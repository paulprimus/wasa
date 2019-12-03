package main

import (
	//"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"github.com/mattn/go-tty"
)

var esc string = "\x1b"
var runeA rune = '\x3d'
var runeLf rune = 0xA
var runeCarriageReturn rune = 0xD
var cmdSend []rune = []rune{0x73, 0x65, 0x6E, 0x64, 0xD}
var cmdExit []rune = []rune{0x65, 0x78, 0x69, 0x74, 0xD}

func main() {

	fmt.Println("Start!")

	pr, pw := io.Pipe()
	c1 := make(chan int)
	go writeToPipe(pw, c1)
	go readFromPipe(pr, c1)

	<-c1

	fmt.Println("Tschüssikovski!")
}

func writeToPipe(pw *io.PipeWriter, ch chan int) {
	var data []rune
	var command []rune
	var err error
	var r rune
	var t *tty.TTY

	if t, err = tty.Open(); err != nil {
		fmt.Println(err)
		panic(err)
	}

	for {
		if r, err = t.ReadRune(); errors.Is(io.EOF, err) {
			fmt.Println("EOF!")
			//command = cmdSend
			//r = runeCarriageReturn
		} else {
			if r == 0x0 {
				//fmt.Println("0x0")
				continue
			}
			command = append(command, r)
			fmt.Println(command)
		}
		if r == runeCarriageReturn {
			fmt.Println("Carriage Return")
			fmt.Println("Kommando: ", string(command))
			fmt.Println("Code: ", command)
			if compareRune(command, cmdSend) {
				fmt.Println("Sending Message", command)
				doWrite(pw, ch, data)
				data = make([]rune, 0) // Daten zum senden zurücksetzen
			} else if compareRune(command, cmdExit) {
				break
			} else {
				data = append(data, command...)
			}
			command = make([]rune, 0) // Neuer Befehl
		}
		if errors.Is(io.EOF, err) {
			break
		}
	}
	fmt.Println("Leitung wird an der Senderseite geschlossen!")
	pw.Close()
	ch <- 1
}

func compareRune(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func doWrite(pw *io.PipeWriter, ch chan int, data []rune) {
	if _, err := pw.Write([]byte(string(data))); errors.Is(io.ErrClosedPipe, err) {
		fmt.Println("Schreiben in die Leitung nicht möglich. Pipe wurde geschlossen!")
		ch <- 1
	}
}

func readFromPipe(pr *io.PipeReader, ch chan int) {

	defer pr.Close()
	if _, err := io.Copy(os.Stdout, pr); errors.Is(err, io.ErrClosedPipe) {
		fmt.Println("Unerwartetes Ende der Leitung. Wurde geschlossen!")
		ch <- 1
	}
	fmt.Println("Pipe wurde geschlossen!")
	ch <- 1
}
