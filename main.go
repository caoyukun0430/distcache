package main

import (
	"flag"
	"fmt"

	"distcache/config"
	"distcache/internal/bussiness/cnf/db"
	"distcache/internal/cache"
	"distcache/pkg/common/logger"
	"distcache/pkg/etcd/discovery"
	"distcache/internal/metrics"
)

var (
	port        = flag.Int("port", 9999, "service node port")
	metricsPort = flag.Int("metricsPort", 2222, "metrics port")
	loggerInstance = logger.NewLogger()
)

// shared updateChan is created for communication between NewServer(to add/update peers in the ring) AND the etcd dynamic node discovery!
// first via DynamicServices the etcd starts to watch updates for nodes prefixed GroupCache
// 2nd SetPeers set up the hash ring, grpcClients and the channels, e.g. update chan
// 3rd Start() calls registerService, which register the svc/addr into etcd endpoints, it triggers DynamicServices on all the peers via update channel, which reconstructs the hash ring defined inside SetPeers
func main() {
	config.InitConfig()
	// Initialize database
	if err := db.InitDB(); err != nil {
		loggerInstance.Errorf("Failed to initialize database: %v", err)
	}
	flag.Parse()

	metrics.StartMetricsServer(*metricsPort)
	loggerInstance.Infof("Metrics server started on port %d", *metricsPort)
	serviceAddr := fmt.Sprintf("localhost:%d", *port)
	gm := cache.NewGroupManager([]string{"metrics"}, serviceAddr)

	updateChan := make(chan struct{})
	svr, err := cache.NewServer(updateChan, serviceAddr)
	if err != nil {
		loggerInstance.Errorf("acquire grpc server instance failed, %v", err)
		return
	}

	go discovery.DynamicServices(updateChan, config.Conf.Services["groupcache"].Name)

	// check if there exists peers already, if so, we need to include them when initially SetPeers
	// if not, SetPeers will be only the node itself
	peers, err := discovery.ListServicePeers(config.Conf.Services["groupcache"].Name)
	if err != nil {
		loggerInstance.Errorf("failed to discover peers: %v", err)
		return
	}

	svr.SetPeers(peers)

	gm["metrics"].RegisterServer(svr)

	if err := svr.Start(); err != nil {
		loggerInstance.Errorf("failed to start server: %v", err)
		return
	}
}
