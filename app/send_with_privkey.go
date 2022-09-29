package app

import (
	"context"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/pkg/errors"
)

var (
	// TODO: update coin type after we register our coin type into BIP44
	hdPath = hd.CreateHDPath(118, 0, 0).String()
)

// GenPrivkeyAndFundResponse is the response returned from GenPrivkeyAndFund
type GenPrivkeyAndFundResponse struct {
	TxHash   string
	Mnemonic string
	Address  string
}

// GenPrivkeyAndFund generates a private key and funds it
func (a App) GenPrivkeyAndFund(ctx context.Context) (GenPrivkeyAndFundResponse, error) {
	kr := keyring.NewInMemory()
	info, mnemonic, err := kr.NewMnemonic("", keyring.English, hdPath, "", hd.Secp256k1)
	if err != nil {
		return GenPrivkeyAndFundResponse{}, errors.Wrapf(ErrUnableToTransferToken, "err:%s", err)
	}
	sdkAddr := info.GetAddress()
	txHash, err := a.batcher.SendToken(ctx, sdkAddr, a.transferAmount)
	if err != nil {
		return GenPrivkeyAndFundResponse{}, errors.Wrapf(ErrUnableToTransferToken, "err:%s", err)
	}

	return GenPrivkeyAndFundResponse{
		TxHash:   txHash,
		Mnemonic: mnemonic,
		Address:  sdkAddr.String(),
	}, nil
}