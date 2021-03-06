package main

import (
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
)

// aggClient is a custom HTTP client for use with
// WeaveWorks Prometheus Aggregation PushGateway
type aggClient struct {
	// We can use the stdlib http.Client
	// and essentially override Do so
	// embed the client here
	*http.Client
}

// Do points to the aggregation pushgateway rather than the standard pushgateway.
func (a *aggClient) Do(req *http.Request) (resp *http.Response, err error) {
	spew.Dump(req)
	return a.Client.Do(req)
}

func pushHistogram(c prometheus.Collector) error {
	return push.New("localhost", "my_job").
		Collector(c).
		Client(&aggClient{&http.Client{}}).
		Format(expfmt.FmtText).
		Push()
}

func main() {
	histogramOptions := prometheus.HistogramOpts{
		Name:      "invocation",
		Namespace: "cluster",
		Subsystem: "installation",

		Help: "cluster_isntallation_invocation represents a call to the installer.",

		ConstLabels: prometheus.Labels{"os": "linux", "command": "create", "target": "cluster"},

		Buckets: []float64{0, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60},
	}

	clusterInstallationInvocation := prometheus.NewHistogram(histogramOptions)

	//Add a value representing how long installation took
	clusterInstallationInvocation.Observe(20)

	err := pushHistogram(clusterInstallationInvocation)
	if err != nil {
		fmt.Println(err)
	}
}

