---
title: "Конфигурация"
---

## Параметры

* `logLevel` — уровень логирования Deckhouse:
    * Возможные варианты: `Debug`, `Info`, `Error`; 
    * По умолчанию `Info`.
* `bundle` — вариант поставки Deckhouse. Определяет включенные по умолчанию модули. 
    * Возможные варианты:
        * `Default` — включает рекомендованный набор модулей для работы кластера: мониторинга, контроля авторизации, организации работы сети и других потребностей. С актуальным списком можно ознакомиться [здесь](https://github.com/deckhouse/deckhouse/blob/master/modules/values-default.yaml).
        * `Minimal` — минимально возможная поставка, которая включает единственный модуль (этот).
        * `Managed` - поставка для managed кластеров от облачных провайдеров, например Google Kubernetes Engine (GKE)
    * По умолчанию `Default`.
* `releaseChannel` — канал обновлений Deckhouse.
    * Возможные варианты в порядке возрастания стабильности обновлений: `Alpha`, `Beta`, `EarlyAccess`, `Stable`, `RockSolid`. 

### Примеры

```yaml
deckhouse: |
  logLevel: Debug
  bundle: Minimal
  releaseChannel: RockSolid
```