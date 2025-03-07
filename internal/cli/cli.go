// Package cli provides the command-line interface for unregex
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/weslien/tempural/pkg/utils"
)

// Run executes the CLI application
func Run() {
	// Define command-line flags

	helpFlag := flag.Bool("help", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version information")

	// Custom usage function
	flag.Usage = func() {

	}

	// Parse command-line flags
	flag.Parse()

	// Show help message and exit
	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	// Show version information and exit
	if *versionFlag {
		fmt.Println(utils.GetVersionInfo())
		os.Exit(0)
	}

	fmt.Printf("Tempural v%s\n\n", utils.Version)

}
