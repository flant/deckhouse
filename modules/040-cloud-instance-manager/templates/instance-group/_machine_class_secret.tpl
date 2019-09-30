{{- define "instance_group_machine_class_secret" }}
{{- $context := index . 0 }}
{{- $ig := index . 1 }}
{{- $zone_name := index . 2 }}
---
apiVersion: v1
kind: Secret
metadata:
  name: machine-class-{{ $ig.name }}-{{ $zone_name }}
  namespace: d8-{{ $context.Chart.Name }}
{{ include "helm_lib_module_labels" (list $context) | indent 2 }}
type: Opaque
data:
  userData: {{ include "instance_group_machine_class_cloud_init_cloud_config" (list $context $ig $zone_name) | b64enc }}
{{- $tpl_context := dict }}
{{- $_ := set $tpl_context "Release" $context.Release }}
{{- $_ := set $tpl_context "Chart" $context.Chart }}
{{- $_ := set $tpl_context "Files" $context.Files }}
{{- $_ := set $tpl_context "Capabilities" $context.Capabilities }}
{{- $_ := set $tpl_context "Template" $context.Template }}
{{- $_ := set $tpl_context "Values" $context.Values }}
{{- $_ := set $tpl_context "instanceGroup" $ig }}
{{- $_ := set $tpl_context "zoneName" $zone_name }}
{{ tpl ($context.Files.Get (list "cloud-providers" $context.Values.cloudInstanceManager.internal.cloudProvider.type "config-for-machine-controller-manager.yaml" | join "/")) $tpl_context | indent 2 }}
{{- end }}