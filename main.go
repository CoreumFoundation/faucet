package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"flag"
	"math/big"
	"os"

	"github.com/CoreumFoundation/faucet/app"
	"github.com/CoreumFoundation/faucet/client/coreum"
	"github.com/CoreumFoundation/faucet/http"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	coreumApp "github.com/CoreumFoundation/coreum/app"
	coreumClient "github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/types"

	"go.uber.org/zap"
)

const (
	flagChainID        = "chain-id"
	flagNode           = "node"
	flagPort           = "port"
	flagTransferAmount = "transfer-amount"
	flagPrivKeyFile    = "key-path"
)

func main() {
	log := logger.New(logger.Config{Format: "console", Verbose: true})
	ctx := context.Background()
	ctx = logger.WithLogger(ctx, log)
	config := processFlags()

	network, err := coreumApp.NetworkByChainID(config.chainID)
	if err != nil {
		log.Fatal(
			"unable to get network config for chain-id",
			zap.Error(err),
			zap.String("chain-id", string(config.chainID)),
		)
	}

	network.SetupPrefixes()
	cl := coreum.New(
		coreumClient.New(network.ChainID(), config.node),
		network,
	)
	transferAmount := types.Coin{
		Amount: big.NewInt(config.transferAmount),
		Denom:  network.TokenSymbol(),
	}
	privateKeys, err := privateKeysFromFile(config.privateKeysFile)
	if err != nil {
		log.Fatal("error parsing private keys from file", zap.Error(err))
	}

	if len(privateKeys) == 0 {
		log.Fatal("private key file is empty", zap.Error(err))
	}

	app := app.New(cl, network, transferAmount, privateKeys[0])
	server := http.New(app)
	err = server.ListenAndServe(ctx, config.port)
	if err != nil {
		log.Fatal("error on ListenAndServe", zap.Error(err))
	}
}

type config struct {
	chainID         coreumApp.ChainID
	node            string
	privateKeysFile string
	port            int
	transferAmount  int64
}

func processFlags() config {
	flagSet := flag.NewFlagSet("faucet", flag.ExitOnError)

	chainID := flagSet.String(flagChainID, string(coreumApp.DefaultChainID), "The network chain ID")
	node := flagSet.String(flagNode, "localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
	port := flagSet.Int(flagPort, 8090, "port to start listening for http requests")
	transferAmount := flagSet.Int64(flagTransferAmount, 1000000, "how much to transfer in each request")
	keyFile := flagSet.String(flagPrivKeyFile, "private_keys_unarmored_hex.txt", "path to file containing hex encoded unarmored private keys, each line must contain one private key")
	//nolint:errcheck // since we have set ExitOnError on flagset, we don't need to check for errors here
	flagSet.Parse(os.Args[1:])
	return config{
		chainID:         coreumApp.ChainID(*chainID),
		node:            *node,
		port:            *port,
		transferAmount:  *transferAmount,
		privateKeysFile: *keyFile,
	}
}

func privateKeysFromFile(path string) ([]types.Secp256k1PrivateKey, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to ope file at %s", path)
	}
	scanner := bufio.NewScanner(file)
	var list []types.Secp256k1PrivateKey
	for scanner.Scan() {
		prvKey, err := hex.DecodeString(scanner.Text())
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse private key")
		}

		list = append(list, prvKey)
	}

	return list, nil
}
