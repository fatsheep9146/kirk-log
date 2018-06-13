package main

import (
	"fmt"
	"os"

	"github.com/fatsheep9146/kirklog/cmd/app"
	"github.com/spf13/pflag"
)

func main() {
	s := app.NewLogManagerServer()
	s.AddFlags(pflag.CommandLine)
	pflag.Parse()

	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
