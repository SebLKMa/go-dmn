package dmn

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// entryMatcher is a compiled S-FEEL input entry (spec.md section 3.1),
// ready to be evaluated repeatedly against different context values without
// re-parsing the entry string (design.md - "Compile once per rule at load
// time").
type entryMatcher interface {
	Match(value any) (bool, error)
}

// compileEntry parses a single inputEntries string for a clause of the
// given type into an evaluable matcher.
func compileEntry(raw string, typ ClauseType) (entryMatcher, error) {
	entry := strings.TrimSpace(raw)

	if entry == "-" {
		return wildcardMatcher{}, nil
	}
	if strings.Contains(entry, "..") && (strings.HasPrefix(entry, "[") || strings.HasPrefix(entry, "(")) {
		return compileInterval(entry, typ)
	}
	if op, rest, ok := splitRelationalOperator(entry); ok {
		threshold, err := toComparable(strings.TrimSpace(rest), typ)
		if err != nil {
			return nil, fmt.Errorf("dmn: invalid relational entry %q: %w", raw, err)
		}
		return relationalMatcher{op: op, threshold: threshold, typ: typ}, nil
	}
	if parts := splitTopLevelComma(entry); len(parts) > 1 {
		matchers := make([]entryMatcher, 0, len(parts))
		for _, part := range parts {
			lit, err := parseLiteral(strings.TrimSpace(part), typ)
			if err != nil {
				return nil, fmt.Errorf("dmn: invalid list entry %q: %w", raw, err)
			}
			matchers = append(matchers, literalMatcher{value: lit})
		}
		return listMatcher{matchers: matchers}, nil
	}

	lit, err := parseLiteral(entry, typ)
	if err != nil {
		return nil, fmt.Errorf("dmn: invalid literal entry %q: %w", raw, err)
	}
	return literalMatcher{value: lit}, nil
}

// wildcardMatcher implements the "-" token: always matches and skips
// evaluation of the context value entirely.
type wildcardMatcher struct{}

func (wildcardMatcher) Match(any) (bool, error) { return true, nil }

// literalMatcher implements exact-literal matching.
type literalMatcher struct {
	value any
}

func (m literalMatcher) Match(value any) (bool, error) {
	return m.value == value, nil
}

// listMatcher implements a comma-separated value list evaluated as a
// logical OR.
type listMatcher struct {
	matchers []entryMatcher
}

func (m listMatcher) Match(value any) (bool, error) {
	for _, sub := range m.matchers {
		ok, err := sub.Match(value)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// relationalMatcher implements <, >, <=, >=, != against a numeric or date
// threshold.
type relationalMatcher struct {
	op        string
	threshold float64
	typ       ClauseType
}

func (m relationalMatcher) Match(value any) (bool, error) {
	v, err := toComparable(value, m.typ)
	if err != nil {
		return false, err
	}
	switch m.op {
	case "<=":
		return v <= m.threshold, nil
	case ">=":
		return v >= m.threshold, nil
	case "!=":
		return v != m.threshold, nil
	case "<":
		return v < m.threshold, nil
	case ">":
		return v > m.threshold, nil
	default:
		return false, fmt.Errorf("dmn: unknown relational operator %q", m.op)
	}
}

// intervalMatcher implements inclusive "[a..b]" and exclusive "(a..b)"
// ranges.
type intervalMatcher struct {
	lower, upper         float64
	lowerIncl, upperIncl bool
	typ                  ClauseType
}

func (m intervalMatcher) Match(value any) (bool, error) {
	v, err := toComparable(value, m.typ)
	if err != nil {
		return false, err
	}
	lowerOK := v > m.lower || (m.lowerIncl && v == m.lower)
	upperOK := v < m.upper || (m.upperIncl && v == m.upper)
	return lowerOK && upperOK, nil
}

func compileInterval(entry string, typ ClauseType) (entryMatcher, error) {
	if len(entry) < 2 {
		return nil, fmt.Errorf("dmn: invalid interval entry %q", entry)
	}
	lowerIncl := entry[0] == '['
	upperIncl := entry[len(entry)-1] == ']'
	inner := entry[1 : len(entry)-1]
	bounds := strings.SplitN(inner, "..", 2)
	if len(bounds) != 2 {
		return nil, fmt.Errorf("dmn: invalid interval entry %q", entry)
	}
	lower, err := toComparable(strings.TrimSpace(bounds[0]), typ)
	if err != nil {
		return nil, fmt.Errorf("dmn: invalid interval lower bound in %q: %w", entry, err)
	}
	upper, err := toComparable(strings.TrimSpace(bounds[1]), typ)
	if err != nil {
		return nil, fmt.Errorf("dmn: invalid interval upper bound in %q: %w", entry, err)
	}
	return intervalMatcher{
		lower: lower, upper: upper,
		lowerIncl: lowerIncl, upperIncl: upperIncl,
		typ: typ,
	}, nil
}

// splitRelationalOperator returns the operator and remaining text if entry
// begins with one of the relational operators. Longer operators (<=, >=,
// !=) are checked before their single-character prefixes.
func splitRelationalOperator(entry string) (op string, rest string, ok bool) {
	for _, candidate := range []string{"<=", ">=", "!=", "<", ">"} {
		if strings.HasPrefix(entry, candidate) {
			return candidate, entry[len(candidate):], true
		}
	}
	return "", "", false
}

// splitTopLevelComma splits on commas that are not inside a quoted string.
func splitTopLevelComma(entry string) []string {
	var parts []string
	var buf strings.Builder
	inQuotes := false
	for _, r := range entry {
		switch {
		case r == '"':
			inQuotes = !inQuotes
			buf.WriteRune(r)
		case r == ',' && !inQuotes:
			parts = append(parts, buf.String())
			buf.Reset()
		default:
			buf.WriteRune(r)
		}
	}
	parts = append(parts, buf.String())
	return parts
}

// parseLiteral parses a single literal token (e.g. `"VIP"`, `100`, `true`)
// into the Go value it should compare equal to for the given clause type.
func parseLiteral(token string, typ ClauseType) (any, error) {
	switch typ {
	case ClauseTypeString:
		return strings.Trim(token, `"`), nil
	case ClauseTypeNumber:
		f, err := strconv.ParseFloat(token, 64)
		if err != nil {
			return nil, fmt.Errorf("not a number: %w", err)
		}
		return f, nil
	case ClauseTypeBoolean:
		b, err := strconv.ParseBool(token)
		if err != nil {
			return nil, fmt.Errorf("not a boolean: %w", err)
		}
		return b, nil
	case ClauseTypeDate:
		t, err := time.Parse(time.RFC3339, strings.Trim(token, `"`))
		if err != nil {
			return nil, fmt.Errorf("not a date: %w", err)
		}
		return t, nil
	default:
		return nil, fmt.Errorf("unsupported clause type %q", typ)
	}
}

// toComparable converts a value (a context value of any accepted dynamic
// type, or a raw token string parsed while compiling an entry) into a
// float64 usable by relational and interval matchers. Numbers compare
// directly; dates compare by Unix timestamp.
func toComparable(value any, typ ClauseType) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case time.Time:
		return float64(v.Unix()), nil
	case string:
		if typ == ClauseTypeDate {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return 0, fmt.Errorf("not a date: %w", err)
			}
			return float64(t.Unix()), nil
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("not a number: %w", err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot compare value of type %T", value)
	}
}
