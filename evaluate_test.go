package dmn

import (
	"reflect"
	"testing"
)

func TestEvaluate_EndToEnd(t *testing.T) {
	t.Run("Unique", func(t *testing.T) {
		def := DecisionTableDefinition{
			TableName: "Discount", HitPolicy: HitPolicyUnique,
			Inputs:  []ClauseDefinition{{ID: "i1", Name: "customerType", Type: ClauseTypeString}},
			Outputs: []ClauseDefinition{{ID: "o1", Name: "discount", Type: ClauseTypeNumber}},
			Rules: []RuleDefinition{
				{RuleID: "r1", InputEntries: []string{`"VIP"`}, OutputEntries: []string{"20"}},
				{RuleID: "r2", InputEntries: []string{`"Regular"`}, OutputEntries: []string{"0"}},
			},
		}
		got, err := Evaluate(def, EvaluationContext{"customerType": "VIP"})
		if err != nil {
			t.Fatalf("Evaluate() error = %v", err)
		}
		want := map[string]any{"discount": 20.0}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Evaluate() = %v, want %v", got, want)
		}
	})

	t.Run("First", func(t *testing.T) {
		def := DecisionTableDefinition{
			TableName: "AgeGroup", HitPolicy: HitPolicyFirst,
			Inputs:  []ClauseDefinition{{ID: "i1", Name: "age", Type: ClauseTypeNumber}},
			Outputs: []ClauseDefinition{{ID: "o1", Name: "group", Type: ClauseTypeString}},
			Rules: []RuleDefinition{
				{RuleID: "r1", InputEntries: []string{"[18..65]"}, OutputEntries: []string{`"adult"`}},
				{RuleID: "r2", InputEntries: []string{"-"}, OutputEntries: []string{`"other"`}},
			},
		}
		got, err := Evaluate(def, EvaluationContext{"age": 30.0})
		if err != nil {
			t.Fatalf("Evaluate() error = %v", err)
		}
		want := map[string]any{"group": "adult"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Evaluate() = %v, want %v", got, want)
		}
	})

	t.Run("Collect sum", func(t *testing.T) {
		def := DecisionTableDefinition{
			TableName: "Fees", HitPolicy: HitPolicyCollectSum,
			Inputs:  []ClauseDefinition{{ID: "i1", Name: "tier", Type: ClauseTypeString}},
			Outputs: []ClauseDefinition{{ID: "o1", Name: "fee", Type: ClauseTypeNumber}},
			Rules: []RuleDefinition{
				{RuleID: "r1", InputEntries: []string{`"Gold", "Silver"`}, OutputEntries: []string{"5"}},
				{RuleID: "r2", InputEntries: []string{"-"}, OutputEntries: []string{"1"}},
			},
		}
		got, err := Evaluate(def, EvaluationContext{"tier": "Gold"})
		if err != nil {
			t.Fatalf("Evaluate() error = %v", err)
		}
		if got != 6.0 {
			t.Errorf("Evaluate() = %v, want 6.0", got)
		}
	})

	t.Run("no match raises CompletenessError", func(t *testing.T) {
		def := DecisionTableDefinition{
			TableName: "Strict", HitPolicy: HitPolicyUnique,
			Inputs:  []ClauseDefinition{{ID: "i1", Name: "tier", Type: ClauseTypeString}},
			Outputs: []ClauseDefinition{{ID: "o1", Name: "fee", Type: ClauseTypeNumber}},
			Rules: []RuleDefinition{
				{RuleID: "r1", InputEntries: []string{`"Gold"`}, OutputEntries: []string{"5"}},
			},
		}
		_, err := Evaluate(def, EvaluationContext{"tier": "Bronze"})
		if err == nil {
			t.Fatal("Evaluate() error = nil, want CompletenessError")
		}
	})

	t.Run("invalid definition is rejected before evaluation", func(t *testing.T) {
		def := DecisionTableDefinition{
			TableName: "Bad", HitPolicy: HitPolicyFirst,
			Inputs:  []ClauseDefinition{{ID: "i1", Name: "tier", Type: ClauseTypeString}},
			Outputs: []ClauseDefinition{{ID: "o1", Name: "fee", Type: ClauseTypeNumber}},
			Rules: []RuleDefinition{
				{RuleID: "r1", InputEntries: []string{`"Gold"`, "extra"}, OutputEntries: []string{"5"}},
			},
		}
		if _, err := Evaluate(def, EvaluationContext{"tier": "Gold"}); err == nil {
			t.Fatal("Evaluate() error = nil, want validation error")
		}
	})
}
