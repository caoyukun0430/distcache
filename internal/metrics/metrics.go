// Package metrics provides cache metrics collection and reporting functionality.
package metrics

import (
	"fmt"
	"net/http"
	"os"

	"distcache/pkg/common/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	instanceName string

	loggerInstance = logger.NewLogger()

	// cache hit related
	cacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "distcache_hits_total",
		Help: "The total number of cache hits",
		ConstLabels: prometheus.Labels{
			"instance": instanceName,
		},
	})

	// miss means either 1st load from db to cache or belongs to other nodes
	cacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "distcache_misses_total",
		Help: "The total number of cache misses",
		ConstLabels: prometheus.Labels{
			"instance": instanceName,
		},
	})

	cacheEvictions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "distcache_evictions_total",
		Help: "The total number of cache evictions",
		ConstLabels: prometheus.Labels{
			"instance": instanceName,
		},
	})

	// 
	databaseHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "database_hits_total",
		Help: "The total number of database hits",
		ConstLabels: prometheus.Labels{
			"instance": instanceName,
		},
	})

	// 
	databaseMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "database_misses_total",
		Help: "The total number of database misses",
		ConstLabels: prometheus.Labels{
			"instance": instanceName,
		},
	})

	requestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "distcache_requests_total",
		Help: "The total number of requests received",
		ConstLabels: prometheus.Labels{
			"instance": instanceName,
		},
	})

	cacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "distcache_size_bytes",
		Help: "The current size of the cache in bytes",
		ConstLabels: prometheus.Labels{
			"instance": instanceName,
		},
	})

	cacheItemCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "distcache_items_total",
		Help: "The total number of items in the cache",
		ConstLabels: prometheus.Labels{
			"instance": instanceName,
		},
	})

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "distcache_request_duration_seconds",
			Help:    "Time spent processing cache requests",
			Buckets: prometheus.ExponentialBuckets(0.00001, 2, 20), // from 10µs to ~5s
		},
		[]string{"operation", "instance"},
	)
)

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	instanceName = hostname
}

// StartMetricsServer 启动指标收集服务器
func StartMetricsServer(port int) {
	mux := http.NewServeMux()

	// register metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// redirect from / to /metrics
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusFound)
	})

	// async start
	go func() {
		addr := fmt.Sprintf(":%d", port)
		loggerInstance.Infof("Starting metrics server on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			loggerInstance.Errorf("Failed to start metrics server: %v", err)
		}
	}()
}

// RecordCacheHit 记录缓存命中
func RecordCacheHit() {
	cacheHits.Inc()
}

// RecordCacheMiss 记录缓存未命中
func RecordCacheMiss() {
	cacheMisses.Inc()
}

// RecordDatabaseHit 
func RecordDatabaseHit() {
	databaseHits.Inc()
}

// RecordDatabaseMiss 
func RecordDatabaseMiss() {
	databaseMisses.Inc()
}

// RecordEviction 记录缓存驱逐
func RecordEviction() {
	cacheEvictions.Inc()
}

// UpdateCacheSize 更新缓存大小（字节）
func UpdateCacheSize(size int64) {
	cacheSize.Set(float64(size))
}

// UpdateCacheItemCount 更新缓存项数量
func UpdateCacheItemCount(count int64) {
	cacheItemCount.Set(float64(count))
}

// ObserveRequestDuration records the duration of a cache operation
func ObserveRequestDuration(operation string, duration float64) {
	requestDuration.WithLabelValues(operation, instanceName).Observe(duration)
}

// RecordRequest increments the total request counter
func RecordRequest() {
	requestsTotal.Inc()
}
