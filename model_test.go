package dmn

import (
	"encoding/json"
	"testing"
)

const sampleDefinitionJSON = `{
	"tableName": "Discount",
	"hitPolicy": "U",
	"inputs": [
		{"id": "i1", "name": "customerType", "type": "string"}
	],
	"outputs": [
		{"id": "o1", "name": "discount", "type": "number"}
	],
	"rules": [
		{"ruleId": "r1", "inputEntries": ["\"VIP\""], "outputEntries": ["10"]}
	]
}`

func TestDecisionTableDefinition_DecodesFromJSON(t *testing.T) {
	var def DecisionTableDefinition
	if err := json.Unmarshal([]byte(sampleDefinitionJSON), &def); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if def.TableName != "Discount" {
		t.Errorf("TableName = %q, want %q", def.TableName, "Discount")
	}
	if def.HitPolicy != HitPolicyUnique {
		t.Errorf("HitPolicy = %q, want %q", def.HitPolicy, HitPolicyUnique)
	}
	if len(def.Inputs) != 1 || def.Inputs[0].Name != "customerType" || def.Inputs[0].Type != ClauseTypeString {
		t.Errorf("Inputs decoded incorrectly: %+v", def.Inputs)
	}
	if len(def.Outputs) != 1 || def.Outputs[0].Name != "discount" || def.Outputs[0].Type != ClauseTypeNumber {
		t.Errorf("Outputs decoded incorrectly: %+v", def.Outputs)
	}
	if len(def.Rules) != 1 || def.Rules[0].RuleID != "r1" {
		t.Errorf("Rules decoded incorrectly: %+v", def.Rules)
	}
}
