package pkg

type MetricType string
type MetricName string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

func (m MetricType) IsValid() bool {
	switch m {
	case Counter, Gauge:
		return true
	default:
		return false
	}
}
