package service

import (
    "context"

    cnfmetricspb "distcache/api/cnfmetricspb"
    "distcache/internal/bussiness/cnf/db"
    "distcache/internal/bussiness/cnf/model"
    "distcache/internal/bussiness/cnf/ecode"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// CnfMetricsSrv implements the CNF Metrics gRPC service.
type CnfMetricsSrv struct {
    cnfmetricspb.UnimplementedCnfMetricsServiceServer
}

// NewCnfMetricsSrv initializes the CNF Metrics service and ensures the database is ready.
func NewCnfMetricsSrv() (*CnfMetricsSrv, error) {
    // Initialize the database connection and migration if required.
    if err := db.InitDB(); err != nil {
        return nil, err
    }
    return &CnfMetricsSrv{}, nil
}

// ShowCnfMetric handles the ShowCnfMetric RPC, retrieving a CNF metric by CNF ID.
func (s *CnfMetricsSrv) ShowCnfMetric(ctx context.Context, req *cnfmetricspb.CnfMetricRequest) (*cnfmetricspb.CnfMetricResponse, error) {
    resp := &cnfmetricspb.CnfMetricResponse{}

    // Fetch the CNF metric from the database.
    cnfMetric, err := db.NewCnfMetricDb(ctx).ShowCnfMetric(req)
    if err != nil {
        // Log the error and set an appropriate error code.
        resp.Code = ecode.ERROR
        return nil, err
    }

    // Map the database model to the protobuf response format.
    resp.Metric = &cnfmetricspb.CnfMetric{
        CnfId:      cnfMetric.CnfId,
        Timestamp:  timestamppb.New(cnfMetric.Timestamp),
        MetricType: cnfMetric.MetricType,
        Value:      cnfMetric.Value,
        Unit:       cnfMetric.Unit,
        Status:     cnfMetric.Status,
    }
    resp.Code = ecode.SUCCESS
    return resp, nil
}

// CreateCnfMetric handles the CreateCnfMetric RPC, inserting a new CNF metric into the database.
func (s *CnfMetricsSrv) CreateCnfMetric(ctx context.Context, req *cnfmetricspb.CreateCnfMetricRequest) (*cnfmetricspb.CnfMetricResponse, error) {
    resp := &cnfmetricspb.CnfMetricResponse{}

    // Map the protobuf request to the database model.
    cnfMetric := model.CnfMetric{
        CnfId:      req.Metric.CnfId,
        Timestamp:  req.Metric.Timestamp.AsTime(),
        MetricType: req.Metric.MetricType,
        Value:      req.Metric.Value,
        Unit:       req.Metric.Unit,
        Status:     req.Metric.Status,
    }

    // Insert the CNF metric into the database.
    err := db.NewCnfMetricDb(ctx).CreateCnfMetric(cnfMetric)
    if err != nil {
        // Log the error and set an appropriate error code.
        resp.Code = ecode.ERROR
        return nil, err
    }

    // Return the created metric in the response.
    resp.Metric = req.Metric
    resp.Code = ecode.SUCCESS
    return resp, nil
}
