package coreum

import (
	"context"
	"sync"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
)

type mockCoreumClient struct {
	mu    sync.Mutex
	calls []clientCall
}

type clientCall struct {
	fromAddress sdk.AccAddress
	requests    []transferRequest
}

func (mc *mockCoreumClient) TransferToken(
	ctx context.Context,
	fromAddress sdk.AccAddress,
	requests ...transferRequest,
) (string, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.calls = append(mc.calls, clientCall{
		fromAddress: fromAddress,
		requests:    requests,
	})
	return fromAddress.String(), nil
}

func TestBatchSend(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	ctx := logger.WithLogger(context.Background(), zaptest.NewLogger(t))
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	amount := sdk.NewCoin("test-denom", sdk.NewInt(13))
	fundingAddresses := []sdk.AccAddress{}
	for i := 0; i < 2; i++ {
		address, err := sdk.AccAddressFromHexUnsafe(secp256k1.GenPrivKey().PubKey().Address().String())
		fundingAddresses = append(fundingAddresses, address)
		requireT.NoError(err)
	}

	mock := &mockCoreumClient{}
	batcher := NewBatcher(mock, fundingAddresses, 10)

	group := parallel.NewGroup(ctx)
	group.Spawn("batcher", parallel.Fail, batcher.Run)
	t.Cleanup(func() {
		group.Exit(nil)
		_ = group.Wait()
	})

	wg := sync.WaitGroup{}
	requestCount := 100
	wg.Add(requestCount)
	for i := 0; i < requestCount; i++ {
		go func() {
			txHash, err := batcher.SendToken(ctx, nil, amount)
			requireT.NoError(err)
			assertT.Greater(len(txHash), 1) //nolint:testifylint // it suggests to use `NotEmpty` which is a nonsense here
			wg.Done()
		}()
	}

	wg.Wait()

	assertT.Less(len(mock.calls), 90)
	assertT.GreaterOrEqual(len(mock.calls), 10)

	totalAddressesCount := 0
	for _, call := range mock.calls {
		totalAddressesCount += len(call.requests)
	}

	assertT.EqualValues(requestCount, totalAddressesCount)
}
