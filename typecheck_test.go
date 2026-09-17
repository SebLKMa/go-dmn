package dmn

import (
	"errors"
	"testing"
	"time"
)

func TestCheckClauseType(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		typ     ClauseType
		wantErr bool
	}{
		{"string matches", "VIP", ClauseTypeString, false},
		{"string mismatch", 42.0, ClauseTypeString, true},
		{"number matches float64", 42.0, ClauseTypeNumber, false},
		{"number matches int", 42, ClauseTypeNumber, false},
		{"number mismatch", "42", ClauseTypeNumber, true},
		{"boolean matches", true, ClauseTypeBoolean, false},
		{"boolean mismatch", "true", ClauseTypeBoolean, true},
		{"date matches time.Time", time.Now(), ClauseTypeDate, false},
		{"date matches RFC3339 string", "2024-01-01T00:00:00Z", ClauseTypeDate, false},
		{"date mismatch", "not-a-date", ClauseTypeDate, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkClauseType("field", tt.value, tt.typ)
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkClauseType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				var mismatch *TypeMismatchError
				if !errors.As(err, &mismatch) {
					t.Errorf("error is not a *TypeMismatchError: %v", err)
				}
			}
		})
	}
}
