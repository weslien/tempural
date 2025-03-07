// Package main provides a simplified entrypoint for the unregex tool
// This allows users to install with: go install github.com/weslien/unregex@v0.1.0
package main

import (
	"os"

	"github.com/weslien/tempural/internal/app"
)

func main() {
	cliApp := app.NewTemporalCLI()
	if err := cliApp.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
