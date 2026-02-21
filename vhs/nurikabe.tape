Output vhs/nurikabe.gif

Set Theme "GitHub Dark"
Set FontSize 14
Set Width 1200
Set Height 840
Set Margin 60
Set MarginFill "#7B6753"
Set WindowBar ColorfulRight
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

Type "./puzzletea new nurikabe easy --theme 'Github Dark'"
Sleep 500ms
Enter
Sleep 4s

# Move and mark sea
Right
Sleep 300ms
Type "x"
Sleep 500ms

# Move and set island
Down
Sleep 300ms
Right
Sleep 300ms
Type "z"
Sleep 500ms

# Demonstrate clearing
Backspace
Sleep 500ms

# Re-apply island
Type "z"
Sleep 500ms

# Continue navigation and marking
Left
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Type "x"
Sleep 500ms
Right
Sleep 300ms
Right
Sleep 300ms
Down
Sleep 300ms
Type "x"
Sleep 500ms

Sleep 3s
