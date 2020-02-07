package main

import (
	"bufio"
	//"bufio"
	"errors"
	"fmt"
	"github.com/mattn/go-tty"
	"io"
	"os"
	//"github.com/atotto/clipboard"
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
var keyArrowDown = []rune{ESC, 91, 66}
var keyArrowUp = []rune{ESC, 91, 65}
var keyArrowRight = []rune{ESC, 91, 67}
var keyArrowLeft = []rune{ESC, 91, 68}

// colors
var bgColorMagenta []rune = []rune{ESC, LEFT_SQUARE_BRACKET, 52, 54, 109}
var bgColorReset []rune = []rune{ESC, LEFT_SQUARE_BRACKET, 48, 109}

type Word struct {
	index       int
	character   []rune
	highlighted bool
}

func main() {

	//fmt.Println("Start!")
	lines := make(map[int][]rune)
	pr, pw := io.Pipe()
	c1 := make(chan int)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	var i int = 0
	for scanner.Scan() {
		text := scanner.Text()
		runes := []rune(text)
		runes = append(runes, LF)
		lines[i] = runes
		i++
	}

	go writeToPipe(pw, c1, lines)
	readFromPipe(pr)
	<-c1
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
	var index int = -1
	var horizontalIndex = -1
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
			if horizontalIndex == -1 {
				WriteToClipboard(string(m[index][:len(m[index])-1]))
			} else {
				sentence := splitSentence(m[index])
				for i := 0; i < len(sentence); i++ {
					if horizontalIndex == i {
						WriteToClipboard(string(sentence[i]))
					}
				}
			}
			break
		}
		input = append(input, key)

		if compareRune(input, keyArrowUp) { // Key up
			horizontalIndex = -1
			if index >= 0 {
				index--
				mh := prepareOutput(m, index, horizontalIndex, false)
				runes = convertToSliceOfRunes(mh)
				doWrite(pw, ch, runes)
			}
		} else if compareRune(input, keyArrowDown) { // Key down
			horizontalIndex = -1
			if index <= maxIndex {
				index++
				mh := prepareOutput(m, index, horizontalIndex, false)
				runes = convertToSliceOfRunes(mh)
				doWrite(pw, ch, runes)
			}
		} else if compareRune(input, keyArrowRight) { // Key right
			if index <= maxIndex {
				sentence := m[index]
				words := splitSentence(sentence)
				if horizontalIndex <= len(words) {
					horizontalIndex++
				}
				mh := prepareOutput(m, index, horizontalIndex, true)
				runes = convertToSliceOfRunes(mh)
				doWrite(pw, ch, runes)
			}

		} else if compareRune(input, keyArrowLeft) { // Key left
			if index <= maxIndex {
				if horizontalIndex >= 0 {
					horizontalIndex--
				}
			}
			mh := prepareOutput(m, index, horizontalIndex, true)
			runes = convertToSliceOfRunes(mh)
			doWrite(pw, ch, runes)
		}

	}
	exitWriteToPipe(pw, ch)
}

func splitSentence(sentence []rune) map[int][]rune {
	wordmap := make(map[int][]rune)
	var word []rune
	//var character []rune
	var wordcount int = 0
	var lastrune rune = 32 // SPACE
	for _, r := range sentence {
		if ((lastrune != 32 && r == 32) || (lastrune == 32 && r != 32)) && len(word) > 0 || r == CR { // neues Wort
			if r == CR {
				word = append(word, CR)
			}
			wordmap[wordcount] = word
			wordcount++
			word = []rune{}
		}
		word = append(word, r)
		lastrune = r
	}
	return wordmap
}

func prepareOutput(originalMap map[int][]rune, index int, horizontalindex int, highlightword bool) map[int][]rune {
	mh := make(map[int][]rune, len(originalMap))
	for i := 0; i < len(originalMap); i++ {
		var consolestream []rune
		if i == 0 {
			consolestream = restoreCursorPosition // Restore Cursor Position
		}
		if index == i {
			if highlightword {
				consolestream = append(consolestream, highlightWord(originalMap[i], horizontalindex)...)
			} else {
				consolestream = append(consolestream, bgColorMagenta...)
				consolestream = append(consolestream, originalMap[i]...)
				consolestream = append(consolestream, bgColorReset...)
			}
		} else {
			consolestream = append(consolestream, originalMap[i]...)
		}
		mh[i] = consolestream
	}

	return mh
}

func highlightWord(row []rune, horizontalindex int) []rune {
	wordmap := make(map[int][]rune)
	var word []rune = bgColorReset
	var wordcount int = 0
	var lastrune rune = 32 // SPACE
	for _, r := range row {
		if (((lastrune != 32 && r == 32) || (lastrune == 32 && r != 32)) && len(word) > 0) || r == LF { // neues Wort
			if r == LF {
				word = append(word, LF)
			}
			if wordcount == horizontalindex {
				if isSpaceWord(word) {
					horizontalindex++
				} else {
					word = append(bgColorMagenta, word...)
					word = append(word, bgColorReset...)
				}
			}
			wordmap[wordcount] = word
			wordcount++
			word = []rune{}
		}
		word = append(word, r)
		lastrune = r
	}
	var highlightedRow []rune
	for i := 0; i < len(wordmap); i++ {
		highlightedRow = append(highlightedRow, wordmap[i]...)
	}
	return highlightedRow
}

func isSpaceWord(word []rune) bool {
	if word == nil {
		return false
	}
	for _, v := range word {
		if v != 32 {
			return false
		}
	}
	return true
}

func exitWriteToPipe(pw *io.PipeWriter, ch chan int) {
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
}
