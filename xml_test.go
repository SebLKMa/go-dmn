package dmn

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

const minimalDecisionTableXML = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" name="definitions" id="_defs">
  <decision id="d1" name="Discount">
    <decisionTable id="dt1" hitPolicy="UNIQUE">
      <input id="i1">
        <inputExpression id="ie1" typeRef="string">
          <text>customerType</text>
        </inputExpression>
      </input>
      <output id="o1" name="discount" typeRef="number"/>
      <rule id="r1">
        <inputEntry id="ie1"><text>"VIP"</text></inputEntry>
        <outputEntry id="oe1"><text>10</text></outputEntry>
      </rule>
      <rule id="r2">
        <inputEntry id="ie2"><text>-</text></inputEntry>
        <outputEntry id="oe2"><text>0</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`

func TestDecodeXML_WellFormed(t *testing.T) {
	def, err := DecodeXML(strings.NewReader(minimalDecisionTableXML))
	if err != nil {
		t.Fatalf("DecodeXML() error = %v", err)
	}

	want := DecisionTableDefinition{
		TableName: "Discount",
		HitPolicy: HitPolicyUnique,
		Inputs: []ClauseDefinition{
			{ID: "i1", Name: "customerType", Type: ClauseTypeString},
		},
		Outputs: []ClauseDefinition{
			{ID: "o1", Name: "discount", Type: ClauseTypeNumber},
		},
		Rules: []RuleDefinition{
			{RuleID: "r1", InputEntries: []string{`"VIP"`}, OutputEntries: []string{"10"}},
			{RuleID: "r2", InputEntries: []string{"-"}, OutputEntries: []string{"0"}},
		},
	}
	if !reflect.DeepEqual(def, want) {
		t.Errorf("DecodeXML() = %+v, want %+v", def, want)
	}
}

func TestDecodeXML_SingleUnnamedOutputTakesDecisionName(t *testing.T) {
	const doc = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/">
  <decision id="d1" name="Discount">
    <decisionTable id="dt1" hitPolicy="UNIQUE">
      <input id="i1">
        <inputExpression id="ie1" typeRef="string"><text>customerType</text></inputExpression>
      </input>
      <output id="o1" typeRef="number"/>
      <rule id="r1">
        <inputEntry id="ie1"><text>"VIP"</text></inputEntry>
        <outputEntry id="oe1"><text>10</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`

	def, err := DecodeXML(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("DecodeXML() error = %v", err)
	}
	if len(def.Outputs) != 1 || def.Outputs[0].Name != "Discount" {
		t.Errorf("DecodeXML() outputs = %+v, want single output named %q", def.Outputs, "Discount")
	}
}

func TestDecodeXML_IgnoresDRDElements(t *testing.T) {
	const doc = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/">
  <inputData id="id1" name="CustomerType"/>
  <knowledgeSource id="ks1" name="Policy Manual"/>
  <decision id="d1" name="Discount">
    <extensionElements/>
    <informationRequirement>
      <requiredInput href="#id1"/>
    </informationRequirement>
    <decisionTable id="dt1" hitPolicy="UNIQUE">
      <input id="i1">
        <inputExpression id="ie1" typeRef="string"><text>customerType</text></inputExpression>
      </input>
      <output id="o1" name="discount" typeRef="number"/>
      <rule id="r1">
        <inputEntry id="ie1"><text>"VIP"</text></inputEntry>
        <outputEntry id="oe1"><text>10</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
  <dmndi:DMNDI xmlns:dmndi="https://www.omg.org/spec/DMN/20191111/DMNDI/">
    <dmndi:DMNDiagram/>
  </dmndi:DMNDI>
</definitions>`

	def, err := DecodeXML(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("DecodeXML() error = %v, want no error for unrelated sibling elements", err)
	}
	if def.TableName != "Discount" || len(def.Inputs) != 1 || len(def.Rules) != 1 {
		t.Errorf("DecodeXML() = %+v, want decision table decoded despite sibling DRD/DMNDI elements", def)
	}
}

func TestDecodeXML_UnsupportedExpressionType(t *testing.T) {
	const doc = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/">
  <decision id="d1" name="Greeting">
    <literalExpression id="le1">
      <text>"hello"</text>
    </literalExpression>
  </decision>
</definitions>`

	_, err := DecodeXML(strings.NewReader(doc))
	if err == nil {
		t.Fatal("DecodeXML() error = nil, want unsupported expression type error")
	}
	if !strings.Contains(err.Error(), "unsupported expression type") {
		t.Errorf("DecodeXML() error = %v, want it to mention unsupported expression type", err)
	}
}

func TestDecodeXML_MalformedXML(t *testing.T) {
	_, err := DecodeXML(strings.NewReader("not xml at all"))
	if err == nil {
		t.Fatal("DecodeXML() error = nil, want an error for invalid XML")
	}
}

func TestDecodeXML_StructuralMismatchRejected(t *testing.T) {
	const doc = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/">
  <decision id="d1" name="Discount">
    <decisionTable id="dt1" hitPolicy="UNIQUE">
      <input id="i1">
        <inputExpression id="ie1" typeRef="string"><text>customerType</text></inputExpression>
      </input>
      <output id="o1" name="discount" typeRef="number"/>
      <rule id="r1">
        <inputEntry id="ie1"><text>"VIP"</text></inputEntry>
        <inputEntry id="ie2"><text>"extra"</text></inputEntry>
        <outputEntry id="oe1"><text>10</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`

	_, err := DecodeXML(strings.NewReader(doc))
	if err == nil {
		t.Fatal("DecodeXML() error = nil, want a ValidationError for mismatched inputEntry count")
	}
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Errorf("DecodeXML() error = %v, want *ValidationError", err)
	}
}

func TestClauseTypeMapping(t *testing.T) {
	for typeRef, want := range feelTypeToClauseType {
		got, err := decodeClauseType(typeRef)
		if err != nil {
			t.Fatalf("decodeClauseType(%q) error = %v", typeRef, err)
		}
		if got != want {
			t.Errorf("decodeClauseType(%q) = %q, want %q", typeRef, got, want)
		}
	}

	if _, err := decodeClauseType("dateTime"); err == nil {
		t.Error("decodeClauseType(\"dateTime\") error = nil, want error for unsupported typeRef")
	}
}

func TestHitPolicyMapping(t *testing.T) {
	t.Run("full word decodes to internal code", func(t *testing.T) {
		hp, err := decodeHitPolicy("PRIORITY", "")
		if err != nil {
			t.Fatalf("decodeHitPolicy() error = %v", err)
		}
		if hp != HitPolicyPriority {
			t.Errorf("decodeHitPolicy(\"PRIORITY\", \"\") = %q, want %q", hp, HitPolicyPriority)
		}
	})

	t.Run("collect with aggregation decodes to matching collect code", func(t *testing.T) {
		hp, err := decodeHitPolicy("COLLECT", "SUM")
		if err != nil {
			t.Fatalf("decodeHitPolicy() error = %v", err)
		}
		if hp != HitPolicyCollectSum {
			t.Errorf("decodeHitPolicy(\"COLLECT\", \"SUM\") = %q, want %q", hp, HitPolicyCollectSum)
		}
	})

	t.Run("missing hitPolicy defaults to Unique", func(t *testing.T) {
		hp, err := decodeHitPolicy("", "")
		if err != nil {
			t.Fatalf("decodeHitPolicy() error = %v", err)
		}
		if hp != HitPolicyUnique {
			t.Errorf("decodeHitPolicy(\"\", \"\") = %q, want %q", hp, HitPolicyUnique)
		}
	})

	t.Run("internal code encodes to matching full-word attributes", func(t *testing.T) {
		attr, err := encodeHitPolicy(HitPolicyCollectMin)
		if err != nil {
			t.Fatalf("encodeHitPolicy() error = %v", err)
		}
		if attr.HitPolicy != "COLLECT" || attr.Aggregation != "MIN" {
			t.Errorf("encodeHitPolicy(C<) = %+v, want {COLLECT MIN}", attr)
		}
	})
}

func TestEncodeXML_WellFormed(t *testing.T) {
	def := DecisionTableDefinition{
		TableName: "Discount",
		HitPolicy: HitPolicyUnique,
		Inputs: []ClauseDefinition{
			{ID: "i1", Name: "customerType", Type: ClauseTypeString},
		},
		Outputs: []ClauseDefinition{
			{ID: "o1", Name: "discount", Type: ClauseTypeNumber},
		},
		Rules: []RuleDefinition{
			{RuleID: "r1", InputEntries: []string{`"VIP"`}, OutputEntries: []string{"10"}},
		},
	}

	var buf bytes.Buffer
	if err := EncodeXML(&buf, def); err != nil {
		t.Fatalf("EncodeXML() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "<decision ") {
		t.Errorf("EncodeXML() output missing <decision> element:\n%s", out)
	}
	if !strings.Contains(out, "<decisionTable ") {
		t.Errorf("EncodeXML() output missing <decisionTable> element:\n%s", out)
	}

	if _, err := DecodeXML(strings.NewReader(out)); err != nil {
		t.Errorf("EncodeXML() produced output that failed to decode: %v\n%s", err, out)
	}
}

func TestXMLRoundTrip(t *testing.T) {
	def := DecisionTableDefinition{
		TableName: "RoundTrip",
		HitPolicy: HitPolicyUnique,
		Inputs: []ClauseDefinition{
			{ID: "i1", Name: "customerType", Type: ClauseTypeString},
			{ID: "i2", Name: "purchaseVolume", Type: ClauseTypeNumber},
			{ID: "i3", Name: "isMember", Type: ClauseTypeBoolean},
			{ID: "i4", Name: "signupDate", Type: ClauseTypeDate},
		},
		Outputs: []ClauseDefinition{
			{ID: "o1", Name: "discount", Type: ClauseTypeNumber},
		},
		Rules: []RuleDefinition{
			{
				RuleID: "r1",
				InputEntries: []string{
					`"VIP", "Gold"`,
					"[100..500]",
					"true",
					">= 2020-01-01T00:00:00Z",
				},
				OutputEntries: []string{"10"},
			},
			{
				RuleID:        "r2",
				InputEntries:  []string{"-", "-", "-", "-"},
				OutputEntries: []string{"0"},
			},
		},
	}

	var buf bytes.Buffer
	if err := EncodeXML(&buf, def); err != nil {
		t.Fatalf("EncodeXML() error = %v", err)
	}

	got, err := DecodeXML(&buf)
	if err != nil {
		t.Fatalf("DecodeXML() error = %v", err)
	}
	if !reflect.DeepEqual(got, def) {
		t.Errorf("round trip mismatch:\n got  = %+v\n want = %+v", got, def)
	}
}

func TestDecodeXML_EvaluatesLikeJSONFixture(t *testing.T) {
	const doc = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/">
  <decision id="d1" name="VolumeDiscount">
    <decisionTable id="dt1" hitPolicy="UNIQUE">
      <input id="i1">
        <inputExpression id="ie1" typeRef="number"><text>purchaseVolume</text></inputExpression>
      </input>
      <output id="o1" name="discount" typeRef="number"/>
      <output id="o2" name="tier" typeRef="string"/>
      <rule id="r1">
        <inputEntry id="e1"><text>[0..499]</text></inputEntry>
        <outputEntry id="oe1"><text>0</text></outputEntry>
        <outputEntry id="oe2"><text>"Standard"</text></outputEntry>
      </rule>
      <rule id="r2">
        <inputEntry id="e2"><text>[500..1499]</text></inputEntry>
        <outputEntry id="oe3"><text>0.05</text></outputEntry>
        <outputEntry id="oe4"><text>"Silver"</text></outputEntry>
      </rule>
      <rule id="r3">
        <inputEntry id="e3"><text>[1500..4999]</text></inputEntry>
        <outputEntry id="oe5"><text>0.10</text></outputEntry>
        <outputEntry id="oe6"><text>"Gold"</text></outputEntry>
      </rule>
      <rule id="r4">
        <inputEntry id="e4"><text>&gt;= 5000</text></inputEntry>
        <outputEntry id="oe7"><text>0.15</text></outputEntry>
        <outputEntry id="oe8"><text>"Platinum"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`

	def, err := DecodeXML(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("DecodeXML() error = %v", err)
	}

	table, err := Compile(def)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	result, err := table.Evaluate(EvaluationContext{"purchaseVolume": 1000.0})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	values, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Evaluate() result type = %T, want map[string]any", result)
	}
	if values["discount"] != 0.05 || values["tier"] != "Silver" {
		t.Errorf("Evaluate() = %+v, want discount=0.05 tier=Silver", values)
	}
}
