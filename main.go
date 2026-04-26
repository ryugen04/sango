package main

import (
	"os"

	"github.com/ryugen04/sango/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
