package multi_tier_caching

import (
	"math"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Helper function to get counter value
func getCounterValue(c prometheus.Counter) float64 {
	var metric dto.Metric
	if err := c.Write(&metric); err != nil {
		return 0
	}
	return metric.Counter.GetValue()
}

// Helper function to get sum of counter vector values
func getCounterValueVec(vec *prometheus.CounterVec) float64 {
	total := 0.0

	// Collect all metrics from the counter vector
	ch := make(chan prometheus.Metric, 10) // Buffered channel to prevent blocking
	go func() {
		vec.Collect(ch)
		close(ch)
	}()

	for metric := range ch {
		dtoMetric := &dto.Metric{}
		if err := metric.Write(dtoMetric); err == nil {
			total += dtoMetric.Counter.GetValue()
		}
	}

	return total
}

func estimateFalsePositiveRate(k uint, m uint, n uint) float64 {
	if m == 0 || n == 0 {
		return 0.0
	}
	return math.Pow(1-math.Exp(-float64(k*n)/float64(m)), float64(k))
}
