package main

import (
	"os"

	"github.com/sheshisheri-hi/token-optimizer/internal/cli"
)

func main() {
	if err := cli.NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
