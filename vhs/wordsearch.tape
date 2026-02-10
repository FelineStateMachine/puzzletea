Output vhs/wordsearch.gif

Set FontSize 13
Set Width 854
Set Height 480
Set Padding 40

Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"
Env CI ""

Hide
Type "go build -o puzzletea"
Enter
Sleep 2s
Ctrl+L
Show

Type "./puzzletea new wordsearch easy-10x10"
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
