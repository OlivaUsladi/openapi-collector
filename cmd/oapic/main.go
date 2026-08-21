package main

import (
	"errors"
	"fmt"
	"openapi-collector/internal/app"
	"os"
)

func main() {
	err := app.Run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		if errors.Is(err, app.ErrIssues) {
			os.Exit(1)
		}
		os.Exit(2)
	}
}
