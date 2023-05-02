Faucet for the Coreum blockchain

## Prerequisites
To use `faucet` you need:
- `go 1.19` or newer
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

### --key-path-mnemonic

path to file containing mnemonics of private keys, each line must contain one mnemonic (default "mnemonic.txt")

### --node (default "localhost:9090")
<host>:<port> to Tendermint GRPC interface for this chain

### --log-format

Format of log output: console | json (default "json")

### --transfer-amount int

How much to transfer in each request (default 100000000)

## API reference

### `fund`

Funds to the specified address.

```shell script
curl --location 'http://localhost:8090/api/faucet/v1/fund' \
--header 'Content-Type: application/json' \
--data '{
    "address": "devcore19tmtuldmuamlzuv4xx704me7ns7yn07crdc4r3"
}'
```

```json
{
    "txHash":"E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"
}
```

### `gen-funded`

Generate funded account.

```shell script
curl --location 'http://localhost:8090/api/faucet/v1/gen-funded' \
--header 'Content-Type: application/json' \
--data '{
    "address": "devcore175m7gdsh9m0rm08a0w3eccz9r895t9jex0abcd"
}'
```

```json
{
  "txHash": "D039E2E8F4318A3C03F2B51D74E8E8CA8CFAFBC02B67E0A9716340B874347778",
  "mnemonic": "day oyster today mechanic soup happy judge matter output asset tiny bundle galaxy theory witness act adapt company thought shock pole explain orchard surround",
  "address": "devcore1lj597uzf689t0tpfxurhra9q9vtkxldezmtvwh"
}
```
