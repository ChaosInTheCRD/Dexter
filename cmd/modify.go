// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
        "context"
	"path/filepath"

	apex "github.com/apex/log"

	"github.com/chaosinthecrd/dexter/pkg/config"
	"github.com/chaosinthecrd/dexter/pkg/files"
	"github.com/chaosinthecrd/dexter/pkg/output"
	"github.com/chaosinthecrd/dexter/internal/log"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var manipulateCmd = &cobra.Command{
	Use:   "manipulate",
	Short: "Inspect working directory for files containing image references and manipulate them to be pinned to the digest",
	RunE: func(cmd *cobra.Command, args []string) error {
           ctx := log.InitLogContext(DebugMode)
           if err := manipulate(ctx); err != nil {
              logs := apex.FromContext(ctx)
              logs.Error("command 'manipulate' failed. Closing.")
           }

           return nil
     },
}

func init() {
	rootCmd.AddCommand(manipulateCmd)
}

func manipulate(ctx context.Context) error {

          logs :=  apex.FromContext(ctx)

          err := output.ValidateOutputMode(OutputMode)
          if err != nil {
             logs.Errorf("Failed to validate output mode")
             return err
          }

          workingDirectory, err := os.Getwd()
          if err != nil {
             logs.Infof("Cannot get working directory: %s", err.Error())
             return err
          }

          logs = log.AddFields(logs, "manipulate", "bar", "foo")

          logs.Debugf("Pinning image references found in %s using config file at", workingDirectory, ConfigFile)

          walker := &files.Walker{}

          conf, err := config.InitialiseConfig(ConfigFile)
          if err != nil {
             logs.Errorf("Failed to initialise dexter config: %s", err.Error())
             return err
          }

          walker.Ignores = conf.Ignores
          walker.Parsers = conf.Parsers
          walker.Context = ctx

          if err := filepath.Walk(workingDirectory, walker.FindImageReferences); err != nil {
             logs.Errorf("Failed to find image references: %s", err.Error()) 
             return err
          }

          numChanges := 0
          for _, file := range walker.Finds {
                  newRefs, err := file.Parser.Parser.Modify(walker.Context, file)
                  if err != nil {
                     logs.Errorf("Failed to modify file with parser: %s", err.Error())
                     return err
                  }
                  file.NewReferences = newRefs
                  if len(newRefs) > 0 {
                          numChanges = numChanges + len(newRefs)
                          for i, n := range file.References {
                                  c := color.New(color.FgGreen)
                                  logs.Infof(file.Parser.Lead + " " + c.Sprintf("%s => %s\n", n, file.NewReferences[i]))
                          }
                  }
          }

          logs.Infof("Found %d files, of which %d had image references", len(walker.Finds), numChanges)

          return nil
}
