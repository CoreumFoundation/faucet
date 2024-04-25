package app

import (
	"context"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// GenMnemonicAndFundResult is the response returned from GenMnemonicAndFund.
type GenMnemonicAndFundResult struct {
	TxHash   string
	Mnemonic string
	Address  string
}

// GenMnemonicAndFund generates a private key and funds it.
func (a App) GenMnemonicAndFund(ctx context.Context) (GenMnemonicAndFundResult, error) {
	kr := a.clientCtx.Keyring()
	info, mnemonic, err := kr.NewMnemonic("", keyring.English, sdk.GetConfig().GetFullBIP44Path(), "", hd.Secp256k1)
	if err != nil {
		return GenMnemonicAndFundResult{}, errors.Wrapf(ErrUnableToTransferToken, "err:%s", err)
	}
	sdkAddr, err := info.GetAddress()
	if err != nil {
		return GenMnemonicAndFundResult{}, errors.Wrapf(ErrUnableToTransferToken, "err:%s", err)
	}
	txHash, err := a.batcher.SendToken(ctx, sdkAddr, a.transferAmount)
	if err != nil {
		return GenMnemonicAndFundResult{}, errors.Wrapf(ErrUnableToTransferToken, "err:%s", err)
	}

	return GenMnemonicAndFundResult{
		TxHash:   txHash,
		Mnemonic: mnemonic,
		Address:  sdkAddr.String(),
	}, nil
}
