ARG BASE_ALPINE
ARG BASE_GOLANG_ALPINE
FROM $BASE_GOLANG_ALPINE as artifact
WORKDIR /go/src/github.com/gardener/autoscaler
COPY "do_not_violate_NodeGroup's_MinSize_when_fixing_NodeGroup.patch" /
COPY "scaling_down_terminated_machines.patch" /

RUN apk add --no-cache linux-headers build-base git mercurial patch && \
    wget https://github.com/gardener/autoscaler/archive/v0.12.0.tar.gz -O - | tar -xz --strip-components=1 -C /go/src/github.com/gardener/autoscaler/ && \
    patch -p1 < "/do_not_violate_NodeGroup's_MinSize_when_fixing_NodeGroup.patch" && \
    patch -p1 < "/scaling_down_terminated_machines.patch" && \
    cd cluster-autoscaler && \
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o cluster-autoscaler

FROM $BASE_ALPINE
COPY --from=artifact /go/src/github.com/gardener/autoscaler/cluster-autoscaler/cluster-autoscaler /
ENTRYPOINT [ "/cluster-autoscaler" ]
