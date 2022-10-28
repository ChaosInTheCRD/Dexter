package files

import (
	"bytes"
	"io"
	"os"
	"strings"

	apex "github.com/apex/log"
	"github.com/chaosinthecrd/dexter/pkg/image"
	"github.com/google/go-containerregistry/pkg/name"
	"golang.org/x/net/context"

	goyaml "github.com/go-yaml/yaml"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var LeadEmoji string = "☸"

type kubernetes struct{}

// Find finds valid Kubernetes yaml files. It does this by trying to parse
// the provided file in the expected format and tries to find an `image` field.
func (_ kubernetes) Find(ctx context.Context, path string) (Found, error) {

   logs := apex.FromContext(ctx)
   logs.Debugf("Running Kubernetes parser against file %s", path)
   references := []string{}
   decode := scheme.Codecs.UniversalDeserializer().Decode

   logs.Debugf("Kubernetes Parser: Reading file %s", path)
   file, err := os.ReadFile(path)
   if err != nil {
      return Found{}, err
   }


   logs.Debugf("Kubernetes Parser: Splitting YAML if multiple documents in one file")
   files, err := SplitYAML(file)
   if err != nil {
      logs.Debugf("Kubernetes Parser: Failed to decode file into YAML. Continuing: %s", err.Error())
   }

   // If context has been cancelled, exit scanning.
   select {
   case <-ctx.Done():
      return Found{}, ctx.Err()
   default:
   }

   for _, file := range(files) {
      // (and again) If context has been cancelled, exit scanning.
      select {
      case <-ctx.Done():
         return Found{}, ctx.Err()
      default:
      }

      obj, gKV, err  := decode(file, nil, nil)
      if err != nil {
         logs.Debugf("Kubernetes Parser: Failed to decode file %s. Continuing: %s", path, err.Error())
         continue
      }

      switch gKV.Kind {
         case "Pod":
            for _, c := range obj.(*corev1.Pod).Spec.Containers {
               references = addRef(references, c.Image)
            }
         case "Deployment":
            for _, c := range obj.(*appsv1.Deployment).Spec.Template.Spec.Containers {
               references = addRef(references, c.Image)
            }
         case "ReplicaSet":
            for _, c := range obj.(*appsv1.ReplicaSet).Spec.Template.Spec.Containers {
               references = addRef(references, c.Image)

            }
         case "StatefulSet":
            for _, c := range obj.(*appsv1.StatefulSet).Spec.Template.Spec.Containers {
               references = addRef(references, c.Image)
            }
         case "DaemonSet":
            for _, c := range obj.(*appsv1.DaemonSet).Spec.Template.Spec.Containers {
               references = addRef(references, c.Image)
            }
         case "Job":
            for _, c := range obj.(*batchv1.Job).Spec.Template.Spec.Containers {
               references = addRef(references, c.Image)
            }
         case "CronJob":
            for _, c := range obj.(*batchv1.CronJob).Spec.JobTemplate.Spec.Template.Spec.Containers {
               references = addRef(references, c.Image)
            }
         default:
            logs.Debugf("Manifest of kind %s does not contain images", gKV.Kind)
            continue
      }
   }

   return Found{Location: path, Parser: Parser{ Lead: LeadEmoji, Name: "kubernetes", Parser: &kubernetes{}}, References: references}, nil
}

func addRef(references []string, image string) []string {
   if strings.Contains(image, "sha256") == true {
      return references
   }
   if _, err := name.ParseReference(image); err == nil {
      references = append(references, image)
   }

   return references
}

func (_ kubernetes) Modify(ctx context.Context, found Found) ([]string, error) {
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
         logs.Warnf("Failed to add digest to reference %s. Failed with error: %s", reference, err.Error())
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

func SplitYAML(resources []byte) ([][]byte, error) {

	dec := goyaml.NewDecoder(bytes.NewReader(resources))

	var res [][]byte
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		valueBytes, err := goyaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, valueBytes)
	}
	return res, nil
}
