package integrationtests

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/app"
	h "github.com/CoreumFoundation/faucet/http"
)

type testConfig struct {
	nodeURI        string
	faucetURI      string
	clientCtx      client.Context
	transferAmount string
	network        app.Network
}

var cfg testConfig

func TestMain(m *testing.M) {
	flag.StringVar(&cfg.nodeURI, "nodeURI", "localhost:26656", "Address of cored node started by znet")
	flag.StringVar(&cfg.faucetURI, "faucetURI", "http://localhost:8090", "Address of the faucet")
	flag.StringVar(&cfg.transferAmount, "transfer-amount", "1000000", "Amount transferred by faucet in each request")
	flag.Parse()
	rpcClient, err := client.NewClientFromNode("tcp://" + cfg.nodeURI)
	must.OK(err)
	cfg.network, _ = app.NetworkByChainID(app.Devnet)
	cfg.network.SetupPrefixes()
	cfg.clientCtx = app.
		NewDefaultClientContext().
		WithChainID(string(cfg.network.ChainID())).
		WithClient(rpcClient)

	m.Run()
}

func TestTransferRequest(t *testing.T) {
	ctx := context.Background()
	address := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	// request fund
	clientCtx := cfg.clientCtx
	txHash, err := requestFunds(ctx, address)
	require.NoError(t, err)
	require.Len(t, txHash, 64)

	// query funds
	bankQueryClient := banktypes.NewQueryClient(clientCtx)
	resp, err := bankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: address})
	require.NoError(t, err)

	// make assertions
	assert.EqualValues(t, cfg.transferAmount, resp.Balances.AmountOf(cfg.network.TokenSymbol()).String())
}

func requestFunds(ctx context.Context, address string) (string, error) {
	url := cfg.faucetURI + "/api/faucet/v1/send-money"
	method := "POST"

	sendMoneyReq := h.SendMoneyRequest{
		Address: address,
	}
	payloadBuffer := bytes.NewBuffer(nil)
	err := json.NewEncoder(payloadBuffer).Encode(sendMoneyReq)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, url, payloadBuffer)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var sendMoneyResponse h.SendMoneyResponse

	err = json.Unmarshal(body, &sendMoneyResponse)
	if err != nil {
		return "", err
	}

	return sendMoneyResponse.TxHash, nil
}
