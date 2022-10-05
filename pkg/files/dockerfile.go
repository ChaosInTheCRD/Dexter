package files

import (
	"bytes"
	"os"

	apex "github.com/apex/log"
	"github.com/chaosinthecrd/dexter/pkg/image"
	"golang.org/x/net/context"
        docker "github.com/moby/buildkit/frontend/dockerfile/parser"

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
      if child.Value == "FROM" {
         references = append(references, child.Next.Value)
         logs.Debugf("Dockerfile Parser: Found image %s in Dockerfile at path", )
         return Found{Location: path, Parser: Parser{ Lead: dockerfileEmoji, Name: "dockerfile", Parser: dockerfile{}}, References: references}, nil
      } 
   }

   return Found{}, nil
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
         return nil, err
      }

      if newRef == reference {
      logs.Debugf("reference %s already has digest specified. Continuing.", reference)
      continue
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

