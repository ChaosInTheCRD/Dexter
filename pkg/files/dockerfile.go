package files

import (
	"bytes"
	"os"
	"strings"
	apex "github.com/apex/log"
	"github.com/chaosinthecrd/dexter/pkg/image"
	"github.com/google/go-containerregistry/pkg/name"
	docker "github.com/moby/buildkit/frontend/dockerfile/parser"
	"golang.org/x/net/context"
)

// Dockerfile instruction.
const (
	FROM = "from"
)

var dockerfileEmoji string = "🐳"
type dockerfile struct{}

// Find finds valid Kubernetes yaml files. It does this by trying to parse
// the provided file in the expected format and tries to find an `image` field.
func (_ dockerfile) Find(ctx context.Context, path string) (Found, error) {

   logs := apex.FromContext(ctx) 
   logs.Debugf("Running Dockerfile parser against file %s", path)
   references := []string{}

   logs.Debugf("Dockerfile Parser: Reading file %s", path)
   file, err := os.Open(path)
   if err != nil {
      return Found{}, err
   }
   

   r, err := docker.Parse(file)
   if err != nil {
      logs.Debugf("Dockerfile Parser: Failed to decode file %s. Continuing: %s", path, err.Error())
      return Found{}, nil
   }

   for _, child := range r.AST.Children {
      var reference string
      if strings.ToUpper(child.Value) == "FROM" {
         reference = child.Next.Value
         logs.Debugf("Dockerfile Parser: Found potential image reference %s in 'FROM' argument of Dockerfile at path %s", reference, path)
      } else if strings.Contains("COPY", child.Value) {
         reference = getCopyRef(child.Flags)
         logs.Debugf("Dockerfile Parser: Found potential image reference %s in 'FROM' argument of Dockerfile at path %s", reference, path)
      }

      if reference == "" {
         continue
      }

      if _, err := name.ParseReference(reference); err == nil {
         references = addRef(references, reference)
      }
   }

   return Found{Location: path, Parser: Parser{ Lead: dockerfileEmoji, Name: "dockerfile", Parser: dockerfile{}}, References: references}, nil
}


func getCopyRef(attrs []string) string {
   for _, n := range attrs {
      if s := strings.Split(n, "="); len(s) == 2 && s[0] == "--from" && s[1] != " " {
         return s[1]
      }
   }

   return ""
}

func (_ dockerfile) Modify(ctx context.Context, found Found) ([]string, error) {
   newReferences := []string{}

   logs := apex.FromContext(ctx)
   file, err := os.ReadFile(found.Location)
   if err != nil {
      return []string{}, err 
   }

   for _, reference := range found.References {
      logs.Debugf("Adding digest to reference %s", reference)
      newRef, err := image.AddDigest(reference)
      if err != nil {
         logs.Warnf("WARNING: Failed to add digest to reference %s. Failed with error: %s", reference, err.Error())
         newRef = reference
      }

      logs.Debugf("Replacing reference %s with reference %s", reference, newRef)
      file = bytes.Replace(file, []byte(reference), []byte(newRef), -1)
      newReferences = append(newReferences, newRef)
   }


   logs.Debugf("Writing reference changes to file %s", found.Location)
   if err = os.WriteFile(found.Location, file, 0666); err != nil {
      return nil, err
   }

   return newReferences, nil
}

