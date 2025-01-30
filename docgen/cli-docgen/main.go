package main

import (
	"log"

	"github.com/jacobmcgowan/simple-scheduler/services/cli/cmd"
)

func main() {
	if err := cmd.GenMarkdownTree("../../services/cli/docs"); err != nil {
		log.Fatalf("Failed to generate documentation: %s", err)
	}
}
