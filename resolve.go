package dmn

import (
	"fmt"
	"reflect"
)

// outputValues parses a rule's OutputEntries into a map keyed by output
// clause name, using the same literal grammar as exact-literal input
// entries (spec.md section 3.1 "Exact Literals").
func outputValues(rule RuleDefinition, outputs []ClauseDefinition) (map[string]any, error) {
	values := make(map[string]any, len(outputs))
	for i, out := range outputs {
		v, err := parseLiteral(rule.OutputEntries[i], out.Type)
		if err != nil {
			return nil, fmt.Errorf("dmn: invalid output entry for rule %q, output %q: %w", rule.RuleID, out.Name, err)
		}
		values[out.Name] = v
	}
	return values, nil
}

// singleNumericOutput extracts the value of the sole output clause for
// hit policies that operate on one numeric column (P ranking, C+, C<, C>).
// Multi-output tables are outside the scope of section 3.3's examples for
// these policies.
func singleNumericOutput(rule RuleDefinition, outputs []ClauseDefinition, hitPolicy HitPolicy) (float64, error) {
	if len(outputs) != 1 {
		return 0, fmt.Errorf("dmn: hit policy %q requires exactly one output clause, table has %d", hitPolicy, len(outputs))
	}
	v, err := parseLiteral(rule.OutputEntries[0], outputs[0].Type)
	if err != nil {
		return 0, fmt.Errorf("dmn: invalid output entry for rule %q: %w", rule.RuleID, err)
	}
	f, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("dmn: hit policy %q requires a numeric output, got %T", hitPolicy, v)
	}
	return f, nil
}

// Resolve applies the decision table's hit policy to a set of matched
// rules, per spec.md section 3.3.
func (t *CompiledTable) Resolve(matched []RuleDefinition) (any, error) {
	switch t.def.HitPolicy {
	case HitPolicyUnique:
		return t.resolveUnique(matched)
	case HitPolicyFirst:
		return t.resolveFirst(matched)
	case HitPolicyAny:
		return t.resolveAny(matched)
	case HitPolicyPriority:
		return t.resolvePriority(matched)
	case HitPolicyCollect:
		return t.resolveCollect(matched)
	case HitPolicyCollectSum, HitPolicyCollectMin, HitPolicyCollectMax, HitPolicyCollectCount:
		return t.resolveCollectNumeric(matched)
	case HitPolicyRuleOrder, HitPolicyOutputOrder:
		return nil, &UnsupportedHitPolicyError{HitPolicy: t.def.HitPolicy}
	default:
		return nil, &UnsupportedHitPolicyError{HitPolicy: t.def.HitPolicy}
	}
}

func (t *CompiledTable) resolveUnique(matched []RuleDefinition) (any, error) {
	if len(matched) > 1 {
		return nil, &ConflictError{TableName: t.def.TableName, HitPolicy: t.def.HitPolicy, Message: "more than one rule matched"}
	}
	return outputValues(matched[0], t.def.Outputs)
}

func (t *CompiledTable) resolveFirst(matched []RuleDefinition) (any, error) {
	return outputValues(matched[0], t.def.Outputs)
}

func (t *CompiledTable) resolveAny(matched []RuleDefinition) (any, error) {
	first, err := outputValues(matched[0], t.def.Outputs)
	if err != nil {
		return nil, err
	}
	for _, rule := range matched[1:] {
		v, err := outputValues(rule, t.def.Outputs)
		if err != nil {
			return nil, err
		}
		if !reflect.DeepEqual(first, v) {
			return nil, &ConflictError{TableName: t.def.TableName, HitPolicy: t.def.HitPolicy, Message: "matched rules produced different outputs"}
		}
	}
	return first, nil
}

func (t *CompiledTable) resolvePriority(matched []RuleDefinition) (any, error) {
	if len(t.def.Outputs) != 1 {
		return nil, fmt.Errorf("dmn: hit policy %q requires exactly one output clause, table has %d", HitPolicyPriority, len(t.def.Outputs))
	}
	allowed := t.def.Outputs[0].AllowedValues
	rank := make(map[string]int, len(allowed))
	for i, v := range allowed {
		rank[v] = i
	}

	best := matched[0]
	bestRank := -1
	for _, rule := range matched {
		lit, err := parseLiteral(rule.OutputEntries[0], t.def.Outputs[0].Type)
		if err != nil {
			return nil, fmt.Errorf("dmn: invalid output entry for rule %q: %w", rule.RuleID, err)
		}
		key := fmt.Sprintf("%v", lit)
		r, ok := rank[key]
		if !ok {
			return nil, fmt.Errorf("dmn: output %q of rule %q is not present in allowedValues", key, rule.RuleID)
		}
		if r > bestRank {
			best, bestRank = rule, r
		}
	}
	return outputValues(best, t.def.Outputs)
}

func (t *CompiledTable) resolveCollect(matched []RuleDefinition) (any, error) {
	results := make([]map[string]any, 0, len(matched))
	for _, rule := range matched {
		v, err := outputValues(rule, t.def.Outputs)
		if err != nil {
			return nil, err
		}
		results = append(results, v)
	}
	return results, nil
}

func (t *CompiledTable) resolveCollectNumeric(matched []RuleDefinition) (any, error) {
	values := make([]float64, 0, len(matched))
	for _, rule := range matched {
		v, err := singleNumericOutput(rule, t.def.Outputs, t.def.HitPolicy)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	switch t.def.HitPolicy {
	case HitPolicyCollectSum:
		var sum float64
		for _, v := range values {
			sum += v
		}
		return sum, nil
	case HitPolicyCollectMin:
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min, nil
	case HitPolicyCollectMax:
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max, nil
	case HitPolicyCollectCount:
		return len(values), nil
	default:
		return nil, &UnsupportedHitPolicyError{HitPolicy: t.def.HitPolicy}
	}
}
