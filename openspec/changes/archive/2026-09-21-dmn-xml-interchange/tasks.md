## 1. Intermediate XML model

- [x] 1.1 Define package-private `xmlDefinitions`/`xmlDecision`/`xmlDecisionTable`/`xmlInput`/`xmlOutput`/`xmlRule`/`xmlEntry` structs in `xml.go` with `encoding/xml` struct tags matching the OMG DMN XML Schema element/attribute names (`decision`, `decisionTable`, `hitPolicy`, `input`, `output`, `typeRef`, `label`, `rule`, `inputEntry`, `outputEntry`, `text`), and verify `go build ./...` succeeds
- [x] 1.2 Implement the fixed two-way type-mapping table between FEEL `typeRef` names (`string`, `number`, `boolean`, `date`) and `ClauseType`, with a lookup function each direction, and verify unit tests cover a supported and an unsupported `typeRef` on decode
- [x] 1.3 Implement the fixed two-way hit-policy mapping table between OMG `hitPolicy`/`aggregation` attribute values (`UNIQUE`, `FIRST`, `ANY`, `PRIORITY`, `COLLECT` with `aggregation` `SUM`/`MIN`/`MAX`/`COUNT`/absent, `RULE ORDER`, `OUTPUT ORDER`) and internal `HitPolicy` codes, with a lookup function each direction and `U` as the decode default when `hitPolicy` is absent, and verify unit tests cover a full-word value, a `COLLECT`+`aggregation` combination, and the missing-attribute default

## 2. Decoding

- [x] 2.1 Implement `DecodeXML(r io.Reader) (DecisionTableDefinition, error)` that unmarshals into `xmlDefinitions`, locates the first `<decision>` with a `<decisionTable>` child, and maps it into a `DecisionTableDefinition` (decision `name` → `tableName`, `hitPolicy`/`aggregation` via the hit-policy mapping table, input `name` from `<inputExpression><text>`, output `name` from its `name` attribute or the decision's `name` when absent on a single output, types via the type-mapping table, rules preserving `<inputEntry>`/`<outputEntry>` document order), and verify a unit test decodes a hand-written minimal DMN XML fixture into the expected `DecisionTableDefinition`
- [x] 2.2 Make `DecodeXML` ignore sibling elements unrelated to decision table content (`<inputData>`, `<knowledgeSource>`, `<businessKnowledgeModel>`, `<informationRequirement>`, DMNDI diagram elements), and verify a unit test decodes a document containing such elements without error
- [x] 2.3 Make `DecodeXML` return a clear "unsupported expression type" error when a `<decision>` element's child is not a `<decisionTable>` (e.g. a literal expression or boxed context), and verify a unit test asserts this error
- [x] 2.4 Make `DecodeXML` reject malformed XML input (not well-formed XML) and structurally inconsistent rules (`<inputEntry>`/`<outputEntry>` count not matching declared `<input>`/`<output>` count) using the same validation error type as the equivalent JSON case (`Validate`/`ValidationError`), and verify unit tests cover both cases

## 3. Encoding

- [x] 3.1 Implement `EncodeXML(w io.Writer, def DecisionTableDefinition) error` that builds the `xmlDefinitions` intermediate tree from a `DecisionTableDefinition` (inverse of the decode mapping in 2.1) and marshals it as a well-formed OMG DMN XML document with a `<definitions>` root in the DMN XML Schema namespace, and verify a unit test asserts the output is well-formed XML containing `<decision>`/`<decisionTable>`
- [x] 3.2 Add a round-trip test that encodes a `DecisionTableDefinition` (covering at least one clause of each `ClauseType` and a rule using an interval, a relational operator, and a wildcard entry) and decodes the result, asserting equality with the original via `reflect.DeepEqual`

## 4. Integration and verification

- [x] 4.1 Verify a `DecisionTableDefinition` decoded via `DecodeXML` compiles and evaluates correctly through the existing unmodified `Compile`/`Evaluate` pipeline, with a test evaluating a decoded XML table equivalent to an existing JSON fixture (e.g. `examples/volume-discounts/volume-discount.dmn.json`) and asserting matching results
- [x] 4.2 Run `go test ./...` and `go vet ./...` and confirm all tests pass with no new vet warnings
