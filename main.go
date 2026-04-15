package main

import (
	"os"

	"github.com/tudou/iconkit/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
