package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	var short bool

	command := cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Long:  "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion(short)
		},
	}
	command.PersistentFlags().BoolVarP(&short, "short", "s", false, "Print version info in short format")

	return &command
}

func printVersion(short bool) {
	const fmat = "  %-10s: %s\n"

	fmt.Printf("%s:\n", appName)
	fmt.Printf(fmat, "Version", version)
	if !short {
		fmt.Printf(fmat, "Commit", commit)
	}
}
