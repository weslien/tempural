package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
)

// Profiling file handles
var (
	cpuProfileFile *os.File
	memProfileFile *os.File
)

// GetCPUProfileFile returns the CPU profile file handle
func GetCPUProfileFile() *os.File {
	return cpuProfileFile
}

// GetMemProfileFile returns the memory profile file handle
func GetMemProfileFile() *os.File {
	return memProfileFile
}

// SetupProfiling initializes profiling based on configuration
func SetupProfiling(config TemporalConfig) error {
	if config.Debug {
		fmt.Println("Debug mode enabled")
	}

	// Setup CPU profiling if requested
	if config.CPUProfile != "" {
		var err error
		cpuProfileFile, err = os.Create(config.CPUProfile)
		if err != nil {
			return fmt.Errorf("could not create CPU profile: %w", err)
		}

		fmt.Printf("CPU profiling enabled, writing to %s\n", config.CPUProfile)
		if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
			cpuProfileFile.Close()
			return fmt.Errorf("could not start CPU profile: %w", err)
		}
	}

	// Setup Memory profiling (will be written at exit)
	if config.MemProfile != "" {
		var err error
		memProfileFile, err = os.Create(config.MemProfile)
		if err != nil {
			return fmt.Errorf("could not create memory profile: %w", err)
		}
		fmt.Printf("Memory profiling enabled, will write to %s on exit\n", config.MemProfile)
	}

	// Start HTTP server for runtime profiling if requested
	if config.EnableProfiling {
		go func() {
			addr := fmt.Sprintf("localhost:%d", config.ProfilePort)
			fmt.Printf("Starting pprof HTTP server on http://%s/debug/pprof/\n", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Printf("Failed to start pprof server: %v\n", err)
			}
		}()
	}

	return nil
}
