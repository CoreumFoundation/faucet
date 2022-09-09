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
	b := &Batcher{
		requestsChan:       make(chan *request),
		logger:             logger,
		client:             client,
		fundingAddressChan: make(chan sdk.AccAddress, len(fundingAddresses)),
		amount:             amount,
		batchSize:          10,
		mut:                &sync.RWMutex{},
		stopped:            false,
	}
	for _, key := range fundingAddresses {
		b.fundingAddressChan <- key
	}

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
	requestsChan       chan *request
	logger             *zap.Logger
	client             coreumClient
	fundingAddressChan chan sdk.AccAddress
	amount             sdk.Coin
	batchSize          int

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
	b.requestsChan <- &req
	return req.responseChan
}

// Start goroutine for batch processing requests
func (b *Batcher) Start(ctx context.Context) {
	var exit bool
	go func() {
		for !exit {
			select {
			case <-ctx.Done():
				b.close()
				return
			default:
				b.batchSendRequests()
			}
		}
	}()
}

func (b *Batcher) batchSendRequests() {
	var accAddresses []sdk.AccAddress
	var rspList []chan result
	fundingAddress := <-b.fundingAddressChan
	defer func() {
		b.fundingAddressChan <- fundingAddress
	}()
	for {
		req := <-b.requestsChan
		// nil means channel is closed, all requests are processed, and we will not process any further requests
		if req == nil {
			break
		}
		accAddresses = append(accAddresses, req.address)
		rspList = append(rspList, req.responseChan)
		if len(accAddresses) >= b.batchSize || len(b.requestsChan) == 0 {
			break
		}
	}

	if len(accAddresses) == 0 {
		return
	}

	ctx := logger.WithLogger(context.Background(), b.logger)
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var rsp = result{}
	txHash, err := b.client.TransferTokenMany(ctx, fundingAddress, b.amount, accAddresses...)
	if err != nil {
		rsp.err = err
	} else {
		rsp.txHash = txHash
	}

	for _, r := range rspList {
		r <- rsp
	}
}
