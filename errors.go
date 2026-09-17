package dmn

import "fmt"

// ValidationError is returned by Validate when a decision table definition
// fails a design-time check (spec.md section 4.1).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// CompletenessError is returned when an evaluation matches zero rules and
// no fallback ("-") rule covers the context (DMNCompletenessException in
// spec.md section 4.2).
type CompletenessError struct {
	TableName string
}

func (e *CompletenessError) Error() string {
	return fmt.Sprintf("dmn: no rule matched context for table %q and no fallback rule exists", e.TableName)
}

// TypeMismatchError is returned when a context value's type does not match
// the declared type of the corresponding input clause (DMNTypeMismatchException
// in spec.md section 4.2).
type TypeMismatchError struct {
	ClauseName string
	Expected   ClauseType
	Value      any
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("dmn: context value for %q has type %T, want %s", e.ClauseName, e.Value, e.Expected)
}

// ConflictError is returned when the U or A hit policies encounter multiple
// matching rules that cannot be resolved to a single result
// (DMNRuntimeConflictException in spec.md section 3.3).
type ConflictError struct {
	TableName string
	HitPolicy HitPolicy
	Message   string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("dmn: conflict resolving hit policy %s for table %q: %s", e.HitPolicy, e.TableName, e.Message)
}

// UnsupportedHitPolicyError is returned when a decision table declares a
// hitPolicy that is accepted by the schema but not implemented by this
// engine (R and O; see proposal.md - Impact).
type UnsupportedHitPolicyError struct {
	HitPolicy HitPolicy
}

func (e *UnsupportedHitPolicyError) Error() string {
	return fmt.Sprintf("dmn: unsupported hit policy %q", e.HitPolicy)
}
