package main

import (
	"fmt"
	"github.com/xiaobogaga/tty"
)

func main() {
	screen := tty.NewScreen([]rune("momoko> "))
	err := screen.Open()
	if err != nil {
		panic(err)
	}
	defer screen.Close()
	reader := screen.Command()
	for {
		command := <-reader
		if command.Error != nil {
			println("exit")
			return
		}
		fmt.Printf("handle command: %s\n", command.Input)
		command.Done <- struct{}{}
	}
}
