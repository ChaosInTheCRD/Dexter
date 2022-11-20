// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
        "context"
	"path/filepath"
        "fmt"

	apex "github.com/apex/log"

	"github.com/chaosinthecrd/dexter/pkg/config"
	"github.com/chaosinthecrd/dexter/pkg/files"
	// "github.com/chaosinthecrd/dexter/pkg/output"
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

          // err := output.ValidateOutputMode(OutputMode)
          // if err != nil {
          //    logs.Errorf("Failed to validate output mode")
          //    return err
          // }

          var err error

          wd := WorkingDirectory
          if wd == "" {
             wd, err = os.Getwd()
             if err != nil {
                return err
             }
          }

          logs = log.AddFields(logs, "manipulate", "bar", "foo")

          logs.Debugf("Pinning image references found in %s using config file at", wd, ConfigFile)

          walker := &files.Walker{}

          conf, err := config.InitialiseConfig(ConfigFile)
          if err != nil {
             logs.Errorf("Failed to initialise dexter config: %s", err.Error())
             return err
          }

          walker.Ignores = conf.Ignores
          walker.Parsers = conf.Parsers
          walker.Context = ctx

          if err := filepath.Walk(wd, walker.FindImageReferences); err != nil {
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
                       for i, n := range file.References {
                         if n == file.NewReferences[i]{
                           continue
                         }

                         numChanges = numChanges + 1

                         c := color.New(color.FgGreen)
                         defer fmt.Println("\n" + file.Parser.Lead + " " + c.Sprintf("%s => %s", n, file.NewReferences[i]))
                       }
                  }
          }

          defer fmt.Printf("Found %d files, of which %d had image references that were manipulated:\n", len(walker.Finds), numChanges)
          defer fmt.Printf("\n")

          return nil
}
