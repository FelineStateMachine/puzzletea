Output vhs/netwalk.gif

Set Theme "catppuccin-mocha"
Set FontSize 14
Set Width 1200
Set Height 800
Set Margin 60
Set MarginFill "#6A5B4A"
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

Type './puzzletea --theme "catppuccin-mocha" new netwalk "Easy 7x7" --with-seed "vhs-netwalk-easy-7x7"'
Sleep 500ms
Enter
Sleep 2s

# Rotate the starting tile and lock it in place.
Type " "
Sleep 500ms
Enter
Sleep 500ms

# Move to a branch and show reverse rotation.
Right
Sleep 350ms
Right
Sleep 350ms
Down
Sleep 350ms
Backspace
Sleep 500ms

# Unlock and rotate a different tile.
Left
Sleep 350ms
Down
Sleep 350ms
Type " "
Sleep 500ms
Enter
Sleep 500ms

# Sweep across the board to show the network changing color/state.
Right
Sleep 300ms
Down
Sleep 300ms
Right
Sleep 300ms
Up
Sleep 300ms
Type " "
Sleep 500ms

Sleep 3s
