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

package licensechecker

import (
	"encoding/json"
	"fmt"

	"github.com/notaryproject/ratify-go"
	"github.com/notaryproject/ratify-verifier-go/licensechecker"

	"github.com/notaryproject/ratify/v2/internal/verifier"
	"github.com/notaryproject/ratify/v2/internal/verifier/verifiercommon"
)

const verifierTypeLicenseChecker = "licensechecker"

// options is the JSON configuration for the license checker verifier.
type options struct {
	// ArtifactTypes is the list of artifact types this verifier handles.
	// Accepts either a single string or an array of strings.
	ArtifactTypes verifiercommon.ArtifactTypes `json:"artifactTypes"`

	// AllowedLicenses is the list of permitted licenses.
	AllowedLicenses []string `json:"allowedLicenses"`
}

func init() {
	verifier.Register(verifierTypeLicenseChecker, func(opts verifier.NewOptions, _ []string) (ratify.Verifier, error) {
		raw, err := json.Marshal(opts.Parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal verifier parameters: %w", err)
		}

		var params options
		if err := json.Unmarshal(raw, &params); err != nil {
			return nil, fmt.Errorf("failed to unmarshal verifier parameters: %w", err)
		}

		return licensechecker.NewVerifier(&licensechecker.VerifierOptions{
			Name:            opts.Name,
			AllowedLicenses: params.AllowedLicenses,
			ArtifactTypes:   params.ArtifactTypes,
		})
	})
}
