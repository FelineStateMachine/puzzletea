Output vhs/lightsout.gif

Set FontSize 13
Set Width 854
Set Height 480
Set Padding 40

Env COLORTERM "truecolor"

Hide
Type "go build -o puzzletea"
Enter
Sleep 2s
Ctrl+L
Show

Type "./puzzletea new lights-out easy"
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
