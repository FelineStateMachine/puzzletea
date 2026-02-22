Output vhs/lightsout.gif

Set Theme "GitHub Dark"
Set FontSize 18
Set Width 800
Set Height 880
Set Margin 60
Set MarginFill "#8A7260"
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

Type './puzzletea --theme "GitHub Dark" new lights-out hard --with-seed "vhs-lightsout-hard"'
Sleep 500ms
Enter
Sleep 2s

# Toggle center cell -- shows cross pattern
Enter
Sleep 800ms

# Move to top-left corner and toggle
Up
Sleep 300ms
Left
Sleep 300ms
Enter
Sleep 800ms

# Move right and toggle
Right
Right
Sleep 300ms
Enter
Sleep 800ms

# Move down and toggle
Down
Sleep 300ms
Enter
Sleep 800ms

Sleep 3s
