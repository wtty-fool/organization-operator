package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	reconcileErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "organization_operator_reconcile_errors_total",
			Help: "The total number of reconciliation errors",
		},
	)

	reconcileDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "organization_operator_reconcile_duration_seconds",
			Help:    "The duration of reconciliation operations",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
		},
	)

	organizationCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "organization_operator_organizations_count",
			Help: "The current number of organizations",
		},
	)

	namespacesExist = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "organization_operator_namespaces_exist",
			Help: "Whether the namespace associated with an organization exists (1) or not (0)",
		},
		[]string{"organization"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(
		reconcileErrors,
		reconcileDuration,
		organizationCount,
		namespacesExist,
	)
}
