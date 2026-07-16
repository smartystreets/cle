package cle

import (
	"bufio"
	"bytes"
	"os"
	"testing"

	"github.com/smarty/assertions/should"
	"github.com/smarty/gunit"
)

func TestCLEFixture(t *testing.T) {
	gunit.Run(new(CLEFixture), t)
}

type CLEFixture struct {
	*gunit.Fixture
}

func (this *CLEFixture) TestOptions() {
	testHistoryFile := "bogus_filename_used_for_testing"
	cleObj := NewCLE(
		HistoryFile(testHistoryFile),
		HistorySize(10),
		HistoryEntryMinimumLength(2),
		ReportErrors(true),
		TestMode(true),
	)
	this.So(cleObj.historyFile, should.Equal, testHistoryFile)
	this.So(cleObj.historyMax, should.Equal, 10)
	this.So(cleObj.historyEntryMinimumLength, should.Equal, 2)
	this.So(cleObj.reportErrors, should.BeTrue)
}

func (this *CLEFixture) TestHandleEnterKey() {
	cleObj := NewCLE(TestMode(true))

	buffer := []byte{32, 0, 0}
	handled := cleObj.handleEnterKey(1, buffer)
	this.So(handled, should.BeFalse)

	buffer = []byte{13, 0, 0}
	cleObj.history.commands = cleObj.history.commands[:0]
	cleObj.data = []rune("123456")
	cleObj.cursorPosition = 2
	handled = cleObj.handleEnterKey(1, buffer)
	this.So(handled, should.BeTrue)
	this.So(cleObj.history.commands[0], should.Resemble, []byte("123456"))
}

func (this *CLEFixture) TestHandleEnterKeyClearHistory() {
	cleObj := NewCLE(TestMode(true))
	buffer := []byte{13, 0, 0}
	cleObj.history.commands = cleObj.history.commands[:0]
	cleObj.data = []rune("this is a history entry")
	cleObj.saveHistoryEntry()
	cleObj.data = []rune("this is a history entry 2")
	cleObj.saveHistoryEntry()
	this.So(len(cleObj.history.commands), should.Equal, 2)
	cleObj.data = []rune("!clear")
	cleObj.handleEnterKey(1, buffer)
	this.So(len(cleObj.history.commands), should.Equal, 0)
}

func (this *CLEFixture) TestHandleDeleteKey() {
	cleObj := NewCLE(TestMode(true))
	buffer := []byte{32, 0, 0}
	handled := cleObj.handleDeleteKey(1, buffer)
	this.So(handled, should.BeFalse)

	buffer = []byte{127, 0, 0}
	cleObj.data = []rune("123456")
	cleObj.cursorPosition = 2
	handled = cleObj.handleDeleteKey(1, buffer)
	this.So(handled, should.BeTrue)
	this.So(cleObj.data, should.Resemble, []rune("13456"))

	buffer = []byte{127, 0, 0}
	cleObj.data = []rune("123456")
	cleObj.cursorPosition = 0
	handled = cleObj.handleDeleteKey(1, buffer)
	this.So(handled, should.BeTrue)
	this.So(cleObj.data, should.Resemble, []rune("123456"))
}

func (this *CLEFixture) TestHandleAnySingleKey() {
	cleObj := NewCLE(TestMode(true))

	buffer := []byte{'a', 'b', 0} // multiple keys are unhandled
	handled := cleObj.handleAnySingleKey(2, buffer)
	this.So(handled, should.BeFalse)

	buffer = []byte{150, 0, 0} // unprintable is handled by not added to data
	handled = cleObj.handleAnySingleKey(1, buffer)
	this.So(handled, should.BeTrue)
	this.So(cleObj.data, should.Resemble, []rune(nil))

	buffer = []byte{'a', 0, 0}
	cleObj.handleAnySingleKey(1, buffer)
	this.So(cleObj.data, should.Resemble, []rune("a"))

	buffer = []byte{'b', 0, 0}
	cleObj.handleAnySingleKey(1, buffer)
	this.So(cleObj.data, should.Resemble, []rune("ab"))

	cleObj.handledLeftArrow()
	buffer = []byte{'c', 0, 0}
	handled = cleObj.handleAnySingleKey(1, buffer)
	this.So(cleObj.data, should.Resemble, []rune("acb"))
}

func (this *CLEFixture) TestHandlePaste() {
	cleObj := NewCLE(TestMode(true))

	buffer := []byte{'a', 'b', 'c'}
	cleObj.handlePaste(buffer)

	buffer = []byte{'d', 0, 0}
	cleObj.handlePaste(buffer)
	this.So(cleObj.data, should.Resemble, []rune("abcd"))
}

func (this *CLEFixture) TestHandleArrowKeys() {
	cleObj := NewCLE(TestMode(true))

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

	cleObj.data = []rune("this is a history entry")
	cleObj.saveHistoryEntry()
	cleObj.data = []rune("this is a history entry2")
	cleObj.saveHistoryEntry()

	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, UP_ARROW}
	cleObj.history.currentPosition = len(cleObj.history.commands)
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)

	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, DOWN_ARROW}
	cleObj.history.currentPosition = 0
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)

	cleObj.data = []rune("abc")
	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, RIGHT_ARROW}
	cleObj.cursorPosition = 0
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)

	buffer = []byte{ESCAPE_KEY, ARROW_KEY_INDICATOR, LEFT_ARROW}
	cleObj.cursorPosition = len(cleObj.data)
	this.So(cleObj.handleArrowKeys(3, buffer), should.BeTrue)
}

func (this *CLEFixture) TestHandleControlKeys() {
	cleObj := NewCLE(TestMode(true))

	buffer := []byte{UP_ARROW, 0, 0}
	this.So(cleObj.handleControlKeys(1, buffer), should.BeFalse)

	cleObj.data = []rune("test command")
	buffer = []byte{CONTROL_A, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.cursorPosition, should.BeZeroValue)

	cleObj.data = []rune("test command")
	buffer = []byte{CONTROL_B, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.cursorPosition, should.Equal, 0)
	this.So(cleObj.data, should.Resemble, []rune("command"))

	cleObj.data = []rune("test command")
	buffer = []byte{CONTROL_D, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.cursorPosition, should.Equal, 5)
	this.So(cleObj.data, should.Resemble, []rune("test ommand"))

	cleObj.data = []rune("test command")
	buffer = []byte{CONTROL_E, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.cursorPosition, should.Equal, len(cleObj.data))

	cleObj.data = []rune("test command")
	buffer = []byte{CONTROL_K, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.data, should.Resemble, []rune("test "))

	cleObj.data = []rune("test command")
	buffer = []byte{CONTROL_N, 0, 0}
	cleObj.cursorPosition = 5
	cleObj.handleControlKeys(1, buffer)
	this.So(cleObj.cursorPosition, should.Equal, 0)
	this.So(len(cleObj.data), should.Equal, 0)
}

func (this *CLEFixture) TestHistoryHandledUpArrow() {
	cleObj := NewCLE(TestMode(true))
	cleObj.history.commands = cleObj.history.commands[:0]
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte(""))

	cleObj.data = []rune("this is a history entry")
	cleObj.saveHistoryEntry()
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte(""))

	cleObj.handledUpArrow()
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte("this is a history entry"))

	cleObj.history.currentPosition = -1
	this.So(cleObj.handledUpArrow(), should.BeFalse)
	this.So(cleObj.history.currentPosition, should.BeZeroValue)
}

func (this *CLEFixture) TestHistoryHandledUpArrowWithDefaultSearchChar() {
	cleObj := NewCLE(TestMode(true))
	cleObj.history.commands = cleObj.history.commands[:0]
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte(""))

	cleObj.data = []rune("this is a history entry 1")
	cleObj.saveHistoryEntry()
	cleObj.data = []rune("this is a history entry 2")
	cleObj.saveHistoryEntry()
	cleObj.data = []rune("another history entry")
	cleObj.saveHistoryEntry()

	cleObj.data = []rune(":is a")
	cleObj.handledUpArrow()
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte("this is a history entry 2"))
	cleObj.handledUpArrow()
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte("this is a history entry 1"))
	cleObj.handledDownArrow()
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte("this is a history entry 2"))
	cleObj.handledDownArrow()
	this.So(cleObj.data, should.Resemble, []rune(":is a"))
	this.So(cleObj.history.currentPosition, should.Equal, 1)
}

func (this *CLEFixture) TestHistoryHandledUpArrowWithOptionSearchChar() {
	cleObj := NewCLE(TestMode(true), SearchModeChar('~'))
	cleObj.history.commands = cleObj.history.commands[:0]
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte(""))

	cleObj.data = []rune("this is a history entry")
	cleObj.saveHistoryEntry()
	cleObj.data = []rune("another history entry")
	cleObj.saveHistoryEntry()

	cleObj.data = []rune("~this")
	cleObj.handledUpArrow()
	this.So(cleObj.getCurrentHistoryEntry(), should.Resemble, []byte("this is a history entry"))
}

func (this *CLEFixture) TestHistoryHandledDownArrow() {
	cleObj := NewCLE(TestMode(true))
	cleObj.history.commands = cleObj.history.commands[:0]
	cleObj.data = []rune("this is a history entry")
	cleObj.saveHistoryEntry()
	this.So(len(cleObj.history.commands), should.Equal, 1)

	cleObj.data = []rune("this is a history entry 2")
	cleObj.saveHistoryEntry()
	this.So(len(cleObj.history.commands), should.Equal, 2)
	cleObj.saveHistoryEntry()
	this.So(len(cleObj.history.commands), should.Equal, 2)

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
	cleObj := NewCLE(TestMode(true))
	cleObj.data = []rune("1234")

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
	cleObj := NewCLE(TestMode(true))
	cleObj.history.commands = cleObj.history.commands[:0]
	cleObj.data = []rune("this is a history entry")
	cleObj.saveHistoryEntry()

	cleObj.data = []rune("this is a history entry 2")
	cleObj.saveHistoryEntry()
	cleObj.history.currentPosition = len(cleObj.history.commands) - 1

	cleObj.populateDataWithHistoryEntry()
	this.So(cleObj.data, should.Resemble, []rune("this is a history entry 2"))
}

func (this *CLEFixture) TestClearHistory() {
	cleObj := NewCLE(TestMode(true))

	cleObj.history.commands = append(cleObj.history.commands, []byte("testing1"))
	cleObj.history.commands = append(cleObj.history.commands, []byte("testing2"))
	this.So(len(cleObj.history.commands), should.Equal, 2)

	cleObj.ClearHistory()
	this.So(len(cleObj.history.commands), should.BeZeroValue)
}

func (this *CLEFixture) TestPrepareHistoryForWriting() {
	cleObj := NewCLE(TestMode(true))

	cleObj.data = []rune("this is a history entry")
	cleObj.saveHistoryEntry()

	cleObj.data = []rune("this is a history entry 2")
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
	cleObj := NewCLE(TestMode(true))

	reader := bytes.NewReader([]byte("history entry\nhistory entry 2"))
	scanner := bufio.NewScanner(reader)
	cleObj.loadHistory(scanner)

	this.So(cleObj.history.currentPosition, should.Equal, 2)
	this.So(len(cleObj.history.commands), should.Equal, 2)
	this.So(bytes.Contains(cleObj.history.commands[1], []byte("entry 2")), should.BeTrue)
}

func (this *CLEFixture) TestInsert() {
	data := []rune("abcdef")

	this.So(insert(data, 2, 'a'), should.Resemble, []rune("abacdef"))
	this.So(insert(data, 0, 'a'), should.Resemble, []rune("aabcdef"))
	this.So(insert(data, 99, 'a'), should.Resemble, []rune("abcdefa"))

	// Multibyte runes are inserted as a single element, not per-byte.
	this.So(insert([]rune("café"), 2, 'X'), should.Resemble, []rune("caXfé"))
}

func (this *CLEFixture) TestRemove() {
	data := []rune("abcdef")

	this.So(remove(data, 2), should.Resemble, []rune("abdef"))
	this.So(remove(data, 0), should.Resemble, []rune("bcdef"))
	this.So(remove(data, 99), should.Resemble, []rune("abcdef"))

	// Removing a multibyte rune removes the whole character.
	this.So(remove([]rune("café"), 3), should.Resemble, []rune("caf"))
}

func (this *CLEFixture) TestHandleEnterKeyWithSearchModePrefix() {
	cleObj := NewCLE(TestMode(true))
	cleObj.data = []rune(":search term")
	handled := cleObj.handleEnterKey(1, []byte{ENTER_KEY, 0, 0})
	this.So(handled, should.BeTrue)
	this.So(len(cleObj.data), should.BeZeroValue)
}

func (this *CLEFixture) TestSaveHistoryEntryMinimumLength() {
	cleObj := NewCLE(TestMode(true))
	cleObj.history.commands = cleObj.history.commands[:0]

	cleObj.data = []rune("hi") // 2 chars, below default minimum of 5
	cleObj.saveHistoryEntry()
	this.So(len(cleObj.history.commands), should.BeZeroValue)

	cleObj.data = []rune("hello world") // 11 chars, above minimum
	cleObj.saveHistoryEntry()
	this.So(len(cleObj.history.commands), should.Equal, 1)
}

func (this *CLEFixture) TestHistoryEntryIndependenceAfterMutation() {
	cleObj := NewCLE(TestMode(true))
	cleObj.history.commands = cleObj.history.commands[:0]

	cleObj.data = []rune("original command")
	cleObj.saveHistoryEntry()

	// Mutate the original backing array in-place
	for i := range cleObj.data {
		cleObj.data[i] = 'X'
	}

	this.So(cleObj.history.commands[0], should.Resemble, []byte("original command"))
}

func (this *CLEFixture) TestHistoryFileRoundTrip() {
	tmpFile, err := os.CreateTemp("", "cle-history-test-*")
	this.So(err, should.BeNil)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	writer := NewCLE(TestMode(true), HistoryFile(tmpFile.Name()))
	writer.data = []rune("first command here")
	writer.saveHistoryEntry()
	writer.data = []rune("second command here")
	writer.saveHistoryEntry()
	writer.SaveHistory()

	reader := NewCLE(TestMode(true), HistoryFile(tmpFile.Name()))
	this.So(len(reader.history.commands), should.Equal, 2)
	this.So(bytes.Contains(reader.history.commands[0], []byte("first")), should.BeTrue)
	this.So(bytes.Contains(reader.history.commands[1], []byte("second")), should.BeTrue)
}

func (this *CLEFixture) TestClearHistoryDeletesFile() {
	tmpFile, err := os.CreateTemp("", "cle-history-test-*")
	this.So(err, should.BeNil)
	tmpFile.Close()

	cleObj := NewCLE(TestMode(true), HistoryFile(tmpFile.Name()))
	cleObj.history.commands = append(cleObj.history.commands, []byte("some command entry"))
	cleObj.ClearHistory()

	this.So(len(cleObj.history.commands), should.BeZeroValue)
	_, statErr := os.Stat(tmpFile.Name())
	this.So(os.IsNotExist(statErr), should.BeTrue)
}

func (this *CLEFixture) TestHandlePasteWithUnprintableBytes() {
	cleObj := NewCLE(TestMode(true))
	cleObj.handlePaste([]byte{'a', 1, 'b'}) // 1 is a control char, should be skipped
	this.So(cleObj.data, should.Resemble, []rune("ab"))
}

func (this *CLEFixture) TestHandlePasteWithAccentedCharacters() {
	cleObj := NewCLE(TestMode(true))

	// "Á" is U+00C1, UTF-8: 0xC3 0x81 -- both bytes are >= 0x80.
	cleObj.handlePaste([]byte("Á"))
	this.So(cleObj.data, should.Resemble, []rune("Á"))
	this.So(string(cleObj.data), should.Equal, "Á")

	// Accented characters mixed with ASCII and control bytes in one paste.
	cleObj = NewCLE(TestMode(true))
	cleObj.handlePaste([]byte("caf\xc3\xa9\x01!")) // "café" + control byte + "!"
	this.So(string(cleObj.data), should.Equal, "café!")
}

func (this *CLEFixture) TestHandlePasteCarriesIncompleteTrailingRune() {
	cleObj := NewCLE(TestMode(true))

	// A two-byte character ("é" = 0xC3 0xA9) split across two reads: the first
	// read ends mid-character, so its trailing byte must be carried over.
	carry := cleObj.handlePaste([]byte{'a', 0xC3})
	this.So(cleObj.data, should.Resemble, []rune("a"))
	this.So(carry, should.Resemble, []byte{0xC3})

	// The completing byte arrives with the next read, prepended to the carry.
	carry = cleObj.handlePaste(append(carry, 0xA9))
	this.So(string(cleObj.data), should.Equal, "aé")
	this.So(carry, should.BeEmpty)
}

func (this *CLEFixture) TestCursorMovementIsRuneAware() {
	cleObj := NewCLE(TestMode(true))
	cleObj.data = []rune("áé") // two 2-byte characters, two cursor positions

	cleObj.cursorPosition = 0
	this.So(cleObj.handledRightArrow(), should.BeTrue)
	this.So(cleObj.cursorPosition, should.Equal, 1) // between the two runes
	this.So(cleObj.handledRightArrow(), should.BeTrue)
	this.So(cleObj.cursorPosition, should.Equal, 2) // end of input
	this.So(cleObj.handledRightArrow(), should.BeFalse)

	this.So(cleObj.handledLeftArrow(), should.BeTrue)
	this.So(cleObj.cursorPosition, should.Equal, 1)
}

func (this *CLEFixture) TestDeleteIsRuneAware() {
	cleObj := NewCLE(TestMode(true))

	// Backspace removes a whole multibyte character, not a single byte.
	cleObj.data = []rune("ábé")
	cleObj.cursorPosition = 2 // just after 'b'
	cleObj.handleDeleteKey(1, []byte{DELETE_KEY, 0, 0})
	this.So(string(cleObj.data), should.Equal, "áé")
	this.So(cleObj.cursorPosition, should.Equal, 1)

	// Ctrl+D (delete forward) removes the whole multibyte character at cursor.
	cleObj.data = []rune("áé")
	cleObj.cursorPosition = 0
	cleObj.handleControlKeys(1, []byte{CONTROL_D, 0, 0})
	this.So(string(cleObj.data), should.Equal, "é")
	this.So(cleObj.cursorPosition, should.Equal, 0)
}

func (this *CLEFixture) TestWordDeleteIsRuneAware() {
	cleObj := NewCLE(TestMode(true))

	// Word delete operates on characters, leaving surrounding multibyte text intact.
	cleObj.data = []rune("café über")
	cleObj.cursorPosition = len(cleObj.data) // past last char
	cleObj.handleControlKeys(1, []byte{CONTROL_W, 0, 0})
	this.So(string(cleObj.data), should.Equal, "café ")
	this.So(cleObj.cursorPosition, should.Equal, 5)
}

func (this *CLEFixture) TestHandledAltLeftArrow() {
	cleObj := NewCLE(TestMode(true))
	cleObj.data = []rune("hello world")

	// inside a word → beginning of that word
	cleObj.cursorPosition = 8 // on 'r' in "world"
	cleObj.handledAltLeftArrow()
	this.So(cleObj.cursorPosition, should.Equal, 6) // 'w'

	// at start of a word → beginning of previous word
	cleObj.cursorPosition = 6 // on 'w'
	cleObj.handledAltLeftArrow()
	this.So(cleObj.cursorPosition, should.Equal, 0) // 'h'

	// in whitespace → beginning of previous word
	cleObj.cursorPosition = 5 // on ' '
	cleObj.handledAltLeftArrow()
	this.So(cleObj.cursorPosition, should.Equal, 0)

	// at beginning → stays at 0
	cleObj.cursorPosition = 0
	cleObj.handledAltLeftArrow()
	this.So(cleObj.cursorPosition, should.BeZeroValue)
}

func (this *CLEFixture) TestHandledAltRightArrow() {
	cleObj := NewCLE(TestMode(true))
	cleObj.data = []rune("hello world")

	// inside a word → first whitespace after that word
	cleObj.cursorPosition = 2 // on 'l' in "hello"
	cleObj.handledAltRightArrow()
	this.So(cleObj.cursorPosition, should.Equal, 5) // the space

	// in whitespace → first whitespace after next word (end of string)
	cleObj.cursorPosition = 5 // on ' '
	cleObj.handledAltRightArrow()
	this.So(cleObj.cursorPosition, should.Equal, 11)

	// at start of last word → end of string
	cleObj.cursorPosition = 6 // on 'w'
	cleObj.handledAltRightArrow()
	this.So(cleObj.cursorPosition, should.Equal, 11)

	// at end → stays at end
	cleObj.cursorPosition = 11
	cleObj.handledAltRightArrow()
	this.So(cleObj.cursorPosition, should.Equal, 11)
}

func (this *CLEFixture) TestHandleArrowKeysAltBackspace() {
	cleObj := NewCLE(TestMode(true))
	cleObj.data = []rune("hello world")
	cleObj.cursorPosition = 11 // past last char

	// ESC DEL: Alt+Backspace deletes word to the left
	cleObj.handleArrowKeys(2, []byte{ESCAPE_KEY, DELETE_KEY, 0, 0, 0, 0})
	this.So(cleObj.data, should.Resemble, []rune("hello "))
	this.So(cleObj.cursorPosition, should.Equal, 6)
}

func (this *CLEFixture) TestHandleArrowKeysWordDeleteRight() {
	cleObj := NewCLE(TestMode(true))
	cleObj.data = []rune("hello world")
	cleObj.cursorPosition = 6 // on 'w'

	// ESC d: readline/xterm-style Alt+D
	cleObj.handleArrowKeys(2, []byte{ESCAPE_KEY, 'd', 0, 0, 0, 0})
	this.So(cleObj.data, should.Resemble, []rune("hello "))

	cleObj.data = []rune("hello world")
	cleObj.cursorPosition = 6

	// macOS Terminal.app Option+D sends ∂ (U+2202, UTF-8: 0xE2 0x88 0x82)
	cleObj.handleArrowKeys(3, []byte{0xE2, 0x88, 0x82, 0, 0, 0})
	this.So(cleObj.data, should.Resemble, []rune("hello "))
}

func (this *CLEFixture) TestHandledWordDeleteRight() {
	cleObj := NewCLE(TestMode(true))
	cleObj.data = []rune("hello world")

	// cursor in the middle of a word → delete from cursor to end of word
	cleObj.cursorPosition = 2 // on 'l' in "hello"
	cleObj.handledWordDeleteRight()
	this.So(cleObj.data, should.Resemble, []rune("he world"))
	this.So(cleObj.cursorPosition, should.Equal, 2)

	// cursor at start of a word → delete the whole word
	cleObj.data = []rune("hello world")
	cleObj.cursorPosition = 6 // on 'w'
	cleObj.handledWordDeleteRight()
	this.So(cleObj.data, should.Resemble, []rune("hello "))
	this.So(cleObj.cursorPosition, should.Equal, 6)

	// cursor on whitespace → skip whitespace, delete next word
	cleObj.data = []rune("hello world")
	cleObj.cursorPosition = 5 // on ' '
	cleObj.handledWordDeleteRight()
	this.So(cleObj.data, should.Resemble, []rune("hello"))
	this.So(cleObj.cursorPosition, should.Equal, 5)

	// cursor at end → nothing deleted
	cleObj.data = []rune("hello world")
	cleObj.cursorPosition = 11 // past last char
	cleObj.handledWordDeleteRight()
	this.So(cleObj.data, should.Resemble, []rune("hello world"))
	this.So(cleObj.cursorPosition, should.Equal, 11)
}

func (this *CLEFixture) TestHandleWordDeleteLeft() {
	cleObj := NewCLE(TestMode(true))

	// Cursor in the middle of a word: delete chars to the left until space (char at cursor is not deleted)
	cleObj.data = []rune("hello world")
	cleObj.cursorPosition = 8 // on 'r'
	cleObj.handleControlKeys(1, []byte{CONTROL_W, 0, 0})
	this.So(cleObj.data, should.Resemble, []rune("hello rld"))
	this.So(cleObj.cursorPosition, should.Equal, 6)

	// Cursor past end: delete word to the left until space
	cleObj.data = []rune("hello world")
	cleObj.cursorPosition = 11 // past last char
	cleObj.handleControlKeys(1, []byte{CONTROL_W, 0, 0})
	this.So(cleObj.data, should.Resemble, []rune("hello "))
	this.So(cleObj.cursorPosition, should.Equal, 6)

	// Cursor at start of word (char to left is space): delete space and word to the left, char at cursor is not deleted
	cleObj.data = []rune("hello world")
	cleObj.cursorPosition = 6 // on 'w', data[5]==' '
	cleObj.handleControlKeys(1, []byte{CONTROL_W, 0, 0})
	this.So(cleObj.data, should.Resemble, []rune("world"))
	this.So(cleObj.cursorPosition, should.BeZeroValue)

	// Cursor past end after trailing space: delete space and word to the left
	cleObj.data = []rune("hello ")
	cleObj.cursorPosition = 6 // past trailing space
	cleObj.handleControlKeys(1, []byte{CONTROL_W, 0, 0})
	this.So(cleObj.data, should.BeEmpty)
	this.So(cleObj.cursorPosition, should.BeZeroValue)

	// No whitespace to left: delete back to beginning of line, char at cursor not deleted
	cleObj.data = []rune("hello")
	cleObj.cursorPosition = 3 // on 'l'
	cleObj.handleControlKeys(1, []byte{CONTROL_W, 0, 0})
	this.So(cleObj.data, should.Resemble, []rune("lo"))
	this.So(cleObj.cursorPosition, should.BeZeroValue)

	// Cursor at position 0: nothing to the left, nothing deleted
	cleObj.data = []rune("hello world")
	cleObj.cursorPosition = 0 // on 'h'
	cleObj.handleControlKeys(1, []byte{CONTROL_W, 0, 0})
	this.So(cleObj.data, should.Resemble, []rune("hello world"))
	this.So(cleObj.cursorPosition, should.BeZeroValue)
}

func (this *CLEFixture) TestHandleControlKeysUnrecognized() {
	cleObj := NewCLE(TestMode(true))
	cleObj.data = []rune("some data")
	cleObj.cursorPosition = 4

	// CTRL+C (3) satisfies isControlKey but has no case in the switch
	handled := cleObj.handleControlKeys(1, []byte{3, 0, 0})
	this.So(handled, should.BeTrue)
	this.So(cleObj.data, should.Resemble, []rune("some data"))
	this.So(cleObj.cursorPosition, should.Equal, 4)
}
