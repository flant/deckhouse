# Based on https://github.com/coredns/coredns/blob/master/Dockerfile
ARG BASE_ALPINE
FROM coredns/coredns:1.6.9@sha256:40ee1b708e20e3a6b8e04ccd8b6b3dd8fd25343eab27c37154946f232649ae21 as artifact

FROM $BASE_ALPINE
COPY --from=artifact /coredns /coredns
ENTRYPOINT [ "/coredns" ]

