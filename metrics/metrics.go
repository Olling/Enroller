package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric definintions
var (
	ServersAdded = promauto.NewCounter(prometheus.CounterOpts{
		Subsystem: "enrolld",
		Name:      "servers_added_total",
		Help:      "The total number of added servers",
	})
)
