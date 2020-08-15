package main

import (
	"bytes"

	"bitbucket.org/smartybryan/cle"
)

func main() {
	commandLineEditor := cle.NewCLE(
		cle.HistoryFile("/tmp/cle_history.txt"),
		cle.ReportErrors(true),
	)

	for {
		data := commandLineEditor.ReadInput("Enter string: ")
		if len(data) == 1 && bytes.ToLower(data)[0] == 'q' {
			commandLineEditor.SaveHistory()
			break
		}
	}
}
