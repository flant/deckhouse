ARG BASE_ALPINE
ARG BASE_GOLANG_ALPINE
FROM $BASE_GOLANG_ALPINE as artifact
WORKDIR /src/
COPY main.go go.mod go.sum /src/
RUN apk add --no-cache git && \
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ingress-conversion-webhook main.go

FROM $BASE_ALPINE
COPY --from=artifact /src/ingress-conversion-webhook /ingress-conversion-webhook
USER nobody
ENTRYPOINT [ "/ingress-conversion-webhook" ]
