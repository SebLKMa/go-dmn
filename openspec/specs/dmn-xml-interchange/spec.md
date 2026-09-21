# DMN XML Interchange Specification

## Purpose

Lets callers load and save decision tables using the OMG DMN XML interchange format, so tables authored in standard DMN modeling tools (or produced by this library) can be exchanged without a hand-written JSON transcription step.

## Requirements

### Requirement: XML Decoding of Decision Tables
The system SHALL decode an OMG DMN XML document containing a `<decision>` element with a `<decisionTable>` child into a `DecisionTableDefinition` equivalent to the one produced by decoding the same table's JSON representation: the decision's `name` attribute SHALL become `tableName`; the `decisionTable` element's `hitPolicy` (and, for `COLLECT`, `aggregation`) attributes SHALL become `hitPolicy` per the Hit Policy Mapping requirement; each `<input>` element SHALL become a `ClauseDefinition` whose `name` is the text content of its `<inputExpression><text>` child (the FEEL variable used as the evaluation context key) with its FEEL `typeRef` mapped per the Type Mapping requirement; each `<output>` element SHALL become a `ClauseDefinition` whose `name` is its own `name` attribute, or, when the decision table has exactly one output clause and that attribute is absent, the enclosing `<decision>` element's `name`, with its `typeRef` attribute mapped per the Type Mapping requirement; and each `<rule>` element SHALL become a `RuleDefinition` whose `inputEntries`/`outputEntries` are the text content of its `<inputEntry>`/`<outputEntry>` children, in document order.

#### Scenario: Well-formed decision table document decodes
- **WHEN** an OMG DMN XML document contains one `<decision>` with a `<decisionTable>` whose inputs, outputs, and rules are all well-formed
- **THEN** the system produces a `DecisionTableDefinition` with matching `tableName`, `hitPolicy`, `inputs`, `outputs`, and `rules`

#### Scenario: Single unnamed output takes the decision's name
- **WHEN** a `<decisionTable>` has exactly one `<output>` element with no `name` attribute
- **THEN** the decoded output `ClauseDefinition`'s `name` is the enclosing `<decision>` element's `name`

#### Scenario: Rule entry order is preserved
- **WHEN** a `<rule>` element contains `<inputEntry>` and `<outputEntry>` children in a given order
- **THEN** the decoded `RuleDefinition`'s `inputEntries` and `outputEntries` preserve that same order

### Requirement: Hit Policy Mapping
The system SHALL map between a `DecisionTableDefinition`'s `hitPolicy` field and the OMG DMN XML Schema's `hitPolicy` (and, for `COLLECT`, `aggregation`) attributes as follows: `UNIQUE` ↔ `U`, `FIRST` ↔ `F`, `ANY` ↔ `A`, `PRIORITY` ↔ `P`, `RULE ORDER` ↔ `R`, `OUTPUT ORDER` ↔ `O`, `COLLECT` with no `aggregation` attribute ↔ `C`, `COLLECT` with `aggregation="SUM"` ↔ `C+`, `COLLECT` with `aggregation="MIN"` ↔ `C<`, `COLLECT` with `aggregation="MAX"` ↔ `C>`, and `COLLECT` with `aggregation="COUNT"` ↔ `C#`. A `<decisionTable>` with no `hitPolicy` attribute SHALL decode as `U`, per the OMG DMN XML Schema default.

#### Scenario: Full-word hitPolicy decodes to internal code
- **WHEN** a `<decisionTable>` element declares `hitPolicy="PRIORITY"`
- **THEN** the decoded `DecisionTableDefinition` has `hitPolicy` `P`

#### Scenario: Collect with aggregation decodes to the matching collect code
- **WHEN** a `<decisionTable>` element declares `hitPolicy="COLLECT"` and `aggregation="SUM"`
- **THEN** the decoded `DecisionTableDefinition` has `hitPolicy` `C+`

#### Scenario: Missing hitPolicy attribute defaults to Unique
- **WHEN** a `<decisionTable>` element has no `hitPolicy` attribute
- **THEN** the decoded `DecisionTableDefinition` has `hitPolicy` `U`

#### Scenario: Internal code encodes to the matching full-word attributes
- **WHEN** a `DecisionTableDefinition` with `hitPolicy` `C<` is encoded to XML
- **THEN** the resulting `<decisionTable>` element declares `hitPolicy="COLLECT"` and `aggregation="MIN"`

### Requirement: XML Encoding of Decision Tables
The system SHALL encode a `DecisionTableDefinition` into a well-formed OMG DMN XML document containing a single `<decision>` element with a `<decisionTable>` child, structured so that decoding the produced document (per the XML Decoding requirement) yields a `DecisionTableDefinition` equal to the original.

#### Scenario: Encoded document round-trips
- **WHEN** a `DecisionTableDefinition` is encoded to XML and the resulting document is then decoded
- **THEN** the decoded `DecisionTableDefinition` is equal to the original

#### Scenario: Encoded document is well-formed OMG DMN XML
- **WHEN** a `DecisionTableDefinition` is encoded to XML
- **THEN** the output is a well-formed XML document with a `<definitions>` root in the OMG DMN XML Schema namespace, containing the `<decision>` and `<decisionTable>` elements

### Requirement: Clause Type Mapping
The system SHALL map between a `ClauseDefinition`'s `type` field and the DMN FEEL type names used in XML `typeRef` attributes as follows: `string` ↔ `string`, `number` ↔ `number`, `boolean` ↔ `boolean`, `date` ↔ `date`. A `typeRef` value outside this set SHALL be rejected when decoding.

#### Scenario: Supported typeRef decodes
- **WHEN** an `<input>` element's `<inputExpression>` child, or an `<output>` element itself, declares `typeRef="number"`
- **THEN** the decoded `ClauseDefinition`'s `type` is `number`

#### Scenario: Unsupported typeRef is rejected
- **WHEN** an `<input>` element's `<inputExpression>` child, or an `<output>` element itself, declares a `typeRef` other than `string`, `number`, `boolean`, or `date`
- **THEN** the system rejects the document with a clear error rather than guessing a type

### Requirement: Decision Requirements Elements Are Out of Scope
The system SHALL ignore DMN XML elements unrelated to decision table content when present alongside a `<decision><decisionTable>` (for example `<inputData>`, `<knowledgeSource>`, `<businessKnowledgeModel>`, `<informationRequirement>`, and DMNDI diagram interchange elements), decoding only the decision table content, consistent with this library's existing scope of decision tables without Decision Requirement Diagrams.

#### Scenario: Sibling DRD elements are ignored
- **WHEN** an OMG DMN XML document's `<definitions>` element contains `<inputData>` or DMNDI diagram elements alongside a `<decision><decisionTable>`
- **THEN** the system decodes the decision table and does not error on the unrelated elements

#### Scenario: Decision without a decision table is rejected
- **WHEN** a `<decision>` element's expression is not a `<decisionTable>` (for example a literal expression or boxed context)
- **THEN** the system rejects the document with a clear "unsupported expression type" error rather than silently producing an empty definition

### Requirement: Malformed XML Is Rejected
The system SHALL reject, before evaluation, an XML document that is not well-formed XML, or that is missing required decision table structure (a `<decisionTable>` with no `<rule>` entries whose `<inputEntry>`/`<outputEntry>` counts do not match its declared `<input>`/`<output>` count), returning the same validation error type used for an equivalently malformed JSON definition.

#### Scenario: Invalid XML syntax is rejected
- **WHEN** the input to the XML decoder is not well-formed XML
- **THEN** the system returns an error and does not produce a partial `DecisionTableDefinition`

#### Scenario: Structurally inconsistent rule is rejected
- **WHEN** a `<rule>` element's `<inputEntry>` or `<outputEntry>` count does not match the number of declared `<input>` or `<output>` elements
- **THEN** the system rejects the definition using the same validation error as the equivalent JSON case
