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

var _ = Describe("Modules :: monitoring-custom :: hooks :: reserved_domain_nodes ::", func() {
	const (
		properResources = `
---
apiVersion: v1
kind: Node
metadata:
  name: system
  labels:
    node-role.deckhouse.io/system: ""
spec:
  taints:
  - effect: NoExecute
    key: dedicated.deckhouse.io
    value: system
`
		resourcesWithReservedLabels = `
---
apiVersion: v1
kind: Node
metadata:
  name: stateful
  labels:
    node-role.deckhouse.io/stateful: ""
`
		resourcesWithReservedTaints = `
---
apiVersion: v1
kind: Node
metadata:
  name: database
spec:
  taints:
  - effect: NoExecute
    key: dedicated.deckhouse.io
    value: database
`
	)
	f := HookExecutionConfigInit(
		`{"monitoringKubernetes":{"internal":{}},"global":{"enabledModules":[]}}`,
		`{}`,
	)

	Context("Empty cluster", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(``))
			f.RunHook()
		})

		It("Hook must not fail", func() {
			Expect(f).To(ExecuteSuccessfully())
		})
	})

	Context("Cluster containing proper Node", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(properResources))
			f.RunHook()
		})

		It("Hook must not fail, labels and taints should be selected", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.BindingContexts.Array()).ShouldNot(BeEmpty())
			Expect(f.BindingContexts.Array()).Should(HaveLen(1))
			Expect(f.BindingContexts.Get("0.snapshots.nodes.0.filterResult").String()).To(MatchJSON(`
{
  "name": "system",
  "usedLabelsAndTaints": [
	"system"
  ]
}
`))
		})
	})

	Context("Cluster with Node having reserved `metadata.labels`", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(resourcesWithReservedLabels))
			f.RunHook()
		})

		It("Hook must not fail, labels should be selected", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.BindingContexts.Array()).ShouldNot(BeEmpty())
			Expect(f.BindingContexts.Array()).Should(HaveLen(1))
			Expect(f.BindingContexts.Get("0.snapshots.nodes.0.filterResult").String()).To(MatchJSON(`
{
  "name": "stateful",
  "usedLabelsAndTaints": [
	"stateful"
  ]
}
`))
		})
	})

	Context("Cluster with Node having reserved `spec.taints`", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(resourcesWithReservedTaints))
			f.RunHook()
		})

		It("Hook must not fail, taints should be selected", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.BindingContexts.Array()).ShouldNot(BeEmpty())
			Expect(f.BindingContexts.Array()).Should(HaveLen(1))
			Expect(f.BindingContexts.Get("0.snapshots.nodes.0.filterResult").String()).To(MatchJSON(`
{
  "name": "database",
  "usedLabelsAndTaints": [
	"database"
  ]
}
`))
		})
	})

})
