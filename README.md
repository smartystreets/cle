#### SMARTY DISCLAIMER: Subject to the terms of the associated license agreement, this software is freely available for your use. This software is FREE, AS IN PUPPIES, and is a gift. Enjoy your new responsibility. This means that while we may consider enhancement requests, we may or may not choose to entertain requests at our sole and absolute discretion.

Command-Line Editor (CLE)
============================

The Command-Line Editor (CLE) Go library provides basic command-line editing and command history functions.

#### Features

 * Command editing
 * Command history; persistence across sessions
 * Support for most Linux/Unix based systems (MS Windows is not supported)
 
## Import
```
import github.com/smartystreets/cle
```

## Usage 
```
commandLineEditor := cle.NewCLE()
command := commandLineEditor.ReadInput("Enter something: ")
``` 

### Options
Specify any number of comma separated options as parameters to `NewCLE()`

#### Command History File
Enable command history load/save. If file is specified, the command history is loaded
automatically when `NewCLE()` is called. You must explicitly save the history as shown below.
```
commandLineEditor := cle.NewCLE(cle.HistoryFile("/tmp/cle_history.txt"))

...

commandLineEditor.SaveHistory()
```

#### Command History Size
Only save/load the specified number of commands in the history file. (Default `100`)

```
cle.HistorySize(50)
```

#### Command History Entry Minimum Length
Only add commands to the history that exceed this length. (Default `5`)

```
cle.HistoryEntryMinimumLength(2)
```

#### Search Mode Character
Set search mode character. (Default ':')

```
cle.SearchModeChar('!')
```

#### Print Errors
Debugging: Print errors to the console. (Default `false`)
 
```
cle.ReportErrors(true)
```

## Command Editing Keys
* `CTL-A` - Move to beginning of line
* `CTL-B` - Delete to beginning of line
* `CTL-D` - Delete current character
* `CTL-E` - Move to end of line
* `CTL-K` - Delete current character to end of line
* `CTL-N` - Delete entire line

## Searching History
You can search through the command stack by typing the Search Mode Character (default is `:`)
then any string to search for, then press `<up arrow>`. (You can change the attention character using the `SearchModeChar` option.)

### Example:
`:mine<up arrow>` will search backwards through history for the first occurrence of the word `mine`

Continue pressing `<up arrow>` or `<down arrow>` to search for other matches.

If you get back to the bottom of the stack it will display your search term. 
One more `<down arrow>` and the search will be cancelled. 
The search will also be cancelled if the `<left arrow>`, `<right arrow>` or `Enter` is pressed at any time during the search.

## Clearing History
Clear the command history by entering the command: `!clear`

## Example Code
See `cmd/main.go` for a fully functional sample.
