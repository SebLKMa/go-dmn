package dmn

import (
	"errors"
	"testing"
)

func discountTable(t *testing.T, hitPolicy HitPolicy, rules []RuleDefinition) *CompiledTable {
	t.Helper()
	def := DecisionTableDefinition{
		TableName: "Discount",
		HitPolicy: hitPolicy,
		Inputs:    []ClauseDefinition{{ID: "i1", Name: "customerType", Type: ClauseTypeString}},
		Outputs:   []ClauseDefinition{{ID: "o1", Name: "discount", Type: ClauseTypeNumber}},
		Rules:     rules,
	}
	ct, err := Compile(def)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	return ct
}

func TestCompiledRule_Matches(t *testing.T) {
	t.Run("all input entries true", func(t *testing.T) {
		ct := discountTable(t, HitPolicyFirst, []RuleDefinition{
			{RuleID: "r1", InputEntries: []string{`"VIP"`}, OutputEntries: []string{"10"}},
		})
		matched, err := ct.MatchingRules(EvaluationContext{"customerType": "VIP"})
		if err != nil {
			t.Fatalf("MatchingRules() error = %v", err)
		}
		if len(matched) != 1 {
			t.Fatalf("MatchingRules() = %v, want 1 match", matched)
		}
	})

	t.Run("one input entry false", func(t *testing.T) {
		ct := discountTable(t, HitPolicyFirst, []RuleDefinition{
			{RuleID: "r1", InputEntries: []string{`"VIP"`}, OutputEntries: []string{"10"}},
		})
		matches, err := ct.computeMatches(EvaluationContext{"customerType": "Regular"})
		if err != nil {
			t.Fatalf("computeMatches() error = %v", err)
		}
		if len(matches) != 0 {
			t.Fatalf("computeMatches() = %v, want no matches", matches)
		}
	})

	t.Run("empty input entries vacuously match", func(t *testing.T) {
		def := DecisionTableDefinition{
			TableName: "NoInputs",
			HitPolicy: HitPolicyFirst,
			Inputs:    []ClauseDefinition{},
			Outputs:   []ClauseDefinition{{ID: "o1", Name: "discount", Type: ClauseTypeNumber}},
			Rules: []RuleDefinition{
				{RuleID: "r1", InputEntries: []string{}, OutputEntries: []string{"5"}},
			},
		}
		ct, err := Compile(def)
		if err != nil {
			t.Fatalf("Compile() error = %v", err)
		}
		matched, err := ct.MatchingRules(EvaluationContext{})
		if err != nil {
			t.Fatalf("MatchingRules() error = %v", err)
		}
		if len(matched) != 1 {
			t.Fatalf("MatchingRules() = %v, want 1 match", matched)
		}
	})

	t.Run("wildcard rule always matches", func(t *testing.T) {
		ct := discountTable(t, HitPolicyFirst, []RuleDefinition{
			{RuleID: "r1", InputEntries: []string{"-"}, OutputEntries: []string{"0"}},
		})
		matched, err := ct.MatchingRules(EvaluationContext{"customerType": "anything"})
		if err != nil {
			t.Fatalf("MatchingRules() error = %v", err)
		}
		if len(matched) != 1 {
			t.Fatalf("MatchingRules() = %v, want 1 match", matched)
		}
	})
}

func TestMatchingRules_CompletenessError(t *testing.T) {
	ct := discountTable(t, HitPolicyFirst, []RuleDefinition{
		{RuleID: "r1", InputEntries: []string{`"VIP"`}, OutputEntries: []string{"10"}},
	})

	_, err := ct.MatchingRules(EvaluationContext{"customerType": "Regular"})
	if err == nil {
		t.Fatal("MatchingRules() = nil error, want CompletenessError")
	}
	var completeness *CompletenessError
	if !errors.As(err, &completeness) {
		t.Errorf("error is not a *CompletenessError: %v", err)
	}
}
