Output vhs/takuzu.gif

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

Type "./puzzletea new takuzu easy"
Sleep 500ms
Enter
Sleep 2s

# Place a filled circle
Type "z"
Sleep 500ms

# Move and place an open circle
Right
Sleep 300ms
Type "x"
Sleep 500ms

# Move down and place another
Down
Sleep 300ms
Type "z"
Sleep 500ms

# Demonstrate clearing
Backspace
Sleep 500ms

# Re-place with opposite value
Type "x"
Sleep 500ms

# Navigate around to show crosshair
Down
Down
Sleep 300ms
Right
Right
Sleep 300ms

# Place a few more
Type "z"
Sleep 500ms
Down
Sleep 300ms
Type "x"
Sleep 500ms

Sleep 3s
