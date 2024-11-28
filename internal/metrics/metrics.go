package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var DeploymentsMatchingSelector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "metadata_reflector_deployments_matching_selector",
	Help: "The total number of deployments matching the selector",
},
	[]string{"selector"},
)

func InitMetrics() {
	metrics.Registry.MustRegister(DeploymentsMatchingSelector)
}

func init() {
	InitMetrics()
}
