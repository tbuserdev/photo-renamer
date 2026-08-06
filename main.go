package main

import (
	"fmt"
	"os"

	"photo-renamer/gui"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("photo-renamer version: %s\n", Version)
		return
	}

	gui.Run(Version)
}
