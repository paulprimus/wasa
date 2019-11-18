package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

var Esc string = "\x1b"

func main() {
	var reader = bufio.NewReader(os.Stdin)
	var output []rune
	fmt.Println("Wasa hat begonnen!")
	//[]byte{0x1b}, '[','s'
	fmt.Sprintf("%s","\x1b[s")
	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}
	writer := bufio.NewWriter(os.Stdout)
	_, err := writer.Write([]byte(string(output)))
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(-1)
	}
	err = writer.Flush()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(-1)
	}
	fmt.Sprintf("%s","\x1b[2K")
	//fmt.Print(restoreCursorPos())
}

//func escape(format string, args ...interface{}) []byte {
//	//return fmt.Sprintf("%s%s", Esc, fmt.Sprintf(format, args...))
//	return []byte{0x1b, '[', 'u'}
//}
//
//func saveCursorPos() string {
//	return escape("[s")
//}

func restoreCursorPos() []byte {
	return []byte{0x1b, '[', 'u'}
}
