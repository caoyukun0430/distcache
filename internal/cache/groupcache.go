package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"distcache/internal/metrics"
	"gorm.io/gorm"
)

var (
	mu           sync.RWMutex
	GroupManager = make(map[string]*Group)
)

// Group represents a cache namespace and associated data/operations.
type Group struct {
	name      string
	cache     *cache
	retriever Retriever
	server    Picker
	flight    *FlightGroup
}

// NewGroup creates a new cache namespace with the specified configuration.
// It returns an existing group if one exists with the same name.
func NewGroup(name string, strategy string, maxBytes int64, retriever Retriever) *Group {
	if retriever == nil {
		panic("retriever is required for group creation")
	}

	mu.RLock()
	if group, exists := GroupManager[name]; exists {
		mu.RUnlock()
		return group
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	cache, err := NewCache(strategy, maxBytes)
	if err != nil {
		loggerInstance.Errorf("failed to create cache with strategy %q: %v", strategy, err)
		return nil
	}

	group := &Group{
		name:      name,
		cache:     cache,
		retriever: retriever,
		flight:    NewFlightGroup(10 * time.Second),
	}

	GroupManager[name] = group
	return group
}

// RegisterServer registers a server picker for distributed cache functionality.
// It panics if a server is already registered.
func (g *Group) RegisterServer(p Picker) {
	if g.server != nil {
		panic("server already registered for group")
	}
	g.server = p
}

// GetGroup retrieves a Group by name from the GroupManager.
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return GroupManager[name]
}

// DestroyGroup removes a Group and stops its associated server.
func DestroyGroup(name string) {
	g := GetGroup(name)
	if g != nil {
		if server, ok := g.server.(*Server); ok {
			if err := server.Stop(); err != nil {
				loggerInstance.Errorf("Failed to stop server: %v", err)
			}
		}
		// Stop the flight group and clear its cache
		if g.flight != nil {
			g.flight.Stop()
		}
		mu.Lock()
		delete(GroupManager, name)
		mu.Unlock()
	}
}

// Get retrieves a value from the cache by key.
// If the key doesn't exist in cache, it loads it using the configured retriever.
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key cannot be empty")
	}

	metrics.RecordRequest()

	if value, ok := g.cache.get(key); ok {
		loggerInstance.Infof("Group %s cache hit ..., key %s...", g.name, key)
		return value, nil
	}

	return g.load(key)
}

// load retrieves data for a key, either from a peer or locally.
// It uses FlightGroup to prevent thundering herd.
func (g *Group) load(key string) (value ByteView, err error) {
	ctx := context.Background()
	viewi, err := g.flight.Do(ctx, key, func() (interface{}, error) {
		if g.server != nil {
			if peer, ok := g.server.Pick(key); ok {
				if value, err = g.fetchFromPeer(peer, key); err == nil {
					return value, nil
				}
				loggerInstance.Warnf("failed to get from peer: %v", err)
			}
		}

		return g.getLocally(key)
	})

	if err != nil {
		return ByteView{}, err
	}

	return viewi.(ByteView), nil
}

// fetchFromPeer retrieves data from a peer cache node.
func (g *Group) fetchFromPeer(peer Fetcher, key string) (ByteView, error) {
	loggerInstance.Infof("fetchFromPeer peer is %+v", peer)
	bytes, err := peer.Fetch(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: cloneBytes(bytes)}, nil
}

// getLocally retrieves data from the configured retriever and populates the cache.
func (g *Group) getLocally(key string) (ByteView, error) {
	// put menas we need to retrieve the data from db and load into the cache
	start := time.Now()
	defer func() {
		metrics.ObserveRequestDuration("put", time.Since(start).Seconds()*1000)
	}()
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		metrics.RecordDatabaseMiss()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Cache empty result to prevent cache penetration
			loggerInstance.Infof("caching empty result for non-existent key %q to prevent cache penetration", key)
			g.populateCache(key, ByteView{})
		}
		return ByteView{}, fmt.Errorf("failed to retrieve key %q locally: %w", key, err)
	} else {
		metrics.RecordDatabaseHit()
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)

	return value, nil
}

// populateCache adds a key-value pair to the cache.
func (g *Group) populateCache(key string, value ByteView) {
	g.cache.put(key, value)
}
