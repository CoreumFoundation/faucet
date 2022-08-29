# Build Stage
FROM golang:1.19-alpine3.16 AS builder

RUN apk add --no-cache gcc libc-dev wget curl

# See https://github.com/CosmWasm/wasmvm/releases
RUN curl -L https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/libwasmvm_muslc.aarch64.a -o /lib/libwasmvm_muslc.aarch64.a
RUN curl -L https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/libwasmvm_muslc.x86_64.a -o /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 7d2239e9f25e96d0d4daba982ce92367aacf0cbd95d2facb8442268f2b1cc1fc
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep f6282df732a13dec836cda1f399dd874b1e3163504dbd9607c6af915b2740479

# Copy the library you want to the final location that will be found by the linker flag `-lwasmvm_muslc`
ARG arch=x86_64
RUN cp /lib/libwasmvm_muslc.${arch}.a /lib/libwasmvm_muslc.a

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o /bin/faucet -tags=muslc -trimpath -ldflags="-w -s -extldflags=-static"

# Deploy Stage
FROM alpine:3.16

WORKDIR /

VOLUME /data

COPY --from=builder /bin/faucet /bin/faucet

ENV KEY_PATH=/data/private_keys_unarmored_hex.txt

EXPOSE 8090

ENTRYPOINT ["/bin/faucet"]
