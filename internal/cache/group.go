package cache

import (
	"context"
	"errors"
	"fmt"
	"encoding/json"
	"time"

	cnfmetricspb "distcache/api/cnfmetricspb"
	"distcache/config"
	"distcache/internal/bussiness/cnf/db"

	"gorm.io/gorm"
)

// NewGroupManager creates and initializes cache groups for the specified CNF metric types.
// Returns a map of group names (metric types) to their corresponding Group instances.
func NewGroupManager(metricTypes []string, currentPeerAddr string) map[string]*Group {
    for _, metricType := range metricTypes {
        retriever := createCnfMetricRetriever()
        group := NewGroup(metricType, config.Conf.GroupManager.Strategy, config.Conf.GroupManager.MaxCacheSize, retriever)
        GroupManager[metricType] = group
        loggerInstance.Infof("Group '%s' created with strategy: '%s'", metricType, config.Conf.GroupManager.Strategy)
    }
    return GroupManager
}

// createCnfMetricRetriever sets up a RetrieveFunc to fetch CNF metric data from the database.
// It logs query execution time and handles errors appropriately.
// when cache is not hit, the group.getLocally func will call the retriever
func createCnfMetricRetriever() RetrieveFunc {
    return func(key string) ([]byte, error) {
        start := time.Now()
        defer func() {
            loggerInstance.Debugf("Database query duration: %v ms", time.Since(start).Milliseconds())
        }()

        ctx := context.Background()
        cnfMetricDb := db.NewCnfMetricDb(ctx)

        // Retrieve CNF metric information by key (CnfId).
        cnfMetric, err := cnfMetricDb.ShowCnfMetric(&cnfmetricspb.CnfMetricRequest{CnfId: key})
        if err != nil {
            // Handle case where the record is not found.
            if errors.Is(err, gorm.ErrRecordNotFound) {
                loggerInstance.Infof("No CNF metric record found for key: '%s'", key)
                return []byte{}, nil // Empty bytes indicate a negative cache result.
            }
            // Log and return other errors.
            loggerInstance.Errorf("Failed to query database for key '%s': %v", key, err)
            return nil, fmt.Errorf("database query error: %w", err)
        }

        loggerInstance.Infof("Successfully retrieved CNF metric record: CnfId='%s'", key)

        // Serialize the full CnfMetric object into JSON for storage in the cache
        metricJSON, err := json.Marshal(cnfMetric) // Calls MarshalJSON internally
        if err != nil {
            loggerInstance.Errorf("Failed to serialize CNF metric for key '%s': %v", key, err)
            return nil, fmt.Errorf("serialization error: %w", err)
        }

        return metricJSON, nil // Return the serialized metric as JSON
    }
}
