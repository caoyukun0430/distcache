# Distributed gRPC Cache (distcache)

In cloud-native networks, many microservices and network functions (often running as CNFs—Cloud Native Network Functions)continuously exchange telemetry data and state information. This high-frequency, concurrent access to operational data (such as performance metrics, resource usage, and routing information) is one of the primary motivations for using an in-memory, distributed cache.

Using a cache in this context ensures that thousands (or even millions) of containers and microservices can rapidly retrieve critical metrics without hitting a centralized database, which would otherwise become a bottleneck during peak loads.
Therefore, we try to build a distributed key-val based cache system, which supports HTTP and gRPC, and service discovery.

## Supported Feature

1. LRU (Least Recently Used) cache eviction and expiration mechanisms.

2. SingleFlight mechanism to manage concurrent read requests, preventing system overload.

3. Consistent Hashing to mitigate cache avalanche and penetration issues.

4. gRPC and HTTP protocols for seamless communication between nodes.

5. Dynamic node management facilitated by the ETCD endpoint manager.

6. collection for observability tools such as Prometheus and Grafana.

## Project Structure

```
.
├── README.md
├── api
│   ├── groupcachepb         // grpc server proto
│   └── cnfmetricspb         // business model proto
├── config                   // global config manage
│   ├── config.go
│   └── config.yml
├── go.mod
├── go.sum
├── internal
│   ├── business/cnf         // business logic
│   │   ├── db
|   |   |   ├── db.go        // init DB and DB data
│   |   |   └── cnfmetricspb // operations for DB CRUD
│   │   ├── ecode
│   │   ├── model
│   │   └── service
|   ├── metrics
|   └── cache                // grpc dist cache service
│       ├── byteview.go      // read-only 
│       ├── cache.go         // concurrency-safe caching
│       ├── eviction         // cache eviction algorithm
│       │   ├── lru
│       │   └── strategy
│       ├── consistenthash   // consistent hash algorithm for load balance
│       ├── group.go         
│       ├── groupcache.go    // group cache imp.
│       ├── grpc_fetcher.go  // grpc client 
│       ├── grpc_picker.go   // grpc server
│       ├── http_fetcher.go  // http proxy
│       ├── http_helper.go   // http api server and http server start helper
│       ├── http_picker.go   // http peer selector
│       ├── interface.go     // grpc peer selector and grpc proxy abstract
│       └── singleflight     // single flight concurrent access control 
├── pkg                
│   ├── etcd 
│   │   ├── cluster         // goerman etcd cluster manage
│   │   └── discovery       // service registration discovery        
│   └── common              // grpc group cache service imp.
│       ├── logger
|       └── validate        // ip address validation
|
├── main.go                 // grpc server default imp
├── test
│   ├── grpc                // grpc clients
│   └── sql                 // sql run sh
|
├── start.sh                // sh to start
└── stop.sh                 // sh to stop

```

## Dependencies

1. etcd and goreman

2. grpc and protobuf

3. mysql

## Reference

1. https://geektutu.com/post/geecache.html

2. https://github.com/mattn/goreman

3. https://grpc.io/docs/languages/go/quickstart

4. https://etcd.io/docs/v3.5/quickstart/