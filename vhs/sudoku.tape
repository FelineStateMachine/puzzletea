Output vhs/sudoku.gif

Set FontSize 13
Set Width 854
Set Height 480
Set Padding 40

Env COLORTERM "truecolor"

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
