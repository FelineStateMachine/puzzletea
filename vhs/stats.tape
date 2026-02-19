Output vhs/stats.gif

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

Type "./puzzletea"
Sleep 500ms
Enter
Sleep 2s

# Navigate to Stats (2nd item in main menu)
Down
Sleep 400ms

# Select Stats
Enter
Sleep 2s

# Scroll down to browse stats cards
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

# Scroll back up
Up
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms

Sleep 1s

# Return to main menu
Escape
Sleep 1s

Sleep 2s
