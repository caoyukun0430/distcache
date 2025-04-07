package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	pb "distcache/api/groupcachepb"
	"distcache/config"
	"distcache/internal/bussiness/cnf/db"
	"distcache/internal/bussiness/cnf/model"
	"distcache/pkg/common/logger"
	discovery "distcache/pkg/etcd/discovery"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrRPCCallNotFound  = "rpc error: code = Unknown desc = record not found"
	MaxRetries          = 3
	InitialRetryWaitSec = 1
)

const (
	NotFoundStatus Status = iota
	ErrorStatus
)

var loggerInstance = logger.NewLogger()

type Status int

type GGCacheClient struct {
	etcdCli     *clientv3.Client
	conn        *grpc.ClientConn
	client      pb.GroupCacheClient
	serviceName string
	connected   bool
	mu          sync.RWMutex
}

func NewGGCacheClient(etcdCli *clientv3.Client, serviceName string) (*GGCacheClient, error) {
	client := &GGCacheClient{
		etcdCli:     etcdCli,
		serviceName: serviceName,
	}
	if err := client.connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *GGCacheClient) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := discovery.Discovery(c.etcdCli, c.serviceName)
	if err != nil {
		return fmt.Errorf("failed to discover service: %v", err)
	}

	c.conn = conn
	c.client = pb.NewGroupCacheClient(conn)
	c.connected = true
	return nil
}

func (c *GGCacheClient) Get(ctx context.Context, group, key string) (*pb.GetResponse, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		if err := c.connect(); err != nil {
			return nil, err
		}
	} else {
		c.mu.RUnlock()
	}

	var lastErr error
	for retry := 0; retry < MaxRetries; retry++ {
		resp, err := c.client.Get(ctx, &pb.GetRequest{
			Group: group,
			Key:   key,
		})

		if err == nil {
			return resp, nil
		}

		lastErr = err
		if status.Code(err) == codes.Unavailable {
			// 连接断开，尝试重连
			c.mu.Lock()
			c.connected = false
			c.mu.Unlock()

			if reconnErr := c.connect(); reconnErr != nil {
				lastErr = reconnErr
			}
		}

		// 使用指数退避等待
		waitTime := time.Duration(backoff(retry)) * time.Second
		loggerInstance.Warnf("第 %d 次重试失败，等待 %v 后重试: %v", retry+1, waitTime, err)
		time.Sleep(waitTime)
	}

	return nil, fmt.Errorf("max retries exceeded: %v", lastErr)
}

func main() {
	config.InitConfig()
	if err := db.InitDB(); err != nil {
		loggerInstance.Errorf("Failed to initialize database: %v", err)
	}

	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		panic(err)
	}

	ggcacheClient, err := NewGGCacheClient(cli, config.Conf.Services["groupcache"].Name)
	if err != nil {
		panic(err)
	}

	// 构造热点数据（20%的key承载80%的访问）
    // 选择指标值高的CNF作为热点数据
     // Define hot keys: These will handle 80% of the traffic
	 hotKeys := []string{
        "CNF-001", "CNF-002", "CNF-003", "CNF-004", "CNF-005",
        "CNF-006", "CNF-007", "CNF-008", "CNF-009", "CNF-010",
    }

    // Define a map to track used keys
    usedKeys := map[string]bool{
        "CNF-001": true, "CNF-002": true, "CNF-003": true,
        "CNF-004": true, "CNF-005": true, "CNF-006": true,
        "CNF-007": true, "CNF-008": true, "CNF-009": true,
        "CNF-010": true,
    }

    // 构造长尾数据（80%的key承载20%的访问）
    // 使用随机生成或较低重要性的CNF数据作为长尾数据
    coldKeys := make([]string, 0)
    // 添加其他CNF ID
    for i := 11; i <= 100; i++ { // Arbitrarily generate up to CNF-100 for cold keys
        cnfId := fmt.Sprintf("CNF-%03d", i) // Format as CNF-006, CNF-007, etc.
        if !usedKeys[cnfId] {
            coldKeys = append(coldKeys, cnfId)
        }
    }

    // 构造最终的请求序列
    totalRequests := 1000 // 总请求数
    cnfIds := make([]string, 0, totalRequests)

    // 添加热点请求（80%的访问量）
    hotRequestCount := int(float64(totalRequests) * 0.8) // 80%的请求量
    for i := 0; i < hotRequestCount; i++ {
        cnfIds = append(cnfIds, hotKeys[i%len(hotKeys)])
    }

    // 添加长尾请求（20%的访问量）
    coldRequestCount := totalRequests - hotRequestCount // 剩余20%的请求量
    for i := 0; i < coldRequestCount; i++ {
        cnfIds = append(cnfIds, coldKeys[i%len(coldKeys)])
    }

    // 随机打散请求顺序
    rand.Shuffle(len(cnfIds), func(i, j int) {
        cnfIds[i], cnfIds[j] = cnfIds[j], cnfIds[i]
    })

    // 模拟请求 CNF metrics
    for {
        for _, cnfId := range cnfIds {
            ctx := context.Background()
            resp, err := ggcacheClient.Get(ctx, "metrics", cnfId)
            if err != nil {
                if ErrorHandle(err) == NotFoundStatus {
                    loggerInstance.Warnf("Not CNF ID %s metrics found", cnfId)
                    continue
                }
                loggerInstance.Errorf("Query CNF ID %s metrics failed: %v", cnfId, err)
                return // 如果不是 NotFound 错误，直接退出程序
            }
            // Deserialize the response into a CnfMetric object
			var cnfMetric model.CnfMetric
			if err := json.Unmarshal(resp.Value, &cnfMetric); err != nil {
				loggerInstance.Errorf("Deserialize CNF ID %s metrics failed: %v", cnfId, err)
				continue
			}

			// Log detailed information about the metric
			loggerInstance.Infof("Query succeed, CNF ID: %s, MetricType: %s, Value: %.2f, Unit: %s, Status: %s",
				cnfMetric.CnfId, cnfMetric.MetricType, cnfMetric.Value, cnfMetric.Unit, cnfMetric.Status)
		}
        time.Sleep(time.Millisecond * 100) // 控制请求频率
    }
}

func ErrorHandle(err error) Status {
	if err.Error() == ErrRPCCallNotFound {
		return NotFoundStatus
	}
	return ErrorStatus
}

// First retry wait 1s
// Second retry wait 2s
// The third retry waits for 4 seconds
func backoff(retry int) int {
	return int(math.Pow(float64(2), float64(retry)))
}
