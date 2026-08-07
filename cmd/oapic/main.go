package main

import (
	"fmt"
	"openapi-collector/internal/app"
	"os"
)

func main() {
	err := app.Run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
}
