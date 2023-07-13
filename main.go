package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"github.com/CoreumFoundation/coreum/pkg/client"
	coreumconfig "github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/faucet/app"
	"github.com/CoreumFoundation/faucet/client/coreum"
	"github.com/CoreumFoundation/faucet/http"
	"github.com/CoreumFoundation/faucet/pkg/config"
	"github.com/CoreumFoundation/faucet/pkg/limiter"
	"github.com/CoreumFoundation/faucet/pkg/logger"
	"github.com/CoreumFoundation/faucet/pkg/signal"
)

const (
	flagChainID           = "chain-id"
	flagNode              = "node"
	flagAddress           = "address"
	flagMonitoringAddress = "monitoring-address"
	flagTransferAmount    = "transfer-amount"
	flagMnemonicFilePath  = "key-path-mnemonic"
	flagIPRateLimit       = "ip-rate-limit"
)

func main() {
	ctx, log, cfg := setup()
	if cfg.help {
		return
	}

	log.Info("Starting faucet",
		zap.String("address", cfg.address),
		zap.String("chainID", cfg.chainID),
		zap.String("mnemonicFilePath", cfg.mnemonicFilePath),
		zap.String("node", cfg.node))

	network, err := coreumconfig.NetworkConfigByChainID(constant.ChainID(cfg.chainID))
	if err != nil {
		log.Fatal(
			"Unable to get network config for chain-id",
			zap.Error(err),
			zap.String("chain-id", cfg.chainID),
		)
	}

	if network.ChainID() == constant.ChainIDMain {
		log.Fatal("running a faucet against mainnet is not allowed")
	}

	network.SetSDKConfig()

	transferAmount := sdk.Coin{
		Amount: sdk.NewInt(cfg.transferAmount),
		Denom:  network.Denom(),
	}

	kr, addresses, err := newKeyringFromFile(cfg.mnemonicFilePath)
	if err != nil {
		log.Fatal(
			"Unable to create keyring",
			zap.Error(err),
			zap.String("chain-id", cfg.chainID),
		)
	}

	var addrList []string
	for _, addr := range addresses {
		addrList = append(addrList, addr.String())
	}
	log.Info("funding account addresses", zap.Strings("addresses", addrList))

	clientCtx := client.NewContext(client.DefaultContextConfig(), config.NewModuleManager()).
		WithChainID(string(network.ChainID())).
		WithBroadcastMode(flags.BroadcastBlock)

	clientCtx = addClient(cfg, log, clientCtx)

	txf := client.Factory{}.
		WithTxConfig(clientCtx.TxConfig()).
		WithKeybase(kr).
		WithChainID(string(network.ChainID())).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)
	cl := coreum.New(
		network,
		clientCtx,
		txf,
	)

	err = parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		batcher := coreum.NewBatcher(cl, addresses, 10)
		application := app.New(batcher, network, transferAmount)
		ipLimiter := limiter.NewWeightedWindowLimiter(cfg.ipRateLimit.howMany, cfg.ipRateLimit.period)
		//nolint:contextcheck
		server := http.New(application, ipLimiter, log)

		spawn("batcher", parallel.Fail, batcher.Run)
		spawn("limiterCleanup", parallel.Fail, ipLimiter.Run)
		spawn("server", parallel.Fail, func(ctx context.Context) error {
			return server.ListenAndServe(ctx, cfg.address)
		})
		spawn("monitoring", parallel.Fail, func(ctx context.Context) error {
			return app.RunMonitoring(ctx, cfg.monitoringAddress, clientCtx, addresses, network.Denom())
		})

		return nil
	})

	if err != nil {
		log.Fatal("Error on ListenAndServe", zap.Error(err))
	}
}

func addClient(cfg cfg, log *zap.Logger, clientCtx client.Context) client.Context {
	nodeURL, err := url.Parse(cfg.node)
	if err != nil {
		log.Fatal(
			"Unable to decode node url",
			zap.Error(err),
			zap.String("url", cfg.node),
		)
	}

	// tls grpc
	if nodeURL.Scheme == "https" {
		grpcClient, err := grpc.Dial(nodeURL.Host, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
		if err != nil {
			panic(err)
		}

		return clientCtx.WithGRPCClient(grpcClient)
	}

	// no-tls grpc
	host := nodeURL.Host
	// it is possible that protocol wasn't provided, in such scenario we use the node as a host to dial
	if host == "" {
		host = cfg.node
	}
	grpcClient, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(
			"Unable to create cosmos grpc client",
			zap.Error(err),
		)
	}

	return clientCtx.WithGRPCClient(grpcClient)
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
	chainID           string
	node              string
	mnemonicFilePath  string
	address           string
	monitoringAddress string
	transferAmount    int64
	ipRateLimit       rateLimit
	help              bool
}

func parseRateLimit(limit string) (rateLimit, error) {
	parts := strings.Split(limit, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return rateLimit{}, errors.New("invalid format")
	}
	howMany, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return rateLimit{}, errors.Wrap(err, "invalid format")
	}
	period, err := time.ParseDuration(parts[1])
	if err != nil {
		return rateLimit{}, errors.Wrap(err, "invalid format")
	}

	return rateLimit{
		howMany: uint64(howMany),
		period:  period,
	}, nil
}

type rateLimit struct {
	howMany uint64
	period  time.Duration
}

func getConfig(log *zap.Logger, flagSet *pflag.FlagSet) cfg {
	var conf cfg
	var ipRateLimit string

	flagSet.StringVar(&conf.chainID, flagChainID, string(constant.ChainIDDev), "The network chain ID")
	flagSet.StringVar(&conf.node, flagNode, "localhost:9090", "<host>:<port> to Tendermint GRPC endpoint for this chain")
	flagSet.StringVar(&conf.address, flagAddress, ":8090", "<host>:<port> address to start listening for http requests")
	flagSet.StringVar(&conf.monitoringAddress, flagMonitoringAddress, ":8091", "<host>:<port> address to expose metrics to")
	flagSet.Int64Var(&conf.transferAmount, flagTransferAmount, 100000000, "how much to transfer in each request")
	flagSet.StringVar(&conf.mnemonicFilePath, flagMnemonicFilePath, "mnemonic.txt", "path to file containing mnemonic for private keys, each line containing one mnemonic")
	flagSet.StringVar(&ipRateLimit, flagIPRateLimit, "2/1h", "limit of requests per IP in the format <num-of-req>/<period>")
	flagSet.BoolVarP(&conf.help, "help", "h", false, "prints help")
	_ = flagSet.Parse(os.Args[1:])

	var err error
	conf.ipRateLimit, err = parseRateLimit(ipRateLimit)
	if err != nil {
		log.Fatal("Error parsing IP rate limit", zap.Error(err))
	}

	err = config.WithEnv(flagSet, "")
	if err != nil {
		log.Fatal("Error getting config", zap.Error(err))
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
		info, err := tempKr.NewAccount("temp", mnemonic, "", sdk.GetConfig().GetFullBIP44Path(), hd.Secp256k1)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "unable to parse mnemonic key")
		}
		address := info.GetAddress()
		addresses = append(addresses, address)
		_, err = kr.NewAccount(address.String(), mnemonic, "", sdk.GetConfig().GetFullBIP44Path(), hd.Secp256k1)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "unable to parse mnemonic key")
		}
	}

	if len(addresses) == 0 {
		return nil, nil, errors.New("could not parse any mnemonic")
	}

	return kr, addresses, nil
}
