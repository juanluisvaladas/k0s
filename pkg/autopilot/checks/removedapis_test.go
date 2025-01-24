// Copyright 2024 k0s authors
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

package checks

import (
	"cmp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestRemovedGVKs(t *testing.T) {
	assert.True(t, slices.IsSortedFunc(removedGVKs[:], func(l, r removedAPI) int {
		if cmp := cmp.Compare(l.group, r.group); cmp != 0 {
			return cmp
		}
		if cmp := cmp.Compare(l.version, r.version); cmp != 0 {
			return cmp
		}
		return cmp.Compare(l.kind, r.kind)
	}), "removedGVKs needs to be sorted, so that it can be used for binary searches")

	// Test two random entries at the top and the bottom of the list
	version, currentVersion := removedInVersion(schema.GroupVersionKind{
		Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta2", Kind: "FlowSchema",
	})
	assert.Equal(t, "v1.29.0", version)
	assert.Equal(t, "v1beta3", currentVersion)

	version, currentVersion = removedInVersion(schema.GroupVersionKind{
		Group: "k0s.k0sproject.example.com", Version: "v1beta1", Kind: "RemovedCRD",
	})
	assert.Equal(t, "v99.99.99", version)
	assert.Equal(t, "", currentVersion)

	version, currentVersion = removedInVersion(schema.GroupVersionKind{
		Group: "k0s.k0sproject.example.com", Version: "v1beta1", Kind: "MustFail",
	})
	assert.Equal(t, "", version)
	assert.Equal(t, "", currentVersion)
}
