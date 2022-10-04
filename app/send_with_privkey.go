package app

import (
	"context"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

var (
	hdPath = sdk.GetConfig().GetFullBIP44Path()
)

// GenPrivKeyAndFundResponse is the response returned from GenPrivkeyAndFund
type GenPrivKeyAndFundResponse struct {
	TxHash   string
	Mnemonic string
	Address  string
}

// GenPrivkeyAndFund generates a private key and funds it
func (a App) GenPrivkeyAndFund(ctx context.Context) (GenPrivKeyAndFundResponse, error) {
	kr := keyring.NewInMemory()
	info, mnemonic, err := kr.NewMnemonic("", keyring.English, hdPath, "", hd.Secp256k1)
	if err != nil {
		return GenPrivKeyAndFundResponse{}, errors.Wrapf(ErrUnableToTransferToken, "err:%s", err)
	}
	sdkAddr := info.GetAddress()
	txHash, err := a.batcher.SendToken(ctx, sdkAddr, a.transferAmount)
	if err != nil {
		return GenPrivKeyAndFundResponse{}, errors.Wrapf(ErrUnableToTransferToken, "err:%s", err)
	}

	return GenPrivKeyAndFundResponse{
		TxHash:   txHash,
		Mnemonic: mnemonic,
		Address:  sdkAddr.String(),
	}, nil
}
