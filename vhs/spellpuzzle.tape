Output vhs/spellpuzzle.gif

Set Theme "Solarized Darcula"
Set FontSize 14
Set Width 1200
Set Height 800
Set Margin 40
Set MarginFill "#9C7A5B"
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

Type './puzzletea --theme "Solarized Darcula" new spell beginner --with-seed "vhs-spell-beginner"'
Sleep 500ms
Enter
Sleep 2s

Right
Enter
Right
Enter
Right
Enter
Right
Enter
Space
Sleep 500ms

Ctrl+R
Sleep 500ms

Sleep 3s
