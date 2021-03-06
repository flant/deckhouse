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
	"encoding/json"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"

	"github.com/deckhouse/deckhouse/deckhouse-controller/pkg/app"
)

// Secrets are loaded as in Kubernetes source code
// https://github.com/kubernetes/kubernetes/blob/v1.16.0/pkg/credentialprovider/config.go

// DockerConfigJSON represents a local docker auth config file
// for pulling images.
type DockerConfigJSON struct {
	Auths DockerConfig `json:"auths"`

	XXX map[string]interface{} `yaml:",inline"`
}

// DockerConfig represents the config file used by the docker CLI.
// This config that represents the credentials that should be used
// when pulling images from specific image repositories.
type DockerConfig map[string]DockerConfigEntry

type DockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Auth     string `json:"auth,omitempty"`

	// Email is an optional value associated with the username.
	// This field is deprecated and will be removed in a later
	// version of docker.
	Email string `json:"email,omitempty"`

	ServerAddress string `json:"serveraddress,omitempty"`

	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`

	// RegistryToken is a bearer token to be sent to a registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

// AnonymousAuth is an empty entry for unknown registries
var AnonymousAuth = DockerConfigEntry{}

// DockerCfgAuths stores all available registries and their auths
var DockerCfgAuths = DockerConfig{}

func LoadDockerRegistrySecret(bytes []byte) error {
	//
	var tmpVar interface{}
	err := json.Unmarshal(bytes, &tmpVar)
	if err != nil {
		return err
	}

	isAuths := false
	if tmpMap, ok := tmpVar.(map[string]interface{}); ok {
		if _, hasKey := tmpMap["auths"]; hasKey {
			isAuths = hasKey
		}
	} else {
		return fmt.Errorf("bad JSON structure: should be an object")
	}

	if isAuths {
		// unmarshal as DockerConfigJson
		tmpConfigJSON := DockerConfigJSON{}
		err := json.Unmarshal(bytes, &tmpConfigJSON)
		if err != nil {
			return err
		}
		DockerCfgAuths = tmpConfigJSON.Auths
	} else {
		// copy all from top keys
		err := json.Unmarshal(bytes, &DockerCfgAuths)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewKeychain() authn.Keychain {
	return &kubeKeychain{}
}

type kubeKeychain struct {
}

func (k *kubeKeychain) Resolve(target authn.Resource) (authn.Authenticator, error) {
	cfg, ok := DockerCfgAuths[target.RegistryStr()]
	if !ok {
		if app.InsecureRegistry == "yes" {
			return authn.Anonymous, nil
		}
		return nil, fmt.Errorf("no auth for registry %s", target.RegistryStr())
	}

	if cfg == AnonymousAuth {
		return authn.Anonymous, nil
	}

	return authn.FromConfig(authn.AuthConfig{
		Username:      cfg.Username,
		Password:      cfg.Password,
		Auth:          cfg.Auth,
		IdentityToken: cfg.IdentityToken,
		RegistryToken: cfg.RegistryToken,
	}), nil

}
