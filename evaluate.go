package dmn

// Evaluate runs a decision table's rules against a context and returns the
// resolved result per its hit policy, or an error (see errors.go for the
// exception types from spec.md section 4.2 and the hit-policy conflicts
// from section 3.3).
func (t *CompiledTable) Evaluate(ctx EvaluationContext) (any, error) {
	matched, err := t.MatchingRules(ctx)
	if err != nil {
		return nil, err
	}
	return t.Resolve(matched)
}

// Evaluate validates and compiles def, then evaluates it against ctx. Callers
// evaluating the same definition repeatedly should call Compile once and
// reuse the returned *CompiledTable instead.
func Evaluate(def DecisionTableDefinition, ctx EvaluationContext) (any, error) {
	table, err := Compile(def)
	if err != nil {
		return nil, err
	}
	return table.Evaluate(ctx)
}
