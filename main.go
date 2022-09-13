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
	flagChainID             = "chain-id"
	flagNode                = "node"
	flagAddress             = "address"
	flagTransferAmount      = "transfer-amount"
	flagPrivKeyFile         = "key-path"
	flagPrivKeyFileMnemonic = "key-path-mnemonic"
)

func main() {
	ctx, log, cfg := setup()
	if cfg.help {
		return
	}

	log.Info("Starting faucet",
		zap.String("address", cfg.address),
		zap.String("chainID", cfg.chainID),
		zap.String("privateKeysFile", cfg.privateKeysFile),
		zap.String("node", cfg.node))

	network, err := coreumapp.NetworkByChainID(coreumapp.ChainID(cfg.chainID))
	if err != nil {
		log.Fatal(
			"Unable to get network config for chain-id",
			zap.Error(err),
			zap.String("chain-id", cfg.chainID),
		)
	}

	if network.ChainID() == coreumapp.Mainnet {
		log.Fatal("running a faucet against mainnet is not allowed")
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

	addresses := make([]string, 0, len(privateKeys))
	for _, privKey := range privateKeys {
		addresses = append(addresses, sdk.AccAddress(privKey.PubKey().Address()).String())
	}
	log.Info("Funding addresses", zap.Strings("addresses", addresses))

	application := app.New(cl, network, transferAmount, privateKeys[0])
	server := http.New(application, log)
	err = server.ListenAndServe(ctx, cfg.address)
	if err != nil {
		log.Fatal("Error on ListenAndServe", zap.Error(err))
	}
}

func setup() (context.Context, *zap.Logger, cfg) {
	loggerConfig, loggerFlagRegistry := logger.ConfigureWithCLI(logger.ServiceDefaultConfig)
	log := logger.New(loggerConfig)
	ctx := logger.WithLogger(context.Background(), log)
	ctx = signal.TerminateSignal(ctx)

	flagSet := pflag.NewFlagSet("faucet", pflag.ExitOnError)
	loggerFlagRegistry(flagSet)
	cfg := getConfig(log, flagSet)
	if cfg.help {
		flagSet.PrintDefaults()
	}
	return ctx, log, cfg
}

type cfg struct {
	chainID                 string
	node                    string
	privateKeysFile         string
	privateKeysFileMnemonic string
	address                 string
	transferAmount          int64
	help                    bool
}

func getConfig(log *zap.Logger, flagSet *pflag.FlagSet) cfg {
	var conf cfg
	flagSet.StringVar(&conf.chainID, flagChainID, string(coreumapp.Devnet), "The network chain ID")
	flagSet.StringVar(&conf.node, flagNode, "localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
	flagSet.StringVar(&conf.address, flagAddress, ":8090", "<host>:<port> address to start listening for http requests")
	flagSet.Int64Var(&conf.transferAmount, flagTransferAmount, 1000000, "how much to transfer in each request")
	flagSet.StringVar(&conf.privateKeysFile, flagPrivKeyFile, "private_keys_unarmored_hex.txt", "path to file containing hex encoded unarmored private keys, each line must contain one private key")
	flagSet.StringVar(&conf.privateKeysFileMnemonic, flagPrivKeyFileMnemonic, "private_keys_mnemonic.txt", "path to file containing mnemonic for private keys, each line containing one mnemonic")
	flagSet.BoolVarP(&conf.help, "help", "h", false, "prints help")
	_ = flagSet.Parse(os.Args[1:])
	err := config.WithEnv(flagSet, "")
	if err != nil {
		log.Fatal("error getting config", zap.Error(err))
	}
	return conf
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
