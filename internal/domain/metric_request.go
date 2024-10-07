package domain

// metric storage object used in requests (YP format requirement)
type MetricRequest struct {
	Type  string
	Name  string
	Value string
}
