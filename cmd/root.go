// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"strings"
	"github.com/chaosinthecrd/dexter/pkg/output"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dexter",
	Short: "Fix container image references so they're pinned to a specific digest",
	Long:  `Dexter is a CLI tool to inspect files to find container image references and manipulate them to be pinned to a specific digest.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var OutputMode string
var ConfigFile string
var WorkingDirectory string

func init() {
	rootCmd.PersistentFlags().StringVarP(&OutputMode, "output", "o", "pretty", "Output mode. Supported modes: "+strings.Join(output.Modes, ", "))
        rootCmd.PersistentFlags().StringVarP(&WorkingDirectory, "directory", "d", "", "Directory to scan.")
        rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config-file", "c", "dexter.yaml", "Config file path.")
}
