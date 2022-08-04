package app

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/pkg/errors"
)

func parseAddress(address string) (string, sdk.AccAddress, error) {
	if len(strings.TrimSpace(address)) == 0 {
		return "", nil, errors.New("empty address string is not allowed")
	}

	hrp, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return "", nil, err
	}

	err = sdk.VerifyAddressFormat(bz)
	if err != nil {
		return "", nil, err
	}

	return hrp, sdk.AccAddress(bz), nil
}
