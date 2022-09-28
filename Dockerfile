# Build Stage
FROM golang:1.19-alpine3.16 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /bin/faucet -tags=muslc -trimpath -ldflags="-w -s"

# Deploy Stage
FROM alpine:3.16

WORKDIR /

VOLUME mnemonic.txt

COPY --from=builder /bin/faucet /bin/faucet

ENV KEY_PATH=mnemonic.txt

EXPOSE 8090

ENTRYPOINT ["/bin/faucet"]
