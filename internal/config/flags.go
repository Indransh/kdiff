package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultLogLevel    = "info"
	DefaultRefreshRate = 2 // secs
)

var (
	DefaultKubeconfig = filepath.Join(KdiffUserHomeDir(), ".kube", "config")
	DefaultConfigFile = filepath.Join(KdiffUserHomeDir(), ".config", "kdiff", "config.yaml")
	DefaultLogFile    = filepath.Join(os.TempDir(), fmt.Sprintf("kdiff-%s.log", KdiffUser()))
)

// kdiff configuration flags.
type ConfigFlags struct {
	LogFile     string
	LogLevel    string
	RefreshRate int
	KubeConfig  string
}
