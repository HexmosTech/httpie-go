package main

import (
	"fmt"
	"os"

	"github.com/HexmosTech/httpie-go"
)

func main() {
	if err := httpie.Main(&httpie.Options{}); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
