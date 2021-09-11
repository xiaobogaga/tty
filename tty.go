package tty

import (
	"errors"
	"fmt"
	"github.com/eiannone/keyboard"
	"sync"
)

type Screen struct {
	prompt         []rune
	commandHistory [][]rune
	cursor         *Cursor
	command        []rune
	closed         bool
	lock           sync.Mutex
}

type Cursor struct {
	row int
	col int
}

const (
	ClearScreen        string = "\u001B[2K"
	CursorMoveToHeader        = "\r"
	CursorPos                 = "\u001B[%dG"
	CursorMoveToLeft          = "\u001B[%dD"
	CursorMoveToRight         = "\u001B[%dC"
	DeleteRight               = "\u001B[%dP"
)

func NewScreen(prompt []rune) *Screen {
	return &Screen{
		prompt:  prompt,
		cursor:  &Cursor{},
		command: prompt,
		closed:  false,
	}
}

func (screen *Screen) Open() error {
	return keyboard.Open()
}

func (screen *Screen) Close() error {
	screen.lock.Lock()
	defer screen.lock.Unlock()
	if screen.closed {
		return nil
	}
	screen.closed = true
	return keyboard.Close()
}

func (screen *Screen) onKeyArrowUp() {
	if screen.cursor.row <= 0 {
		return
	}
	screen.cursor.row--
	fmt.Printf("%s%s%s", ClearScreen, CursorMoveToHeader, string(screen.commandHistory[screen.cursor.row]))
	screen.setCommandToHistory(screen.cursor.row)
	screen.cursor.col = len(screen.command)
}

func (screen *Screen) setCommandToHistory(row int) {
	screen.command = make([]rune, len(screen.commandHistory[row]))
	copy(screen.command, screen.commandHistory[row])
}

func (screen *Screen) onKeyArrowDown() {
	switch {
	case screen.cursor.row < len(screen.commandHistory)-1:
		screen.cursor.row++
		fmt.Printf("%s%s%s", ClearScreen, CursorMoveToHeader, string(screen.commandHistory[screen.cursor.row]))
		screen.setCommandToHistory(screen.cursor.row)
		screen.cursor.col = len(screen.command)
	case screen.cursor.row == len(screen.commandHistory)-1:
		screen.command = screen.prompt
		fmt.Printf("%s%s%s", ClearScreen, CursorMoveToHeader, string(screen.command))
		screen.cursor.row++
		screen.cursor.col = len(screen.command)
	}
}

func (screen *Screen) onKeyEnter() {
	fmt.Printf(CursorPos+"\n", len(string(screen.command)))
	screen.commandHistory = append(screen.commandHistory, screen.command)
	screen.command = screen.prompt
	screen.cursor.row = len(screen.commandHistory)
	return
}

func (screen *Screen) onKeyArrowLeft() {
	if screen.cursor.col <= len(screen.prompt) {
		return
	}
	r := screen.command[screen.cursor.col-1]
	screen.cursor.col--
	fmt.Printf(CursorMoveToLeft, len(string(r)))
}

func (screen *Screen) onKeyArrowRight() {
	if screen.cursor.col >= len(screen.command) {
		screen.command = append(screen.command, ' ')
		fmt.Printf(" ")
		screen.cursor.col++
		return
	}
	r := screen.command[screen.cursor.col]
	fmt.Printf(CursorMoveToRight, len(string(r)))
	screen.cursor.col++
}

func (screen *Screen) onKeyBackSpace() {
	if screen.cursor.col <= len(screen.prompt) {
		return
	}
	r := screen.command[screen.cursor.col-1]
	fmt.Printf(CursorMoveToLeft, len(string(r)))
	fmt.Printf(DeleteRight, len(string(r)))
	screen.command = append(screen.command[:screen.cursor.col-1], screen.command[screen.cursor.col:]...)
	screen.cursor.col--
}

func (screen *Screen) onCharacter(r rune) {
	if r == 0 {
		return
	}
	if screen.cursor.col >= len(screen.command) {
		fmt.Printf("%s", string(r))
		screen.command = append(screen.command, r)
		screen.cursor.col++
		return
	}
	newCursorColPos := len(string(screen.command[:screen.cursor.col])) + len(string(r))
	screen.command = append(screen.command[:screen.cursor.col], append([]rune{r}, screen.command[screen.cursor.col:]...)...)
	fmt.Printf("%s%s%s", ClearScreen, CursorMoveToHeader, string(screen.command))
	fmt.Printf(CursorPos, newCursorColPos+1)
	screen.cursor.col++
	return
}

func (screen *Screen) onKeyHome() {
	fmt.Printf(CursorPos, len(screen.prompt)+1)
	screen.cursor.col = len(screen.prompt)
}

func (screen *Screen) onKeyEnd() {
	fmt.Printf(CursorPos, len(screen.command)+1)
	screen.cursor.col = len(screen.command)
}

type Command struct {
	Input string
	Error error
	Done  chan struct{}
}

var ErrExit = errors.New("Exit")

func (screen *Screen) Command() <-chan Command {
	ret := make(chan Command)
	go func() {
		for {
		NEXT:
			fmt.Printf("%s", string(screen.command))
			screen.cursor.col = len(screen.command)
			for {
				r, key, err := keyboard.GetSingleKey()
				if err != nil {
					ret <- Command{Error: err}
					return
				}
				switch key {
				case keyboard.KeyArrowUp:
					screen.onKeyArrowUp()
				case keyboard.KeyArrowDown:
					screen.onKeyArrowDown()
				case keyboard.KeyEnter:
					screen.onKeyEnter()
					command := Command{Input: string(screen.commandHistory[len(screen.commandHistory)-1][len(screen.prompt):]), Done: make(chan struct{})}
					ret <- command
					<-command.Done
					goto NEXT
				case keyboard.KeyArrowLeft:
					screen.onKeyArrowLeft()
				case keyboard.KeyArrowRight:
					screen.onKeyArrowRight()
				case keyboard.KeyCtrlD, keyboard.KeyCtrlC:
					screen.onKeyEnd()
					ret <- Command{Error: ErrExit}
					return
				case keyboard.KeyBackspace, keyboard.KeyBackspace2:
					screen.onKeyBackSpace()
				case keyboard.KeySpace:
					screen.onCharacter(' ')
				case keyboard.KeyTab:
					screen.onCharacter('\t')
				case keyboard.KeyEnd:
					screen.onKeyEnd()
				case keyboard.KeyHome:
					screen.onKeyHome()
				default:
					screen.onCharacter(r)
				}
			}
		}
	}()
	return ret
}
