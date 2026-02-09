package main

import (
	"log"

	"github.com/FelineStateMachine/puzzletea/store"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	s, err := store.Open(store.DefaultDBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	p := tea.NewProgram(initialModel(s))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
