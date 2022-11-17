package config

import (
	"os"
	"os/user"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	DefaultDirMod  os.FileMode = 0755
	DefaultFileMod os.FileMode = 0600
)

// EnsureDir ensures the given path is a directory
func EnsureDir(path string, mode os.FileMode) {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, mode); err != nil {
			log.Fatalf("Unable to create dir %q %v", path, err)
		}
	}
}

// KdiffUser returns current user or fails.
func KdiffUser() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Could not retrive current user info!")
	}
	return usr.Username
}

// KdiffUserHomeDir returns current user or fails.
func KdiffUserHomeDir() string {
	usrDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Could not retrive user home directory info!")
	}
	return usrDir
}
