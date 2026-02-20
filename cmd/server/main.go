package main

import (
	"fmt"
	"os"

	"gogogo/cmd/cli/commands/server"
)

func main() {
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
