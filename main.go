// Package main provides a simplified entrypoint for the tempural tool
// This allows users to install with: go install github.com/weslien/unregex@v0.1.0
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	_ "net/http/pprof" // Import for side effects (registers HTTP handlers)

	"github.com/weslien/tempural/internal/app"
)

func main() {
	// Create CLI app
	cliApp := app.NewTemporalCLI()

	// Start the application
	err := cliApp.Run(os.Args)

	// Always close profile files if they exist
	if cpuFile := app.GetCPUProfileFile(); cpuFile != nil {
		pprof.StopCPUProfile()
		cpuFile.Close()
		fmt.Println("CPU profile written")
	}

	if memFile := app.GetMemProfileFile(); memFile != nil {
		// Force garbage collection to get up-to-date memory info
		runtime.GC()
		if err := pprof.WriteHeapProfile(memFile); err != nil {
			fmt.Printf("Failed to write memory profile: %v\n", err)
		}
		memFile.Close()
		fmt.Println("Memory profile written")
	}

	// Exit with error code if there was an error
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
