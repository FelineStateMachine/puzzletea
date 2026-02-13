Output vhs/nonogram.gif

Set Theme "Monokai Pro"
Set FontSize 14
Set Width 1200
Set Height 860
Set Margin 60
Set MarginFill "#A85C40"
Set WindowBar ColorfulRight

Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"
Env CI ""

Hide
Type "go build -o puzzletea"
Enter
Sleep 2s
Ctrl+L
Show

Type "./puzzletea new nonogram easy-20x20"
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
Down
Down
Right
Down
Down
Down
Type "z"
Sleep 400ms
Right
Right
Right
Down
Down
Down
Right
Right
Type "z"
Sleep 400ms
Right
Type "z"
Sleep 500ms
Right
Type "z"
Sleep 300ms
Right
Type "z"
Sleep 400ms

# Demonstrate clearing a cell
Backspace
Sleep 900ms

# Refill it
Type "z"

Sleep 4s
