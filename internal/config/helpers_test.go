package config_test

import (
	"kdiff/internal/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureDir(t *testing.T) {
	testFile1 := filepath.Join(os.TempDir(), "kdiff-testEnsureDir-dir1", "emptyFile.txt")
	defer os.RemoveAll(filepath.Dir(testFile1))

	// Test case: Missing directory
	config.EnsureDir(testFile1, 0755)
	require.DirExists(t, filepath.Dir(testFile1))

	// Test case: Existing directory
	config.EnsureDir(testFile1, 0755)
	require.DirExists(t, filepath.Dir(testFile1))
}
