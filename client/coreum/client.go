package coreum

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/faucet/pkg/logger"
)

// New returns an instance of the Client interface.
func New(network config.NetworkConfig, clientCtx client.Context, txf client.Factory) Client {
	return Client{
		network:   network,
		clientCtx: clientCtx,
		txf:       txf,
	}
}

// Client is used to communicate with coreum blockchain.
type Client struct {
	clientCtx client.Context
	network   config.NetworkConfig
	txf       tx.Factory
}

type transferRequest struct {
	amount      sdk.Coin
	destAddress sdk.AccAddress
}

// TransferToken transfers amount to a list of destination addresses in single tx.
func (c Client) TransferToken(
	ctx context.Context,
	fromAddress sdk.AccAddress,
	requests ...transferRequest,
) (string, error) {
	toAddressList := []string{}
	for _, rq := range requests {
		toAddressList = append(toAddressList, rq.destAddress.String())
	}
	log := logger.Get(ctx).With(zap.Stringer("fromAddress", fromAddress), zap.Strings("toAddresses", toAddressList))
	log.Info("Sending tokens")

	msg := &banktypes.MsgMultiSend{}
	sum := sdk.NewCoins()
	for _, rq := range requests {
		sum = sum.Add(rq.amount)
		msg.Outputs = append(msg.Outputs, banktypes.Output{
			Address: rq.destAddress.String(),
			Coins:   sdk.NewCoins(rq.amount),
		})
	}
	msg.Inputs = []banktypes.Input{{
		Address: fromAddress.String(),
		Coins:   sum,
	}}
	clientCtx := c.clientCtx.
		WithFromName(fromAddress.String()).
		WithFromAddress(fromAddress)

	result, err := client.BroadcastTx(ctx, clientCtx, c.txf.WithSimulateAndExecute(true), msg)
	if err != nil {
		return "", err
	}

	log.Info("Tokens sent")
	return result.TxHash, nil
}
