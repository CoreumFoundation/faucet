package coreum

import (
	"context"
	"fmt"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/faucet/pkg/logger"
)

// New returns an instance of the Client interface
func New(network app.Network, clientCtx cosmosclient.Context, txf tx.Factory) Client {
	return Client{
		network:   network,
		clientCtx: clientCtx,
		txf:       txf,
	}
}

// Client is used to communicate with coreum blockchain
type Client struct {
	clientCtx cosmosclient.Context
	network   app.Network
	txf       tx.Factory
}

func zapStringers[T fmt.Stringer](key string, list []T) zap.Field {
	var strList []string
	for _, l := range list {
		strList = append(strList, l.String())
	}
	return zap.Strings(key, strList)
}

// TransferToken transfers amount to a list of destination addresses in single tx
func (c Client) TransferToken(
	ctx context.Context,
	fromAddress sdk.AccAddress,
	amount sdk.Coin,
	destAddresses ...sdk.AccAddress,
) (string, error) {
	var msgs []sdk.Msg
	log := logger.Get(ctx).With(zap.Stringer("from_address", fromAddress), zapStringers("to_addresses", destAddresses))
	log.Info("Sending tokens")
	for _, destAddress := range destAddresses {
		msg := &banktypes.MsgSend{
			FromAddress: fromAddress.String(),
			ToAddress:   destAddress.String(),
			Amount:      []sdk.Coin{amount},
		}
		msgs = append(msgs, msg)
	}
	clientCtx := c.clientCtx.
		WithFromName(fromAddress.String()).
		WithFrom(fromAddress.String()).
		WithFromAddress(fromAddress)

	txf := c.txf.
		WithGas(c.network.DeterministicGas().BankSend * uint64(len(msgs))).
		WithGasPrices(c.network.FeeModel().Params().InitialGasPrice.String() + c.network.TokenSymbol())
	result, err := tx.BroadcastTx(ctx, clientCtx, txf, msgs...)
	if err != nil {
		return "", err
	}

	log.Info("Tokens sent")
	return result.TxHash, nil
}
