package main

import (
	"log"

	"github.com/backtesting-org/kronos-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
