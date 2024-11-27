# Environment Variables

## Config

 - `BACKGROUND_REFLECTION_INTERVAL` (default: `5m`) - the interval of the background propagation task
 - `DEPLOYMENT_SELECTOR` - a deployment selector to limit the watched resources
should be provided in this format https://pkg.go.dev/k8s.io/apimachinery/pkg/labels#Parse
if empty, all deployments will match
 - `NAMESPACES` (comma-separated) - a comma-separated list of namespaces where to watch the deployments
if empty, all namespaces will be watched
 - `PROMETHEUS_METRICS_PORT` (default: `9090`) - the port on which the Prometheus server should be exposed
 - `HEALTH_CHECK_PORT` (default: `8083`) - the port for health checking
 - `ENABLE_LEADER_ELECTION` (default: `false`) - whether to enable leader election

