package main

import (
	"context"
	"math/big"

	coreumApp "github.com/CoreumFoundation/coreum/app"
	coreumClient "github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/CoreumFoundation/faucet/app"
	"github.com/CoreumFoundation/faucet/client/coreum"
	"github.com/CoreumFoundation/faucet/http"
)

func main() {
	ctx := context.Background()
	network, _ := coreumApp.NetworkByChainID(coreumApp.Devnet)
	network.SetupPrefixes()
	cl := coreum.New(coreumClient.New(network.ChainID(), "localhost:26657"))
	transferAmount := types.Coin{
		Amount: big.NewInt(10_000),
		Denom:  network.TokenSymbol(),
	}
	fundsPrivatekey := types.Secp256k1PrivateKey{0x9b, 0xc9, 0xd0, 0x15, 0x11, 0x2, 0x94, 0x9, 0x92, 0xfd, 0x2b, 0xad, 0xbe, 0x36, 0x63, 0x1f, 0xed, 0x30, 0x10, 0xd, 0x6e, 0x24, 0xb1, 0xc2, 0x58, 0xb4, 0xfd, 0xe4, 0xaf, 0xdd, 0xf2, 0x40}
	app := app.New(cl, network, transferAmount, fundsPrivatekey)
	server := http.New(app)
	server.ListenAndServe(ctx, 8080)
}
