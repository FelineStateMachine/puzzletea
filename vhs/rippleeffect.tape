Output vhs/rippleeffect.gif

Set Theme "catppuccin-mocha"
Set FontSize 18
Set Width 1200
Set Height 940
Set Margin 80
Set MarginFill "#6E6352"
Set WindowBar ColorfulRight

Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"
Env CI ""

Hide
Type "go build -o puzzletea"
Enter
Sleep 2s
Ctrl+L
Show

Type './puzzletea --theme "catppuccin-mocha" new ripple-effect "Medium 7x7" --with-seed "vhs-ripple-medium"'
Sleep 500ms
Enter
Sleep 2s

Type "2"
Sleep 500ms
Right
Sleep 300ms
Type "1"
Sleep 500ms
Down
Sleep 300ms
Type "3"
Sleep 500ms
Left
Sleep 300ms
Backspace
Sleep 500ms
Type "2"
Sleep 500ms
Down
Sleep 300ms
Right
Right
Sleep 300ms
Type "1"
Sleep 500ms

Sleep 3s
