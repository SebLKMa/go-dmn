## Why

Today the engine only accepts decision tables as hand-written JSON
(`.dmn.json`, see `examples/volume-discounts/`), a format this library
invented for itself. The OMG DMN standard's actual interchange format is
XML, and real-world decision tables (authored in tools like Camunda
Modeler, Trisotech, or the DMN TCK test suite) are distributed as `.dmn`
XML files. Without an XML reader, every such table must be manually
transcribed into the JSON shape before this engine can evaluate it, which
is error-prone and blocks interop with the wider DMN ecosystem.

## What Changes

- Add an XML decoder that reads the `<decision><decisionTable>` subtree of
  an OMG DMN XML document (`dmn:definitions` root, DMN XML Schema
  namespace) and produces the existing `DecisionTableDefinition` struct
  (`model.go`), unchanged.
- Add an XML encoder that serializes a `DecisionTableDefinition` back out
  as a well-formed OMG DMN XML document containing an equivalent
  `<decision><decisionTable>` subtree.
- Scope is decision tables only: DRD wiring (`<inputData>`,
  `<knowledgeSource>`, `<businessKnowledgeModel>`, requirement links),
  `<informationRequirement>`, Boxed Expressions other than decision table,
  and DMNDI diagram layout are out of scope, consistent with this
  library's existing non-goals (`dmn.go` package doc; original
  `dmn_decision_table` design.md Non-Goals).
- The XML path is additive: `.dmn.json` remains fully supported, `Compile`
  and `Evaluate` are unchanged, and no existing example or test is
  modified to use XML.
- Uses only `encoding/xml` from the standard library, preserving the
  package's zero-dependency stance (see `dmn_decision_table`'s design.md
  decision on `encoding/json`).

## Capabilities

### New Capabilities
- `dmn-xml-interchange`: Decode an OMG DMN XML `<decision><decisionTable>`
  document into a `DecisionTableDefinition`, and encode a
  `DecisionTableDefinition` back into an equivalent OMG DMN XML document.

### Modified Capabilities
(none — decision table evaluation semantics in `dmn_decision_table` are
unchanged; this only adds a new input/output encoding.)

## Impact

- **New code**: an XML decode/encode module (new file(s) in the `dmn`
  package, e.g. `xml.go`) exposing something like `DecodeXML(io.Reader)
  (DecisionTableDefinition, error)` and `EncodeXML(io.Writer,
  DecisionTableDefinition) error`, following the existing flat
  single-package layout.
- **No changes** to `model.go`, `validate.go`, `table.go`, `feel.go`,
  `typecheck.go`, `resolve.go`, `evaluate.go`, or `errors.go` — a decoded
  definition flows into the existing `Compile`/`Evaluate` pipeline
  unmodified.
- **Dependencies**: none added; `encoding/xml` is part of the Go standard
  library.
- **Examples/docs**: none of the existing examples are required to
  change; a new fixture/example may be added later to demonstrate loading
  a `.dmn` XML file, but is not required by this change.
