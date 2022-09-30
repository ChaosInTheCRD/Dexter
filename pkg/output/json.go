// SPDX-License-Identifier: Apache-2.0

package output

type JSONOutput struct {
	Modifications []Modification `json:"certificates"`
}

type Modification struct {
	FileLocation      string `json:"fileLocation"`
	OldImageReference string `json:"oldImageReference"`
	NewImageReference string `json:"newImageReference"`
	Digest            string `json:"digest"`
}

