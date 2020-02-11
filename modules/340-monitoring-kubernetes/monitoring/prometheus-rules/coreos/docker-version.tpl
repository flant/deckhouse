- name: docker.version
  rules:
  - alert: UnsupportedDockerVersion
{{- if and (semverCompare ">=1.10" .Values.global.discovery.clusterVersion) (semverCompare "<1.14" .Values.global.discovery.clusterVersion) }}
    expr: sum by (container_runtime_version, job, kernel_version, kubelet_version, kubeproxy_version, node, os_image) (kube_node_info{kubelet_version=~"v1.1[0-3].+", container_runtime_version!~"docker://1\\.1[1-3]\\..*|docker://1[7-9]\\..*|docker://3\\..*"})
{{- else if semverCompare ">=1.14" .Values.global.discovery.clusterVersion }}
    expr: sum by (container_runtime_version, job, kernel_version, kubelet_version, kubeproxy_version, node, os_image) (kube_node_info{kubelet_version=~"v1.1[4-9].+", container_runtime_version!~"docker://1\\.13\\..*|docker://1[7-9]\\..*|docker://3\\..*"})
{{- end }}
    for: 20m
    labels:
      impact: negligible
      likelihood: certain
      tier: cluster
    annotations:
      polk_flant_com_markup_format: markdown
      description: |-
        Unsupported version {{`{{$labels.container_runtime_version}}`}} of Docker installed on {{`{{$labels.node}}`}} node.
        Supported version of Docker for kubernetes {{`{{$labels.kubelet_version}}`}} version:
{{- if and (semverCompare ">=1.10" .Values.global.discovery.clusterVersion) (semverCompare "<1.14" .Values.global.discovery.clusterVersion) }}
        * 1.11.x
        * 1.12.x
        * 1.13.x
        * 17.x
        * 18.x
        * 3.x (for moby project in Azure)
{{- else if semverCompare ">=1.14" .Values.global.discovery.clusterVersion }}
        * 1.13.x
        * 17.x
        * 18.x
        * 3.x (for moby project in Azure)
{{- end }}
      summary: >
        Unsupported version of Docker {{`{{$labels.container_runtime_version}}`}} installed for Kubernetes version: {{`{{$labels.kubelet_version}}`}}
- name: kubernetes.version
  rules:
  - alert: ControlPlaneAndKubeletVersionsDiffer
    expr: sum by (node, gitVersion, instance, job) (kubernetes_build_info{gitVersion!="v{{ .Values.global.discovery.clusterVersion }}", job!~"kube-dns|coredns"})
    for: 20m
    labels:
      impact: negligible
      likelihood: certain
      tier: cluster
    annotations:
      polk_flant_com_markup_format: markdown
      description: |-
        kube-apiserver is at version {{ .Values.global.discovery.clusterVersion }}, but cluster component {{`{{$labels.job}}`}} on {{`{{$labels.node}}`}} is at version {{`{{$labels.gitVersion}}`}}.
        1. Check it: `kubectl get nodes`
        2. Correct {{`{{$labels.job}}`}} version or control plane version on kubernetes master static pod manifests
      summary: >
        Different version of {{`{{$labels.job}}`}} on {{`{{$labels.node}}`}} node and kubernetes apiserver version
{{- if and (eq .Values.global.discovery.clusterType "Manual") (semverCompare ">=1.12" .Values.global.discovery.clusterVersion) }}
- name: kernel.version
  rules:
  - alert: UbuntuUnsupportedKernel
    expr: sum by (node, kernel_version, os_image) (kube_node_info{kernel_version!~"4.1[6-9]\\..*|5\\..*", os_image=~".*(Ubuntu|ubuntu).*"})
    for: 20m
    labels:
      impact: negligible
      likelihood: certain
      tier: cluster
    annotations:
      polk_flant_com_markup_format: markdown
      description: |-
        Unsupported kernel version of ubuntu OS on {{`{{$labels.node}}`}}.
        * For actual supported Ubuntu 16.04 kernel launch: `apt install linux-image-unsigned-4.18.0-20-generic linux-modules-4.18.0-20-generic linux-modules-extra-4.18.0-20-generic linux-headers-4.18.0-20-generic linux-headers-4.18.0-20`
        * For actual supported Ubuntu 18.04 just install latest 5.* kernel
{{- end }}