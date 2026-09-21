## Why

`go-dmn` now has an OMG DMN XML decoder (`DecodeXML`, capability
`dmn-xml-interchange`) but no runnable example showing it in use. The
existing `examples/volume-discounts/` example demonstrates the same
volume-discount pricing clause loaded from JSON via `dmn.Compile`; a parallel
XML-backed example gives future users a template for consuming a real `.dmn`
XML file instead of this library's own JSON shape, and exercises `DecodeXML`
end-to-end beyond unit tests.

## What Changes

- Add a new example directory `examples/volume-discounts-xml/` containing a
  Go CLI program with the same behavior as `examples/volume-discounts/`
  (same contract clause, same discount tiers, same `Unit-Price`/`Discount`/
  `Net-Unit-Price`/`Volume-Discounted-Price` output), but loading its
  decision table from an OMG DMN XML file (`volume-discount.dmn.xml`) via
  `dmn.DecodeXML` instead of from JSON via `encoding/json`.
- Duplicate `contract-clause.md` into the new example directory so it stays
  self-contained like the existing example (each `examples/*/` directory
  already carries its own supporting docs and fixture file).
- No changes to the `dmn` package itself — this only exercises the existing
  `DecodeXML`/`Compile`/`Evaluate` pipeline.

## Capabilities

### New Capabilities
- `volume-discount-xml-example`: a runnable example (OMG DMN XML file + Go
  CLI) that evaluates the contract clause's volume-discount tiers, loaded
  via `dmn.DecodeXML`, and prints the resulting pricing breakdown for a
  given purchase volume.

### Modified Capabilities
(none — this only consumes the existing `dmn_decision_table` and
`dmn-xml-interchange` capabilities; it does not change their requirements)

## Impact

- Affected code: new files under `examples/volume-discounts-xml/` only (a
  `.dmn.xml` decision table file, a `main.go` CLI, and a copy of
  `contract-clause.md`). No changes to the root `dmn` package or to
  `examples/volume-discounts/`.
- Depends on the `dmn-xml-interchange` capability's `DecodeXML` and the
  `dmn_decision_table` capability's `Compile`/`Evaluate`, both already
  implemented.
- Picked up automatically by `make build` (any `examples/*/main.go`).
- No new external dependencies, no breaking changes.
