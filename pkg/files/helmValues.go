package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	apex "github.com/apex/log"
	"github.com/chaosinthecrd/dexter/pkg/image"
	"github.com/google/go-containerregistry/pkg/name"
	"golang.org/x/net/context"

	"github.com/sigstore/cosign/pkg/oci/remote"
	yaml "gopkg.in/yaml.v3"
)

var helmEmoji string = "⚙️"

type helm struct{}

// Find finds container image references in helm values files. Given the unstructured nature
// of helm values files, this function can only look for references in standard 'known' locations
// such as inside an `image` object which contains `registry`,  `repository` and `tag` fields.
func (_ helm) Find(ctx context.Context, path string) (Found, error) {
	logs := apex.FromContext(ctx)
	logs.Debugf("Running Helm values parser against file %s", path)
	references := []string{}

	name := strings.ToLower(filepath.Base(path))
	if !strings.Contains(name, "value") && strings.ToLower(filepath.Ext(path)) != ".yaml" {
		logs.Debugf("Helm Values Parser: filename %s at path %s doesn't look like a values file, skipping", name, filepath.Dir(path))
		return Found{}, nil
	}

	logs.Debugf("Helm Values Parser: Reading file %s", path)
	file, err := os.ReadFile(path)
	if err != nil {
		return Found{}, err
	}

	body := make(map[string]interface{})
	err = yaml.Unmarshal(file, body)
	if err != nil {
		return Found{}, err
	}

	processData(body, &references)
	fmt.Println(references)
	return Found{Location: path, Parser: Parser{Lead: dockerfileEmoji, Name: "dockerfile", Parser: dockerfile{}}, References: references}, nil
}

func (_ helm) Manipulate(ctx context.Context, refCache map[string]string, found *Found) (map[string]string, error) {
	logs := apex.FromContext(ctx)
	file, err := os.ReadFile(found.Location)
	if err != nil {
		return map[string]string{}, err
	}

	for _, reference := range found.References {
		logs.Debugf("Adding digest to reference %s", reference)
		newRef, err := image.AddDigest(reference)
		if err != nil {
			logs.Warnf("WARNING: Failed to add digest to reference %s. Failed with error: %s", reference, err.Error())
			newRef = reference
		}

		// We need to handle things slightly differently here.
		// Given that the reference is split up in the helm values between `registry`, `repository` and `tag`,
		// we are going to pin the digest to the `tag` field.

		// Note: given that references are usually split up in helm values files, we really have no choice but to walk through the file again... this isn't great.
		// For now, we are just going to do this, but this probably needs to be addressed as it isn't efficient. Maybe we need to consider manipulating at the point of expection
		// on all parsers.
		logs.Debugf("Replacing reference %s with reference %s", reference, newRef)
		body := make(map[string]interface{})
		err = yaml.Unmarshal(file, body)
		if err != nil {
			return map[string]string{}, err
		}

		body, err = manipulateData(body, reference)
		if err != nil {
			return map[string]string{}, err
		}

		file = 

		found.NewReferences = append(found.NewReferences, newRef)

		logs.Debugf("Adding reference %s to cache against original reference %s", newRef, reference)
		refCache[reference] = newRef
		found.NewReferences = append(found.NewReferences, newRef)
	}

	logs.Debugf("Writing reference changes to file %s", found.Location)
	if err = os.WriteFile(found.Location, file, 0666); err != nil {
		return nil, err
	}

	return refCache, nil
}

func processData(data map[string]interface{}, references *[]string) {
	for key, value := range data {
		dataValue, ok := data[key]
		if !ok {
			continue
		}
		if key == "image" {
			if _, ok := data[key].(map[string]interface{}); ok {
				nestedData, ok := dataValue.(map[string]interface{})
				if !ok {
					continue
				}
				reg, ok := nestedData["registry"]
				if !ok {
					continue
				}
				repo, ok := nestedData["repository"]
				if !ok {
					continue
				}
				tag, ok := nestedData["tag"]
				if !ok {
					continue
				}

				ref := fmt.Sprintf("%s/%s:%s", reg, repo, tag)
				*references = append(*references, ref)

				return
			}
		} else {
			// Determine the type of the value in the schema
			switch value.(type) {
			case map[string]interface{}:
				// Nested object, recursively process
				if nestedData, ok := dataValue.(map[string]interface{}); ok {
					processData(nestedData, references)
				} else {
				}
			case []interface{}:
				// Array of values, process each element
				if arrayData, ok := dataValue.([]interface{}); ok {
					for _, element := range arrayData {
						if nestedData, ok := element.(map[string]interface{}); ok {
							processData(nestedData, references)
						} else {
						}
					}
				}
			default:
				// Leaf node, do nothing
			}
		}
	}
}

func manipulateData(data map[string]interface{}, reference string) (map[string]interface{}, error) {
	ref, err := name.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	for key, value := range data {
		dataValue, ok := data[key]
		if !ok {
			continue
		}
		if key == "image" {
			if _, ok := data[key].(map[string]interface{}); ok {
				nestedData, ok := dataValue.(map[string]interface{})
				if !ok {
					continue
				}
				reg, ok := nestedData["registry"]
				if !ok && ref.Context().RegistryStr() != reg {
					continue
				}
				repo, ok := nestedData["repository"]
				if !ok && ref.Context().RepositoryStr() != repo {
					continue
				}
				_, ok = nestedData["tag"]
				if !ok {
					continue
				}

				digRef, err := remote.ResolveDigest(ref)
				if err != nil {
					return nil, err
				}

				dig := strings.SplitAfter(fmt.Sprintf("%s", digRef), "@sha256:")
				nestedData["tag"] = fmt.Sprintf("%s@sha256:", dig)

				return data, nil
			}
		} else {
			// Determine the type of the value in the schema
			switch value.(type) {
			case map[string]interface{}:
				// Nested object, recursively process
				if nestedData, ok := dataValue.(map[string]interface{}); ok {
					data, err = manipulateData(nestedData, reference)
					if err != nil {
						return nil, err
					}
				}
			case []interface{}:
				// Array of values, process each element
				if arrayData, ok := dataValue.([]interface{}); ok {
					for _, element := range arrayData {
						if nestedData, ok := element.(map[string]interface{}); ok {
							data, err = manipulateData(nestedData, reference)
							if err != nil {
								return nil, err
							}
						}
					}
				}
			default:
				// Leaf node, do nothing
			}
		}
	}

	return nil, nil
}
