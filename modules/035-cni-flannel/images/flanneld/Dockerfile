FROM quay.io/coreos/flannel:v0.11.0-amd64
COPY entrypoint.sh /
RUN apk add --no-cache curl jq
ENTRYPOINT ["/entrypoint.sh"]