package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	k8smetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const metricNamespace string = "jitaccess"
const metricLabelScope string = "scope"
const metricLabelTargetNamespace string = "target_namespace"
const metricLabelSubject string = "subject"

var (
	BuildInfo prometheus.Gauge

	RequestsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "requests_created",
			Help:      "Number of access requests created",
		},
		[]string{metricLabelScope, metricLabelTargetNamespace, metricLabelSubject},
	)

	RequestsApproved = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "requests_approved",
			Help:      "Number of access requests approved",
		},
		[]string{metricLabelScope, metricLabelTargetNamespace, metricLabelSubject},
	)

	RequestStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "request_status",
			Help:      "Status of access requests (0: pending, 1: approved, 2: denied)",
		},
		[]string{metricLabelScope, metricLabelTargetNamespace, "request", metricLabelSubject},
	)

	RolesGranted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "roles_granted",
			Help:      "Number of roles granted",
		},
		[]string{metricLabelScope, metricLabelTargetNamespace, metricLabelSubject, "roleKind", "role"},
	)

	PermissionsGranted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "permissions_granted",
			Help:      "Number of permissions granted",
		},
		[]string{metricLabelScope, metricLabelTargetNamespace, metricLabelSubject, "apiGroup", "resource", "verb", "resourceName"},
	)

	GrantDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "grant_duration_seconds",
			Help:      "Duration of grants in seconds",
		},
		[]string{metricLabelScope, metricLabelTargetNamespace, "grant", metricLabelSubject},
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
	k8smetrics.Registry.MustRegister(RequestStatus)
	k8smetrics.Registry.MustRegister(RolesGranted)
	k8smetrics.Registry.MustRegister(PermissionsGranted)

	k8smetrics.Registry.MustRegister(GrantDuration)
}
