# Build Stage
FROM golang:1.19-alpine3.16 AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/faucet

# Deploy Stage
FROM alpine:3.16

WORKDIR /

COPY --from=builder /bin/faucet /bin/faucet

EXPOSE 8090

ENTRYPOINT ["/bin/faucet"]
