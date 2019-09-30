FROM golang:1.12
RUN apt update && apt install -qfy \
  bash make git patch ca-certificates openssh-client openssl
RUN mkdir /coreos && cd /coreos \
  && git clone -b "v0.31.1" --single-branch https://github.com/coreos/prometheus-operator
WORKDIR /coreos/prometheus-operator
ADD tls-config.patch .
RUN patch -p1 < tls-config.patch && make operator

# todo: подобрать подходящий образ
FROM quay.io/prometheus/busybox:latest
COPY --from=0 /coreos/prometheus-operator/operator /bin/operator
# On busybox 'nobody' has uid `65534'
USER 65534
ENTRYPOINT ["/bin/operator"]