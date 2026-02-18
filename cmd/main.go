package main

import (
	"bytes"

	"github.com/smartystreets/cle"
)

func main() {
	commandLineEditor := cle.NewCLE(
		cle.HistoryFile("/tmp/cle_history.txt"),
		cle.ReportErrors(true),
	)

	for {
		data := commandLineEditor.ReadInput("Enter string: ")
		commandLineEditor.SaveHistory()
		if len(data) == 1 && bytes.ToLower(data)[0] == 'q' {
			break
		}
	}
}
