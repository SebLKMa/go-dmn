package dmn

import "testing"

func mustCompile(t *testing.T, entry string, typ ClauseType) entryMatcher {
	t.Helper()
	m, err := compileEntry(entry, typ)
	if err != nil {
		t.Fatalf("compileEntry(%q, %q) error = %v", entry, typ, err)
	}
	return m
}

func TestWildcardMatcher(t *testing.T) {
	m := mustCompile(t, "-", ClauseTypeString)
	for _, v := range []any{"anything", 42.0, true} {
		ok, err := m.Match(v)
		if err != nil || !ok {
			t.Errorf("Match(%v) = %v, %v; want true, nil", v, ok, err)
		}
	}
}

func TestLiteralMatcher(t *testing.T) {
	tests := []struct {
		name  string
		entry string
		typ   ClauseType
		value any
		want  bool
	}{
		{"string match", `"VIP"`, ClauseTypeString, "VIP", true},
		{"string mismatch", `"VIP"`, ClauseTypeString, "Regular", false},
		{"number match", "100", ClauseTypeNumber, 100.0, true},
		{"boolean match", "true", ClauseTypeBoolean, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mustCompile(t, tt.entry, tt.typ)
			got, err := m.Match(tt.value)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Match(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestRelationalMatcher(t *testing.T) {
	tests := []struct {
		entry string
		value float64
		want  bool
	}{
		{"< 50", 40, true},
		{"< 50", 50, false},
		{"<= 50", 50, true},
		{"> 50", 60, true},
		{">= 50", 50, true},
		{"!= 50", 51, true},
		{"!= 50", 50, false},
	}
	for _, tt := range tests {
		m := mustCompile(t, tt.entry, ClauseTypeNumber)
		got, err := m.Match(tt.value)
		if err != nil {
			t.Fatalf("Match() error = %v", err)
		}
		if got != tt.want {
			t.Errorf("entry %q, value %v: Match() = %v, want %v", tt.entry, tt.value, got, tt.want)
		}
	}
}

func TestIntervalMatcher(t *testing.T) {
	tests := []struct {
		name  string
		entry string
		value float64
		want  bool
	}{
		{"inclusive lower bound", "[18..65]", 18, true},
		{"inclusive upper bound", "[18..65]", 65, true},
		{"inclusive inside", "[18..65]", 30, true},
		{"inclusive outside", "[18..65]", 17, false},
		{"exclusive lower bound excluded", "(18..65)", 18, false},
		{"exclusive upper bound excluded", "(18..65)", 65, false},
		{"exclusive inside", "(18..65)", 30, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mustCompile(t, tt.entry, ClauseTypeNumber)
			got, err := m.Match(tt.value)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Match(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestListMatcher(t *testing.T) {
	m := mustCompile(t, `"Gold", "Silver"`, ClauseTypeString)

	got, err := m.Match("Silver")
	if err != nil || !got {
		t.Errorf("Match(Silver) = %v, %v; want true, nil", got, err)
	}

	got, err = m.Match("Bronze")
	if err != nil || got {
		t.Errorf("Match(Bronze) = %v, %v; want false, nil", got, err)
	}
}

func TestCompileEntry_CachesParsing(t *testing.T) {
	m := mustCompile(t, "[18..65]", ClauseTypeNumber)
	for i := 0; i < 1000; i++ {
		if _, err := m.Match(float64(i)); err != nil {
			t.Fatalf("Match() error = %v", err)
		}
	}
}
