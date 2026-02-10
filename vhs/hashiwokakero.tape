Output vhs/hashiwokakero.gif

Set FontSize 13
Set Width 854
Set Height 480
Set Padding 40
Set Theme "Catppuccin Mocha"

Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"
Env CI ""

Hide
Type "go build -o puzzletea"
Enter
Sleep 2s
Ctrl+L
Show

Type "./puzzletea new hashi easy-7x7"
Sleep 500ms
Enter
Sleep 2s

# Navigate between islands
Right
Sleep 400ms
Right
Sleep 400ms

# Select island to enter bridge mode
Enter
Sleep 500ms

# Build a bridge to the right
Right
Sleep 500ms

# Deselect
Enter
Sleep 500ms

# Navigate to another island
Down
Down
Sleep 400ms

# Select and build downward bridge
Enter
Sleep 500ms
Down
Sleep 500ms

# Build double bridge (cycle 1 -> 2)
Down
Sleep 500ms

# Deselect
Escape
Sleep 500ms

Sleep 3s
