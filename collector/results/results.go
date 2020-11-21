package results

import (
	"time"
)

// Result of a monitor run
type Result struct {
	Status    int
	Timestamp int64
	Metrics   map[string]interface{}
	Errors    []string
}

// StatusOK indicates that the monitor ran, and the conditions were OK
const StatusOK = 1

// StatusFailed indicates that the monitor ran, but the conditions of the monitor indicate a failure
const StatusFailed = 0

// StatusFatal means there was a serious problem in running the monitor
const StatusFatal = -1

// NewResult creates a new Result struct
func NewResult() Result {
	now := time.Now()
	return Result{
		Status:    StatusOK,
		Timestamp: now.Unix(),
		Metrics:   make(map[string]interface{}, 0),
		Errors:    make([]string, 0),
	}
}

// AddMetric is a convenience method for adding metrics to the result
func (r *Result) AddFloat(name string, value float64) {
	r.Metrics[name] = value
}

// AddError is a convenience method for adding errors to the result
func (r *Result) AddError(err string) {
	r.Errors = append(r.Errors, err)
}

// Fail is a convenience method for failing this result
func (r *Result) Fail(err string) {
	r.Errors = append(r.Errors, err)
	r.Status = StatusFailed
}

// Fatal is a convenience method for setting this result as fatal
func (r *Result) Fatal(err string) {
	r.Errors = append(r.Errors, err)
	r.Status = StatusFatal
}
