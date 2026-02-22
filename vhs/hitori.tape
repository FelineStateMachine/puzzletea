Output vhs/hitori.gif

Set Theme "catppuccin-mocha"
Set FontSize 18
Set Width 1200
Set Height 940
Set Margin 80
Set MarginFill "#6E6352"
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

Type './puzzletea --theme "catppuccin-mocha" new hitori medium --with-seed "vhs-hitori-medium"'
Sleep 500ms
Enter
Sleep 2s

# Shade a cell
Type "x"
Sleep 500ms

# Move and circle a cell as confirmed white
Right
Sleep 300ms
Type "z"
Sleep 500ms

# Move down and shade another
Down
Sleep 300ms
Type "x"
Sleep 500ms

# Demonstrate clearing
Backspace
Sleep 500ms

# Re-shade
Type "x"
Sleep 500ms

# Navigate around to show crosshair
Down
Down
Sleep 300ms
Right
Right
Sleep 300ms

# Shade a few more
Type "x"
Sleep 500ms
Down
Sleep 300ms
Type "z"
Sleep 500ms

Sleep 3s
