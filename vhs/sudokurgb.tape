Output vhs/sudokurgb.gif

Set Theme "GitHub Dark"
Set FontSize 16
Set Width 900
Set Height 900
Set Margin 60
Set MarginFill "#2F3E5C"
Set WindowBar Rings
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

Type './puzzletea --theme "GitHub Dark" new "sudoku rgb" medium --with-seed "vhs-sudokurgb-medium"'
Sleep 500ms
Enter
Sleep 2s

# Move through the grid to show cursor and house highlighting.
Right
Right
Down
Down
Sleep 400ms

# Place a few values, then clear one to show the interaction loop.
Type "1"
Sleep 500ms
Right
Type "2"
Sleep 500ms
Down
Type "3"
Sleep 700ms
Backspace
Sleep 500ms

# Toggle full help briefly, then continue navigating.
Ctrl+H
Sleep 1s
Ctrl+H
Sleep 500ms

Left
Up
Up
Right
Right
Down
Down
Down
Sleep 800ms

Sleep 3s
