Output vhs/hitori.gif

Set Theme "Monokai Pro"
Set FontSize 14
Set Width 800
Set Height 700
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

Type "./puzzletea new hitori medium"
Sleep 500ms
Enter
Sleep 2s

# Shade a cell
Type "z"
Sleep 500ms

# Move and shade another
Right
Right
Type "z"
Sleep 500ms

# Clear a cell
Backspace
Sleep 500ms

# Refill it
Type "z"
Sleep 500ms

# Navigate around
Down
Right
Right
Down

# Shade more
Type "z"
Sleep 500ms

Right
Right
Down
Type "z"
Sleep 500ms

Sleep 3s
