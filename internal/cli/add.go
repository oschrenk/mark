package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/oschrenk/mark/internal/config"
	"github.com/oschrenk/mark/internal/remote"
	"github.com/spf13/cobra"
)

var (
	flagTags   []string
	flagAt     string
	flagTarget string
	flagDryRun bool
)

var addCmd = &cobra.Command{
	Use:   "add <description>",
	Short: "Record one event",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringSliceVar(&flagTags, "tag", nil, "tag to attach, repeatable")
	addCmd.Flags().StringVar(&flagAt, "at", "", "event time as RFC3339, defaults to now")
	addCmd.Flags().StringVar(&flagTarget, "target", "", "target from the config, defaults to the one marked default")
	addCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "print the event and send nothing")
	rootCmd.AddCommand(addCmd)
}

func runAdd(_ *cobra.Command, args []string) error {
	at := time.Now()
	if flagAt != "" {
		parsed, err := time.Parse(time.RFC3339, flagAt)
		if err != nil {
			return fmt.Errorf("parsing --at %q as RFC3339: %w", flagAt, err)
		}
		at = parsed
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	target, err := cfg.Target(flagTarget)
	if err != nil {
		return err
	}

	event := remote.Event{
		Metric:      target.Metric,
		Description: args[0],
		Tags:        flagTags,
		At:          at,
	}

	if flagDryRun {
		printEvent(event, target.URL)
		return nil
	}
	return event.Send(target.URL)
}

func printEvent(e remote.Event, url string) {
	out := os.Stdout
	fmt.Fprintf(out, "would POST to %s\n\n", url)

	width := 0
	for _, l := range e.Labels() {
		if n := len(l.Name); n > width {
			width = n
		}
	}
	for _, l := range e.Labels() {
		fmt.Fprintf(out, "  %-*s  %s\n", width, l.Name, l.Value)
	}

	fmt.Fprintf(out, "\n  %-*s  %g\n", width, "value", remote.Value)
	fmt.Fprintf(out, "  %-*s  %d  (%s)\n", width, "timestamp", e.TimestampMS(), e.At.Format(time.RFC3339))

	encoded := e.Encode()
	fmt.Fprintf(out, "\n  %-*s  %d bytes protobuf, before snappy\n", width, "payload", len(encoded))

	fmt.Fprintf(out, "\n  as promql: %s\n", promQL(e))
}

// promQL shows the selector a Perses annotation would use to find this event.
func promQL(e remote.Event) string {
	var parts []string
	for _, l := range e.Labels() {
		if l.Name == "__name__" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%q", l.Name, l.Value))
	}
	if len(parts) == 0 {
		return e.Metric
	}
	return fmt.Sprintf("%s{%s}", e.Metric, strings.Join(parts, ", "))
}
