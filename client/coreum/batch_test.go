package coreum

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestBatchSize(t *testing.T) {
	assertT := assert.New(t)
	log := zaptest.NewLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batchSize := 10
	batcher := &Batcher{
		requestBuffer:    make(chan request, batchSize),
		logger:           log,
		fundingAddresses: []sdk.AccAddress{},
		batchSize:        batchSize,
		batchChan:        make(chan batch),
	}

	go func() {
		for i := 0; i < 10; i++ {
			select {
			case batcher.requestBuffer <- request{}:
			case <-time.After(time.Second):
			}
		}
	}()

	batcher.Start(ctx)

	select {
	case batch, ok := <-batcher.batchChan:
		assertT.True(ok)
		assertT.Len(batch.addresses, 10)
		assertT.Len(batch.responses, 10)
	case <-time.After(10 * time.Second):
		assertT.Fail("test timed out")
	}
}

type mockCoreumClient struct {
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
	mc.calls = append(mc.calls, clientCall{
		fromAddress:   fromAddress,
		amount:        amount,
		destAddresses: destAddresses,
	})
	return "", nil
}

func TestBatchSend(t *testing.T) {
	assertT := assert.New(t)
	log := zaptest.NewLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	amount := sdk.NewCoin("test-denom", sdk.NewInt(13))
	fundingAddresses := []sdk.AccAddress{}
	for i := 0; i < 2; i++ {
		address, _ := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
		fundingAddresses = append(fundingAddresses, address)
	}

	mock := &mockCoreumClient{}
	batcher := *NewBatcher(log, mock, fundingAddresses, amount, 10)
	batcher.Start(ctx)

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			_, _ = batcher.TransferToken(ctx, nil)
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

	assertT.EqualValues(100, totalAddressesCount)
}
