## Разработка Prometheus Rules

### Схема метрик

Exporter extended-monitoring экспортирует метрики в следующем формате:
```
extended_monitoring_{0}_threshold{{namespace="{1}", threshold="{2}", {3}="{4}"}} {5}
```

0. Kind Kubernetes объекта в нижнем регистре;
1. Namespace, где находится Kubernetes объект. `None` для non-namespaced объектов;
2. Имя threshold аннотации;
3. Kind Kubernetes объекта в нижнем регистре. Дублируется для удобства работы с PromQL;
4. Имя Kubernetes объекта;
5. Значение, полученное из value аннотации или из стандартных значений, закрепленных в исходном коде экспортера.

## Добавление стандартных аннотаций и их значений

В файле [extended-monitoring.py](modules/350-extended-monitoring/images/extended-monitoring/src/extended-monitoring.py) достаточно добавить в аттрибут `default_annotations` в Annotated классе, соответствующий типу Kubernetes объекта, необходимые аннотации.