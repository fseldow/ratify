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

package sbom

import (
	"encoding/json"
	"fmt"

	"github.com/notaryproject/ratify-go"
	"github.com/notaryproject/ratify-verifier-go/sbom"

	"github.com/notaryproject/ratify/v2/internal/verifier"
	"github.com/notaryproject/ratify/v2/internal/verifier/verifiercommon"
)

const verifierTypeSBOM = "sbom"

// options is the JSON configuration for the SBOM verifier.
type options struct {
	// ArtifactTypes is the list of artifact types this verifier handles.
	// Accepts either a single string or an array of strings.
	ArtifactTypes verifiercommon.ArtifactTypes `json:"artifactTypes"`

	// DisallowedLicenses is the list of licenses that are not permitted.
	DisallowedLicenses []string `json:"disallowedLicenses,omitempty"`

	// DisallowedPackages is the list of packages that are not permitted.
	DisallowedPackages []sbom.PackageInfo `json:"disallowedPackages,omitempty"`
}

func init() {
	verifier.Register(verifierTypeSBOM, func(opts verifier.NewOptions, _ []string) (ratify.Verifier, error) {
		raw, err := json.Marshal(opts.Parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal verifier parameters: %w", err)
		}

		var params options
		if err := json.Unmarshal(raw, &params); err != nil {
			return nil, fmt.Errorf("failed to unmarshal verifier parameters: %w", err)
		}

		return sbom.NewVerifier(&sbom.VerifierOptions{
			Name:               opts.Name,
			DisallowedLicenses: params.DisallowedLicenses,
			DisallowedPackages: params.DisallowedPackages,
			ArtifactTypes:      params.ArtifactTypes,
		})
	})
}
