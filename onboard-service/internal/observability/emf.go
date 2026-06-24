package observability

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

const (
	Namespace = "CardOnboarding"
	Component = "onboard-service"

	UnitCount        = "Count"
	UnitMilliseconds = "Milliseconds"

	MetricRequestCount                 = "onboard_service.request.count"
	MetricSuccessCount                 = "onboard_service.success.count"
	MetricFailedCount                  = "onboard_service.failed.count"
	MetricCustomerRegisterSuccessCount = "onboard_service.customer_register_success.count"
	MetricCustomerRegisterFailedCount  = "onboard_service.customer_register_failed.count"
	MetricInterestDetailsSuccessCount  = "onboard_service.interest_details_success.count"
	MetricInterestDetailsFailedCount   = "onboard_service.interest_details_failed.count"
	MetricResumeCount                  = "onboard_service.resume.count"
	MetricDBWriteFailedCount           = "onboard_service.db_write_failed.count"
	MetricResponseTimeMilliseconds     = "onboard_service.response_time_ms"
)

type Metric struct {
	Name  string
	Value float64
	Unit  string
}

type Fields struct {
	Environment   string
	Component     string
	CorrelationID string
	JobID         string
	RecordID      string
	CustomerID    string
	SourceFile    string
	RowNumber     int
	Step          string
	Status        string
	DurationMs    int64
	ErrorCode     string
	ErrorMessage  string
	Operation     string
	TableName     string
}

type awsMetadata struct {
	Timestamp         int64                    `json:"Timestamp"`
	CloudWatchMetrics []cloudWatchMetricConfig `json:"CloudWatchMetrics"`
}

type cloudWatchMetricConfig struct {
	Namespace  string             `json:"Namespace"`
	Dimensions [][]string         `json:"Dimensions"`
	Metrics    []cloudWatchMetric `json:"Metrics"`
}

type cloudWatchMetric struct {
	Name string `json:"Name"`
	Unit string `json:"Unit,omitempty"`
}

func NewFields() Fields {
	return Fields{
		Environment: os.Getenv("ENVIRONMENT_NAME"),
		Component:   Component,
	}
}

func LogMetric(metric Metric, fields Fields) {
	message, err := BuildMetric(metric, fields, time.Now())
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println(message)
}

func BuildMetric(metric Metric, fields Fields, timestamp time.Time) (string, error) {
	if fields.Component == "" {
		fields.Component = Component
	}

	event := map[string]any{
		"_aws": awsMetadata{
			Timestamp: timestamp.UnixMilli(),
			CloudWatchMetrics: []cloudWatchMetricConfig{
				{
					Namespace:  Namespace,
					Dimensions: [][]string{{"Environment", "Component"}},
					Metrics: []cloudWatchMetric{
						{
							Name: metric.Name,
							Unit: metric.Unit,
						},
					},
				},
			},
		},
		"Environment": fields.Environment,
		"Component":   fields.Component,
		metric.Name:   metric.Value,
	}

	addString(event, "correlationId", fields.CorrelationID)
	addString(event, "jobId", fields.JobID)
	addString(event, "recordId", fields.RecordID)
	addString(event, "customerId", fields.CustomerID)
	addString(event, "sourceFile", fields.SourceFile)
	addInt(event, "rowNumber", fields.RowNumber)
	addString(event, "step", fields.Step)
	addString(event, "status", fields.Status)
	addInt64(event, "durationMs", fields.DurationMs)
	addString(event, "errorCode", fields.ErrorCode)
	addString(event, "errorMessage", fields.ErrorMessage)
	addString(event, "operation", fields.Operation)
	addString(event, "tableName", fields.TableName)

	payload, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}

func LogCount(name string, fields Fields) {
	LogMetric(Metric{Name: name, Value: 1, Unit: UnitCount}, fields)
}

func LogDuration(name string, duration time.Duration, fields Fields) {
	fields.DurationMs = duration.Milliseconds()
	LogMetric(Metric{Name: name, Value: float64(duration.Milliseconds()), Unit: UnitMilliseconds}, fields)
}

func addString(event map[string]any, key string, value string) {
	if value != "" {
		event[key] = value
	}
}

func addInt(event map[string]any, key string, value int) {
	if value != 0 {
		event[key] = value
	}
}

func addInt64(event map[string]any, key string, value int64) {
	if value != 0 {
		event[key] = value
	}
}
