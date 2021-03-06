// Copyright 2021 Flant CJSC
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

package docker_registry_watcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsValidImageDigest(t *testing.T) {
	good := []string{
		"sha256:2a8b0e16c845d9a9521d6ea2534096bb095c0ad1ff6a65fe6397158ac9537057",
		"2a8b0e16c845d9a9521d6ea2534096bb095c0ad1ff6a65fe6397158ac9537057",
	}

	bad := []string{
		"sha256:2a8b0e16c845d9a9521d6ea2534096bb095c0ad1ff6a65fe6397158ac953705",
		"sha25:2a8b0e16c845d9a9521d6ea2534096bb095c0ad1ff6a65fe6397158ac953705",
		"sha256-2a8b0e16c845d9a9521d6ea2534096bb095c0ad1ff6a65fe6397158ac953705",
	}

	for _, name := range good {
		res := IsValidImageDigest(name)
		assert.Truef(t, res, "%s should be valid image digest", name)
	}
	for _, name := range bad {
		res := IsValidImageDigest(name)
		assert.Falsef(t, res, "%s should not be valid image digest", name)
	}
}

func Test_FindImageDigest(t *testing.T) {
	good := []string{
		"docker.io/library/alpine@sha256:7746df395af22f04212cd25a92c1d6dbc5a06a0ca9579a229ef43008d4d1302a",
		"docker-pullable://docker.io/library/alpine@sha256:2a8b0e16c845d9a9521d6ea2534096bb095c0ad1ff6a65fe6397158ac9537057",
	}

	bad := []string{
		"docker://docker.io/library/alpine@sha256:7746df395af22f04212cd25a92c1d6dbc5a06a0ca9579a229ef43008d4d1302a",
		"sha256:2a8b0e16c845d9a9521d6ea2534096bb095c0ad1ff6a65fe6397158ac953705",
		"docker-pullable://docker.io/library/alpine@sha256:2a8b0e16c845d9a9521d6ea2534096bb095c0ad1ff6a65fe6397158ac95370",
	}

	for _, name := range good {
		_, err := FindImageDigest(name)
		assert.NoError(t, err, "%s should be valid imageId with digest", name)
	}
	for _, name := range bad {
		_, err := FindImageDigest(name)
		assert.Errorf(t, err, "%s should not be valid imageId with digest", name)
	}
}
