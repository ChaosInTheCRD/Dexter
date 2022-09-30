package files

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/chaosinthecrd/dexter/pkg/image"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
        goyaml "github.com/go-yaml/yaml"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

type kubernetes struct{}

// Find finds valid Kubernetes yaml files. It does this by trying to parse
// the provided file in the expected format and tries to find an `image` field.
func (_ kubernetes) Find(ctx context.Context, location string) (Found, error) {

   references := []string{}
   decode := scheme.Codecs.UniversalDeserializer().Decode
   file, err := os.ReadFile(location)
   if err != nil {
      return Found{}, err
   }

   files, err := SplitYAML(file)
   if err != nil {
      return Found{}, err
   }

   fmt.Println(location)

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
         return Found{}, err
      }

      fmt.Println(gKV.Kind)
      switch gKV.Kind {
         case "Pod":
            for _, c := range obj.(*corev1.Pod).Spec.Containers {
               references = append(references, c.Image) 
            }
         case "Deployment":
            for _, c := range obj.(*appsv1.Deployment).Spec.Template.Spec.Containers {
               references = append(references, c.Image)

            }
         case "ReplicaSet":
            for _, c := range obj.(*appsv1.ReplicaSet).Spec.Template.Spec.Containers {
               references = append(references, c.Image)

            }
         case "StatefulSet":
            for _, c := range obj.(*appsv1.StatefulSet).Spec.Template.Spec.Containers {
               references = append(references, c.Image)
            }
         case "DaemonSet":
            for _, c := range obj.(*appsv1.DaemonSet).Spec.Template.Spec.Containers {
               references = append(references, c.Image)
            }
         case "Job":
            for _, c := range obj.(*batchv1.Job).Spec.Template.Spec.Containers {
               references = append(references, c.Image)
            }
         case "CronJob":
            for _, c := range obj.(*batchv1.CronJob).Spec.JobTemplate.Spec.Template.Spec.Containers {
               references = append(references, c.Image)
            }
         default:
            return Found{}, fmt.Errorf("Kubernetes Parser: Failed to parse file as with kubernetes parser")
      }
   }

   return Found{Location: location, Parser: &kubernetes{}, References: references}, nil
}

func (_ kubernetes) Modify(found Found) (int, []string, error) {
   fmt.Println("What is going")
   newReferences := []string{}

   file, err := os.ReadFile(found.Location)
   if err != nil {
           fmt.Println(err)
           os.Exit(1)
   }

   for _, reference := range found.References {
      newRef, err := image.AddDigest(reference)
      if err != nil {
         return 0, nil, err
      }

      file = bytes.Replace(file, []byte(reference), []byte(newRef), -1)
      newReferences = append(newReferences, newRef)
   }

   if err = os.WriteFile(found.Location, file, 0666); err != nil {
      return 0, nil, err
   }

   return len(found.References), newReferences, nil
}

func splitYaml(file []byte) ([][]byte, error) {

   yamls := [][]byte{}
   dec := yaml.NewDecoder(bytes.NewReader(file))
   for {
        var node yaml.Node
        err := dec.Decode(&node)
        if errors.Is(err, io.EOF) {
            break
        }
        if err != nil {
           return [][]byte{}, fmt.Errorf("Kubernetes Parser: Failed to decode yaml file %s", err)
        }

        fmt.Println(node)
        nodeByte, err := yaml.Marshal(node)
        if err != nil {
           return [][]byte{}, fmt.Errorf("Kubernetes Parser: Failed to marshal yaml file node: %s", err)
        }

        yamls = append(yamls, nodeByte)
   }

   return yamls, nil
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
