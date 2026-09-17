package dmn

import "testing"

func baseValidDefinition() DecisionTableDefinition {
	return DecisionTableDefinition{
		TableName: "T",
		HitPolicy: HitPolicyUnique,
		Inputs:    []ClauseDefinition{{ID: "i1", Name: "age", Type: ClauseTypeNumber}},
		Outputs:   []ClauseDefinition{{ID: "o1", Name: "tier", Type: ClauseTypeString}},
		Rules: []RuleDefinition{
			{RuleID: "r1", InputEntries: []string{"[18..65]"}, OutputEntries: []string{"\"adult\""}},
		},
	}
}

func TestValidateStructuralCongruence(t *testing.T) {
	t.Run("matching lengths pass", func(t *testing.T) {
		if err := Validate(baseValidDefinition()); err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
	})

	t.Run("mismatched inputEntries fails", func(t *testing.T) {
		def := baseValidDefinition()
		def.Rules[0].InputEntries = []string{"[18..65]", "extra"}
		if err := Validate(def); err == nil {
			t.Fatal("Validate() = nil, want error for mismatched inputEntries length")
		}
	})

	t.Run("mismatched outputEntries fails", func(t *testing.T) {
		def := baseValidDefinition()
		def.Rules[0].OutputEntries = []string{}
		if err := Validate(def); err == nil {
			t.Fatal("Validate() = nil, want error for mismatched outputEntries length")
		}
	})
}

func TestValidateRequiredFields(t *testing.T) {
	t.Run("valid definition passes", func(t *testing.T) {
		if err := Validate(baseValidDefinition()); err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
	})

	t.Run("missing tableName fails", func(t *testing.T) {
		def := baseValidDefinition()
		def.TableName = ""
		if err := Validate(def); err == nil {
			t.Fatal("Validate() = nil, want error for missing tableName")
		}
	})

	t.Run("missing hitPolicy fails", func(t *testing.T) {
		def := baseValidDefinition()
		def.HitPolicy = ""
		if err := Validate(def); err == nil {
			t.Fatal("Validate() = nil, want error for missing hitPolicy")
		}
	})
}

func TestValidatePriorityBound(t *testing.T) {
	t.Run("priority with allowedValues passes", func(t *testing.T) {
		def := baseValidDefinition()
		def.HitPolicy = HitPolicyPriority
		def.Outputs[0].AllowedValues = []string{"child", "adult"}
		if err := Validate(def); err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
	})

	t.Run("priority with missing allowedValues fails", func(t *testing.T) {
		def := baseValidDefinition()
		def.HitPolicy = HitPolicyPriority
		def.Outputs[0].AllowedValues = nil
		if err := Validate(def); err == nil {
			t.Fatal("Validate() = nil, want error for missing allowedValues")
		}
	})

	t.Run("priority with empty allowedValues fails", func(t *testing.T) {
		def := baseValidDefinition()
		def.HitPolicy = HitPolicyPriority
		def.Outputs[0].AllowedValues = []string{}
		if err := Validate(def); err == nil {
			t.Fatal("Validate() = nil, want error for empty allowedValues")
		}
	})

	t.Run("non-priority hit policy skips check", func(t *testing.T) {
		def := baseValidDefinition()
		def.HitPolicy = HitPolicyFirst
		def.Outputs[0].AllowedValues = nil
		if err := Validate(def); err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
	})
}
