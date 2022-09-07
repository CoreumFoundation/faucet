package coreum

import (
	"context"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
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
	fromAddress := sdk.AccAddress(sourcePrivateKey.PubKey().Address()).String()

	log := logger.Get(ctx).With(zap.String("from", fromAddress), zap.String("to", destAddress.String()))
	log.Info("Sending tokens")

	msg := banktypes.MsgSend{
		FromAddress: fromAddress,
		ToAddress:   destAddress.String(),
		Amount:      []sdk.Coin{amount},
	}
	signedTx, err := c.client.Sign(
		ctx,
		tx.BaseInput{
			Signer:   types.Wallet{Key: sourcePrivateKey.Key},
			GasLimit: c.network.DeterministicGas().BankSend,
			GasPrice: types.Coin{Amount: c.network.FeeModel().Params().InitialGasPrice.BigInt(), Denom: c.network.TokenSymbol()},
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

	log.Info("Tokens sent")
	return result.TxHash, nil
}
