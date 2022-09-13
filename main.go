package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/ignite/cli/ignite/pkg/cosmoscmd"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	coreumapp "github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/tx"
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

	transferAmount := sdk.Coin{
		Amount: sdk.NewInt(cfg.transferAmount),
		Denom:  network.TokenSymbol(),
	}

	kr, addresses, err := newKeyringFromFile(cfg.privateKeysFile)
	if err != nil {
		log.Fatal(
			"Unable to create keyring",
			zap.Error(err),
			zap.String("chain-id", cfg.chainID),
		)
	}
	for _, addr := range addresses {
		fmt.Println(addr.String())
	}

	rpcClient, err := client.NewClientFromNode(cfg.node)
	if err != nil {
		log.Fatal(
			"Unable to create cosmos rpc client",
			zap.Error(err),
		)
	}

	mbm := module.NewBasicManager(
		bank.AppModuleBasic{},
		auth.AppModuleBasic{},
	)
	encodingConfig := cosmoscmd.MakeEncodingConfig(mbm)
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithNodeURI(cfg.node).
		WithChainID(string(network.ChainID())).
		WithBroadcastMode(flags.BroadcastBlock).
		WithClient(rpcClient)

	txf := tx.Factory{}.
		WithTxConfig(clientCtx.TxConfig).
		WithKeybase(kr).
		WithChainID(string(network.ChainID())).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT) //nolint:nosnakecase
	cl := coreum.New(
		network,
		clientCtx,
		txf,
	)

	application := app.New(ctx, log, cl, network, transferAmount, addresses)
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
	flagSet.StringVar(&conf.node, flagNode, "tcp://localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
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

func newKeyringFromFile(path string) (keyring.Keyring, []sdk.AccAddress, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "unable to open file at %s", path)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	kr := keyring.NewInMemory()
	var addresses []sdk.AccAddress
	for scanner.Scan() {
		mnemonic := scanner.Text()
		tempKr := keyring.NewInMemory()
		info, err := tempKr.NewAccount("temp", mnemonic, "", "", hd.Secp256k1)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "unable to parse mnemonic key")
		}
		address := info.GetAddress()
		addresses = append(addresses, address)
		_, _ = kr.NewAccount(address.String(), mnemonic, "", "", hd.Secp256k1)
	}

	if len(addresses) == 0 {
		return nil, nil, errors.New("could not parse any funding private keys")
	}

	return kr, addresses, nil
}
