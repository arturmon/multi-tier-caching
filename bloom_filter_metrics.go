package multi_tier_caching

import "github.com/prometheus/client_golang/prometheus"

var (
	bloomCapacityGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "bloom_filter_capacity",
			Help: "Current capacity of the Bloom filter",
		},
	)

	bloomCountGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "bloom_filter_element_count",
			Help: "Approximate number of elements in the Bloom filter",
		},
	)

	bloomHashFunctionsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "bloom_filter_hash_functions",
			Help: "Number of hash functions used in the Bloom filter",
		},
	)

	bloomFalsePositiveGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "bloom_filter_false_positive_rate",
			Help: "Estimated false positive rate of the Bloom filter",
		},
	)

	bloomLoadFactorGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "bloom_filter_load_factor",
			Help: "Current load factor of the Bloom filter",
		},
	)

	bloomLastAdjustmentGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "bloom_filter_last_adjustment_timestamp",
			Help: "Timestamp of last Bloom filter adjustment",
		},
	)
)
