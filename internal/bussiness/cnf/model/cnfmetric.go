package model

import (
    "encoding/json"
	"fmt"
    "time"
)

// CnfMetric represents a single record of CNF metric data.
type CnfMetric struct {
    CnfId      string    `gorm:"primaryKey;type:varchar(50);not null"` // Aligned with Protobuf naming
    Timestamp  time.Time `gorm:"not null"`
    MetricType string    `gorm:"type:varchar(100);not null"`
    Value      float64   `gorm:"not null"`
    Unit       string    `gorm:"type:varchar(50);not null"`
    Status     string    `gorm:"type:varchar(20);not null"`
}

// Specify the table name for this model.
func (CnfMetric) Table() string {
    return "cnfMetric"
}

func (c *CnfMetric) MarshalJSON() ([]byte, error) {
    temp := struct {
        CnfId      string  `json:"cnf_id"`
        Timestamp  string  `json:"timestamp"`
        MetricType string  `json:"metric_type"`
        Value      float64 `json:"value"`
        Unit       string  `json:"unit"`
        Status     string  `json:"status"`
    }{
        CnfId:      c.CnfId,
        Timestamp:  c.Timestamp.Format(time.RFC3339),
        MetricType: c.MetricType,
        Value:      c.Value,
        Unit:       c.Unit,
        Status:     c.Status,
    }

    return json.Marshal(temp)
}

func (c *CnfMetric) UnmarshalJSON(data []byte) error {
    temp := struct {
        CnfId      string  `json:"cnf_id"`
        Timestamp  string  `json:"timestamp"` // Parse timestamp from string
        MetricType string  `json:"metric_type"`
        Value      float64 `json:"value"`
        Unit       string  `json:"unit"`
        Status     string  `json:"status"`
    }{}

    if err := json.Unmarshal(data, &temp); err != nil {
        return fmt.Errorf("failed to unmarshal JSON: %v", err)
    }

    // Parse timestamp string
    parsedTime, err := time.Parse(time.RFC3339, temp.Timestamp)
    if err != nil {
        return fmt.Errorf("failed to parse timestamp: %v", err)
    }

    // Populate fields
    c.CnfId = temp.CnfId
    c.Timestamp = parsedTime // Assuming Timestamp is time.Time
    c.MetricType = temp.MetricType
    c.Value = temp.Value
    c.Unit = temp.Unit
    c.Status = temp.Status

    return nil
}
