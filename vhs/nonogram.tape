Output vhs/nonogram.gif

Set Theme "catppuccin-mocha"
Set FontSize 14
Set Width 1200
Set Height 800
Set Margin 60
Set MarginFill "#A85C40"
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
