package coreum

import (
	"context"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

// NewBatcher returns new instance of Batcher type
func NewBatcher(
	ctx context.Context,
	logger *zap.Logger,
	client coreumClient,
	fundingAddresses []sdk.AccAddress,
	amount sdk.Coin,
) *Batcher {
	requestBufferSize := 10 // number of requests that will be buffered to be batched
	batchSize := 10
	b := &Batcher{
		requestsChan:     make(chan request, requestBufferSize),
		logger:           logger,
		client:           client,
		fundingAddresses: append([]sdk.AccAddress{}, fundingAddresses...),
		amount:           amount,
		batchSize:        batchSize,
		mut:              &sync.RWMutex{},
		stopped:          false,
	}
	b.start(ctx)

	return b
}

// coreumClient is the interface that provides Coreum coreumClient functionality
type coreumClient interface {
	TransferTokenMany(
		ctx context.Context,
		fromAddress sdk.AccAddress,
		amount sdk.Coin,
		destAddresses ...sdk.AccAddress,
	) (string, error)
}

// Batcher exposes functionality to batch many transfer requests
type Batcher struct {
	requestsChan     chan request
	logger           *zap.Logger
	client           coreumClient
	fundingAddresses []sdk.AccAddress
	amount           sdk.Coin
	batchSize        int

	mut     *sync.RWMutex
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

// TransferToken receives a singer transfer token request, batch sends them and returns the result
func (b *Batcher) TransferToken(ctx context.Context, destAddress sdk.AccAddress) (string, error) {
	res := <-b.requestFund(destAddress)
	if res.err != nil {
		return "", res.err
	}
	return res.txHash, nil
}

func (b *Batcher) close() {
	b.mut.Lock()
	defer b.mut.Unlock()
	if b.stopped {
		return
	}
	close(b.requestsChan)
	b.stopped = true
}

func (b *Batcher) isClosed() bool {
	b.mut.RLock()
	defer b.mut.RUnlock()
	return b.stopped
}

func (b *Batcher) requestFund(address sdk.AccAddress) chan result {
	req := request{
		responseChan: make(chan result, 1),
		address:      address,
	}
	if b.isClosed() {
		req.responseChan <- result{
			err: errors.New("request processor is closed"),
		}
		return req.responseChan
	}
	b.requestsChan <- req
	return req.responseChan
}

// start starts goroutines for batch processing requests
func (b *Batcher) start(ctx context.Context) {
	go func() {
		<-ctx.Done()
		b.close()
	}()
	batchChan := make(chan batch)
	go func() {
		b.createBatches(batchChan)
	}()

	for _, fundingAddress := range b.fundingAddresses {
		go func(addr sdk.AccAddress) {
			b.processBatches(addr, batchChan)
		}(fundingAddress)
	}
}

type batch struct {
	addresses []sdk.AccAddress
	responses []chan result
}

func (b *Batcher) processBatches(fromAddress sdk.AccAddress, batchCh <-chan batch) {
	for {
		ba, ok := <-batchCh
		if !ok {
			break
		}

		ctx := logger.WithLogger(context.Background(), b.logger)
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		var rsp = result{}
		txHash, err := b.client.TransferTokenMany(ctx, fromAddress, b.amount, ba.addresses...)
		if err != nil {
			rsp.err = err
		} else {
			rsp.txHash = txHash
		}

		for _, r := range ba.responses {
			r <- rsp
		}
	}
}

func (b *Batcher) createBatches(batchCh chan<- batch) {
	var ba batch
	var exit bool
	for !exit {
		req, ok := <-b.requestsChan
		if !ok {
			exit = true
		} else {
			ba.addresses = append(ba.addresses, req.address)
			ba.responses = append(ba.responses, req.responseChan)
		}

		if len(ba.addresses) >= b.batchSize || len(b.requestsChan) == 0 || exit {
			batchCh <- ba
			ba = batch{}
		}
	}
	close(batchCh)
}
