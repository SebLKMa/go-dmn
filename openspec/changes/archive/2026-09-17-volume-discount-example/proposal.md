## Why

`go-dmn` now has a working `dmn_decision_table` engine but no runnable example showing it applied to a real-world rule set. `examples/volume-discounts/contract-clause.md` describes a concrete volume-discount pricing clause (tiered discounts by cumulative purchase volume) that maps directly onto a single decision table. A worked example — a DMN decision table file plus a small Go CLI that evaluates it — demonstrates the library end-to-end and gives future users a template to copy.

## What Changes

- Add a DMN decision table definition file (JSON, per the `dmn_decision_table` data model) encoding the contract clause's discount tiers: cumulative purchase volume as the sole input, discount rate (and tier name) as outputs, using hit policy `U` (the tiers are non-overlapping).
- Add a Go CLI program in `examples/volume-discounts/` that:
  - Accepts a purchase-volume as a command-line argument.
  - Loads the DMN decision table file and evaluates it via the existing `dmn` package (`dmn.Compile`/`dmn.Evaluate`).
  - Computes and prints `Unit-Price`, `Discount`, `Net-Unit-Price`, and `Volume-Discounted-Price`.
- `Unit-Price` ($50.00, per the clause's "Base Pricing" clause) is a fixed program constant, not a decision-table output — the contract clause treats it as sourced from the Product Catalog, independent of volume tier. `Net-Unit-Price` (`Unit-Price * (1 - Discount)`) and `Volume-Discounted-Price` (`Net-Unit-Price * purchase-volume`) are derived in the Go program from the decision table's `Discount` output and the input volume; they are not decision-table outputs themselves, since they depend on the input volume rather than being a lookup.
- Assumption: purchase volumes below the Silver tier's 500-unit threshold receive a 0% discount ("Standard" tier) — the clause doesn't name this tier explicitly, but its existence is implied as the baseline before any discount applies.

## Capabilities

### New Capabilities
- `volume-discount-example`: a runnable example (DMN file + Go CLI) that evaluates the contract clause's volume-discount tiers and prints the resulting pricing breakdown for a given purchase volume.

### Modified Capabilities
(none — this only consumes the existing `dmn_decision_table` engine, it does not change its requirements)

## Impact

- Affected code: new files under `examples/volume-discounts/` only (a `.dmn.json` decision table file and a `main.go` CLI). No changes to the root `dmn` package.
- Depends on the `dmn_decision_table` capability's public `Compile`/`Evaluate` API and JSON decoding of `DecisionTableDefinition`, both already implemented.
- No new external dependencies (`encoding/json`, `os`, `strconv`, `fmt` from the standard library are sufficient).
- No breaking changes.
