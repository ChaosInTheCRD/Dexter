package files

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
        "strings"

	"github.com/apex/log"
	apex "github.com/apex/log"
	"github.com/chaosinthecrd/dexter/pkg/config"
	"github.com/chaosinthecrd/dexter/pkg/util"
)

// walker is the struct that is used to hold information gathered during the walk
type Walker struct {
   Finds []Found
   Ignores config.Ignores
   Parsers []string
   Context context.Context
}

// parser is the interface implemented by image file parsers.
type parser interface {
	Find(context.Context, string) (Found, error)
        Modify(context.Context, Found) ([]string, error)
}

// NewParser takes a string and initialises the respective parser for which it represents
func NewParser(parser string) (parser, error) {
   switch parser {
    case "kubernetes":
       k := kubernetes{}
       return k, nil
    case "dockerfile":
       d := dockerfile{}
       return d, nil
    default:
       return nil, fmt.Errorf("Parser %s declared in config file not recognised.", parser)
    }
}

func FindParsers(parsers []string) ([]parser, error) {
   newParsers := []parser{}
   for _, p := range(parsers) {
      newParser, err := NewParser(p)
      if err != nil {
         return []parser{}, err
      }

      newParsers = append(newParsers, newParser)
   }


   return newParsers, nil
}

func (w *Walker) FindImageReferences(path string, info os.FileInfo, err error) error {

   wd, err := os.Getwd()
   if err != nil {
      return fmt.Errorf("Failed to get working directory: %s", err.Error())
   }

   wd = fmt.Sprintf("%s/", wd)

   if util.Contains(w.Ignores.Files, strings.ReplaceAll(path, wd, "")) {
      log.Infof("file %s is in ignore list, skipping.", strings.ReplaceAll(path, wd, ""))
      return nil 
   }

   logs := apex.FromContext(w.Context)

   if err != nil {
      return err
   }

   if info.IsDir() {

      if info.Name() == ".git" {
         logs.Debugf("Skipping directory %s as it is a git directory", path)
         return filepath.SkipDir
      }

      logs.Debugf("Skipping %s as it is a directory", path)
      return nil
   }

   parsers, err := FindParsers(w.Parsers)
   if err != nil {
      return fmt.Errorf("Failed to find parsers: %s", err.Error())
   }

   for {
    var (
            wg         sync.WaitGroup
            lock       sync.Mutex
    )

    wg.Add(len(parsers))

    // Run all parsers.
    for _, p := range parsers {
      go func(p parser) {
      defer wg.Done()
      res, err := p.Find(w.Context, path)
      if err != nil {
         logs.Errorf("Failed to find references at path %s: %s", err.Error())
      }
      lock.Lock()
      defer lock.Unlock()

      for _, n := range w.Ignores.References {
         var clean bool
         res.References, clean = util.Clean(res.References, n)
         if clean == true {
            log.Infof("reference %s is in ignore list, skipping.", n)
         }
      }
      if res.References != nil {
      w.Finds = append(w.Finds, res)
      }
      }(p)
    }

    wg.Wait()

    select {
    case <-w.Context.Done():
            return w.Context.Err()
    default:
    }

    return nil

   }

}
