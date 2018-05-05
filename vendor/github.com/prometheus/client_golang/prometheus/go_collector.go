package prometheus

import (
	"runtime"
)

type goCollector struct {
	goroutinesDesc *Desc
	threadsDesc    *Desc
	gcDesc         *Desc

	// metrics to describe and collect
	metrics memStatsMetrics
}

// NewGoCollector returns a collector which exports metrics about the current
// go process.
func NewGoCollector() Collector {
	return &goCollector{}
}

// Describe returns all descriptions of the collector.
func (c *goCollector) Describe(ch chan<- *Desc) {

}

// Collect returns the current state of all metrics of the collector.
func (c *goCollector) Collect(ch chan<- Metric) {

}

// memStatsMetrics provide description, value, and value type for memstat metrics.
type memStatsMetrics []struct {
	desc    *Desc
	eval    func(*runtime.MemStats) float64
	valType ValueType
}
