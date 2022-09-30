package image

import (
   "fmt"
   "github.com/sigstore/cosign/pkg/oci/remote"
   "github.com/google/go-containerregistry/pkg/name"
)


func AddDigest(reference string) (string, error) {
   ref, err := name.ParseReference(reference)
   if err != nil {
      return "", err
   }

   dig, err := remote.ResolveDigest(ref)
   if err != nil {
      return "", err
   }

   return fmt.Sprintf("%s@sha256:%s", ref, dig), nil
}
