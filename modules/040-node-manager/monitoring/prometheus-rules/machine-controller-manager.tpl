{{- if include "machine_controller_manager_enabled" . }}
- name: d8.machine-controller-manager.availability
  rules:
  - alert: D8MachineControllerManagerPodIsNotReady
    expr: min by (pod) (kube_pod_status_ready{condition="false", namespace="d8-cloud-instance-manager", pod=~"machine-controller-manager-.*"}) > 0
    labels:
      severity_level: "8"
      tier: cluster
      d8_module: node-manager
      d8_component: machine-controller-manager
    annotations:
      plk_protocol_version: "1"
      plk_markup_format: "markdown"
      plk_pending_until_firing_for: "10m"
      plk_grouped_by__d8_machine_controller_manager_unavailable: "D8MachineControllerManagerUnavailable,tier=cluster,prometheus=deckhouse"
      plk_labels_as_annotations: "pod"
      summary: Под {{`{{$labels.pod}}`}} находится в состоянии НЕ Ready

  - alert: D8MachineControllerManagerPodIsNotRunning
    expr: max by (namespace, pod, phase) (kube_pod_status_phase{namespace="d8-cloud-instance-manager",phase!="Running",pod=~"machine-controller-manager-.*"} > 0)
    labels:
      severity_level: "8"
      tier: cluster
      d8_module: node-manager
      d8_component: machine-controller-manager
    annotations:
      plk_protocol_version: "1"
      plk_markup_format: "markdown"
      plk_pending_until_firing_for: "10m"
      plk_grouped_by__d8_machine_controller_manager_unavailable: "D8MachineControllerManagerUnavailable,tier=cluster,prometheus=deckhouse"
      plk_labels_as_annotations: "phase"
      summary: Под machine-controller-manager находится в состоянии НЕ Running
      description: |-
        Под {{`{{$labels.pod}}`}} находится в состоянии {{`{{$labels.phase}}`}}. Для проверки статуса пода необходимо выполнить:
        1. `kubectl -n {{`{{$labels.namespace}}`}} get pods {{`{{$labels.pod}}`}} -o json | jq .status`

  - alert: D8MachineControllerManagerTargetDown
    expr: max by (job) (up{job="machine-controller-manager", namespace="d8-cloud-instance-manager"} == 0)
    labels:
      severity_level: "8"
      tier: cluster
      d8_module: deckhouse
      d8_component: machine-controller-manager
    annotations:
      plk_protocol_version: "1"
      plk_markup_format: "markdown"
      plk_pending_until_firing_for: "5m"
      plk_grouped_by__d8_machine_controller_manager_unavailable: "D8MachineControllerManagerUnavailable,tier=cluster,prometheus=deckhouse"
      plk_labels_as_annotations: "instance,pod"
      plk_ignore_labels: "job"
      summary: Prometheus не может получить метрики cluster machine-controller-manager'a.

  - alert: D8MachineControllerManagerTargetAbsent
    expr: absent(up{job="machine-controller-manager", namespace="d8-cloud-instance-manager"} == 1)
    labels:
      severity_level: "8"
      tier: cluster
      d8_module: prometheus
      d8_component: machine-controller-manager
    annotations:
      plk_markup_format: "markdown"
      plk_protocol_version: "1"
      plk_pending_until_firing_for: "5m"
      plk_grouped_by__d8_machine_controller_manager_unavailable: "D8MachineControllerManagerUnavailable,tier=cluster,prometheus=deckhouse"
      summary: >
        В таргетах prometheus нет machine-controller-manager
      description: |-
        Machine controller manager используется для управления эфемерными нодами в кластере, его недоступность не позволит удалять и
        добавлять ноды.

        Необходимо выполнить следующие действия:
        1. Проверить наличие и состояние подов machine-controller-manager `kubectl -n d8-cloud-instance-manager get pods -l app=machine-controller-manager`
        2. Проверить наличие deployment'a machine-controller-manager `kubectl -n d8-cloud-instance-manager get deploy machine-controller-manager`
        3. Посмотреть состояние deployment'a machine-controller-manager `kubectl -n d8-cloud-instance-manager describe deploy machine-controller-manager`

  - alert: D8MachineControllerManagerUnavailable
    expr: |
      count(ALERTS{alertname=~"D8MachineControllerManagerPodIsNotReady|D8MachineControllerManagerPodIsNotRunning|D8MachineControllerManagerTargetAbsent|D8MachineControllerManagerTargetDown", alertstate="firing"}) > 0
      OR
      count(ALERTS{alertname=~"KubernetesDeploymentReplicasUnavailable", namespace="d8-cloud-instance-manager", deployment="machine-controller-manager", alertstate="firing"}) > 0
      OR
      count(ALERTS{alertname=~"KubernetesDeploymentStuck", namespace="d8-cloud-instance-manager", deployment="machine-controller-manager", alertstate="firing"}) > 0
    labels:
      tier: cluster
      d8_module: node-manager
      d8_component: machine-controller-manager
    annotations:
      plk_protocol_version: "1"
      plk_markup_format: "markdown"
      plk_alert_type: "group"
      plk_group_for__machine_controller_manager_replicas_unavailable: "KubernetesDeploymentReplicasUnavailable,namespace=d8-cloud-instance-manager,prometheus=deckhouse,deployment=machine-controller-manager"
      plk_group_for__machine_controller_manager_stuck: "KubernetesDeploymentStuck,namespace=d8-cloud-instance-manager,prometheus=deckhouse,deployment=machine-controller-manager"
      plk_grouped_by__d8_machine_controller_manager_malfunctioning: "D8MachineControllerManagerMalfunctioning,tier=cluster,prometheus=deckhouse"
      summary: Machine controller manager не работает
      description: |
        Machine controller manager не работает. Что именно с ним не так можно узнать в одном из связанных алертов.

- name: d8.machine-controller-manager.malfunctioning
  rules:
  - alert: D8MachineControllerManagerPodIsRestartingTooOften
    expr: max by (pod) (increase(kube_pod_container_status_restarts_total{namespace="d8-cloud-instance-manager", pod=~"machine-controller-manager-.*"}[1h]) and kube_pod_container_status_restarts_total{namespace="d8-cloud-instance-manager", pod=~"machine-controller-manager-.*"}) > 5
    labels:
      severity_level: "9"
      tier: cluster
      d8_module: node-manager
      d8_component: machine-controller-manager
    annotations:
      plk_protocol_version: "1"
      plk_markup_format: "markdown"
      plk_grouped_by__d8_machine_controller_manager_malfunctioning: "D8MachineControllerManagerMalfunctioning,tier=cluster,prometheus=deckhouse"
      plk_labels_as_annotations: "pod"
      summary: Machine controller manager слишком часто перезагружается
      description: |
        Количество перезапусков за последний час: {{`{{ $value }}`}}.

        Частый перезапуск Machine controller manager не является нормальной ситуацией, он должен быть постоянно запущена и работать.
        Необходимо посмотреть логи:
        1. `kubectl -n d8-cloud-instance-manager logs -f -l app=machine-controller-manager -c controller`

  - alert: D8MachineControllerManagerMalfunctioning
    expr: |
      count(ALERTS{alertname=~"D8MachineControllerManagerPodIsRestartingTooOften", alertstate="firing"}) > 0
    labels:
      tier: cluster
      d8_module: node-manager
      d8_component: machine-controller-manager
    annotations:
      plk_protocol_version: "1"
      plk_markup_format: "markdown"
      plk_alert_type: "group"
      summary: Machine controller manager работает некорректно
      description: |
        Machine controller manager работает некорректно. Что именно с ним не так можно узнать в одном из связанных алертов.
{{- else }}
[]
{{- end }}
