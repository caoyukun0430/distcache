package db

import (
    "context"
    "errors"
    "fmt"
    "math/rand"
    "os"
    "time"

    "distcache/config"
    "distcache/pkg/common/logger"
    "distcache/internal/bussiness/cnf/model"

    cnfmetricspb "distcache/api/cnfmetricspb"

    "google.golang.org/protobuf/types/known/timestamppb"

    "gorm.io/driver/mysql" // Gorm MySQL driver
    "gorm.io/gorm"         // Gorm ORM
    "gorm.io/gorm/schema"  // For Gorm naming strategy
)

var (
    _db    *gorm.DB
    loggerInstance = logger.NewLogger()
)

type DBConfig struct {
	Host         string
	Port         string
	Database     string
	Username     string
	Password     string
	Charset      string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
}

// InitDB initializes the database connection
func InitDB() error {
	cfg := DBConfig{
		Host:         config.Conf.Mysql.Host,
		Port:         config.Conf.Mysql.Port,
		Database:     config.Conf.Mysql.Database,
		Username:     config.Conf.Mysql.UserName,
		Password:     config.Conf.Mysql.Password,
		Charset:      config.Conf.Mysql.Charset,
		MaxLifetime:  time.Hour,
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		loggerInstance.Errorf("failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		loggerInstance.Errorf("failed to get sql.DB: %v", err)
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

	_db = db
	ititData()

	return nil
}

func NewDBClient(ctx context.Context) *gorm.DB {
	if _db == nil {
		panic("database not initialized")
	}
	return _db.WithContext(ctx)
}

// handles the database init for the CNF metrics table.
func ititData() {
    // Don't repeat if the table already exists
    if IsHasTable("cnfMetric") {
        loggerInstance.Warnln("Table cnfMetric already exists, skipping migration.")
        return
    }

    // Migrate the CNF metrics table using Gorm
    err := _db.Set("gorm:table_options", "charset=utf8mb4").AutoMigrate(&model.CnfMetric{})
    if err != nil {
        loggerInstance.Errorf("Failed to register table cnfMetric: %v", err)
        os.Exit(1)
    }

    // Initialize test data after successful table creation
    InitializeCnfMetricTestData()

    loggerInstance.Infoln("Table cnfMetric successfully registered and test data initialized.")
}

// IsHasTable checks if a table exists in the database.
func IsHasTable(tableName string) bool {
    return _db.Migrator().HasTable(tableName)
}

// InitializeCnfMetricTestData populates the CNF metrics table with test data.
func InitializeCnfMetricTestData() {
    db := NewCnfMetricDb(context.Background())

    // Define test CNF metric types and their base values
    testData := []struct {
        CnfId      string
        MetricType string
        BaseValue  float64
        Unit       string
        Status     string
    }{
        {"CNF-001", "Ingress Latency", 20.0, "ms", "Normal"},
        {"CNF-002", "Service Response Time", 50.0, "ms", "Normal"},
        {"CNF-003", "Container CPU Utilization", 85.0, "%", "Warning"},
        {"CNF-004", "Egress Throughput", 2.3, "Gbps", "Normal"},
        {"CNF-005", "Packet Drop Rate", 1.2, "%", "Warning"},
        {"CNF-006", "Memory Usage", 78.0, "%", "Normal"},
        {"CNF-007", "Container Restart Count", 3.0, "restarts", "Info"},
        {"CNF-008", "Service Mesh Latency", 15.0, "ms", "Normal"},
        {"CNF-009", "API Gateway Throughput", 1.8, "Gbps", "Normal"},
        {"CNF-010", "Security Inspection Rate", 3000.0, "packets/sec", "Normal"},
    }

    // Randomly generate metrics
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    now := time.Now()

    for _, data := range testData {
		if err := db.DeleteCnfMetric(data.CnfId); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			loggerInstance.Errorf("Failed to delete existing CNF metric with ID %s: %v", data.CnfId, err)
			continue // If deletion fails for unexpected reasons, skip this entry
		}
		// Add a small random variation to the base value
		variation := (r.Float64() - 0.5) * 0.2 * data.BaseValue
		value := data.BaseValue + variation
	
		// Create the Protobuf CreateCnfMetricRequest type
		metricRequest := &cnfmetricspb.CreateCnfMetricRequest{
			Metric: &cnfmetricspb.CnfMetric{
				CnfId:      data.CnfId,
				Timestamp:  timestamppb.New(now.Add(-time.Duration(r.Intn(100)) * time.Minute)), // Random historical timestamps
				MetricType: data.MetricType,
				Value:      value,
				Unit:       data.Unit,
				Status:     data.Status,
			},
		}
	
		// Insert the metric into the database using the Protobuf request type
		if err := db.CreateCnfMetric(metricRequest); err != nil {
			loggerInstance.Errorf("Failed to create CNF metric with ID %s: %v", data.CnfId, err)
		}
	}
	
	// Dynamically generate additional test data to ensure at least 100 metrics
	for i := 11; i <= 100; i++ { // Create CNF-011 to CNF-100
		cnfId := fmt.Sprintf("CNF-%03d", i) // Format as CNF-011, CNF-012, ..., CNF-100
		metricType := fmt.Sprintf("Generated Metric Type %03d", i)
		baseValue := float64(r.Intn(100) + 1) // Random base value between 1 and 100
		unit := "unit"                        // Default unit for generated data
		status := "Normal"                   // Default status for generated data
	
		// Add small random variation to the base value
		variation := (r.Float64() - 0.5) * 0.2 * baseValue
		value := baseValue + variation
		// Delete any existing record with the same cnfId
		if err := db.DeleteCnfMetric(cnfId); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			loggerInstance.Errorf("Failed to delete existing CNF metric with ID %s: %v", cnfId, err)
			continue // Skip insertion if deletion fails unexpectedly
		}
	
		// Create the Protobuf CreateCnfMetricRequest type
		metricRequest := &cnfmetricspb.CreateCnfMetricRequest{
			Metric: &cnfmetricspb.CnfMetric{
				CnfId:      cnfId,
				Timestamp:  timestamppb.New(now.Add(-time.Duration(r.Intn(500)) * time.Minute)), // Random timestamp within ~8 hours
				MetricType: metricType,
				Value:      value,
				Unit:       unit,
				Status:     status,
			},
		}
	
		// Insert the generated metric into the database using the Protobuf request type
		if err := db.CreateCnfMetric(metricRequest); err != nil {
			loggerInstance.Errorf("Failed to create CNF metric with ID %s: %v", cnfId, err)
		}
	}
    loggerInstance.Infoln("Test data for CNF metrics initialized successfully.")
}
