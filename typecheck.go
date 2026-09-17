package dmn

import "time"

// checkClauseType verifies that a context value's dynamic Go type matches
// the declared type of an input clause (spec.md section 4.2,
// DMNTypeMismatchException), returning a *TypeMismatchError on mismatch.
func checkClauseType(clauseName string, value any, typ ClauseType) error {
	var ok bool
	switch typ {
	case ClauseTypeString:
		_, ok = value.(string)
	case ClauseTypeNumber:
		switch value.(type) {
		case float64, int:
			ok = true
		}
	case ClauseTypeBoolean:
		_, ok = value.(bool)
	case ClauseTypeDate:
		switch v := value.(type) {
		case time.Time:
			ok = true
		case string:
			_, err := time.Parse(time.RFC3339, v)
			ok = err == nil
		}
	}
	if !ok {
		return &TypeMismatchError{ClauseName: clauseName, Expected: typ, Value: value}
	}
	return nil
}
