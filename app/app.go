package app

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/app"
)

// App implements core functionality
type App struct {
	client         CoreumClient
	transferAmount sdk.Coin
	network        app.Network
}

// New returns a new instance of the App
func New(
	client CoreumClient,
	network app.Network,
	transferAmount sdk.Coin,
) App {
	return App{
		client:         client,
		network:        network,
		transferAmount: transferAmount,
	}
}

// CoreumClient indicates the required functionality to connect to coreum blockchain
type CoreumClient interface {
	TransferToken(ctx context.Context, destAddress sdk.AccAddress) (string, error)
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

	txHash, err := a.client.TransferToken(ctx, sdkAddr)
	if err != nil {
		return "", errors.Wrapf(ErrUnableToTransferToken, "err:%s", err)
	}

	return txHash, nil
}
