package coreum

import (
	"context"
	"sync"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

type mockCoreumClient struct {
	mu    sync.Mutex
	calls []clientCall
}

type clientCall struct {
	fromAddress   sdk.AccAddress
	amount        sdk.Coin
	destAddresses []sdk.AccAddress
}

func (mc *mockCoreumClient) TransferToken(
	ctx context.Context,
	fromAddress sdk.AccAddress,
	amount sdk.Coin,
	destAddresses ...sdk.AccAddress,
) (string, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.calls = append(mc.calls, clientCall{
		fromAddress:   fromAddress,
		amount:        amount,
		destAddresses: destAddresses,
	})
	return "", nil
}

func TestBatchSend(t *testing.T) {
	assertT := assert.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	amount := sdk.NewCoin("test-denom", sdk.NewInt(13))
	fundingAddresses := []sdk.AccAddress{}
	for i := 0; i < 2; i++ {
		address, _ := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
		fundingAddresses = append(fundingAddresses, address)
	}

	mock := &mockCoreumClient{}
	batcher := NewBatcher(mock, fundingAddresses, amount, 10)
	batcher.Start(ctx)

	wg := sync.WaitGroup{}
	requestCount := 100
	wg.Add(requestCount)
	for i := 0; i < requestCount; i++ {
		go func() {
			_, err := batcher.TransferToken(ctx, nil)
			assertT.NoError(err)
			wg.Done()
		}()
	}

	wg.Wait()

	assertT.Less(len(mock.calls), 20)
	assertT.GreaterOrEqual(len(mock.calls), 10)

	totalAddressesCount := 0
	for _, call := range mock.calls {
		totalAddressesCount += len(call.destAddresses)
	}

	assertT.EqualValues(requestCount, totalAddressesCount)
}
