// Copyright © 2020 Atomist
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vent

import (
	"testing"
)

func TestGenerateSignature(t *testing.T) {
	s1, err := generateSignature([]byte(`{"jason":"isbell"}`), "The400Unit")
	if err != nil {
		t.Errorf("failed to create signature: %v", err)
	}
	e1 := "sha1=634212a9128672522f8d9ac32657d996d80ef7be"
	if s1 != e1 {
		t.Errorf("failed to generate proper signature: '%s' (expected: '%s')", s1, e1)
	}
}
