## Purpose
Describe menu, view, theme, stats, help, and contrast behavior.

## Requirements

### Requirement: Menu And View Layout
The Bubble Tea UI SHALL provide navigable menu, create, continue, daily, weekly, stats, themes, and help screens with keyboard-driven focus management.

#### Scenario: Navigate screen
- **WHEN** the user changes menu selection or opens a submenu
- **THEN** PuzzleTea renders the new view with consistent layout and active focus indication

### Requirement: Theme Lookup And Contrast
Theme selection MUST resolve configured theme names, apply palette values, and enforce readable contrast for foreground/background combinations.

#### Scenario: Preview theme
- **WHEN** the user moves through the theme picker
- **THEN** PuzzleTea previews the selected theme and keeps text legible using contrast handling

### Requirement: Help And Stats Screens
Help and stats screens SHALL render structured game help, profile progress, category progress, streaks, and weekly status using shared UI components.

#### Scenario: Toggle help during game
- **WHEN** the user requests full help
- **THEN** PuzzleTea sends the help toggle message and renders game-specific key bindings
