Output vhs/menu.gif

Set Theme "Treehouse"
Set FontSize 14
Set Width 1200
Set Height 800
Set Margin 60
Set MarginFill "#8B6B4A"
Set WindowBar RingsRight
Set BorderRadius 14


Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"
Env CI ""

Hide
Type "go build -o puzzletea"
Enter
Sleep 2s
Ctrl+L
Show

Type './puzzletea --theme "Treehouse"'
Sleep 500ms
Enter
Sleep 2s

# Browse main menu (Play, Stats, Options, Quit)
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms
Up
Up
Up
Sleep 500ms

# Select "Play" (first item)
Enter
Sleep 1s

# Browse Play submenu (Daily Puzzle, Generate, Continue, Find)
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms
Up
Up
Up
Sleep 500ms

# Select "Generate" (second item in Play submenu)
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

# Select Nonogram (fourth item)
Up
Enter
Sleep 1s

# Browse modes
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Down
Sleep 600ms


# Go back through menus
Escape
Sleep 1s
Escape
Sleep 500ms
Escape
Sleep 500ms

# Show Options submenu
Down
Down
Enter
Sleep 1s

# Browse Options (Theme, Guides)
Down
Sleep 500ms
Up
Sleep 500ms

Escape
Sleep 500ms

Sleep 2s
