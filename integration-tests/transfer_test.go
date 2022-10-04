//go:build integration
// +build integration

package integrationtests

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"go.uber.org/zap/zaptest"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	coreumconfig "github.com/CoreumFoundation/coreum/pkg/config"
	coreumtx "github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/faucet/http"
	"github.com/CoreumFoundation/faucet/pkg/config"
)

type testConfig struct {
	coredAddress   string
	faucetAddress  string
	clientCtx      coreumtx.ClientContext
	transferAmount string
	network        coreumconfig.Network
}

var cfg testConfig

func TestMain(m *testing.M) {
	flag.StringVar(&cfg.coredAddress, "cored-address", "tcp://localhost:26657", "Address of cored node started by znet")
	flag.StringVar(&cfg.faucetAddress, "faucet-address", "http://localhost:8090", "Address of the faucet")
	flag.StringVar(&cfg.transferAmount, "transfer-amount", "1000000", "Amount transferred by faucet in each request")
	flag.Parse()
	rpcClient, err := client.NewClientFromNode(cfg.coredAddress)
	must.OK(err)
	cfg.network, _ = coreumconfig.NetworkByChainID(coreumconfig.Devnet)
	cfg.network.SetSDKConfig()
	cfg.clientCtx = coreumtx.NewClientContext(config.NewModuleManager()).
		WithChainID(string(cfg.network.ChainID())).
		WithClient(rpcClient).
		WithBroadcastMode(flags.BroadcastBlock)

	m.Run()
}

func TestTransferRequest(t *testing.T) {
	log := zaptest.NewLogger(t)
	ctx := logger.WithLogger(context.Background(), log)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(cancel)
	address := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	// request fund
	clientCtx := cfg.clientCtx
	txHash, err := requestFunds(ctx, address)
	require.NoError(t, err)
	require.Len(t, txHash, 64)

	err = waitForTxInclusionAndSync(ctx, clientCtx, txHash)
	require.NoError(t, err)

	// query funds
	bankQueryClient := banktypes.NewQueryClient(clientCtx)
	resp, err := bankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: address})
	require.NoError(t, err)

	// make assertions
	assert.EqualValues(t, cfg.transferAmount, resp.Balances.AmountOf(cfg.network.TokenSymbol()).String())
}

// waitForTxInclusionAndSync waits for one block, so all nodes are synced
// it is possible that the tx and query requests go to different nodes and although
// the tx is complete, the query will go to a different node which is not yet synced up
// and does not have the transaction included in a block and its state will be different.
// By waiting for one block we make sure that all nodes are synced on the previous block.
func waitForTxInclusionAndSync(ctx context.Context, clientCtx coreumtx.ClientContext, txHash string) error {
	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return errors.WithStack(err)
	}
	var resultTx *ctypes.ResultTx
	err = retry.Do(ctx, 200*time.Millisecond, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		resultTx, err = clientCtx.Client().Tx(requestCtx, txHashBytes, false)
		if err != nil {
			return retry.Retryable(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = retry.Do(ctx, 200*time.Millisecond, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		height := resultTx.Height + 1
		_, err := clientCtx.Client().Block(requestCtx, &height)
		if err != nil {
			return retry.Retryable(err)
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func TestTransferRequestWithGenPrivkey(t *testing.T) {
	log := zaptest.NewLogger(t)
	ctx := logger.WithLogger(context.Background(), log)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(cancel)

	// request fund
	clientCtx := cfg.clientCtx
	response, err := requestFundsWithPrivkey(ctx)
	require.NoError(t, err)
	require.Len(t, response.TxHash, 64)

	err = waitForTxInclusionAndSync(ctx, clientCtx, response.TxHash)
	require.NoError(t, err)

	// query funds
	bankQueryClient := banktypes.NewQueryClient(clientCtx)
	resp, err := bankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: response.Address})
	require.NoError(t, err)

	// make assertions
	assert.EqualValues(t, cfg.transferAmount, resp.Balances.AmountOf(cfg.network.TokenSymbol()).String())
}

func TestTransferRequest_WrongAddress(t *testing.T) {
	ctx := context.Background()
	address := "core1hrlnys435ph2gehthddlg2g2s246my30q0gfs2"

	// request fund
	clientCtx := cfg.clientCtx
	txHash, err := requestFunds(ctx, address)
	assert.Error(t, err)
	assert.Len(t, txHash, 0)

	// query funds
	bankQueryClient := banktypes.NewQueryClient(clientCtx)
	resp, err := bankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: address})
	require.Error(t, err)

	// make assertions
	assert.Nil(t, resp)
}

func requestFunds(ctx context.Context, address string) (string, error) {
	url := cfg.faucetAddress + "/api/faucet/v1/fund"
	method := "POST"

	sendMoneyReq := http.FundRequest{
		Address: address,
	}
	payloadBuffer := bytes.NewBuffer(nil)
	err := json.NewEncoder(payloadBuffer).Encode(sendMoneyReq)
	if err != nil {
		return "", errors.WithStack(err)
	}

	client := &nethttp.Client{}
	req, err := nethttp.NewRequestWithContext(ctx, method, url, payloadBuffer)
	if err != nil {
		return "", errors.WithStack(err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return "", errors.Errorf("non 2xx response, body: %s", body)
	}

	decoder := json.NewDecoder(res.Body)
	var sendMoneyResponse http.FundResponse
	err = decoder.Decode(&sendMoneyResponse)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return sendMoneyResponse.TxHash, nil
}

func requestFundsWithPrivkey(ctx context.Context) (http.GenFundedResponse, error) {
	url := cfg.faucetAddress + "/api/faucet/v1/gen-funded"
	method := "POST"

	client := &nethttp.Client{}
	req, err := nethttp.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return http.GenFundedResponse{}, errors.WithStack(err)
	}

	res, err := client.Do(req)
	if err != nil {
		return http.GenFundedResponse{}, errors.WithStack(err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return http.GenFundedResponse{}, errors.Errorf("non 2xx response, body: %s", body)
	}

	decoder := json.NewDecoder(res.Body)
	var responseStruct http.GenFundedResponse
	err = decoder.Decode(&responseStruct)
	if err != nil {
		return http.GenFundedResponse{}, errors.WithStack(err)
	}

	return responseStruct, nil
}
