package utils

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	// Just verify it's not empty
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestGetVersionInfo(t *testing.T) {
	info := GetVersionInfo()

	// Check that it contains the version
	if !strings.Contains(info, Version) {
		t.Errorf("GetVersionInfo() = %q, should contain version %q", info, Version)
	}

	// Check that it contains the git commit
	if !strings.Contains(info, GitCommit) {
		t.Errorf("GetVersionInfo() = %q, should contain git commit %q", info, GitCommit)
	}

	// Check that it contains the build date
	if !strings.Contains(info, BuildDate) {
		t.Errorf("GetVersionInfo() = %q, should contain build date %q", info, BuildDate)
	}
}
