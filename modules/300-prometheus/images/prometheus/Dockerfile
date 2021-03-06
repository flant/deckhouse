ARG BASE_GOLANG_BUSTER
ARG BASE_ALPINE
FROM $BASE_GOLANG_BUSTER as artifact

RUN curl -sL https://deb.nodesource.com/setup_12.x | bash - &&  \
  apt install -y nodejs && \
  npm update -g npm && \
  npm install webpack -g && \
  npm config set registry http://registry.npmjs.org/ && \
  apt-key adv --fetch-keys http://dl.yarnpkg.com/debian/pubkey.gpg && \
  echo "deb https://dl.yarnpkg.com/debian/ stable main" > /etc/apt/sources.list.d/yarn.list && \
  apt update && apt install -y yarn && \
  apt install -y make bash git patch ca-certificates openssl openssh-client

RUN mkdir /prometheus && cd /prometheus \
  && git clone -b "v2.28.0" --single-branch https://github.com/prometheus/prometheus
WORKDIR /prometheus/prometheus
COPY sample_limit_annotation.patch ./
COPY successfully_sent_metric.patch ./

RUN patch -p1 < sample_limit_annotation.patch && \
  patch -p1 < successfully_sent_metric.patch && \
  make build

FROM $BASE_ALPINE
COPY --from=artifact /prometheus/prometheus/prometheus                             /bin/prometheus
COPY --from=artifact /prometheus/prometheus/promtool                               /bin/promtool
COPY --from=artifact /prometheus/prometheus/documentation/examples/prometheus.yml  /etc/prometheus/prometheus.yml
COPY --from=artifact /prometheus/prometheus/console_libraries/                     /usr/share/prometheus/console_libraries/
COPY --from=artifact /prometheus/prometheus/consoles/                              /usr/share/prometheus/consoles/

RUN apk --no-cache add curl
RUN ln -s /usr/share/prometheus/console_libraries /usr/share/prometheus/consoles/ /etc/prometheus/
RUN mkdir -p /prometheus && \
    chown -R 65534:2000 etc/prometheus /prometheus

USER       65534
EXPOSE     9090
VOLUME     [ "/prometheus" ]
WORKDIR    /prometheus
ENTRYPOINT [ "/bin/prometheus" ]
CMD        [ "--config.file=/etc/prometheus/prometheus.yml", \
             "--storage.tsdb.path=/prometheus", \
             "--web.console.libraries=/usr/share/prometheus/console_libraries", \
             "--web.console.templates=/usr/share/prometheus/consoles" ]
