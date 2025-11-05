package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	k8smetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const metricNamespace string = "kairos_controller"

var (
	BuildInfo prometheus.Gauge

	CounterTest = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "test_counter",
			Help:      "A test counter for kairos metrics",
		},
	)

	RequestsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "requests_created",
			Help:      "Number of access requests created",
		},
		[]string{"scope", "namespace", "subject"},
	)

	RequestsApproved = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "requests_approved",
			Help:      "Number of access requests approved",
		},
		[]string{"scope", "namespace", "subject"},
	)

	RolesGranted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "roles_granted",
			Help:      "Number of roles granted",
		},
		[]string{"scope", "namespace", "subject", "roleKind", "role"},
	)

	PermissionsGranted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "permissions_granted",
			Help:      "Number of permissions granted",
		},
		[]string{"scope", "namespace", "subject", "apiGroup", "resource", "verb"},
	)
)

func RegisterMetrics(version string) {
	BuildInfo = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   metricNamespace,
			Name:        "build_info",
			Help:        "Build information.",
			ConstLabels: prometheus.Labels{"revision": version},
		},
	)

	k8smetrics.Registry.MustRegister(BuildInfo)
	k8smetrics.Registry.MustRegister(collectors.NewBuildInfoCollector())

	k8smetrics.Registry.MustRegister(RequestsCreated)
	k8smetrics.Registry.MustRegister(RequestsApproved)
	k8smetrics.Registry.MustRegister(RolesGranted)
	k8smetrics.Registry.MustRegister(PermissionsGranted)
	k8smetrics.Registry.MustRegister(CounterTest)
}
