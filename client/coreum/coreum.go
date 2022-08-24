package coreum

import (
	"context"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/app"
	coreumClient "github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// Client is the interface that expose Coreum client functionality
type Client interface {
	TransferToken(
		ctx context.Context,
		sourcePrivateKey secp256k1.PrivKey,
		amount sdk.Coin,
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
	sourcePrivateKey secp256k1.PrivKey,
	amount sdk.Coin,
	destAddress sdk.AccAddress,
) (string, error) {
	sourcePrivateKey.PubKey().Address()
	fromAddress := sdk.AccAddress(sourcePrivateKey.PubKey().Bytes()).String()
	msg := banktypes.MsgSend{
		FromAddress: fromAddress,
		ToAddress:   destAddress.String(),
		Amount:      []sdk.Coin{amount},
	}
	signedTx, err := c.client.Sign(
		ctx,
		tx.BaseInput{
			Signer:   types.Wallet{Key: types.Secp256k1PrivateKey(sourcePrivateKey.Key)},
			GasLimit: c.network.DeterministicGas().BankSend,
			GasPrice: types.Coin{Amount: c.network.InitialGasPrice(), Denom: c.network.TokenSymbol()},
		},
		&msg,
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
