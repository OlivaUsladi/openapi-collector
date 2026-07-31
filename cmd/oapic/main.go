package main

import (
	"fmt"
	"openapi-collector/internal/app"
	"os"
)

func main() {
	err := app.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

}
