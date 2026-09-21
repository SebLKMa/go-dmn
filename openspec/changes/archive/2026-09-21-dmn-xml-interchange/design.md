## Context

The engine's public model (`DecisionTableDefinition`, `ClauseDefinition`, `RuleDefinition` in `model.go`) is currently populated only via `encoding/json`. `Compile`/`Validate`/`Evaluate` operate on that Go struct and are format-agnostic already — they never see JSON or XML directly. See proposal.md - Why / What Changes for motivation and scope (decision tables only, parse + serialize, additive to JSON).

## Goals / Non-Goals

**Goals:**
- Decode the `<decision><decisionTable>` subtree of an OMG DMN XML document into the existing `DecisionTableDefinition`, reusing `Validate`/`Compile`/`Evaluate` unchanged.
- Encode a `DecisionTableDefinition` back into a well-formed OMG DMN XML document that round-trips through the decoder.
- Keep the package dependency-free by using only `encoding/xml` from the standard library.

**Non-Goals:**
- No DRD graph modeling (`<inputData>`, `<businessKnowledgeModel>`, `<informationRequirement>` wiring) beyond tolerating and ignoring these elements when decoding.
- No support for boxed expressions other than decision table (literal expression, boxed context, relation, etc.) — a `<decision>` using one of these is a decode error, not a silent no-op.
- No DMNDI diagram interchange (layout/shape/edge elements) — ignored on decode, never emitted on encode.
- No XML Schema validation against the OMG `.xsd` — structural checks are the same hand-written checks `Validate` already performs on the decoded struct, not schema conformance.

## Decisions

**New file `xml.go` in the existing flat `dmn` package**, exposing `DecodeXML(r io.Reader) (DecisionTableDefinition, error)` and `EncodeXML(w io.Writer, def DecisionTableDefinition) error`. Mirrors the existing single-package layout (design.md of `dmn_decision_table`: "nothing external needs to import [parser/model] independently").

**Decode via `encoding/xml` struct tags on a package-private intermediate type (`xmlDefinitions`), not directly into `DecisionTableDefinition`.**
The OMG DMN XML element names (`decision`, `decisionTable`, `input`, `output`, `rule`, `inputEntry`, `outputEntry`, `text`) and attribute names (`typeRef`, `hitPolicy`, `label`) don't match the JSON struct tags already on `DecisionTableDefinition` (`tableName`, `id`, `inputEntries`, ...), and `<inputEntry>`/`<outputEntry>` wrap their literal as a nested `<text>` element rather than being the entry text directly. A private `xmlDefinitions`/`xmlDecision`/`xmlDecisionTable`/`xmlInput`/`xmlOutput`/`xmlRule` tree decoded via `xml:"..."` tags, then mapped field-by-field into `DecisionTableDefinition`, keeps the public model free of XML-specific shape and keeps both decode and encode symmetric (marshal the same intermediate tree back out). Alternative considered: a streaming `xml.Decoder` token-by-token walk — rejected as more code for no benefit given the subtree is small and fits comfortably in memory.

**Hit policy mapping table (OMG `hitPolicy`/`aggregation` attribute values ↔ internal `HitPolicy` codes) is a fixed two-way map, discovered necessary during implementation.**
The OMG DMN XML Schema's `hitPolicy` attribute uses full words (`UNIQUE`, `FIRST`, `ANY`, `PRIORITY`, `COLLECT`, `RULE ORDER`, `OUTPUT ORDER`), not this library's internal short codes (`U`, `F`, `A`, `P`, `C`, `R`, `O`), and `COLLECT`'s sum/min/max/count variants (`C+`/`C<`/`C>`/`C#`) are expressed as a separate `aggregation` attribute (`SUM`/`MIN`/`MAX`/`COUNT`) rather than as distinct `hitPolicy` values. Decoding the attribute value directly into `HitPolicy` (as originally planned) would only work for XML this library itself produced, defeating the proposal's interop goal of reading files from real DMN tools. A fixed map (word+aggregation → code, and its inverse for encoding) is applied the same way as the clause type mapping below. Alternative considered: keep `HitPolicy`'s Go representation as the OMG words instead of the existing short codes — rejected because `HitPolicy` is an existing public type used throughout `resolve.go`'s switch statement and every existing test/example; changing its values would be a breaking change far outside this change's scope.

**Clause `name` is derived from the FEEL variable reference (input `<inputExpression><text>`, output `name` attribute), not from the display `label` attribute — also discovered during implementation.**
`ClauseDefinition.Name` doubles as the evaluation-context lookup key (`table.go`'s `ctx[input.Name]`) and the key of the output map `Evaluate` returns. Real OMG DMN XML separates the human-readable `label` attribute from the actual FEEL variable name, which for inputs lives in the `<inputExpression>` child's `<text>` content and for outputs is the `name` attribute (required by the OMG spec once a table has more than one output; optional, falling back to the decision's own `name`, for a single output). Mapping `label` to `name` — the original plan — would decode successfully but produce a `DecisionTableDefinition` whose input names don't match any real caller's context keys, silently breaking evaluation. Taking the `<inputExpression>` text at face value (without evaluating it as FEEL) mirrors the existing non-goal of not implementing a full FEEL compiler: this library only supports the case where that text is a plain variable reference, the same restriction already implied for input/output entries.

**Type mapping table (`string`/`number`/`boolean`/`date` FEEL type names ↔ `ClauseType`) is a fixed two-way map, not a fallback/guessing heuristic.**
An unrecognized `typeRef` fails decode immediately (Clause Type Mapping requirement) rather than defaulting to `string`, so a table using an unsupported FEEL type (e.g. `dateTime`, `duration`, or a custom item definition) fails loudly instead of silently mis-evaluating.

**Unsupported expression types (`<decision>` without a `<decisionTable>` child) are a decode error, not a skipped decision.**
Silently dropping such a decision would let a caller believe an XML file's full rule set was loaded when part of it was discarded. Ignoring genuinely unrelated sibling elements (`<inputData>`, DMNDI) is safe because they carry no rule content this engine evaluates; skipping a `<decision>` with a real expression this library can't represent is not.

**Round-trip equality is validated at the `DecisionTableDefinition` level, not by comparing XML bytes.**
Byte-for-byte XML equality would be brittle (namespace prefix choice, attribute order, self-closing tags are all valid variations of the same document). The contract is "decode(encode(def)) == def", tested with `reflect.DeepEqual` the same way the JSON round-trip is implicitly tested today via existing example fixtures.

**Namespace handling: accept the OMG DMN XML Schema namespace on decode without requiring a specific version; emit the same namespace on encode.**
Go's `encoding/xml` matches local element names by default; the namespace URI is read but not used to branch behavior, since this change targets one subtree shape rather than multiple schema versions. If a future DMN XML Schema version changes the decision table subtree shape, that's a follow-up change.

## Risks / Trade-offs

- [OMG DMN XML documents in the wild vary in how literals are wrapped (e.g. CDATA vs plain text inside `<text>`, whitespace formatting) which could make decode brittle against real tool output] → Mitigate with test fixtures drawn from more than one authoring pattern (hand-written minimal XML, and, if available, output resembling common modeling tools), trimming whitespace around `<text>` content the same way `feel.go` already trims input-entry strings.
- [A fixed type-mapping table will reject legitimate DMN files using types this engine doesn't evaluate (e.g. `dateTime`)] → Accepted trade-off, consistent with the JSON model's existing four-type scope (`model.go`); documented in the Clause Type Mapping requirement rather than silently coerced.
- [Encoding produces XML this library can decode, but may not be byte-identical to what other DMN tools expect for full interoperability] → Acceptable per proposal.md scope (decision tables only, not full interchange fidelity); documented as a known limitation rather than solved here.
