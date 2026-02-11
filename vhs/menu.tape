Output vhs/menu.gif

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

Type "./puzzletea"
Sleep 500ms
Enter
Sleep 2s

# Browse main menu
Down
Sleep 500ms
Down
Sleep 500ms
Up
Up
Sleep 500ms

# Select "Generate"
Down
Enter
Sleep 1s

# Browse game categories
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms

# Select Nonogram (third item)
Up
Up
Enter
Sleep 1s

# Browse modes
Down
Sleep 400ms
Down
Sleep 400ms

# Go back through menus
Escape
Sleep 500ms
Escape
Sleep 500ms

Sleep 2s
