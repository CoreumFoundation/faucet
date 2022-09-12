package app

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/faucet/client/coreum"
)

// App implements core functionality
type App struct {
	batcher        *coreum.Batcher
	transferAmount sdk.Coin
	network        app.Network
}

// New returns a new instance of the App
func New(
	ctx context.Context,
	logger *zap.Logger,
	client coreum.Client,
	network app.Network,
	transferAmount sdk.Coin,
	fundingAddresses []sdk.AccAddress,
) App {
	batcher := coreum.NewBatcher(ctx, logger, client, fundingAddresses, transferAmount)
	return App{
		batcher:        batcher,
		network:        network,
		transferAmount: transferAmount,
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

	txHash, err := a.batcher.TransferToken(ctx, sdkAddr)
	if err != nil {
		return "", errors.Wrapf(ErrUnableToTransferToken, "err:%s", err)
	}

	return txHash, nil
}
