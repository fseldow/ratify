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

package main

import (
	"encoding/json"
	"testing"

	"github.com/notaryproject/ratify/v2/internal/executor"
)

// TestPluginVerifierExecutorConfigs validates that the executor configurations
// used by the migrated K8s plugin e2e tests (licensechecker, sbom,
// vulnerabilityreport) parse and construct a working ScopedExecutor through the
// same code path used by the CLI and the Gatekeeper provider controller. This
// gives offline confidence that the Executor CRs applied by the bats tests are
// structurally valid and that the verifiers are registered.
func TestPluginVerifierExecutorConfigs(t *testing.T) {
	tests := []struct {
		name   string
		config string
	}{
		{
			name: "licensechecker complete",
			config: `{
  "executors": [{
    "scopes": ["registry:5000"],
    "stores": [{"type": "registry-store", "parameters": {"credential": {"provider": "static", "username": "u", "password": "p"}, "plainHttp": true}}],
    "verifiers": [{
      "name": "licensechecker-1",
      "type": "licensechecker",
      "parameters": {"artifactTypes": "application/vnd.ratify.spdx.v0", "allowedLicenses": ["MIT", "Apache-2.0", "GPL-2.0-only"]}
    }],
    "policyEnforcer": {"type": "threshold-policy", "parameters": {"policy": {"threshold": 1, "rules": [{"verifierName": "licensechecker-1"}]}}}
  }]
}`,
		},
		{
			name: "sbom denylist",
			config: `{
  "executors": [{
    "scopes": ["registry:5000"],
    "stores": [{"type": "registry-store", "parameters": {"credential": {"provider": "static", "username": "u", "password": "p"}, "plainHttp": true}}],
    "verifiers": [{
      "name": "sbom-1",
      "type": "sbom",
      "parameters": {"artifactTypes": "application/spdx+json", "disallowedLicenses": ["MPL"], "disallowedPackages": [{"name": "zlib", "version": "1.2.13-r0"}]}
    }],
    "policyEnforcer": {"type": "threshold-policy", "parameters": {"policy": {"threshold": 1, "rules": [{"verifierName": "sbom-1"}]}}}
  }]
}`,
		},
		{
			name: "vulnerabilityreport denylist",
			config: `{
  "executors": [{
    "scopes": ["registry:5000"],
    "stores": [{"type": "registry-store", "parameters": {"credential": {"provider": "static", "username": "u", "password": "p"}, "plainHttp": true}}],
    "verifiers": [{
      "name": "vulnerabilityreport-1",
      "type": "vulnerabilityreport",
      "parameters": {"artifactTypes": "application/sarif+json", "maximumAge": "8760h", "denylistCVEs": ["CVE-2021-44228"]}
    }],
    "policyEnforcer": {"type": "threshold-policy", "parameters": {"policy": {"threshold": 1, "rules": [{"verifierName": "vulnerabilityreport-1"}]}}}
  }]
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts executor.Options
			if err := json.Unmarshal([]byte(tt.config), &opts); err != nil {
				t.Fatalf("failed to unmarshal executor config: %v", err)
			}
			if _, err := executor.NewScopedExecutor(opts); err != nil {
				t.Fatalf("failed to build scoped executor: %v", err)
			}
		})
	}
}
