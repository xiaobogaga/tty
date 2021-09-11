an embeddable tty for golang application. Basically you can use it to read command from stand input
and user can use arrowKey to move input, up and down key to select history command.

## Usage

```golang
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
```

This code can be found in examples

## Try

```golang
go install -u github.com/xiaobogaga/tty/examples
```

Note: currently only reliable on linux, macos.