FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY api ./api
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/configaudit ./cmd/configaudit

FROM alpine:3.20

RUN adduser -D -H -u 10001 appuser

WORKDIR /work

COPY --from=builder /out/configaudit /usr/local/bin/configaudit
COPY testdata ./testdata

EXPOSE 8080 9090

USER appuser

ENTRYPOINT ["configaudit"]
