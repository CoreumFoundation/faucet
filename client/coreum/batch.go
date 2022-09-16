package coreum

import (
	"context"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

// NewBatcher returns new instance of Batcher type
func NewBatcher(
	client coreumClient,
	fundingAddresses []sdk.AccAddress,
	amount sdk.Coin,
	batchSize int,
) *Batcher {
	requestBufferSize := batchSize // number of requests that will be buffered to be batched
	b := &Batcher{
		requestBuffer:    make(chan request, requestBufferSize),
		client:           client,
		fundingAddresses: fundingAddresses,
		amount:           amount,
		batchSize:        batchSize,
		batchChan:        make(chan batch),
		mu:               sync.RWMutex{},
	}

	return b
}

// coreumClient is the interface that provides Coreum coreumClient functionality
type coreumClient interface {
	TransferToken(
		ctx context.Context,
		fromAddress sdk.AccAddress,
		amount sdk.Coin,
		destAddresses ...sdk.AccAddress,
	) (string, error)
}

// Batcher exposes functionality to batch many transfer requests
type Batcher struct {
	requestBuffer    chan request
	client           coreumClient
	fundingAddresses []sdk.AccAddress
	amount           sdk.Coin
	batchSize        int
	batchChan        chan batch

	mu      sync.RWMutex
	stopped bool
}

type result struct {
	txHash string
	err    error
}

type request struct {
	responseChan chan result
	address      sdk.AccAddress
}

// TransferToken receives a single transfer token request, batch sends them and returns the result
func (b *Batcher) TransferToken(ctx context.Context, destAddress sdk.AccAddress) (string, error) {
	resChan, err := b.requestFund(destAddress)
	if err != nil {
		return "", err
	}
	select {
	case res := <-resChan:
		return res.txHash, res.err
	case d := <-ctx.Done():
		return "", errors.Errorf("request aborted, %v", d)
	}
}

func (b *Batcher) close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.stopped {
		return
	}
	close(b.requestBuffer)
	b.stopped = true
}

func (b *Batcher) isClosed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.stopped
}

func (b *Batcher) requestFund(address sdk.AccAddress) (<-chan result, error) {
	if b.isClosed() {
		return nil, errors.New("request processor is closed")
	}
	req := request{
		responseChan: make(chan result, 1),
		address:      address,
	}
	b.requestBuffer <- req
	return req.responseChan, nil
}

// Start starts goroutines for batch processing requests
func (b *Batcher) Start(ctx context.Context) {
	go func() {
		<-ctx.Done()
		b.close()
	}()
	go b.createBatches()

	for _, fundingAddress := range b.fundingAddresses {
		go func(addr sdk.AccAddress) {
			b.processBatches(ctx, addr)
		}(fundingAddress)
	}
}

type batch struct {
	addresses []sdk.AccAddress
	responses []chan result
}

func (b *Batcher) processBatches(ctx context.Context, fromAddress sdk.AccAddress) {
	for {
		ba, ok := <-b.batchChan
		if !ok {
			break
		}

		b.sendBatch(ctx, fromAddress, ba)
	}
}

func (b *Batcher) sendBatch(ctx context.Context, fromAddress sdk.AccAddress, ba batch) {
	log := logger.Get(ctx)
	ctx = logger.WithLogger(context.Background(), log)
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var rsp = result{}
	txHash, err := b.client.TransferToken(ctx, fromAddress, b.amount, ba.addresses...)
	if err != nil {
		rsp.err = err
	} else {
		rsp.txHash = txHash
	}

	for _, r := range ba.responses {
		r <- rsp
	}
}

func (b *Batcher) createBatches() {
	var ba batch
	for {
		req, ok := <-b.requestBuffer
		if ok {
			ba.addresses = append(ba.addresses, req.address)
			ba.responses = append(ba.responses, req.responseChan)
		}

		if (len(ba.addresses) >= b.batchSize || len(b.requestBuffer) == 0 || !ok) && len(ba.addresses) > 0 {
			b.batchChan <- ba
			ba = batch{}
		}

		if !ok {
			break
		}
	}
	close(b.batchChan)
}
