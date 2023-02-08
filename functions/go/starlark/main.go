package main

import (
	"os"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
