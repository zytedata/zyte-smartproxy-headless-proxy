###############################################################################
# BUILD STAGE

FROM golang:1.14-alpine AS build-env

WORKDIR /app

RUN set -x \
  && apk --no-cache --update add \
    bash \
    ca-certificates \
    git \
    make \
    upx

ADD https://docs.zyte.com/_downloads/1c2c73cfb44b56a4756294649a2e0f2d/zyte-smartproxy-ca.crt /usr/local/share/ca-certificates/crawlera-ca.crt
RUN set -x \
  && sha1sum /usr/local/share/ca-certificates/crawlera-ca.crt | cut -f1 -d' ' | \
    while read -r sum _; do \
      if [ "${sum}" != "5798e59f6f7ecad3c0e1284f42b07dcaa63fbd37" ]; then \
        echo "Incorrect CA certificate checksum ${sum}"; \
        exit 1; \
    fi; done \
  && update-ca-certificates

COPY . /app

RUN set -x \
  && make static

ARG upx=
RUN set -x \
  && if [ -n "$upx" ]; then \
    upx --ultra-brute -qq ./crawlera-headless-proxy; \
  fi


###############################################################################
# PACKAGE STAGE

FROM scratch

ENTRYPOINT ["/crawlera-headless-proxy"]
ENV CRAWLERA_HEADLESS_BINDIP=0.0.0.0 \
    CRAWLERA_HEADLESS_BINDPORT=3128 \
    CRAWLERA_HEADLESS_PROXYAPIIP=0.0.0.0 \
    CRAWLERA_HEADLESS_PROXYAPIPORT=3130 \
    CRAWLERA_HEADLESS_CONFIG=/config.toml
EXPOSE 3128 3130

COPY --from=build-env \
  /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-env \
  /app/crawlera-headless-proxy \
  /app/config.toml \
  /
