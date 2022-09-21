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
	batchSize int,
) *Batcher {
	requestBufferSize := batchSize // number of requests that will be buffered to be batched
	b := &Batcher{
		requestBuffer:    make(chan request, requestBufferSize),
		client:           client,
		fundingAddresses: fundingAddresses,
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
		requests ...transferRequest,
	) (string, error)
}

// Batcher exposes functionality to batch many transfer requests
type Batcher struct {
	requestBuffer    chan request
	client           coreumClient
	fundingAddresses []sdk.AccAddress
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
	req          transferRequest
}

// SendToken receives a single transfer token request, batch sends them and returns the result
func (b *Batcher) SendToken(ctx context.Context, destAddress sdk.AccAddress, amount sdk.Coin) (string, error) {
	resChan, err := b.requestFund(destAddress, amount)
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

func (b *Batcher) requestFund(address sdk.AccAddress, amount sdk.Coin) (<-chan result, error) {
	if b.isClosed() {
		return nil, errors.New("request processor is closed")
	}
	req := request{
		responseChan: make(chan result, 1),
		req: transferRequest{
			destAddress: address,
			amount:      amount,
		},
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

type batch []request

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
	requests := []transferRequest{}
	for _, r := range ba {
		requests = append(requests, r.req)
	}
	// TODO: retry can be implemented to make it more resilient to network errors.
	//nolint:contextcheck // We don't want to cancel requests on shutdown sequence
	txHash, err := b.client.TransferToken(ctx, fromAddress, requests...)
	if err != nil {
		rsp.err = err
	} else {
		rsp.txHash = txHash
	}

	for _, rq := range ba {
		rq.responseChan <- rsp
	}
}

func (b *Batcher) createBatches() {
	var ba batch
	for {
		req, ok := <-b.requestBuffer
		if ok {
			ba = append(ba, req)
		}

		if (len(ba) >= b.batchSize || len(b.requestBuffer) == 0 || !ok) && len(ba) > 0 {
			b.batchChan <- ba
			ba = batch{}
		}

		if !ok {
			break
		}
	}
	close(b.batchChan)
}
