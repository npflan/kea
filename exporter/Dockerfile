FROM golang:1.13.0-alpine3.10 AS builder

WORKDIR /app
RUN apk add --no-cache --virtual .go-deps git gcc musl-dev openssl pkgconf
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY exporter.go .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' exporter.go

ENTRYPOINT ["/app/exporter"]

# Build the smallest image possible
FROM scratch AS runner
COPY --from=builder /app/exporter /bin/exporter
ENTRYPOINT ["/bin/exporter"]
