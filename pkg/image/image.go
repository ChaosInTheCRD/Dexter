package image

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sigstore/cosign/pkg/oci/remote"
)


func AddDigest(reference string) (string, error) {
   // Check if the digest is already present
   if strings.Contains(reference, "@sha256:") {
      return reference, nil
   }

   ref, err := name.ParseReference(reference)
   if err != nil {
      return "", err
   }

   dig, err := remote.ResolveDigest(ref)
   if err != nil {
      return "", err
   }

   return fmt.Sprintf("%s", dig), nil
}
