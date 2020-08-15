package cle

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/term"
)

const (
	TTY = "/dev/tty" // Microsoft Windows is not supported

	HISTORY_MAX_DEFAULT           = 100
	HISTORY_ENTRY_LEN_MIN_DEFAULT = 5
	REPORT_ERRORS_DEFAULT         = false

	ENTER_KEY           = 13
	ESCAPE_KEY          = 27
	UP_ARROW            = 65
	DOWN_ARROW          = 66
	RIGHT_ARROW         = 67
	LEFT_ARROW          = 68
	ARROW_KEY_INDICATOR = 91
	DELETE_KEY          = 127
)

type CLE struct {
	data           []byte
	terminal       *term.Term
	prompt         string
	cursorPosition int
	history        CommandHistory

	historyFile               string
	historyMax                int
	historyEntryMinimumLength int
	reportErrors              bool
}

type CommandHistory struct {
	commands        [][]byte
	currentPosition int
}

func NewCLE(options ...Option) *CLE {
	return new(CLE).configure(options)
}

func (this *CLE) configure(options []Option) *CLE {
	this.historyMax = HISTORY_MAX_DEFAULT
	this.historyEntryMinimumLength = HISTORY_ENTRY_LEN_MIN_DEFAULT
	this.reportErrors = REPORT_ERRORS_DEFAULT
	this.history = CommandHistory{}

	for _, configure := range options {
		configure(this)
	}

	this.loadHistory()
	return this
}

func (this *CLE) ReadInput(prompt string) []byte {
	this.prompt = prompt
	this.data = []byte{}
	this.cursorPosition = 0
	this.repaint(this.data)

	this.openTty()
	defer this.closeTty()

	for {
		work := make([]byte, 3)
		numRead, err := this.terminal.Read(work)
		if this.handleError(err) {
			continue
		}

		if this.handleArrowKeys(numRead, work) {
			continue
		}

		if this.handleDeleteKey(numRead, work) {
			continue
		}

		if this.handleEnterKey(numRead, work) {
			return this.data
		}

		if this.handleAnySingleKey(numRead, work) {
			continue
		}

		this.handlePaste(work)
	}
}

func (this *CLE) handleArrowKeys(numRead int, work []byte) bool {
	if numRead != 3 || work[0] != ESCAPE_KEY || work[1] != ARROW_KEY_INDICATOR {
		return false
	}

	switch work[2] {
	case UP_ARROW:
		if !this.handledUpArrow() {
			return true
		}
		this.populateDataWithHistoryEntry()
		this.repaint(this.data)

	case DOWN_ARROW:
		if !this.handledDownArrow() {
			this.clearInputData()
			this.repaint(this.data)
			return true
		}
		this.populateDataWithHistoryEntry()
		this.repaint(this.data)

	case RIGHT_ARROW:
		if !this.handledRightArrow() {
			return true
		}
		this.repaint(this.data)

	case LEFT_ARROW:
		if !this.handledLeftArrow() {
			return true
		}
		this.repaint(this.data)
	}

	return true
}

func (this *CLE) handleEnterKey(numRead int, work []byte) bool {
	if numRead != 1 || work[0] != ENTER_KEY {
		return false
	}

	crlf()

	if string(this.data) == "!clear" {
		this.ClearHistory()
		this.clearInputData()
		return true
	}

	this.saveHistoryEntry()
	return true
}

func (this *CLE) handleDeleteKey(numRead int, work []byte) bool {
	if numRead != 1 || work[0] != DELETE_KEY {
		return false
	}

	if this.cursorPosition == 0 {
		return true
	}

	this.cursorPosition--
	this.data = remove(this.data, this.cursorPosition)
	this.repaint(this.data)
	return true
}

func (this *CLE) handleAnySingleKey(numRead int, work []byte) bool {
	if numRead != 1 {
		return false
	}

	if !isPrintable(work[0]) {
		return true
	}
	this.data = insert(this.data, this.cursorPosition, work[0])
	this.cursorPosition++
	this.repaint(this.data)
	return true
}

func (this *CLE) handlePaste(work []byte) {
	for _, c := range work {
		if !isPrintable(c) {
			continue
		}
		this.data = insert(this.data, this.cursorPosition, c)
		this.cursorPosition++
	}
	this.repaint(this.data)
}

func (this *CLE) repaint(data []byte) {
	fmt.Printf("%c%c%c%c", 27, '[', '2', 'K')                 // VT100 clear line
	fmt.Printf("%c%s%s%c", 13, this.prompt, string(data), 32) // go to beginning and print data
	for i := len(data) + 1; i > this.cursorPosition; i-- {    // backspace to the current cursor position
		fmt.Printf("%c", 8)
	}
}

func (this *CLE) openTty() {
	this.terminal, _ = term.Open(TTY)
	this.handleError(term.RawMode(this.terminal))
}

func (this *CLE) closeTty() {
	this.handleError(this.terminal.Restore())
	this.handleError(this.terminal.Close())
}

func (this *CLE) clearInputData() {
	this.data = this.data[:0]
	this.cursorPosition = 0
}

func (this *CLE) handledLeftArrow() bool {
	if this.cursorPosition <= 0 {
		return false
	}
	this.cursorPosition--
	return true
}

func (this *CLE) handledRightArrow() bool {
	if this.cursorPosition > len(this.data)-1 {
		return false
	}
	this.cursorPosition++
	return true
}

func (this *CLE) handledUpArrow() bool {
	this.history.currentPosition--
	if this.history.currentPosition < 0 {
		this.history.currentPosition = 0
		return false
	}
	return true
}

func (this *CLE) handledDownArrow() bool {
	this.history.currentPosition++
	if this.history.currentPosition >= len(this.history.commands) {
		this.history.currentPosition = len(this.history.commands)
		return false
	}
	return true
}

func (this *CLE) populateDataWithHistoryEntry() {
	this.clearInputData()
	this.data = append(this.data, this.getCurrentHistoryEntry()...)
	this.cursorPosition = len(this.data)
}

func (this *CLE) saveHistoryEntry() {
	if len(this.data) > this.historyEntryMinimumLength {
		this.history.commands = append(this.history.commands, this.data)
		this.history.currentPosition = len(this.history.commands)
	}
}

func (this *CLE) getCurrentHistoryEntry() []byte {
	if len(this.history.commands) == 0 {
		return []byte("")
	}
	if this.history.currentPosition < 0 || this.history.currentPosition >= len(this.history.commands) {
		return []byte("")
	}

	return this.history.commands[this.history.currentPosition]
}

func (this *CLE) SaveHistory() {
	var history []byte
	// save the last n commands
	startIndex := len(this.history.commands) - this.historyMax
	if startIndex < 0 {
		startIndex = 0
	}
	this.history.commands = this.history.commands[startIndex:]
	for _, historyLine := range this.history.commands {
		history = append(history, historyLine...)
		history = append(history, '\n')
	}

	this.handleError(ioutil.WriteFile(this.historyFile, history, 0644))
}

func (this *CLE) loadHistory() {
	if len(this.historyFile) == 0 {
		return
	}

	history, err := ioutil.ReadFile(this.historyFile)
	if this.handleError(err) {
		return
	}
	scanner := bufio.NewScanner(bytes.NewReader(history))
	for scanner.Scan() {
		this.history.commands = append(this.history.commands, scanner.Bytes())
	}
	this.history.currentPosition = len(this.history.commands)
}

func (this *CLE) ClearHistory() {
	this.history.commands = this.history.commands[:0]
	this.history.currentPosition = 0
	this.handleError(os.Remove(this.historyFile))
}

func (this *CLE) handleError(err error) bool {
	if err != nil && this.reportErrors {
		fmt.Println(err)
	}
	return err != nil
}

////////////////////////////////////////////

func crlf() {
	fmt.Printf("%c%c", 10, 13)
}

func isPrintable(c byte) bool {
	return c >= 32 && c <= 126
}

func insert(slice []byte, position int, character byte) []byte {
	if position > len(slice)-1 {
		return append(slice, character)
	}
	slice = append(slice, 0)
	copy(slice[position+1:], slice[position:])
	slice[position] = character
	return slice
}

func remove(slice []byte, position int) []byte {
	if position > len(slice)-1 {
		return slice
	}
	ret := make([]byte, 0)
	ret = append(ret, slice[:position]...)
	return append(ret, slice[position+1:]...)
}
