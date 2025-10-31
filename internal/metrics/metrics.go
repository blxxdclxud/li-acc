package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// File processing counters
	FileProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "liacc_file_processed_total",
			Help: "Total number of files processed",
		},
		// `result=("success"|"failure")`, `error_stage` (empty or one of the stage names below), `type=("email"|"payers")`
		[]string{"result", "error_stage", "type"},
	)

	// End-to-end latency histograms
	PayersFileLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "liacc_payers_file_total_duration_seconds",
			Help:    "End-to-end latency per payers file",
			Buckets: []float64{0.5, 1, 2, 5, 10, 20, 30, 60, 120},
		},
		[]string{"result"},
	)

	EmailsFileLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "liacc_emails_file_total_duration_seconds",
			Help:    "End-to-end latency per emails file",
			Buckets: []float64{0.5, 1, 2, 5, 10, 20, 30, 60, 120},
		},
		[]string{"result"},
	)
)

// Parse payers stage
var (
	ParsePayersDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "liacc_stage_parse_payers_duration_seconds",
			Help:    "Parse payers stage duration",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"result"},
	)

	ParsePayersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "liacc_stage_parse_payers_total",
			Help: "Total parse payers operations",
		},
		[]string{"result", "error_type"},
	)

	PayersParsedCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "liacc_payers_parsed_count",
			Help:    "Number of payers parsed per file",
			Buckets: []float64{10, 50, 100, 500, 1000, 5000, 10000},
		},
		[]string{"result"},
	)
)

// Parse org settings
var (
	ParseOrgDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "liacc_stage_parse_org_duration_seconds",
			Help:    "Parse organization settings duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)

	ParseOrgTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "liacc_stage_parse_org_total",
			Help: "Total parse org operations",
		},
		[]string{"result", "error_type"},
	)
)

// Generate receipts
var (
	GenerateReceiptsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "liacc_stage_generate_receipts_duration_seconds",
			Help:    "Generate receipts duration",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"result"},
	)

	GenerateReceiptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "liacc_stage_generate_receipts_total",
			Help: "Total generate receipts operations",
		},
		[]string{"result", "error_type"},
	)

	ReceiptsGeneratedCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "liacc_receipts_generated_count",
			Help:    "Number of receipts generated per operation",
			Buckets: []float64{10, 50, 100, 500, 1000, 5000},
		},
		[]string{"result"}, // Removed file_id to avoid high cardinality
	)
)

// Send mails
var (
	SendMailsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "liacc_stage_send_mails_duration_seconds",
			Help:    "Send mails stage duration",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"result"},
	)

	SendMailsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "liacc_stage_send_mails_total",
			Help: "Total send mails operations",
		},
		[]string{"result", "error_type"},
	)

	MailsSentCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "liacc_mails_sent_count",
			Help:    "Number of mails sent per operation",
			Buckets: []float64{10, 50, 100, 500, 1000, 5000},
		},
		// `status=("success"|"partial"|"failure")`
		[]string{"status"},
	)

	// for partial success cases
	MailsFailedCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "liacc_mails_failed_count",
			Help:    "Number of mails failed per operation",
			Buckets: []float64{10, 50, 100, 500, 1000, 5000},
		},
		[]string{"status"},
	)
)
