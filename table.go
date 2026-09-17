package dmn

// CompiledRule is a rule definition with its input entries pre-parsed into
// evaluable matchers (design.md - "compile once per rule at load time").
type CompiledRule struct {
	def      RuleDefinition
	matchers []entryMatcher
}

// CompiledTable is a decision table definition whose rules have been
// validated and compiled, ready for repeated evaluation.
type CompiledTable struct {
	def   DecisionTableDefinition
	rules []CompiledRule
}

// Compile validates a decision table definition and compiles every rule's
// input entries into evaluable matchers.
func Compile(def DecisionTableDefinition) (*CompiledTable, error) {
	if err := Validate(def); err != nil {
		return nil, err
	}

	rules := make([]CompiledRule, len(def.Rules))
	for i, rule := range def.Rules {
		matchers := make([]entryMatcher, len(rule.InputEntries))
		for j, entry := range rule.InputEntries {
			m, err := compileEntry(entry, def.Inputs[j].Type)
			if err != nil {
				return nil, err
			}
			matchers[j] = m
		}
		rules[i] = CompiledRule{def: rule, matchers: matchers}
	}
	return &CompiledTable{def: def, rules: rules}, nil
}

// matches reports whether every input entry of the rule evaluates to true
// against ctx (logical AND, per the Rule Matching requirement). A wildcard
// entry skips both type checking and evaluation for its input, per the
// S-FEEL Input Entry Grammar requirement.
func (r CompiledRule) matches(inputs []ClauseDefinition, ctx EvaluationContext) (bool, error) {
	for i, input := range inputs {
		if _, isWildcard := r.matchers[i].(wildcardMatcher); isWildcard {
			continue
		}

		value := ctx[input.Name]
		if err := checkClauseType(input.Name, value, input.Type); err != nil {
			return false, err
		}

		ok, err := r.matchers[i].Match(value)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// computeMatches returns every rule whose input entries all match ctx,
// without applying the completeness check.
func (t *CompiledTable) computeMatches(ctx EvaluationContext) ([]RuleDefinition, error) {
	var matched []RuleDefinition
	for _, rule := range t.rules {
		ok, err := rule.matches(t.def.Inputs, ctx)
		if err != nil {
			return nil, err
		}
		if ok {
			matched = append(matched, rule.def)
		}
	}
	return matched, nil
}

// MatchingRules returns every rule whose input entries all match ctx. If no
// rule matches, it returns a *CompletenessError (spec.md section 4.2,
// DMNCompletenessException) - a rule consisting entirely of wildcard ("-")
// entries would itself match, so a genuine catch-all rule never triggers
// this error.
func (t *CompiledTable) MatchingRules(ctx EvaluationContext) ([]RuleDefinition, error) {
	matched, err := t.computeMatches(ctx)
	if err != nil {
		return nil, err
	}
	if len(matched) == 0 {
		return nil, &CompletenessError{TableName: t.def.TableName}
	}
	return matched, nil
}
