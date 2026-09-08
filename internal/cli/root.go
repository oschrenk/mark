package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:          "mark",
	Short:        "Record an event in Prometheus",
	Long:         "Write a timestamped event, with a description and tags, into Prometheus by remote write.",
	SilenceUsage: true,
	Version:      "dev",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func Execute() {
	rootCmd.Version = fmt.Sprintf("%s (%s, %s) %s", Version, Commit, BuildDate, runtime.Version())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
