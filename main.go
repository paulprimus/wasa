package main

import (
	"bufio"
	//"bufio"
	"errors"
	"fmt"
	"github.com/mattn/go-tty"
	"io"
	"os"
	"github.com/mattn/go-colorable"
)

const ESC rune = 0x1b

var csi rune = 0x9b // CSI
const LEFT_SQUARE_BRACKET rune = 0x5B
const SEMI_COLUMN rune = 0x3B
const M_HEX rune = 0x6D

var runeA rune = '\x3d'
var runeLf rune = 0xA
var runeCarriageReturn rune = 0xD
var cmdSend []rune = []rune{0x73, 0x65, 0x6E, 0x64, 0xD}
var cmdExit []rune = []rune{0x65, 0x78, 0x69, 0x74, 0xD}

//var bgColorWhite []rune = []rune{ESC, LEFT_SQUARE_BRACKET, 0x20, SEMI_COLUMN, 0x31,  0x6D} // esc [
var bgColorWhite []rune = []rune{ESC, LEFT_SQUARE_BRACKET, 0x21, M_HEX}
//var bgColorWhite []rune = []rune{ESC, LEFT_SQUARE_BRACKET, 0x2}

func main() {

	fmt.Println("Start!")
	var lines []string
	pr, pw := io.Pipe()
	c1 := make(chan int)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	m := convertToMap(lines)
	go writeToPipe(pw, c1, m)
	readFromPipe(pr, c1)
	<-c1

	fmt.Println("Tschüssikovski!")
}

func convertToMap(str []string) map[int][]rune {
	m := make(map[int][]rune)
	for i, s := range str {
		m[i] = []rune(s)
	}
	return m
}

func writeToPipe(pw *io.PipeWriter, ch chan int, v map[int][]rune) {
	var err error
	var input string
	var t *tty.TTY
	n := append(bgColorWhite, v[5]...)
	fmt.Println(n)
	if t, err = tty.Open(); err != nil {
		fmt.Println(err)
		panic(err)
	}

	for {
		if input, err = t.ReadString(); errors.Is(io.EOF, err) {
			fmt.Println("EOF!")
			exitWriteToPipe(pw, ch)
		}

		if input == "exit" {
			break
		}
		doWrite(pw, ch, n)
	}
	exitWriteToPipe(pw, ch)
}

func exitWriteToPipe(pw *io.PipeWriter, ch chan int) {
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
	// var b []byte
	////	for _, v := range data {
	//		//b = append(b,[]byte(v)...)
	////	}
	fmt.Println(data)
	if _, err := pw.Write([]byte(string(data))); errors.Is(io.ErrClosedPipe, err) {
		fmt.Println("Schreiben in die Leitung nicht möglich. Pipe wurde geschlossen!")
		ch <- 1
	}
}

func readFromPipe(pr *io.PipeReader, ch chan int) {

	defer pr.Close()
	//stdout := colorable.NewColorableStdout()
	newColorable := colorable.NewColorable(os.Stdout)
	//newColorable.Write(io.C)
	//writer := bufio.NewWriter(os.Stdout)
	if _, err := io.Copy(newColorable, pr); err != nil {
		fmt.Println("Unerwartetes Ende der Leitung. Wurde geschlossen!")
		//ch <- 1
	}
	fmt.Println("Pipe wurde geschlossen!")
	//ch <- 1
}
