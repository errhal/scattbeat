package main

import (
	"os"

	"github.com/errhal/scattbeat/cmd"

	_ "github.com/errhal/scattbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
