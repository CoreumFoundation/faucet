package app

import "github.com/pkg/errors"

// error type produced by app
var (
	ErrInvalidAddressFormat     = errors.New("invalid address format")
	ErrAddressPrefixUnsupported = errors.New("address prefix is not supported by this chain")
	ErrUnableToTransferToken    = errors.New("unable to transfer tokens")
)
