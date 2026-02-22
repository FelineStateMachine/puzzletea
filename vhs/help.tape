Output vhs/help.gif

Set Theme "Solarized Darcula"
Set FontSize 14
Set Width 1400
Set Height 800
Set Margin 32
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

Type './puzzletea --theme "Solarized Darcula" -h'
Enter
Sleep 5s
