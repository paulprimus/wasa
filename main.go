package main

import (
	"bufio"
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

	//fmt.Println("Start!")
	lines := make(map[int][]rune)
	pr, pw := io.Pipe()
	c1 := make(chan int)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	var i int = 0
	for scanner.Scan() {
		line := scanner.Text()

		//if utf8.ValidString(line) == false {
		//	fmt.Println("Kein UTF-8!")
		//}
		runes := []rune(line)
		runes = append(runes, LF)
		lines[i] = runes
		i++
	}
	//fmt.Println(lines)

	go writeToPipe(pw, c1, lines)
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
	var maxIndex int = len(m) - 1
	for {
		if key, err = t.ReadRune(); errors.Is(io.EOF, err) {
			fmt.Println("EOF!")
			exitWriteToPipe(pw, ch)
		}
		if key == 0 {
			continue
		}
		if key == 27 || len(input) > 3 {
			input = make([]rune, 0)
		} else if key == CR {
			WriteToClipboard("Paul du hast es geschafft!")
			break
		}
		input = append(input, key)

		if compareRune(input, keyArrowUp) { // Key up
			if index >= 0 {
				index--
				mh := prepareOutput(m, index)
				runes = convertToSliceOfRunes(mh)
				doWrite(pw, ch, runes)
			}
		} else if compareRune(input, keyArrowDown) { // Key down
			if index < maxIndex {
				index++
				mh := prepareOutput(m, index)
				runes = convertToSliceOfRunes(mh)
				doWrite(pw, ch, runes)
			}
		}
	}
	exitWriteToPipe(pw, ch)
}

func prepareOutput(originalMap map[int][]rune, index int) map[int][]rune {
	mh := make(map[int][]rune, len(originalMap))
	for i := 0; i < len(originalMap); i++ {
		//for k, v := range originalMap {
		var consolestream []rune
		var escapeseqence []rune
		if i == 0 {
			escapeseqence = restoreCursorPosition // Restore Cursor Position
		}
		if index == i {
			escapeseqence = append(escapeseqence, bgColorMagenta...)
		}
		consolestream = append(escapeseqence, originalMap[i]...) // Mark Line
		consolestream = append(consolestream, bgColorReset...)
		mh[i] = consolestream

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
	for i := 0; i < len(m); i++ {
		//for _, v := range m {
		data = append(data, m[i]...)
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
	if _, err := io.Copy(os.Stdout, pr); err != nil {
		fmt.Println("Unerwartetes Ende der Leitung. Wurde geschlossen!")
	}
	fmt.Println("Pipe wurde geschlossen!")
}
