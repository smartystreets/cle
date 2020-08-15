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

#### Options
Specify any number of comma separated options in `NewCLE()`

 * Enable command history load/save. If file is specified, the command history is loaded
 automatically when `NewCLE()` is called. 
 You must explicitly save the history as shown below.
```
commandLineEditor := cle.NewCLE(cle.HistoryFile("/tmp/cle_history.txt"))

...

commandLineEditor.SaveHistory()
```

 * Only save/load the specified number of commands in the history file. (Default `100`)

```
cle.HistorySize(50)
```

 * Only add commands to the history that exceed this length. (Default `5`)

```
cle.HistoryEntryMinimumLength(2)
```

 * Debugging: Print errors to the console. (Default `false`)
 
```
cle.ReportErrors(true)
```

#### Example
See `cmd/main.go` for a fully functional sample.
