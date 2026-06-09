package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var (
	rootCmd = &cobra.Command{
		Use:   "unibrow",
		Short: "Unibrow - Unified AI Agent CLI Hub",
		Long: `Unibrow orchestrates multiple AI coding agents (Claude Code, Codex, Gemini CLI)
from a single terminal with shared context and intelligent routing.`,
		Version: version,
	}
)

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(memoryCmd)
	rootCmd.AddCommand(sessionCmd)
	rootCmd.AddCommand(contextCmd)
}

// GetVersion returns the current version.
func GetVersion() string {
	return version
}