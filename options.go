package cle

// Option is a func type received by CLE.
// Each one allows configuration of the CLE.
type Option func(*CLE)

func HistoryFile(fileName string) Option {
	return func(c *CLE) { c.historyFile = fileName }
}

func HistorySize(historyMax int) Option {
	return func(c *CLE) { c.historyMax = historyMax }
}

func HistoryEntryMinimumLength(historyEntryMinLen int) Option {
	return func(c *CLE) { c.historyEntryMinimumLength = historyEntryMinLen }
}

func ReportErrors(reportErrors bool) Option {
	return func(c *CLE) { c.reportErrors = reportErrors }
}

func SearchModeChar(searchMode byte) Option {
	return func(c *CLE) { c.searchModeChar = searchMode }
}

// disables terminal output for testing
func TestMode(testMode bool) Option {
	return func(c *CLE) { c.testMode = testMode }
}
