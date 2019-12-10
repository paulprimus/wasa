package main

import (
	"bufio"
	"unicode/utf8"

	//"bufio"
	"errors"
	"fmt"
	"github.com/mattn/go-tty"
	"io"
	"os"
)

const ESC rune = 27
const LEFT_SQUARE_BRACKET rune = 91
const LF rune = 10
const CR rune = 13
const sRune rune = 115
const uRune rune = 117

// cursor
var saveCursorPosition = []rune{ESC, LEFT_SQUARE_BRACKET, sRune}
var restoreCursorPosition = []rune{ESC, LEFT_SQUARE_BRACKET, uRune}

// keys
var keyArrowDown = []rune{27, 91, 66}
var keyArrowUp = []rune{27, 91, 65}

// colors
var bgColorMagenta []rune = []rune{ESC, LEFT_SQUARE_BRACKET, 52, 54, 109}
var bgColorReset []rune = []rune{ESC, LEFT_SQUARE_BRACKET, 48, 109}

func main() {

	fmt.Println("Start!")
	var lines []string
	pr, pw := io.Pipe()
	c1 := make(chan int)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if utf8.ValidString(line) == false {
			fmt.Println("Kein UTF-8!")
		}
		lines = append(lines, line)
	}
	m := convertToMap(lines)
	go writeToPipe(pw, c1, m)
	readFromPipe(pr)
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

func writeToPipe(pw *io.PipeWriter, ch chan int, m map[int][]rune) {
	var err error
	var key rune
	var input []rune
	var t *tty.TTY

	if t, err = tty.Open(); err != nil {
		fmt.Println(err)
		panic(err)
	}
	runes := convertToSliceOfRunes(m)
	runes = append(saveCursorPosition, runes...)
	doWrite(pw, ch, runes)
	var index int = 0
	var maxIndex int = len(m)
	for {
		if key, err = t.ReadRune(); errors.Is(io.EOF, err) {
			fmt.Println("EOF!")
			exitWriteToPipe(pw, ch)
		}
		if key == 0 {
			continue
		}
		//fmt.Println("Rune:", key)
		if key == 27 || len(input) > 3 {
			input = make([]rune, 0)
		}
		input = append(input, key)

		if compareRune(input, keyArrowUp) {
			if index > 0 {
                mh := prepareOutput(m, index)
                runes = convertToSliceOfRunes(mh)
                doWrite(pw, ch, runes)
				index--
			}
		} else if compareRune(input, keyArrowDown) { // Key down
			if index <= maxIndex {
                mh := prepareOutput(m, index)
                runes = convertToSliceOfRunes(mh)
                doWrite(pw, ch, runes)
				index++
			}
		}
	}
	exitWriteToPipe(pw, ch)
}

func prepareOutput(originalMap map[int][]rune, index int) map[int][]rune {
	mh := make(map[int][]rune, len(originalMap))
	for k, v := range originalMap {
		if k == 0 {
			mh[k] = append(restoreCursorPosition, v...)
		} else if k == index {
			highlighted := append(bgColorMagenta, v...)
			highlighted = append(highlighted, bgColorReset...)
			mh[k] = highlighted
		} else {
			mh[k] = v
		}
	}
	return mh
}

func exitWriteToPipe(pw *io.PipeWriter, ch chan int) {
	fmt.Println("Leitung wird an der Senderseite geschlossen!")
	pw.Close()
	ch <- 1
}

func convertToSliceOfRunes(m map[int][]rune) []rune {
	var data []rune
	for _, v := range m {
		//v = append(v, LF, CR)
		data = append(data, v...)

	}
	return data
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

func readFromPipe(pr *io.PipeReader) {

	defer pr.Close()
	//var buf []byte
	//writer := bufio.NewWriter(os.Stdout)
	if _, err := io.Copy(os.Stdout, pr); err != nil {
		fmt.Println("Unerwartetes Ende der Leitung. Wurde geschlossen!")
		//ch <- 1
	}
	fmt.Println("Pipe wurde geschlossen!")
	//ch <- 1
}
