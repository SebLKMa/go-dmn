package dmn

// HitPolicy identifies how a decision table resolves multiple matching
// rules into a single result.
type HitPolicy string

const (
	HitPolicyUnique       HitPolicy = "U"
	HitPolicyFirst        HitPolicy = "F"
	HitPolicyAny          HitPolicy = "A"
	HitPolicyPriority     HitPolicy = "P"
	HitPolicyCollect      HitPolicy = "C"
	HitPolicyCollectSum   HitPolicy = "C+"
	HitPolicyCollectMin   HitPolicy = "C<"
	HitPolicyCollectMax   HitPolicy = "C>"
	HitPolicyCollectCount HitPolicy = "C#"
	HitPolicyRuleOrder    HitPolicy = "R"
	HitPolicyOutputOrder  HitPolicy = "O"
)

// ClauseType identifies the expected Go value shape for a clause.
type ClauseType string

const (
	ClauseTypeString  ClauseType = "string"
	ClauseTypeNumber  ClauseType = "number"
	ClauseTypeBoolean ClauseType = "boolean"
	ClauseTypeDate    ClauseType = "date"
)

// ClauseDefinition describes a single input or output column of a decision
// table.
type ClauseDefinition struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Type          ClauseType `json:"type"`
	AllowedValues []string   `json:"allowedValues,omitempty"`
}

// RuleDefinition describes a single row of a decision table.
type RuleDefinition struct {
	RuleID        string   `json:"ruleId"`
	InputEntries  []string `json:"inputEntries"`
	OutputEntries []string `json:"outputEntries"`
}

// DecisionTableDefinition is the structural blueprint of a decision table,
// as described in spec.md section 2.1.
type DecisionTableDefinition struct {
	TableName string             `json:"tableName"`
	HitPolicy HitPolicy          `json:"hitPolicy"`
	Inputs    []ClauseDefinition `json:"inputs"`
	Outputs   []ClauseDefinition `json:"outputs"`
	Rules     []RuleDefinition   `json:"rules"`
}
