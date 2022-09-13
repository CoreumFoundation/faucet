Faucet for the Coreum blockchain

## Prerequisites
To use `crust` you need:
- `go 1.17` or newer
- `tmux`
- `docker`

## Executing `faucet`

`faucet` is an http server that can send funds to a given address.

## Building

Build all the required binaries by running:

```
$ go build -o faucet
```

## Flags

All the flags are optional. Execute

```
$ faucet --help
```

to see what the default values are.

### --address

<host>:<port> address to start listening for http requests (default ":8090")

### --chain-id

The network chain ID (default "coreum-devnet-1")

### --key-path

path to file containing hex encoded unarmored private keys, each line must contain one private key (default "private_keys_unarmored_hex.txt")

### --node
<host>:<port> to Tendermint RPC interface for this chain (default "localhost:26657")

### --log-format

Format of log output: console | json (default "json")

### --transfer-amount int

How much to transfer in each request (default 1000000)

## API reference

### `send_money`

Returns XRP and SOLO balance of an address.

```shell script
curl -H "Content-Type: application/json" -X POST \
     -d '{"address":"devcore175m7gdsh9m0rm08a0w3eccz9r895t9jex0abcd"}' \
     http://localhost:8090/api/faucet/v1/send-money
```

```json
{
    "txHash":"E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"
}
```