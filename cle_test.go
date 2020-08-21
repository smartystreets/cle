package cle

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/smartystreets/assertions/should"
	"github.com/smartystreets/gunit"
)

func TestCLEFixture(t *testing.T) {
	gunit.Run(new(CLEFixture), t)
}

type CLEFixture struct {
	*gunit.Fixture
}

func (this *CLEFixture) TestOptions() {
	cleObj := NewCLE(
		HistoryFile("testfile"),
		HistorySize(10),
		HistoryEntryMinimumLength(2),
		ReportErrors(true),
	)
	this.So(cleObj.historyFile, should.Equal, "testfile")
	this.So(cleObj.historyMax, should.Equal, 10)
	this.So(cleObj.historyEntryMinimumLength, should.Equal, 2)
	this.So(cleObj.reportErrors, should.BeTrue)
}

func (this *CLEFixture) TestHandleEnterKey() {
	cleObj := NewCLE()

	buffer := []byte{32, 0, 0}
	handled := cleObj.handleEnterKey(1, buffer)
	this.So(handled, should.BeFalse)

	buffer = []byte{13, 0, 0}
	cleObj.history.commands = cleObj.history.commands[:0]
	cleObj.data = []byte("123456")
	cleObj.cursorPosition = 2
	handled = cleObj.handleEnterKey(1, buffer)
	this.So(handled, should.BeTrue)
	this.So(cleObj.history.commands[0], should.Resemble, []byte("123456"))
}

func (this *CLEFixture) TestHandleEnterKeyClearHistory() {
	cleObj := NewCLE()
	buffer := []byte{13, 0, 0}
	cleObj.history.commands = cleObj.history.commands[:0]
	cleObj.data = []byte("this is a history entry")
	cleObj.saveHistoryEntry()
	cleObj.data = []byte("this is a history entry 2")
	cleObj.saveHistoryEntry()
	this.So(len(cleObj.history.commands), should.Equal, 2)
	cleObj.data = []byte("!clear")
	cleObj.handleEnterKey(1, buffer)
	this.So(len(cleObj.history.commands), should.Equal, 0)
}

func (this *CLEFixture) TestHandleDeleteKey() {
	cleObj := NewCLE()
	buffer := []byte{32, 0, 0}
	handled := cleObj.handleDeleteKey(1, buffer)
	this.So(handled, should.BeFalse)

	buffer = []byte{127, 0, 0}
	cleObj.data = []byte("123456")
	cleObj.cursorPosition = 2
	handled = cleObj.handleDeleteKey(1, buffer)
	this.So(handled, should.BeTrue)
	this.So(cleObj.data, should.Resemble, []byte("13456"))

	buffer = []byte{127, 0, 0}
	cleObj.data = []byte("123456")
	cleObj.cursorPosition = 0
	handled = cleObj.handleDeleteKey(1, buffer)
	this.So(handled, should.BeTrue)
	this.So(cleObj.data, should.Resemble, []byte("123456"))
}

func (this *CLEFixture) TestHandleAnySingleKey() {
	cleObj := NewCLE()

	buffer := []byte{'a', 'b', 0} // multiple keys are unhandled
	handled := cleObj.handleAnySingleKey(2, buffer)
	this.So(handled, should.BeFalse)

	buffer = []byte{150, 0, 0} // unprintable is handled by not added to data
	handled = cleObj.handleAnySingleKey(1, buffer)
	this.So(handled, should.BeTrue)
	this.So(cleObj.data, should.Resemble, []byte(nil))

	buffer = []byte{'a', 0, 0}
	cleObj.handleAnySingleKey(1, buffer)
	this.So(cleObj.data, should.Resemble, []byte("a"))

	buffer = []byte{'b', 0, 0}
	cleObj.handleAnySingleKey(1, buffer)
	this.So(cleObj.data, should.Resemble, []byte("ab"))

	cleObj.handledLeftArrow()
	buffer = []byte{'c', 0, 0}
	handled = cleObj.handleAnySingleKey(1, buffer)
	this.So(cleObj.data, should.Resemble, []byte("acb"))
}

func (this *CLEFixture) TestHandlePaste() {
	cleObj := NewCLE()

	buffer := []byte{'a', 'b', 'c'}
	cleObj.handlePaste(buffer)

	buffer = []byte{'d', 0, 0}
	cleObj.handlePaste(buffer)
	this.So(cleObj.data, should.Resemble, []byte("abcd"))
}

func (this *CLEFixture) TestHandleArrowKeys() {
	cleObj := NewCLE()

	buffer := []byte{ESCAPE_KEY, 0, 0}
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeFalse)

	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, UP_ARROW}
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)

	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, DOWN_ARROW}
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)

	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, RIGHT_ARROW}
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)

	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, LEFT_ARROW}
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)

	cleObj.data = []byte("this is a history entry")
	cleObj.saveHistoryEntry()
	cleObj.data = []byte("this is a history entry2")
	cleObj.saveHistoryEntry()

	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, UP_ARROW}
	cleObj.history.currentPosition = len(cleObj.history.commands)
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)

	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, DOWN_ARROW}
	cleObj.history.currentPosition = 0
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)

	cleObj.data = []byte("abc")
	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, RIGHT_ARROW}
	cleObj.cursorPosition = 0
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)

	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, LEFT_ARROW}
	cleObj.cursorPosition = len(cleObj.data)
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)
}

func (this *CLEFixture) TestHandleControlKeys() {
	cleObj := NewCLE()

	buffer := []byte{UP_ARROW, 0, 0}
	this.So(cleObj.handleControlKeys(1, buffer), should.BeFalse)

	cleObj.data = []byte("test command")
	buffer = []byte{CONTROL_A, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.cursorPosition, should.BeZeroValue)

	cleObj.data = []byte("test command")
	buffer = []byte{CONTROL_B, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.cursorPosition, should.Equal, 0)
	this.So(cleObj.data, should.Resemble, []byte("command"))

	cleObj.data = []byte("test command")
	buffer = []byte{CONTROL_D, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.cursorPosition, should.Equal, 5)
	this.So(cleObj.data, should.Resemble, []byte("test ommand"))

	cleObj.data = []byte("test command")
	buffer = []byte{CONTROL_E, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.cursorPosition, should.Equal, len(cleObj.data))

	cleObj.data = []byte("test command")
	buffer = []byte{CONTROL_K, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.data, should.Resemble, []byte("test "))

	cleObj.data = []byte("test command")
	buffer = []byte{CONTROL_N, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.cursorPosition, should.Equal, 0)
	this.So(len(cleObj.data), should.Equal, 0)
}

func (this *CLEFixture) TestHistoryHandledUpArrow() {
	cleObj := NewCLE()
	cleObj.history.commands = cleObj.history.commands[:0]
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte(""))

	cleObj.data = []byte("this is a history entry")
	cleObj.saveHistoryEntry()
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte(""))

	cleObj.handledUpArrow()
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte("this is a history entry"))

	cleObj.history.currentPosition = -1
	this.So(cleObj.handledUpArrow(), should.BeFalse)
	this.So(cleObj.history.currentPosition, should.BeZeroValue)
}

func (this *CLEFixture) TestHistoryHandledDownArrow() {
	cleObj := NewCLE()
	cleObj.history.commands = cleObj.history.commands[:0]
	cleObj.data = []byte("this is a history entry")
	cleObj.saveHistoryEntry()
	cleObj.data = []byte("this is a history entry 2")
	cleObj.saveHistoryEntry()

	cleObj.handledUpArrow()
	cleObj.handledUpArrow()
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte("this is a history entry"))
	cleObj.handledDownArrow()
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte("this is a history entry 2"))

	cleObj.history.currentPosition = len(cleObj.history.commands)
	this.So(cleObj.handledDownArrow(), should.BeFalse)
	this.So(cleObj.history.currentPosition, should.Equal, len(cleObj.history.commands))
}

func (this *CLEFixture) TestHandledLeftAndRightArrows() {
	cleObj := NewCLE()
	cleObj.data = []byte("1234")

	cleObj.cursorPosition = 0
	cleObj.handledRightArrow()
	cleObj.handledRightArrow()
	cleObj.handledRightArrow()
	cleObj.handledRightArrow()
	cleObj.handledRightArrow()
	this.So(cleObj.cursorPosition, should.Equal, 4)

	cleObj.handledLeftArrow()
	this.So(cleObj.data[cleObj.cursorPosition], should.Equal, '4')

	cleObj.handledLeftArrow()
	cleObj.handledLeftArrow()
	cleObj.handledLeftArrow()
	cleObj.handledLeftArrow()
	cleObj.handledLeftArrow()
	this.So(cleObj.data[cleObj.cursorPosition], should.Equal, '1')
}

func (this *CLEFixture) TestPopulateDataWithHistoryEntry() {
	cleObj := NewCLE()
	cleObj.history.commands = cleObj.history.commands[:0]
	cleObj.data = []byte("this is a history entry")
	cleObj.saveHistoryEntry()

	cleObj.data = []byte("this is a history entry 2")
	cleObj.saveHistoryEntry()
	cleObj.history.currentPosition = len(cleObj.history.commands) - 1

	cleObj.populateDataWithHistoryEntry()
	this.So(cleObj.data, should.Resemble, []byte("this is a history entry 2"))
}

func (this *CLEFixture) TestClearHistory() {
	cleObj := NewCLE()

	cleObj.history.commands = append(cleObj.history.commands, []byte("testing1"))
	cleObj.history.commands = append(cleObj.history.commands, []byte("testing2"))
	this.So(len(cleObj.history.commands), should.Equal, 2)

	cleObj.ClearHistory()
	this.So(len(cleObj.history.commands), should.BeZeroValue)
}

func (this *CLEFixture) TestPrepareHistoryForWriting() {
	cleObj := NewCLE()

	cleObj.data = []byte("this is a history entry")
	cleObj.saveHistoryEntry()

	cleObj.data = []byte("this is a history entry 2")
	cleObj.saveHistoryEntry()

	history := cleObj.prepareHistoryForWriting()
	this.So(len(history), should.Equal, 50)
	this.So(bytes.Contains(history, []byte("entry 2")), should.BeTrue)

	cleObj.historyMax = 1
	history = cleObj.prepareHistoryForWriting()
	this.So(len(history), should.Equal, 26)
	this.So(bytes.Contains(history, []byte("entry 2")), should.BeTrue)
}

func (this *CLEFixture) TestLoadHistory() {
	cleObj := NewCLE()

	reader := bytes.NewReader([]byte("history entry\nhistory entry 2"))
	scanner := bufio.NewScanner(reader)
	cleObj.loadHistory(scanner)

	this.So(cleObj.history.currentPosition, should.Equal, 2)
	this.So(len(cleObj.history.commands), should.Equal, 2)
	this.So(bytes.Contains(cleObj.history.commands[1], []byte("entry 2")), should.BeTrue)
}

func (this *CLEFixture) TestInsert() {
	data := []byte("abcdef")

	this.So(insert(data, 2, 'a'), should.Resemble, []byte("abacdef"))
	this.So(insert(data, 0, 'a'), should.Resemble, []byte("aabcdef"))
	this.So(insert(data, 99, 'a'), should.Resemble, []byte("abcdefa"))
}

func (this *CLEFixture) TestRemove() {
	data := []byte("abcdef")

	this.So(remove(data, 2), should.Resemble, []byte("abdef"))
	this.So(remove(data, 0), should.Resemble, []byte("bcdef"))
	this.So(remove(data, 99), should.Resemble, []byte("abcdef"))
}
