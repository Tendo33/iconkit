package main

import (
	"os"

	"github.com/Tendo33/iconkit/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
