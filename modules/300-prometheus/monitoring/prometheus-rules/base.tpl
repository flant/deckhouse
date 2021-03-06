- name: d8.prometheus.base
  rules:
    - alert: D8MainPrometheusMalfunctioning
      expr: max(ALERTS{alertname="PrometheusMalfunctioning", namespace="d8-monitoring", service="prometheus", alertstate="firing"})
      labels:
        tier: cluster
        d8_module: prometheus
        d8_component: prometheus-main
      annotations:
        plk_markup_format: "markdown"
        plk_protocol_version: "1"
        plk_alert_type: "group"
        plk_ignore_labels: "pod"
        plk_group_for__prometheus_malfunctioning: "PrometheusMalfunctioning,prometheus=deckhouse,namespace=d8-monitoring,service=prometheus"
        plk_grouped_by__main: "D8PrometheusMalfunctioning,tier=cluster,prometheus=deckhouse"
        description: |-
          Служебный prometheus работает некорректно. Что именно с ним не так можно узнать в одном из связанных алертов.
        summary: Служебный prometheus работает некорректно
{{- if .Values.prometheus.longtermRetentionDays }}
    - alert: D8LongtermPrometheusMalfunctioning
      expr: max(ALERTS{alertname="PrometheusMalfunctioning", namespace="d8-monitoring", service="prometheus-longterm", alertstate="firing"})
      labels:
        tier: cluster
        d8_module: prometheus
        d8_component: prometheus-longterm
      annotations:
        plk_markup_format: "markdown"
        plk_protocol_version: "1"
        plk_alert_type: "group"
        plk_ignore_labels: "pod"
        plk_group_for__prometheus_malfunctioning: "PrometheusMalfunctioning,prometheus=deckhouse,namespace=d8-monitoring,service=prometheus-longterm"
        plk_grouped_by__main: "D8PrometheusMalfunctioning,tier=cluster,prometheus=deckhouse"
        description: |
          Служебный prometheus longterm работает некорректно. Что именно с ним не так можно узнать в одном из связанных алертов.
        summary: Служебный prometheus longterm работает некорректно
{{- end }}

    - alert: D8PrometheusMalfunctioning
      expr: |
        count(ALERTS{alertname=~"D8MainPrometheusMalfunctioning|D8LongtermPrometheusMalfunctioning", alertstate="firing"}) > 0
        OR
        count(ALERTS{alertname=~"IngressResponses5xx", namespace="d8-monitoring", service="trickster", alertstate="firing"}) > 0
        OR
        count(ALERTS{alertname=~"IngressResponses5xx", namespace="d8-monitoring", service="prometheus", alertstate="firing"}) > 0
      labels:
        tier: cluster
        d8_module: prometheus
      annotations:
        plk_markup_format: "markdown"
        plk_protocol_version: "1"
        plk_alert_type: "group"
        plk_group_for__trickster_responses_5xx: "IngressResponses5xx,namespace=d8-monitoring,prometheus=deckhouse,service=trickster"
        plk_group_for__prometheus_responses_5xx: "IngressResponses5xx,namespace=d8-monitoring,prometheus=deckhouse,service=prometheus"
        description: |
          Какой-то из deckhouse prometheus работает некорректно. Что именно и с каким именно prometheus не так можно узнать в одном из связанных алертов.
        summary: Какой-то из deckhouse prometheus работает некорректно

{{- if .Values.prometheus.longtermRetentionDays }}
    - alert: D8PrometheusLongtermTargetAbsent
      expr: absent(up{job="prometheus", namespace="d8-monitoring", service="prometheus-longterm"} == 1)
      labels:
        severity_level: "7"
        tier: cluster
        d8_module: prometheus
        d8_component: prometheus-longterm
      annotations:
        plk_markup_format: "markdown"
        plk_protocol_version: "1"
        plk_pending_until_firing_for: "30m"
        plk_grouped_by__main: "D8LongtermPrometheusMalfunctioning,tier=cluster,prometheus=deckhouse"
        description: |-
          Данный Prometheus используется только для отображения исторических данных и его недоступность может быть совсем некритичной, однако если он будет долгое время недоступен, в будущем вы не сможете посмотреть статистику.

          Чаще всего у данного пода проблемы из-за недоступности диска (например, под переехал на ноду, где диск не цепляется).

          Куда следует смотреть:
          1. Посмотреть информацию о Statefulset: `kubectl -n d8-monitoring describe statefulset prometheus-longterm`
          2. В каком статусе находится его PVC (если он используется): `kubectl -n d8-monitoring describe pvc prometheus-longterm-db-prometheus-longterm-0`
          3. В каком состоянии находится сам под: `kubectl -n d8-monitoring describe pod prometheus-longterm-0`
        summary: >
          В таргетах prometheus нет prometheus longterm
{{- end }}

    - alert: D8TricksterTargetAbsent
      expr: (max(up{job="prometheus", service="prometheus"}) == 1) * absent(up{job="trickster", namespace="d8-monitoring"} == 1)
      labels:
        severity_level: "5"
        tier: cluster
        d8_module: prometheus
        d8_component: trickster
      annotations:
        plk_markup_format: "markdown"
        plk_protocol_version: "1"
        plk_pending_until_firing_for: "2m"
        plk_grouped_by__main: "D8PrometheusMalfunctioning,tier=cluster,prometheus=deckhouse"
        description: |-
          Данный компонент используют:
          * `prometheus-metrics-adapter` — мы остаемся без работающего HPA (автоскейлинг) и не можем посмотреть потребление ресурсов с помощью `kubectl`.
          * `vertical-pod-autoscaler` — для него этот инцидент не так страшен, так как VPA смотрит историю потребления за 8 дней.
          * `grafana` — по умолчанию все дашборды используют trickster для кеширования запросов к Prometheus. Можно забирать данные напрямую из Prometheus, минуя trickster, однако это может привести к повышенному потреблению памяти Prometheus и, соответственно, к недоступности.

          Куда смотреть:
          1. Информация о deployment: `kubectl -n d8-monitoring describe deployment trickster`
          2. Информация о pod: `kubectl -n d8-monitoring describe pod -l app=trickster`
          3. Чаще всего trickster становится недоступным из-за проблем с самим Prometheus, так как readinessProbe trickster'а проверяет доступность Prometheus. Поэтому, убедитесь, что prometheus работает: `kubectl -n d8-monitoring describe pod -l app=prometheus,prometheus=main`
        summary: >
          В таргетах prometheus нет trickster

    - alert: D8TricksterTargetAbsent
      expr: absent(up{job="trickster", namespace="d8-monitoring"} == 1)
      for: 5m
      labels:
        severity_level: "5"
        tier: cluster
        d8_module: prometheus
        d8_component: trickster
      annotations:
        plk_markup_format: "markdown"
        plk_protocol_version: "1"
        plk_grouped_by__main: "D8PrometheusMalfunctioning,tier=cluster,prometheus=deckhouse"
        description: |-
          Данный компонент используют:
          * `prometheus-metrics-adapter` — мы остаемся без работающего HPA (автоскейлинг) и не можем посмотреть потребление ресурсов с помощью `kubectl`.
          * `vertical-pod-autoscaler` — для него этот инцидент не так страшен, так как VPA смотрит историю потребления за 8 дней.
          * `grafana` — по умолчанию все дашборды используют trickster для кеширования запросов к Prometheus. Можно забирать данные напрямую из Prometheus, минуя trickster, однако это может привести к повышенному потреблению памяти Prometheus и, соответственно, к недоступности.

          Куда смотреть:
          1. Информация о deployment: `kubectl -n d8-monitoring describe deployment trickster`
          2. Информация о pod: `kubectl -n d8-monitoring describe pod -l app=trickster`
          3. Чаще всего trickster становится недоступным из-за проблем с самим Prometheus, так как readinessProbe trickster'а проверяет доступность Prometheus. Поэтому, убедитесь, что prometheus работает: `kubectl -n d8-monitoring describe pod -l app=prometheus,prometheus=main`
        summary: >
          В таргетах prometheus нет trickster
