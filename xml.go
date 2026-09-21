package dmn

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// dmnXMLNamespace is the OMG DMN XML Schema namespace this package writes
// on encode. Decoding does not require or branch on a specific namespace
// (design.md - Namespace handling): the decision table subtree shape is
// stable across the DMN 1.1-1.4 XML Schemas this targets, and Go's
// encoding/xml matches elements by local name when a struct tag omits a
// namespace.
const dmnXMLNamespace = "https://www.omg.org/spec/DMN/20191111/MODEL/"

// xmlDefinitions, xmlDecision, xmlDecisionTable, xmlInput, xmlOutput,
// xmlRule, and xmlEntry are a package-private mirror of the subset of the
// OMG DMN XML Schema this package understands (design.md - "intermediate
// type, not directly into DecisionTableDefinition"). Element/attribute names
// follow the schema, not the JSON struct tags on the public model types.
type xmlDefinitions struct {
	XMLName   xml.Name      `xml:"definitions"`
	Decisions []xmlDecision `xml:"decision"`
}

type xmlDecision struct {
	Name          string            `xml:"name,attr"`
	DecisionTable *xmlDecisionTable `xml:"decisionTable"`
}

type xmlDecisionTable struct {
	HitPolicy   string      `xml:"hitPolicy,attr"`
	Aggregation string      `xml:"aggregation,attr"`
	Inputs      []xmlInput  `xml:"input"`
	Outputs     []xmlOutput `xml:"output"`
	Rules       []xmlRule   `xml:"rule"`
}

type xmlInput struct {
	ID              string             `xml:"id,attr"`
	InputExpression xmlInputExpression `xml:"inputExpression"`
}

type xmlInputExpression struct {
	TypeRef string `xml:"typeRef,attr"`
	Text    string `xml:"text"`
}

type xmlOutput struct {
	ID      string `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	TypeRef string `xml:"typeRef,attr"`
}

type xmlRule struct {
	ID            string     `xml:"id,attr"`
	InputEntries  []xmlEntry `xml:"inputEntry"`
	OutputEntries []xmlEntry `xml:"outputEntry"`
}

type xmlEntry struct {
	Text string `xml:"text"`
}

// feelTypeToClauseType and clauseTypeToFEELType are the fixed, two-way
// mapping between FEEL typeRef names used in DMN XML and this package's
// ClauseType (spec.md - "Clause Type Mapping"). A typeRef outside this map
// is rejected rather than guessed at.
var feelTypeToClauseType = map[string]ClauseType{
	"string":  ClauseTypeString,
	"number":  ClauseTypeNumber,
	"boolean": ClauseTypeBoolean,
	"date":    ClauseTypeDate,
}

var clauseTypeToFEELType = map[ClauseType]string{
	ClauseTypeString:  "string",
	ClauseTypeNumber:  "number",
	ClauseTypeBoolean: "boolean",
	ClauseTypeDate:    "date",
}

func decodeClauseType(typeRef string) (ClauseType, error) {
	t, ok := feelTypeToClauseType[typeRef]
	if !ok {
		return "", fmt.Errorf("dmn: unsupported typeRef %q", typeRef)
	}
	return t, nil
}

func encodeClauseType(typ ClauseType) (string, error) {
	feelType, ok := clauseTypeToFEELType[typ]
	if !ok {
		return "", fmt.Errorf("dmn: unsupported clause type %q", typ)
	}
	return feelType, nil
}

// hitPolicyAttr pairs the OMG DMN XML hitPolicy attribute value with the
// aggregation attribute value (only meaningful when hitPolicy is COLLECT).
type hitPolicyAttr struct {
	HitPolicy   string
	Aggregation string
}

// hitPolicyToXML and xmlToHitPolicy are the fixed, two-way mapping between
// this package's internal HitPolicy codes and the OMG DMN XML Schema's
// hitPolicy/aggregation attribute values (spec.md - "Hit Policy Mapping").
// The XML Schema spells out hitPolicy as full words and represents the
// COLLECT sum/min/max/count variants via a separate aggregation attribute,
// rather than as distinct hitPolicy values.
var hitPolicyToXML = map[HitPolicy]hitPolicyAttr{
	HitPolicyUnique:       {HitPolicy: "UNIQUE"},
	HitPolicyFirst:        {HitPolicy: "FIRST"},
	HitPolicyAny:          {HitPolicy: "ANY"},
	HitPolicyPriority:     {HitPolicy: "PRIORITY"},
	HitPolicyRuleOrder:    {HitPolicy: "RULE ORDER"},
	HitPolicyOutputOrder:  {HitPolicy: "OUTPUT ORDER"},
	HitPolicyCollect:      {HitPolicy: "COLLECT"},
	HitPolicyCollectSum:   {HitPolicy: "COLLECT", Aggregation: "SUM"},
	HitPolicyCollectMin:   {HitPolicy: "COLLECT", Aggregation: "MIN"},
	HitPolicyCollectMax:   {HitPolicy: "COLLECT", Aggregation: "MAX"},
	HitPolicyCollectCount: {HitPolicy: "COLLECT", Aggregation: "COUNT"},
}

var xmlToHitPolicy = func() map[hitPolicyAttr]HitPolicy {
	m := make(map[hitPolicyAttr]HitPolicy, len(hitPolicyToXML))
	for hp, attr := range hitPolicyToXML {
		m[attr] = hp
	}
	return m
}()

func decodeHitPolicy(hitPolicy, aggregation string) (HitPolicy, error) {
	if hitPolicy == "" {
		return HitPolicyUnique, nil
	}
	hp, ok := xmlToHitPolicy[hitPolicyAttr{HitPolicy: hitPolicy, Aggregation: aggregation}]
	if !ok {
		return "", fmt.Errorf("dmn: unsupported hitPolicy/aggregation combination %q/%q", hitPolicy, aggregation)
	}
	return hp, nil
}

func encodeHitPolicy(hp HitPolicy) (hitPolicyAttr, error) {
	attr, ok := hitPolicyToXML[hp]
	if !ok {
		return hitPolicyAttr{}, fmt.Errorf("dmn: unsupported hit policy %q", hp)
	}
	return attr, nil
}

// DecodeXML decodes an OMG DMN XML document's first <decision> element with
// a <decisionTable> child into a DecisionTableDefinition. DRD elements
// unrelated to decision table content (inputData, knowledgeSource,
// businessKnowledgeModel, informationRequirement, DMNDI diagrams) are
// ignored if present (spec.md - "Decision Requirements Elements Are Out of
// Scope"). A <decision> whose expression is not a <decisionTable> is a
// decode error, not a skipped decision (spec.md - "Decision without a
// decision table is rejected").
func DecodeXML(r io.Reader) (DecisionTableDefinition, error) {
	var defs xmlDefinitions
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&defs); err != nil {
		return DecisionTableDefinition{}, fmt.Errorf("dmn: invalid XML: %w", err)
	}
	if len(defs.Decisions) == 0 {
		return DecisionTableDefinition{}, fmt.Errorf("dmn: no decision element found")
	}

	decision := defs.Decisions[0]
	if decision.DecisionTable == nil {
		return DecisionTableDefinition{}, fmt.Errorf("dmn: decision %q does not contain a decisionTable (unsupported expression type)", decision.Name)
	}
	dt := decision.DecisionTable

	hitPolicy, err := decodeHitPolicy(dt.HitPolicy, dt.Aggregation)
	if err != nil {
		return DecisionTableDefinition{}, err
	}

	inputs := make([]ClauseDefinition, len(dt.Inputs))
	for i, in := range dt.Inputs {
		typ, err := decodeClauseType(in.InputExpression.TypeRef)
		if err != nil {
			return DecisionTableDefinition{}, fmt.Errorf("dmn: input %d: %w", i, err)
		}
		inputs[i] = ClauseDefinition{
			ID:   in.ID,
			Name: strings.TrimSpace(in.InputExpression.Text),
			Type: typ,
		}
	}

	outputs := make([]ClauseDefinition, len(dt.Outputs))
	for i, out := range dt.Outputs {
		typ, err := decodeClauseType(out.TypeRef)
		if err != nil {
			return DecisionTableDefinition{}, fmt.Errorf("dmn: output %d: %w", i, err)
		}
		name := out.Name
		if name == "" && len(dt.Outputs) == 1 {
			name = decision.Name
		}
		outputs[i] = ClauseDefinition{
			ID:   out.ID,
			Name: name,
			Type: typ,
		}
	}

	rules := make([]RuleDefinition, len(dt.Rules))
	for i, r := range dt.Rules {
		inputEntries := make([]string, len(r.InputEntries))
		for j, e := range r.InputEntries {
			inputEntries[j] = strings.TrimSpace(e.Text)
		}
		outputEntries := make([]string, len(r.OutputEntries))
		for j, e := range r.OutputEntries {
			outputEntries[j] = strings.TrimSpace(e.Text)
		}
		rules[i] = RuleDefinition{
			RuleID:        r.ID,
			InputEntries:  inputEntries,
			OutputEntries: outputEntries,
		}
	}

	def := DecisionTableDefinition{
		TableName: decision.Name,
		HitPolicy: hitPolicy,
		Inputs:    inputs,
		Outputs:   outputs,
		Rules:     rules,
	}
	if err := Validate(def); err != nil {
		return DecisionTableDefinition{}, err
	}
	return def, nil
}

// EncodeXML encodes def as a well-formed OMG DMN XML document containing a
// single <decision> element with a <decisionTable> child, structured so
// that decoding the result with DecodeXML yields a DecisionTableDefinition
// equal to def (spec.md - "XML Encoding of Decision Tables"). Each output's
// name attribute is always emitted explicitly (never relying on the
// single-unnamed-output decode fallback), which is what makes decode(encode(def))
// == def hold regardless of whether an output's name happens to equal
// TableName.
func EncodeXML(w io.Writer, def DecisionTableDefinition) error {
	attr, err := encodeHitPolicy(def.HitPolicy)
	if err != nil {
		return err
	}

	inputs := make([]xmlInput, len(def.Inputs))
	for i, in := range def.Inputs {
		feelType, err := encodeClauseType(in.Type)
		if err != nil {
			return fmt.Errorf("dmn: input %d: %w", i, err)
		}
		inputs[i] = xmlInput{
			ID:              in.ID,
			InputExpression: xmlInputExpression{TypeRef: feelType, Text: in.Name},
		}
	}

	outputs := make([]xmlOutput, len(def.Outputs))
	for i, out := range def.Outputs {
		feelType, err := encodeClauseType(out.Type)
		if err != nil {
			return fmt.Errorf("dmn: output %d: %w", i, err)
		}
		outputs[i] = xmlOutput{ID: out.ID, Name: out.Name, TypeRef: feelType}
	}

	rules := make([]xmlRule, len(def.Rules))
	for i, r := range def.Rules {
		inputEntries := make([]xmlEntry, len(r.InputEntries))
		for j, e := range r.InputEntries {
			inputEntries[j] = xmlEntry{Text: e}
		}
		outputEntries := make([]xmlEntry, len(r.OutputEntries))
		for j, e := range r.OutputEntries {
			outputEntries[j] = xmlEntry{Text: e}
		}
		rules[i] = xmlRule{ID: r.RuleID, InputEntries: inputEntries, OutputEntries: outputEntries}
	}

	defs := xmlDefinitions{
		XMLName: xml.Name{Space: dmnXMLNamespace, Local: "definitions"},
		Decisions: []xmlDecision{
			{
				Name: def.TableName,
				DecisionTable: &xmlDecisionTable{
					HitPolicy:   attr.HitPolicy,
					Aggregation: attr.Aggregation,
					Inputs:      inputs,
					Outputs:     outputs,
					Rules:       rules,
				},
			},
		},
	}

	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(defs); err != nil {
		return fmt.Errorf("dmn: encoding XML: %w", err)
	}
	return nil
}
