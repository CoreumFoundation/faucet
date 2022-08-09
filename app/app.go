package app

import (
	"context"

	"github.com/CoreumFoundation/faucet/client/coreum"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// App implements core functionality
type App struct {
	fundsPrivateKey types.Secp256k1PrivateKey
	client          coreum.Client
	transferAmount  types.Coin
	network         app.Network
}

// New returns a new instance of the App
func New(
	client coreum.Client,
	network app.Network,
	transferAmount types.Coin,
	fundsPrivateKey types.Secp256k1PrivateKey,
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
