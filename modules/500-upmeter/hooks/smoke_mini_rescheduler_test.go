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
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = Describe("Modules :: upmeter :: hooks :: smoke_mini_rescheduler ::", func() {
	Context("Empty cluster", func() {
		f := HookExecutionConfigInit(`{
"upmeter":{
  "internal":{
    "smokeMini":{
      "sts":{"a":{},"b":{},"c":{}}
    }
  }
}}`, `{}`)

		DescribeTable("version change",
			func(state string) {
				f.BindingContexts.Set(f.KubeStateSet(state))
				f.RunHook()
				Expect(f).To(ExecuteSuccessfully())
			},
			Entry("One node, no pods", `
---
apiVersion: v1
kind: Node
metadata:
  labels:
    kubernetes.io/hostname: node-a-1
  name: node-a-1
status:
  conditions:
  - status: "True"
    type: Ready
`),
			Entry("One node and a pod on it", `
---
apiVersion: v1
kind: Node
metadata:
  labels:
    kubernetes.io/hostname: node-a-1
  name: node-a-1
status:
  conditions:
  - status: "True"
    type: Ready
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  annotations:
    node: node-a-1
    zone: nova
  labels:
    app: smoke-mini
    module: upmeter
  name: smoke-mini-a
  namespace: d8-upmeter
spec:
  selector:
    matchLabels:
      smoke-mini: a
  serviceName: smoke-mini-a
  template:
    metadata:
      labels:
        app: smoke-mini
        smoke-mini: a
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - nodeSelectorTerms:
              matchExpressions:
              - key: kubernetes.io/hostname
                operator: In
                values:
                - node-a-1
            weight: 1
      containers:
      - image: registry.deckhouse.io/deckhouse/ce/upmeter/smoke-mini:whatever
        name: smoke-mini
status:
  collisionCount: 0
  currentReplicas: 1
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1
---
apiVersion: v1
kind: Pod
metadata:
  name: smoke-mini-a-0
  namespace: d8-upmeter
spec:
  nodeName: node-a-1
  containers:
  - image: registry.deckhouse.io/deckhouse/ce/upmeter/smoke-mini:whatever
    name: smoke-mini
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - weight: 1
        nodeSelectorTerms:
          matchExpressions:
          - key: kubernetes.io/hostname
            operator: In
            values:
            - node-a-1
`),
			Entry("Two nodes and a pod", `
---
apiVersion: v1
kind: Node
metadata:
  labels:
    kubernetes.io/hostname: node-a-1
  name: node-a-1
status:
  conditions:
  - status: "True"
    type: Ready
---
apiVersion: v1
kind: Node
metadata:
  labels:
    kubernetes.io/hostname: node-a-2
  name: node-a-2
status:
  conditions:
  - status: "True"
    type: Ready
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  annotations:
    node: node-a-1
    zone: nova
  labels:
    app: smoke-mini
    module: upmeter
  name: smoke-mini-a
  namespace: d8-upmeter
spec:
  selector:
    matchLabels:
      smoke-mini: a
  serviceName: smoke-mini-a
  template:
    metadata:
      labels:
        app: smoke-mini
        smoke-mini: a
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - nodeSelectorTerms:
              matchExpressions:
              - key: kubernetes.io/hostname
                operator: In
                values:
                - node-a-1
            weight: 1
      containers:
      - image: registry.deckhouse.io/deckhouse/ce/upmeter/smoke-mini:whatever
        name: smoke-mini
status:
  collisionCount: 0
  currentReplicas: 1
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1
---
apiVersion: v1
kind: Pod
metadata:
  name: smoke-mini-a-0
  namespace: d8-upmeter
spec:
  nodeName: node-a-1
  containers:
  - image: registry.deckhouse.io/deckhouse/ce/upmeter/smoke-mini:whatever
    name: smoke-mini
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - weight: 1
        nodeSelectorTerms:
          matchExpressions:
          - key: kubernetes.io/hostname
            operator: In
            values:
            - node-a-1
`),
			Entry("Unscheduled pod", `
---
apiVersion: v1
kind: Node
metadata:
  labels:
    kubernetes.io/hostname: node-a-1
  name: node-a-1
status:
  conditions:
  - status: "True"
    type: Ready
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  annotations:
    node: node-a-1
    zone: nova
  labels:
    app: smoke-mini
    module: upmeter
  name: smoke-mini-a
  namespace: d8-upmeter
spec:
  selector:
    matchLabels:
      smoke-mini: a
  serviceName: smoke-mini-a
  template:
    metadata:
      labels:
        app: smoke-mini
        smoke-mini: a
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - nodeSelectorTerms:
              matchExpressions:
              - key: kubernetes.io/hostname
                operator: In
                values:
                - node-a-1
            weight: 1
      containers:
      - image: registry.deckhouse.io/deckhouse/ce/upmeter/smoke-mini:whatever
        name: smoke-mini
status:
  collisionCount: 0
  currentReplicas: 1
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1
---
apiVersion: v1
kind: Pod
metadata:
  name: smoke-mini-a-0
  namespace: d8-upmeter
spec:
  nodeName: ""
  containers:
    - image: registry.deckhouse.io/deckhouse/ce/upmeter/smoke-mini:whatever
      name: smoke-mini
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - weight: 1
        nodeSelectorTerms:
          matchExpressions:
          - key: kubernetes.io/hostname
            operator: In
            values:
            - node-a-1
`))
	})

	Context("Empty cluster", func() {
		f := HookExecutionConfigInit(`{"upmeter":{"smokeMiniDisabled": true}}`, `{}`)

		It("Should execute successfully", func() {
			f.BindingContexts.Set(f.KubeStateSet(`
---
apiVersion: v1
kind: Node
metadata:
  labels:
    kubernetes.io/hostname: node-a-1
  name: node-a-1
status:
  conditions:
  - status: "True"
    type: Ready
`))
			f.RunHook()
			Expect(f).To(ExecuteSuccessfully())
		})
	})
})
