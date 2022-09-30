package files

import (
	"context"
	"fmt"
        "strings"
	"os"
	"sync"

	"github.com/chaosinthecrd/dexter/pkg/config"
)

// walker is the struct that is used to hold information gathered during the walk
type Walker struct {
   Finds []Found
   Ignores []config.Ignore
   Parsers []string
}

// Parser is the interface implemented by image file parsers.
type Parser interface {
	Find(context.Context, string) (Found, error)
        Modify(Found) (int, []string, error)
}

// NewParser takes a string and initialises the respective parser for which it represents
func NewParser(parser string) (Parser, error) {
   switch parser {
    case "kubernetes":
       fmt.Println("Found kubernetes parser!")
       k := kubernetes{}
       return k, nil
    default:
       return nil, fmt.Errorf("Parser %s declared in config file not recognised.", parser)
    }
}

func FindParsers(parsers []string) ([]Parser, error) {
   fmt.Printf("Parsers: %s \n", parsers)
   newParsers := []Parser{}
   for _, p := range(parsers) {
      fmt.Printf("P IS: %s \n", p)
      parser, err := NewParser(p)
      if err != nil {
         return []Parser{}, err
      }

      fmt.Printf("Just set up parser %s \n", parser)

      newParsers = append(newParsers, parser)
   }

   fmt.Printf("Initialised Parsers: %s \n", newParsers)

   return newParsers, nil
}

func (w *Walker) FindImageReferences(path string, info os.FileInfo, err error) error {
   if err != nil {
      return err
   }

   if info.IsDir() {
      return nil
   }
   ctx := context.TODO()

   fmt.Printf("Parsers: %s\n", w.Parsers)

   parsers, err := FindParsers(w.Parsers)
   if err != nil {
      return err
   }

   fmt.Printf("Parsers Ready: %s\n", parsers)

   for {
    var (
            wg         sync.WaitGroup
            lock       sync.Mutex
            errs       []string
    )

    wg.Add(len(parsers))

    // Run all parsers.
    for _, p := range parsers {
      go func(p Parser) {
      defer wg.Done()
      fmt.Printf("Running the %s parser...\n", p)
      fmt.Printf("Inspecting File: %s\n", path)
      res, err := p.Find(ctx, path)
      if err != nil {
         fmt.Println(err)
      }
      lock.Lock()
      defer lock.Unlock()
      // if err != nil {
      //    errs = append(errs, err.Error())
      // }
      if res.References != nil {
      w.Finds = append(w.Finds, res)
      }
      }(p)
    }

    wg.Wait()

    select {
    case <-ctx.Done():
            return ctx.Err()
    default:
    }

    if len(errs) > 0 {
            return fmt.Errorf("parser error finding images: %s", strings.Join(errs, "; "))
    }

	return nil

   }

}
