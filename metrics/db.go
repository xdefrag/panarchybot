package metrics

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	dbAcquireCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "db_connection_acquire_count",
		Help: "The total number of connection acquire attempts",
	})

	dbAcquireDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "db_connection_acquire_duration_seconds",
		Help:    "The duration of connection acquire attempts",
		Buckets: prometheus.DefBuckets,
	})

	dbMaxConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_max_connections",
		Help: "The maximum size of the connection pool",
	})

	dbMinConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_min_connections",
		Help: "The minimum size of the connection pool",
	})

	dbTotalConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_total_connections",
		Help: "The current size of the connection pool",
	})

	dbIdleConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_idle_connections",
		Help: "The current number of idle connections",
	})
)

// StartPoolMetrics запускает периодическое обновление метрик пула
func StartPoolMetrics(ctx context.Context, pool *pgxpool.Pool) {
	// Обновляем статические метрики сразу
	dbMaxConns.Set(float64(pool.Config().MaxConns))
	dbMinConns.Set(float64(pool.Config().MinConns))

	// Запускаем периодическое обновление динамических метрик
	go func() {
		ticker := time.NewTicker(time.Second * 15)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats := pool.Stat()
				dbTotalConns.Set(float64(stats.TotalConns()))
				dbIdleConns.Set(float64(stats.IdleConns()))
				dbAcquireCount.Add(float64(stats.AcquireCount()))
				dbAcquireDuration.Observe(stats.AcquireDuration().Seconds())
			}
		}
	}()
}
