package app

import (
	"context"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	faucethttp "github.com/CoreumFoundation/faucet/pkg/http"
)

// RunMonitoring runs monitoring service.
func RunMonitoring(
	ctx context.Context,
	listenAddress string,
	clientCtx client.Context,
	addresses []sdk.AccAddress,
	denom string,
) error {
	metricRecorder := newRecorder()
	registry := metricRecorder.Registry()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	server := &http.Server{Addr: listenAddress, Handler: mux}

	return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		spawn("server", parallel.Fail, func(ctx context.Context) error {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return errors.WithStack(err)
			}
			return errors.WithStack(ctx.Err())
		})
		spawn("watchdog", parallel.Fail, func(ctx context.Context) error {
			<-ctx.Done()

			ctx, cancel := context.WithTimeout(faucethttp.NewReopenedCtx(ctx), 10*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				return errors.WithStack(err)
			}
			return errors.WithStack(ctx.Err())
		})
		spawn("balances", parallel.Fail, func(ctx context.Context) error {
			log := logger.Get(ctx)
			bankClient := banktypes.NewQueryClient(clientCtx)
			for {
				for _, addr := range addresses {
					resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
						Address: addr.String(),
						Denom:   denom,
					})
					if err != nil {
						log.Error("Error occurred while probing balance", zap.Error(err))
						continue
					}
					metricRecorder.Balance(addr).Set(float64(resp.Balance.Amount.Uint64()))
				}

				select {
				case <-ctx.Done():
					return errors.WithStack(ctx.Err())
				case <-time.After(time.Minute):
				}
			}
		})

		return nil
	})
}

// recorder is metrics recorder.
type recorder struct {
	registry     *prometheus.Registry
	balanceGauge *prometheus.GaugeVec
}

// newRecorder returns a new instance of the recorder.
func newRecorder() *recorder {
	registry := prometheus.NewRegistry()
	balanceGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "balance",
		Help: "Faucet address balance",
	}, []string{"address"})

	registry.MustRegister(balanceGauge)

	return &recorder{
		registry:     registry,
		balanceGauge: balanceGauge,
	}
}

// Registry returns metrics registry.
func (r *recorder) Registry() *prometheus.Registry {
	return r.registry
}

// Balance returns gauge for measuring the balance of the address.
func (r *recorder) Balance(address sdk.AccAddress) prometheus.Gauge {
	return r.balanceGauge.With(prometheus.Labels{
		"address": address.String(),
	})
}
