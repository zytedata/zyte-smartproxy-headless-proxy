###############################################################################
# BUILD STAGE

FROM golang:alpine AS build-env

RUN set -x \
  && apk --no-cache --update add \
    bash \
    ca-certificates \
    curl \
    git \
    make \
  && curl -fsL -o /usr/local/share/ca-certificates/crawlera-ca.crt https://doc.scrapinghub.com/_downloads/crawlera-ca.crt \
  && sha1sum /usr/local/share/ca-certificates/crawlera-ca.crt | cut -f1 -d' ' | \
  while read -r sum _; do \
    if [ "${sum}" != "5798e59f6f7ecad3c0e1284f42b07dcaa63fbd37" ]; then \
      echo "Incorrect CA certificate checksum ${sum}"; \
      exit 1; \
  fi; done

ADD . /go/src/bitbucket.org/scrapinghub/crawlera-headless-proxy

RUN set -x \
  && cd /go/src/bitbucket.org/scrapinghub/crawlera-headless-proxy \
  && make clean \
  && make -j 4 static \
  && cp ca.crt /usr/local/share/ca-certificates/own-cert.crt \
  && update-ca-certificates


###############################################################################
# PACKAGE STAGE

FROM scratch

ENTRYPOINT ["/usr/local/bin/crawlera-headless-proxy"]
ENV CRAWLERA_HEADLESS_BINDIP=0.0.0.0 \
    CRAWLERA_HEADLESS_BINDPORT=3128 \
    CRAWLERA_HEADLESS_PROXYAPIIP=0.0.0.0 \
    CRAWLERA_HEADLESS_PROXYAPIPORT=3130 \
    CRAWLERA_HEADLESS_CONFIG=/config.toml
EXPOSE 3128 3130

COPY --from=build-env \
  /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-env \
  /go/src/bitbucket.org/scrapinghub/crawlera-headless-proxy/crawlera-headless-proxy /usr/local/bin/crawlera-headless-proxy
COPY --from=build-env \
  /go/src/bitbucket.org/scrapinghub/crawlera-headless-proxy/config.toml /
