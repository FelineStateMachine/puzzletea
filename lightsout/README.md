# Lights Out

Toggle lights in a cross pattern to turn them all off.

![Lights Out gameplay](../vhs/lightsout.gif)

## How to Play

The grid starts with some lights on and some off. Toggling a cell flips
that cell and its four orthogonal neighbors (up, down, left, right). The
goal is to turn every light off.

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| `Enter` / `Space` | Toggle light |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Ctrl+N` | Return to main menu |

## Modes

| Mode | Grid | Description |
|------|------|-------------|
| Easy | 3x3 | Small grid |
| Medium | 5x5 | Classic size |
| Hard | 7x7 | Large grid |
| Extreme | 9x9 | Maximum size |

## Quick Start

```bash
puzzletea new lights-out easy
puzzletea new lights-out medium
puzzletea new lights-out extreme
```
