###############################################################################
# BUILD STAGE

FROM golang:alpine AS build-env

RUN set -x \
  && apk --update add git make bash

ADD . /go/src/github.com/9seconds/crawlera-headless-proxy

RUN set -x \
  && cd /go/src/github.com/9seconds/crawlera-headless-proxy \
  && make clean \
  && make -j 4


###############################################################################
# PACKAGE STAGE

FROM alpine:latest
LABEL maintainer="Sergey Arkhipov <arkhipov@scrapinghub.com>" version="0.0.1"

ENTRYPOINT ["/crawlera-headless-proxy"]
ENV CRAWLERA_HEADLESS_IP=0.0.0.0 CRAWLERA_HEADLESS_PORT=3128
CMD ["-c", "/config.toml"]
EXPOSE 3128

RUN set -x \
  && apk add --no-cache --update ca-certificates curl \
  && curl -fsL -o /usr/local/share/ca-certificates/crawlera-ca.crt https://doc.scrapinghub.com/_downloads/crawlera-ca.crt \
  && sha1sum /usr/local/share/ca-certificates/crawlera-ca.crt | cut -f1 -d' ' | \
  while read -r sum _; do \
    if [ "${sum}" != "5798e59f6f7ecad3c0e1284f42b07dcaa63fbd37" ]; then \
      echo "Incorrect CA certificate checksum ${sum}"; \
      exit 1; \
  fi; done \
  && apk del --purge curl \
  && update-ca-certificates \
  && rm -rf /var/cache/apk/*

COPY --from=build-env \
  /go/src/github.com/9seconds/crawlera-headless-proxy/crawlera-headless-proxy \
  /go/src/github.com/9seconds/crawlera-headless-proxy/config.toml \
  /
