package coreum

import (
	"context"

	"github.com/CoreumFoundation/coreum/app"
	coreumClient "github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Client is the interface that expose Coreum client functionality
type Client interface {
	TransferToken(
		ctx context.Context,
		sourcePrivateKey types.Secp256k1PrivateKey,
		amount types.Coin,
		destAddress sdk.AccAddress,
	) (txHash string, err error)
}

// New returns an instance of the Client interface
func New(c coreumClient.Client, network app.Network) Client {
	return client{
		client:  c,
		network: network,
	}
}

type client struct {
	client  coreumClient.Client
	network app.Network
}

func (c client) TransferToken(
	ctx context.Context,
	sourcePrivateKey types.Secp256k1PrivateKey,
	amount types.Coin,
	destAddress sdk.AccAddress,
) (string, error) {
	fromAddress, err := sdk.AccAddressFromBech32(sourcePrivateKey.Address())
	if err != nil {
		return "", err
	}
	msg := banktypes.NewMsgSend(fromAddress, destAddress, sdk.Coins{{
		Denom:  amount.Denom,
		Amount: sdk.NewIntFromBigInt(amount.Amount),
	}})

	signedTx, err := c.client.Sign(
		ctx,
		tx.BaseInput{
			Signer:   types.Wallet{Key: sourcePrivateKey},
			GasLimit: c.network.DeterministicGas().BankSend,
			GasPrice: types.Coin{Amount: c.network.InitialGasPrice(), Denom: c.network.TokenSymbol()},
		},
		msg,
	)
	if err != nil {
		return "", err
	}

	result, err := c.client.Broadcast(ctx, c.client.Encode(signedTx))
	if err != nil {
		return "", err
	}

	return result.TxHash, nil
}
