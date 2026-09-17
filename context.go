package dmn

// EvaluationContext maps input clause names (or IDs) to the value supplied
// for that input. Accepted dynamic value types are string, float64, int,
// bool, and time.Time (or an RFC3339 date/time string) for "date" inputs.
type EvaluationContext map[string]any

// EvaluationRequest is the payload passed into the engine to run an
// evaluation, as described in spec.md section 2.2.
type EvaluationRequest struct {
	Context EvaluationContext `json:"context"`
}
