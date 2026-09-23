/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package verifiercommon provides helpers shared by the verifier factories that
// wrap the plugin verifiers from ratify-verifier-go.
package verifiercommon

import (
	"encoding/json"
	"fmt"
)

// ArtifactTypes is a helper type that unmarshals a JSON value that may be either
// a single string or an array of strings into a slice of strings. The v1 Ratify
// verifier configs express `artifactTypes` as a single string, while the
// library options accept a slice.
type ArtifactTypes []string

// UnmarshalJSON implements json.Unmarshaler to accept both a string and a string
// array.
func (a *ArtifactTypes) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		if single == "" {
			*a = nil
			return nil
		}
		*a = ArtifactTypes{single}
		return nil
	}

	var multi []string
	if err := json.Unmarshal(data, &multi); err != nil {
		return fmt.Errorf("artifactTypes must be a string or an array of strings: %w", err)
	}
	*a = multi
	return nil
}
