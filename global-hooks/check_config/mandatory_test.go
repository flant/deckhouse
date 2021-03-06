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

/*

User-stories:
1. There are mandatory fields `global.project` and `global.clusterName`. Hook must fail when the parameters aren't set.

*/

package hooks

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "")
}

const (
	initValuesString       = `{}`
	initConfigValuesString = `{"global": {}}`
)

var _ = Describe("Global hooks :: cluster_is_bootstraped ::", func() {
	f := HookExecutionConfigInit(initValuesString, initConfigValuesString)

	Context("Both `global.project` and `global.clusterName` aren't set", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.GenerateOnStartupContext())
			f.RunHook()
		})

		It("Hook must fail", func() {
			Expect(f).To(Not(ExecuteSuccessfully()))
		})
	})

	Context("`global.project` is set; `global.clusterName` isn't set", func() {
		BeforeEach(func() {
			f.ConfigValuesSet("global.project", "ppp")
			f.BindingContexts.Set(f.GenerateOnStartupContext())
			f.RunHook()
		})

		It("Hook must fail", func() {
			Expect(f).To(Not(ExecuteSuccessfully()))
		})
	})

	Context("`global.project` isn't set; `global.clusterName` is set", func() {
		BeforeEach(func() {
			f.ConfigValuesSet("global.clusterName", "ccc")
			f.BindingContexts.Set(f.GenerateOnStartupContext())
			f.RunHook()
		})

		It("Hook must fail", func() {
			Expect(f).To(Not(ExecuteSuccessfully()))
		})
	})

	Context("`global.project` isn't set; `global.clusterName` is set", func() {
		BeforeEach(func() {
			f.ConfigValuesSet("global.project", "ppp")
			f.ConfigValuesSet("global.clusterName", "ccc")
			f.BindingContexts.Set(f.GenerateOnStartupContext())
			f.RunHook()
		})

		It("Hook must not fail", func() {
			Expect(f).To(ExecuteSuccessfully())
		})
	})

})
