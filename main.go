package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

var Esc string = "\x1b"
var RUNE_A rune = '\x3d'

func main() {

	fmt.Println("Start!")

	pr, pw := io.Pipe()
	c1 := make(chan int)
	//c2 := make(chan int)
	go writeToPipe(pw, c1)
	go readFromPipe(pr, c1)

	<-c1
	//<-c2

	fmt.Println("Bye")
}

func writeToPipe(pw *io.PipeWriter, ch chan int) {
	var data []rune
	var err error
	var r rune

	reader := bufio.NewReader(os.Stdin)
	for {
		if r, _, err = reader.ReadRune(); err != nil {
			panic(err)
		}
		data = append(data, r)
		if r == 's' {
			if _, err = pw.Write([]byte(string(data))); err != nil {
				panic(err)
			}
		} else if r == 'q' {
			pw.Close()
			break
		}
	}
	ch <- 1

}

func readFromPipe(pr *io.PipeReader, ch chan int) {
	//var in []byte
	//bw := bufio.NewWriter(os.Stdout)
	if _, err := io.Copy(os.Stdout, pr); errors.Is(err, io.EOF) {
		//fmt.Println(in)
		pr.Close()
		//bw.Flush()
		ch <- 1
	}
}
