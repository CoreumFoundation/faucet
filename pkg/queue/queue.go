package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"github.com/CoreumFoundation/faucet/client/coreum"
)

type Request struct {
	Address  string
	TxHashCh chan<- string
}

type batch []Request

const maxBatchSize = 20

func Run(ctx context.Context, reqCh <-chan Request, privKeys []secp256k1.PrivKey, cl coreum.Client) error {
	return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		batchCh := make(chan batch)
		spawn("batch", parallel.Fail, func(ctx context.Context) error {
			for {
				b := make(batch, 0, maxBatchSize)

				select {
				case <-ctx.Done():
					return ctx.Err()
				case req := <-reqCh:
					b = append(b, req)
				}

				timeout := time.After(100 * time.Millisecond)
			loop:
				for i := 1; i < maxBatchSize; i++ {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-timeout:
						break loop
					case req := <-reqCh:
						b = append(b, req)
					}
				}

				batchCh <- b
			}
		})

		for i := 0; i < len(privKeys); i++ {
			privKey := privKeys[i]
			spawn(fmt.Sprintf("worker-%d", i), parallel.Fail, func(ctx context.Context) error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case b := <-batchCh:
						txHash, err := cl.TransferTokenManyAsyc(ctx, privKey, ...)
						if err != nil {
							return err
						}
						for _, req := range b {
							select {
							case <-ctx.Done():
								return ctx.Err()
							case req.TxHashCh<-txHash:
							}
						}

						if err := cl.AwaitTx(txHash); err != nil {
							return err
						}
					}
				}
			})
		}

		return nil
	})

	// 2. many goroutines (one per priv key) funding batches
}
