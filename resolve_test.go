package dmn

import (
	"errors"
	"reflect"
	"testing"
)

func numericTable(t *testing.T, hitPolicy HitPolicy, allowedValues []string, rules []RuleDefinition) *CompiledTable {
	t.Helper()
	def := DecisionTableDefinition{
		TableName: "T",
		HitPolicy: hitPolicy,
		Inputs:    []ClauseDefinition{{ID: "i1", Name: "tier", Type: ClauseTypeString}},
		Outputs:   []ClauseDefinition{{ID: "o1", Name: "value", Type: ClauseTypeNumber, AllowedValues: allowedValues}},
		Rules:     rules,
	}
	ct, err := Compile(def)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	return ct
}

func TestResolve_Unique(t *testing.T) {
	rules := []RuleDefinition{
		{RuleID: "r1", InputEntries: []string{`"A"`}, OutputEntries: []string{"10"}},
	}

	t.Run("single match returns output", func(t *testing.T) {
		ct := numericTable(t, HitPolicyUnique, nil, rules)
		got, err := ct.Resolve([]RuleDefinition{rules[0]})
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		want := map[string]any{"value": 10.0}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Resolve() = %v, want %v", got, want)
		}
	})

	t.Run("multiple matches raise conflict", func(t *testing.T) {
		ct := numericTable(t, HitPolicyUnique, nil, rules)
		_, err := ct.Resolve([]RuleDefinition{rules[0], rules[0]})
		var conflict *ConflictError
		if !errors.As(err, &conflict) {
			t.Errorf("Resolve() error = %v, want *ConflictError", err)
		}
	})
}

func TestResolve_First(t *testing.T) {
	first := RuleDefinition{RuleID: "r1", InputEntries: []string{`"A"`}, OutputEntries: []string{"10"}}
	second := RuleDefinition{RuleID: "r2", InputEntries: []string{"-"}, OutputEntries: []string{"20"}}

	ct := numericTable(t, HitPolicyFirst, nil, []RuleDefinition{first, second})
	got, err := ct.Resolve([]RuleDefinition{first, second})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := map[string]any{"value": 10.0}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Resolve() = %v, want %v", got, want)
	}
}

func TestResolve_Any(t *testing.T) {
	same1 := RuleDefinition{RuleID: "r1", InputEntries: []string{"-"}, OutputEntries: []string{"10"}}
	same2 := RuleDefinition{RuleID: "r2", InputEntries: []string{"-"}, OutputEntries: []string{"10"}}
	different := RuleDefinition{RuleID: "r3", InputEntries: []string{"-"}, OutputEntries: []string{"20"}}

	t.Run("identical outputs return that output", func(t *testing.T) {
		ct := numericTable(t, HitPolicyAny, nil, []RuleDefinition{same1, same2})
		got, err := ct.Resolve([]RuleDefinition{same1, same2})
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		want := map[string]any{"value": 10.0}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Resolve() = %v, want %v", got, want)
		}
	})

	t.Run("differing outputs raise conflict", func(t *testing.T) {
		ct := numericTable(t, HitPolicyAny, nil, []RuleDefinition{same1, different})
		_, err := ct.Resolve([]RuleDefinition{same1, different})
		var conflict *ConflictError
		if !errors.As(err, &conflict) {
			t.Errorf("Resolve() error = %v, want *ConflictError", err)
		}
	})
}

func TestResolve_Priority(t *testing.T) {
	low := RuleDefinition{RuleID: "r1", InputEntries: []string{"-"}, OutputEntries: []string{"10"}}
	high := RuleDefinition{RuleID: "r2", InputEntries: []string{"-"}, OutputEntries: []string{"20"}}

	ct := numericTable(t, HitPolicyPriority, []string{"10", "20"}, []RuleDefinition{low, high})
	got, err := ct.Resolve([]RuleDefinition{low, high})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := map[string]any{"value": 20.0}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Resolve() = %v, want %v (last allowedValues element = highest priority)", got, want)
	}
}

func TestResolve_Collect(t *testing.T) {
	r1 := RuleDefinition{RuleID: "r1", InputEntries: []string{"-"}, OutputEntries: []string{"10"}}
	r2 := RuleDefinition{RuleID: "r2", InputEntries: []string{"-"}, OutputEntries: []string{"20"}}

	ct := numericTable(t, HitPolicyCollect, nil, []RuleDefinition{r1, r2})
	got, err := ct.Resolve([]RuleDefinition{r1, r2})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := []map[string]any{{"value": 10.0}, {"value": 20.0}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Resolve() = %v, want %v", got, want)
	}
}

func TestResolve_CollectNumeric(t *testing.T) {
	r1 := RuleDefinition{RuleID: "r1", InputEntries: []string{"-"}, OutputEntries: []string{"10"}}
	r2 := RuleDefinition{RuleID: "r2", InputEntries: []string{"-"}, OutputEntries: []string{"20"}}
	matched := []RuleDefinition{r1, r2}

	tests := []struct {
		hitPolicy HitPolicy
		want      any
	}{
		{HitPolicyCollectSum, 30.0},
		{HitPolicyCollectMin, 10.0},
		{HitPolicyCollectMax, 20.0},
		{HitPolicyCollectCount, 2},
	}
	for _, tt := range tests {
		t.Run(string(tt.hitPolicy), func(t *testing.T) {
			ct := numericTable(t, tt.hitPolicy, nil, matched)
			got, err := ct.Resolve(matched)
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolve_UnsupportedHitPolicy(t *testing.T) {
	r1 := RuleDefinition{RuleID: "r1", InputEntries: []string{"-"}, OutputEntries: []string{"10"}}

	for _, hp := range []HitPolicy{HitPolicyRuleOrder, HitPolicyOutputOrder} {
		t.Run(string(hp), func(t *testing.T) {
			def := DecisionTableDefinition{
				TableName: "T",
				HitPolicy: hp,
				Inputs:    []ClauseDefinition{{ID: "i1", Name: "tier", Type: ClauseTypeString}},
				Outputs:   []ClauseDefinition{{ID: "o1", Name: "value", Type: ClauseTypeNumber, AllowedValues: []string{"10"}}},
				Rules:     []RuleDefinition{r1},
			}
			ct, err := Compile(def)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			_, err = ct.Resolve([]RuleDefinition{r1})
			var unsupported *UnsupportedHitPolicyError
			if !errors.As(err, &unsupported) {
				t.Errorf("Resolve() error = %v, want *UnsupportedHitPolicyError", err)
			}
		})
	}
}
