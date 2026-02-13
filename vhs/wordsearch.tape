Output vhs/wordsearch.gif

Set Theme "Solarized Darcula"
Set FontSize 16
Set Width 1200
Set Height 800
Set Margin 40
Set MarginFill "#9C7A5B"
Set WindowBar ColorfulRight
Set BorderRadius 10

Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"
Env CI ""

Hide
Type "go build -o puzzletea"
Enter
Sleep 2s
Ctrl+L
Show

Type "./puzzletea new word-search hard-20x20"
Sleep 500ms
Enter
Sleep 2s

# Navigate through the grid
Down
Down
Right
Right
Right
Sleep 500ms

# Select start of a word
Enter
Sleep 500ms

# Extend selection to the right
Right
Right
Right
Sleep 500ms

# Show hint toggle
Ctrl+H
Sleep 1s
Ctrl+H
Sleep 500ms
Ctrl+H

# Confirm end of selection
Enter
Sleep 500ms

# Navigate elsewhere
Down
Down
Sleep 300ms

# Start another selection
Enter
Sleep 300ms

# Extend diagonally
Right
Down
Right
Down
Sleep 300ms

# Cancel the selection
Escape
Sleep 500ms

Sleep 3s
