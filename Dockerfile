FROM theshamuel/baseimg-go-build:1.15.1 as builder

ARG VER
ARG SKIP_TESTS
ENV GOFLAGS="-mod=vendor"

RUN apk --no-cache add tzdata zip ca-certificates git

ADD .. /build/datga-sec-exchanger
ADD backend/.golangci.yml /build/datga-sec-exchanger/app/.golangci.yml
WORKDIR /build/datga-sec-exchanger

#test
RUN \
    if [ -z "$SKIP_TESTS" ] ; then \
        go test -timeout=30s ./...; \
    else echo "[WARN] Skip tests" ; fi

#linter GolangCI
RUN \
    if [ -z "$SKIP_TESTS" ] ; then \
        golangci-lint run --skip-dirs vendor --config .golangci.yml ./...; \
    else echo "[WARN] Skip GolangCI linter" ; fi

RUN \
    version="test"; \
    if [ -n "$VER" ] ; then \
    version=${VER}_$(date +%Y%m%d-%H:%M:%S); fi; \
    echo "version=$version"; \
    go build -o datga-sec-exchanger -ldflags "-X main.version=${version} -s -w" ./app

FROM theshamuel/baseimg-go-app:1.0-alpine3.10

WORKDIR /srv
COPY --from=builder /build/datga-sec-exchanger/datga-sec-exchanger /srv/datga-sec-exchanger

RUN chown -R appuser:appuser /srv
USER appuser

CMD [ "/srv/datga-sec-exchanger", "server" ]