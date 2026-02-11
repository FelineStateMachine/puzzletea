Output vhs/sudoku.gif

Set Theme "catppuccin-mocha"
Set FontSize 14
Set Width 1200
Set Height 800
Set Margin 60
Set MarginFill "#6B7B5E"
Set WindowBar Colorful

Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"
Env CI ""

Hide
Type "go build -o puzzletea"
Enter
Sleep 2s
Ctrl+L
Show

Type "./puzzletea new sudoku beginner"
Sleep 500ms
Enter
Sleep 2s

# Navigate to find an empty cell
Down
Down
Right
Sleep 300ms

# Fill with a number
Type "5"
Sleep 500ms

# Move to another cell
Right
Right
Sleep 300ms
Type "3"
Sleep 500ms

# Demonstrate clearing and re-entering
Backspace
Sleep 500ms
Type "8"
Sleep 500ms

# Navigate around to show highlighting
Down
Down
Down
Sleep 300ms
Right
Right
Sleep 300ms

Sleep 3s
