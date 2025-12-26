FROM golang:1.24 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /budget-import ./cmd/budget-import


FROM scratch AS runtime

LABEL org.opencontainers.image.source="https://github.com/markis/budget-importer" \
    org.opencontainers.image.description="Budget importer - imports financial data from SimpleFIN to Google Sheets" \
    org.opencontainers.image.licenses="MIT"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /budget-import /budget-import

USER 65534:65534
ENTRYPOINT ["/budget-import"]
