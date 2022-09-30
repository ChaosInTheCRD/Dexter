// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/chaosinthecrd/dexter/pkg/config"
	"github.com/chaosinthecrd/dexter/pkg/files"
	"github.com/chaosinthecrd/dexter/pkg/output"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var manipulateCmd = &cobra.Command{
	Use:   "manipulate",
	Short: "Inspect working directory for files containing image references and manipulate them to be pinned to the digest",
	RunE: func(cmd *cobra.Command, args []string) error {
          err := output.ValidateOutputMode(OutputMode)
          if err != nil {
                  return err
          }

          workingDirectory, err := os.Getwd()
          if err != nil {
             log.Println(err.Error())
          }

          fmt.Printf("Working Dir: %s \n", workingDirectory)
          walker := &files.Walker{}

          conf, err := config.InitialiseConfig(ConfigFile)
          if err != nil {
             log.Println(err.Error())
          }
          walker.Ignores = conf.Ignores
          walker.Parsers = conf.Parsers

          if err := filepath.Walk(workingDirectory, walker.FindImageReferences); err != nil {
            log.Println(err.Error()) 
          }

          numChanges := 0

          if walker.Finds == nil {
             fmt.Println("Nothing")
             return nil
          }

          for _, file := range walker.Finds {
                  changes, newRefs, err := file.Parser.Modify(file)
                  if err != nil {
                     log.Println(err.Error())
                  }
                  file.NewReferences = newRefs
                  if changes > 0 {
                          numChanges = numChanges + changes
                          log.Printf("File %s\n", file.Location)
                          for i, n := range file.References {
                                  var lead string
                                  if i == changes-1 {
                                          lead = "┗"
                                  } else {
                                          lead = "┣"
                                  }
                                  c := color.New(color.FgGreen)
                                  c.Println(lead + " " + c.Sprintf("%s => %s\n", n, file.NewReferences[i]))
                          }
                  }
          }
          fmt.Printf("Found %d files, of which %d had image references\n", len(walker.Finds), numChanges)

          return nil
	},
}

func init() {
	rootCmd.AddCommand(manipulateCmd)
}
