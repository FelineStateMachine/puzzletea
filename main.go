package main

import (
	"os"

	"github.com/FelineStateMachine/puzzletea/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
