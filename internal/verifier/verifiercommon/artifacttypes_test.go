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

package verifiercommon

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestArtifactTypesUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ArtifactTypes
		wantErr bool
	}{
		{name: "single string", input: `"application/spdx+json"`, want: ArtifactTypes{"application/spdx+json"}},
		{name: "empty string", input: `""`, want: nil},
		{name: "array", input: `["a","b"]`, want: ArtifactTypes{"a", "b"}},
		{name: "empty array", input: `[]`, want: ArtifactTypes{}},
		{name: "invalid", input: `123`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ArtifactTypes
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Unmarshal() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
