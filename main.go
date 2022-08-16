package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	coreumApp "github.com/CoreumFoundation/coreum/app"
	coreumClient "github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/CoreumFoundation/faucet/app"
	"github.com/CoreumFoundation/faucet/client/coreum"
	"github.com/CoreumFoundation/faucet/http"
	"github.com/CoreumFoundation/faucet/pkg/config"
	"github.com/CoreumFoundation/faucet/pkg/logger"
	"github.com/CoreumFoundation/faucet/pkg/signal"
)

const (
	flagChainID        = "chain-id"
	flagNode           = "node"
	flagAddress        = "address"
	flagTransferAmount = "transfer-amount"
	flagPrivKeyFile    = "key-path"
)

func main() {
	loggerConfig, loggerFlagRegistry := logger.ConfigureWithCLI(logger.ServiceDefaultConfig)
	log := logger.New(loggerConfig)
	ctxQuit := signal.TerminateSignal(
		signal.WithForceTimeout(30*time.Second),
		signal.WithForceOnSecondSignal(true),
	)

	ctx := context.Background()
	ctx = logger.WithLogger(ctx, log)
	flagSet := pflag.NewFlagSet("faucet", pflag.ExitOnError)
	loggerFlagRegistry(flagSet)
	config := getConfig(log, flagSet)

	network, err := coreumApp.NetworkByChainID(config.chainID)
	if err != nil {
		log.Fatal(
			"Unable to get network config for chain-id",
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
		log.Fatal("Error parsing private keys from file", zap.Error(err))
	}

	if len(privateKeys) == 0 {
		log.Fatal("Private key file is empty", zap.Error(err))
	}

	app := app.New(cl, network, transferAmount, privateKeys[0])
	server := http.New(app, log)
	err = server.ListenAndServe(ctx, config.address, ctxQuit.Done())
	if err != nil {
		log.Fatal("Error on ListenAndServe", zap.Error(err))
	}
}

type cfg struct {
	chainID         coreumApp.ChainID
	node            string
	privateKeysFile string
	address         string
	transferAmount  int64
}

func getConfig(log *zap.Logger, flagSet *pflag.FlagSet) cfg {
	chainID := flagSet.String(flagChainID, string(coreumApp.DefaultChainID), "The network chain ID")
	node := flagSet.String(flagNode, "localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
	listeningAddress := flagSet.String(flagAddress, ":8090", "<host>:<port> address to start listening for http requests")
	transferAmount := flagSet.Int64(flagTransferAmount, 1000000, "how much to transfer in each request")
	keyFile := flagSet.String(flagPrivKeyFile, "private_keys_unarmored_hex.txt", "path to file containing hex encoded unarmored private keys, each line must contain one private key")
	_ = flagSet.Parse(os.Args[1:])
	err := config.WithEnv(flagSet, "")
	if err != nil {
		log.Fatal("error getting config", zap.Error(err))
	}
	return cfg{
		chainID:         coreumApp.ChainID(*chainID),
		node:            *node,
		address:         *listeningAddress,
		transferAmount:  *transferAmount,
		privateKeysFile: *keyFile,
	}
}

func privateKeysFromFile(path string) ([]types.Secp256k1PrivateKey, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open file at %s", path)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var list []types.Secp256k1PrivateKey
	for scanner.Scan() {
		privKey, err := hex.DecodeString(scanner.Text())
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse private key")
		}

		list = append(list, privKey)
	}

	return list, nil
}
