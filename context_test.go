package dmn

import (
	"encoding/json"
	"testing"
)

func TestEvaluationRequest_DecodesFromJSON(t *testing.T) {
	const payload = `{"context": {"customerType": "VIP", "age": 42, "active": true}}`

	var req EvaluationRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got, want := req.Context["customerType"], "VIP"; got != want {
		t.Errorf("Context[customerType] = %v, want %v", got, want)
	}
	if got, want := req.Context["age"], float64(42); got != want {
		t.Errorf("Context[age] = %v, want %v", got, want)
	}
	if got, want := req.Context["active"], true; got != want {
		t.Errorf("Context[active] = %v, want %v", got, want)
	}
}
