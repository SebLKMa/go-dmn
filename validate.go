package dmn

import "fmt"

// Validate performs design-time validation of a decision table definition
// per spec.md section 4.1: structural congruence between rule entry counts
// and declared clauses, and a populated allowedValues for the Priority hit
// policy.
func Validate(def DecisionTableDefinition) error {
	if err := validateRequiredFields(def); err != nil {
		return err
	}
	if err := validateStructuralCongruence(def); err != nil {
		return err
	}
	if err := validatePriorityBound(def); err != nil {
		return err
	}
	return nil
}

func validateRequiredFields(def DecisionTableDefinition) error {
	switch {
	case def.TableName == "":
		return &ValidationError{Message: "tableName is required"}
	case def.HitPolicy == "":
		return &ValidationError{Message: "hitPolicy is required"}
	}
	return nil
}

func validateStructuralCongruence(def DecisionTableDefinition) error {
	for _, rule := range def.Rules {
		if len(rule.InputEntries) != len(def.Inputs) {
			return &ValidationError{Message: fmt.Sprintf(
				"rule %q has %d inputEntries, want %d to match table inputs",
				rule.RuleID, len(rule.InputEntries), len(def.Inputs),
			)}
		}
		if len(rule.OutputEntries) != len(def.Outputs) {
			return &ValidationError{Message: fmt.Sprintf(
				"rule %q has %d outputEntries, want %d to match table outputs",
				rule.RuleID, len(rule.OutputEntries), len(def.Outputs),
			)}
		}
	}
	return nil
}

func validatePriorityBound(def DecisionTableDefinition) error {
	if def.HitPolicy != HitPolicyPriority && def.HitPolicy != HitPolicyOutputOrder {
		return nil
	}
	for _, out := range def.Outputs {
		if len(out.AllowedValues) == 0 {
			return &ValidationError{Message: fmt.Sprintf(
				"hitPolicy %q requires a non-empty allowedValues on output %q",
				def.HitPolicy, out.Name,
			)}
		}
	}
	return nil
}
