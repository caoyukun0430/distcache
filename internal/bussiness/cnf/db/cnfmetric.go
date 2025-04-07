package db

import (
    "context"

    cnfmetricspb "distcache/api/cnfmetricspb"
    "distcache/internal/bussiness/cnf/model"

    "gorm.io/gorm"
)

// CnfMetricDb exposes methods to access CNF metric data.
type CnfMetricDb struct {
    *gorm.DB
}

// NewCnfMetricDb creates a new instance for CNF Metrics.
func NewCnfMetricDb(ctx context.Context) *CnfMetricDb {
    return &CnfMetricDb{NewDBClient(ctx)}
}

// ShowCnfMetric retrieves a CNF Metric record by CNF ID.
func (db *CnfMetricDb) ShowCnfMetric(req *cnfmetricspb.CnfMetricRequest) (*model.CnfMetric, error) {
    var metric model.CnfMetric
    err := db.Model(&model.CnfMetric{}).Where("cnf_id=?", req.CnfId).First(&metric).Error
    if err != nil {
        loggerInstance.Errorf("Failed to retrieve CNF metric with ID %s: %v", req.CnfId, err)
        return nil, err
    }

    // Map the database model to the proto message.
    return &metric, nil
}

// CreateCnfMetric inserts a new CNF Metric record into the database.
func (db *CnfMetricDb) CreateCnfMetric(req *cnfmetricspb.CreateCnfMetricRequest) error {
    // Map the proto message to the database model.
    metric := model.CnfMetric{
        CnfId:      req.Metric.CnfId,
        Timestamp:  req.Metric.Timestamp.AsTime(),
        MetricType: req.Metric.MetricType,
        Value:      req.Metric.Value,
        Unit:       req.Metric.Unit,
        Status:     req.Metric.Status,
    }

    // Insert the metric into the database.
    err := db.Model(&model.CnfMetric{}).Create(&metric).Error
    if err != nil {
        loggerInstance.Errorf("Failed to insert CNF metric: %v", err)
        return err
    }
    return nil
}

func (db *CnfMetricDb) DeleteCnfMetric(cnfId string) error {
	err := db.DB.Where("cnf_id = ?", cnfId).Delete(&model.CnfMetric{}).Error
    if err != nil {
        loggerInstance.Errorf("Failed to DELETE CNF metric with ID %s: %v", cnfId, err)
        return err
    }
    return nil
}