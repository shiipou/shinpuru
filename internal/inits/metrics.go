package inits

import (
	"github.com/sarulabs/di/v2"
	"github.com/sirupsen/logrus"
	"github.com/zekroTJA/shinpuru/internal/config"
	"github.com/zekroTJA/shinpuru/internal/services/metrics"
	"github.com/zekroTJA/shinpuru/internal/util/static"
)

func InitMetrics(container di.Container) (ms *metrics.MetricsServer) {
	var err error

	cfg := container.Get(static.DiConfig).(*config.Config)

	if cfg.Metrics != nil && cfg.Metrics.Enable {
		if cfg.Metrics.Addr == "" {
			cfg.Metrics.Addr = ":9091"
		}

		ms, err = metrics.NewMetricsServer(cfg.Metrics.Addr)
		if err != nil {
			logrus.WithError(err).Fatal("failed initializing metrics server")
		}

		go func() {
			logrus.WithField("addr", cfg.Metrics.Addr).Info("Metrics server started")
			if err := ms.ListenAndServeBlocking(); err != nil {
				logrus.WithError(err).Fatal("failed setting up metrics server")
			}
		}()
	}

	return
}
