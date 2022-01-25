###############################################################################
# BUILD STAGE

FROM golang:1.16-alpine AS build-env

WORKDIR /app

RUN set -x \
  && apk --no-cache --update add \
    bash \
    ca-certificates \
    git \
    make \
    upx

ADD https://docs.zyte.com/_static/zyte-smartproxy-ca.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates

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
