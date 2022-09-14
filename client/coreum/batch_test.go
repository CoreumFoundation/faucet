package coreum

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

func TestBatchSize(t *testing.T) {
	assertT := assert.New(t)
	log := logger.New(logger.ToolDefaultConfig)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batchSize := 10
	batcher := &Batcher{
		requestChan:      make(chan request, batchSize),
		logger:           log,
		fundingAddresses: []sdk.AccAddress{},
		batchSize:        batchSize,
		batchChan:        make(chan batch),
	}

	for i := 0; i < 10; i++ {
		go func() { batcher.requestChan <- request{} }()
	}
	time.Sleep(10 * time.Millisecond)

	batcher.Start(ctx)

	batch, ok := <-batcher.batchChan
	assertT.True(ok)
	assertT.Len(batch.addresses, 10)
	assertT.Len(batch.responses, 10)
}
