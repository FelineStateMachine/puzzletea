Output vhs/shikaku.gif

Set Theme "catppuccin-mocha"
Set FontSize 14
Set Width 800
Set Height 800
Set Margin 60
Set MarginFill "#7A6855"
Set WindowBar Colorful
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

Type "./puzzletea new shikaku easy-7x7"
Sleep 500ms
Enter
Sleep 2s

# Navigate to a clue
Right
Sleep 400ms
Down
Sleep 400ms

# Select clue to enter expansion mode
Enter
Sleep 500ms

# Expand rightward
Right
Sleep 400ms
Right
Sleep 400ms

# Expand downward
Down
Sleep 400ms

# Confirm placement
Enter
Sleep 500ms

# Navigate to another clue
Down
Sleep 400ms
Down
Sleep 400ms

# Select and expand
Enter
Sleep 500ms
Right
Sleep 400ms
Right
Sleep 400ms
Right
Sleep 400ms

# Shrink one step back
Type "L"
Sleep 400ms

# Confirm
Enter
Sleep 500ms

# Navigate to another clue
Left
Sleep 400ms
Left
Sleep 400ms

# Select and expand vertically
Enter
Sleep 500ms
Down
Sleep 400ms
Down
Sleep 400ms

# Cancel to discard preview
Escape
Sleep 500ms

Sleep 3s
