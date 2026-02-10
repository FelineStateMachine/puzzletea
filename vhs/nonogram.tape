Output vhs/nonogram.gif

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

Type "./puzzletea new nonogram easy-5x5"
Sleep 500ms
Enter
Sleep 2s

# Fill a cell
Type "z"
Sleep 400ms

# Move right and fill
Right
Type "z"
Sleep 400ms

# Move right and mark
Right
Type "x"
Sleep 400ms

# Move down and left, fill more cells
Down
Left
Left
Type "z"
Sleep 400ms
Down
Type "z"
Sleep 400ms
Right
Type "z"
Sleep 400ms
Right
Type "z"
Sleep 400ms

# Demonstrate clearing a cell
Backspace
Sleep 400ms

# Refill it
Type "z"
Sleep 400ms

Sleep 3s
