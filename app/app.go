package app

import (
	"context"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/faucet/client/coreum"
)

// App implements core functionality
type App struct {
	fundsPrivateKey secp256k1.PrivKey
	client          coreum.Client
	transferAmount  sdk.Coin
	network         app.Network
}

// New returns a new instance of the App
func New(
	client coreum.Client,
	network app.Network,
	transferAmount sdk.Coin,
	fundsPrivateKey secp256k1.PrivKey,
) App {
	return App{
		client:          client,
		fundsPrivateKey: fundsPrivateKey,
		network:         network,
		transferAmount:  transferAmount,
	}
}

// GiveFunds gives funds to people asking for it
func (a App) GiveFunds(ctx context.Context, address string) (string, error) {
	prefix, sdkAddr, err := parseAddress(address)
	if err != nil {
		return "", errors.Wrapf(ErrInvalidAddressFormat, "err:%s", err)
	}

	if prefix != a.network.AddressPrefix() {
		return "", errors.Wrapf(
			ErrAddressPrefixUnsupported,
			"account prefix (%s) does not match expected prefix (%s)",
			prefix,
			a.network.AddressPrefix(),
		)
	}

	txHash, err := a.client.TransferToken(ctx, a.fundsPrivateKey, a.transferAmount, sdkAddr)
	if err != nil {
		return "", errors.Wrapf(ErrUnableToTransferToken, "err:%s", err)
	}

	return txHash, nil
}
