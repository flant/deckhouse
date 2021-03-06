/*
Copyright 2021 Flant CJSC

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

package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = Describe("Modules :: deckhouse-web :: hooks :: get_deckhouse_tag_and_version ::", func() {

	const (
		initValuesString       = `{"deckhouseWeb":{"deckhouseTag":"","deckhouseVersion":"","internal":{}}}`
		initConfigValuesString = `{}`

		stateWithStableChannel = `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    core.deckhouse.io/version: "20.20"
  name: deckhouse
  namespace: d8-system
spec:
  template:
    spec:
      containers:
      - name: deckhouse
        image: registry.deckhouse.io/deckhouse/ce:stable
`
		stateWithAbsentAnnotation = `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deckhouse
  namespace: d8-system
spec:
  template:
    spec:
      containers:
      - name: deckhouse
        image: registry.deckhouse.io/deckhouse/ce:sometag
`
	)

	f := HookExecutionConfigInit(initValuesString, initConfigValuesString)

	Context("Empty cluster", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(``))
			f.RunHook()
		})

		It("Hook must not fail", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.ValuesGet("deckhouseWeb.deckhouseTag").String()).To(Equal(""))
			Expect(f.ValuesGet("deckhouseWeb.deckhouseVersion").String()).To(Equal(""))
		})
	})

	Context("Absent core.deckhouse.io/version annotation", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(stateWithAbsentAnnotation))
			f.RunHook()
		})

		It("Hook must not fail with an absent version annotation", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.BindingContexts.Array()).ShouldNot(BeEmpty())
			Expect(f.ValuesGet("deckhouseWeb.deckhouseVersion").String()).To(Equal("unknown"))
			Expect(f.ValuesGet("deckhouseWeb.deckhouseTag").String()).To(Equal("sometag"))
		})
	})

	Context("Deckhouse on update channel", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(stateWithStableChannel))
			f.RunHook()
		})

		It("Hook must not fail, version and channel should be set", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.BindingContexts.Array()).ShouldNot(BeEmpty())
			Expect(f.BindingContexts.Get("0.snapshots.d8_deployment.0.filterResult").String()).To(MatchJSON(`
{
	"tag": "stable",
	"version": "20.20"
}
`))
			Expect(f.ValuesGet("deckhouseWeb.deckhouseTag").String()).To(Equal("stable"))
			Expect(f.ValuesGet("deckhouseWeb.deckhouseVersion").String()).To(Equal("20.20"))
		})
	})

})
