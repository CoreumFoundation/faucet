package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	coreumapp "github.com/CoreumFoundation/coreum/app"
	coreumclient "github.com/CoreumFoundation/coreum/pkg/client"
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
	ctx := signal.TerminateSignal(context.Background())

	ctx = logger.WithLogger(ctx, log)
	flagSet := pflag.NewFlagSet("faucet", pflag.ExitOnError)
	loggerFlagRegistry(flagSet)
	cfg := getConfig(log, flagSet)

	network, err := coreumapp.NetworkByChainID(cfg.chainID)
	if err != nil {
		log.Fatal(
			"Unable to get network config for chain-id",
			zap.Error(err),
			zap.String("chain-id", string(cfg.chainID)),
		)
	}

	network.SetupPrefixes()
	cl := coreum.New(
		coreumclient.New(network.ChainID(), cfg.node),
		network,
	)
	transferAmount := sdk.Coin{
		Amount: sdk.NewInt(cfg.transferAmount),
		Denom:  network.TokenSymbol(),
	}
	privateKeys, err := privateKeysFromFile(cfg.privateKeysFile)
	if err != nil {
		log.Fatal("Error parsing private keys from file", zap.Error(err))
	}

	if len(privateKeys) == 0 {
		log.Fatal("Private key file is empty", zap.Error(err))
	}

	application := app.New(cl, network, transferAmount, privateKeys[0])
	server := http.New(application, log)
	err = server.ListenAndServe(ctx, cfg.address)
	if err != nil {
		log.Fatal("Error on ListenAndServe", zap.Error(err))
	}
}

type cfg struct {
	chainID         coreumapp.ChainID
	node            string
	privateKeysFile string
	address         string
	transferAmount  int64
}

func getConfig(log *zap.Logger, flagSet *pflag.FlagSet) cfg {
	chainID := flagSet.String(flagChainID, string(coreumapp.DefaultChainID), "The network chain ID")
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
		chainID:         coreumapp.ChainID(*chainID),
		node:            *node,
		address:         *listeningAddress,
		transferAmount:  *transferAmount,
		privateKeysFile: *keyFile,
	}
}

func privateKeysFromFile(path string) ([]secp256k1.PrivKey, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open file at %s", path)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var list []secp256k1.PrivKey
	for scanner.Scan() {
		privKey, err := hex.DecodeString(scanner.Text())
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse private key")
		}

		list = append(list, secp256k1.PrivKey{Key: privKey})
	}

	return list, nil
}
