## Context

The `dmn` package (root of this module) already implements `DecisionTableDefinition`, `Compile`, and `Evaluate` per the `dmn_decision_table` capability. This change only adds example artifacts under `examples/volume-discounts/`; see proposal.md - Why for motivation. The source contract clause is `examples/volume-discounts/contract-clause.md`.

## Goals / Non-Goals

**Goals:**
- Show a complete, realistic use of the `dmn` package: a hand-authored DMN decision table file, loaded and evaluated from a small Go `main` program.
- Keep the decision table focused on the one thing that's actually a *decision* (which discount tier applies) rather than encoding constants that don't vary with the input.

**Non-Goals:**
- No general-purpose CLI framework or flag parsing library — a single positional argument is simple enough for `os.Args`.
- No persistence, no reading the DMN file from anywhere other than a fixed path relative to the example program.
- Not a change to the `dmn` package's public API or behavior.

## Decisions

**Decision table scope: input is `purchaseVolume` (number); outputs are `discount` (number) and `tier` (string). `Unit-Price` is a Go constant, not a table output.**
The contract clause's base price ($50.00) comes from the Product Catalog and does not vary with volume — it isn't a decision. Modeling it as a table output would mean every rule row repeats the same literal, adding noise without adding decision logic. Keeping it as a named constant in `main.go` (with a comment citing the clause) is more direct. Alternative considered: making `unitPrice` a table output for symmetry with the other three printed values — rejected because it misrepresents a constant as if it were tier-dependent.

**Hit policy: `U` (Unique).**
The four tiers (`< 500`, `[500..1499]`, `[1500..4999]`, `>= 5000`) are constructed to be non-overlapping and exhaustive for any non-negative volume, so exactly one rule should ever match. `U` makes that invariant explicit and gives a loud `ConflictError` if the ranges are ever edited to overlap by mistake, rather than silently picking a first-match under `F`.

**Baseline "Standard" (0%) tier is added explicitly, with an interval upper-bounded at 499, rather than a wildcard `-` catch-all row.**
The contract clause never names a below-500 tier, but the pricing logic needs a defined discount for that range. An explicit `[0..499]` row keeps the table's coverage visible at a glance (matches the spec's own tier-boundary language) instead of relying on a wildcard whose meaning ("everything else") is less self-documenting for a worked example.

**DMN file format: JSON matching `DecisionTableDefinition`'s struct tags exactly, loaded via `encoding/json` + `dmn.Compile`.**
No new parsing code is needed; this exercises exactly the public decode path documented in the `dmn_decision_table` capability. File extension `.dmn.json` (not `.dmn`) to be explicit that it's JSON, not the OMG DMN XML interchange format (which this library does not implement).

**CLI: single positional argument via `os.Args[1]`, parsed with `strconv.ParseFloat`.**
Matches the "sole command-line argument" requirement without pulling in `flag` for one value. A non-numeric or missing argument prints a usage line to stderr and exits via `os.Exit(1)`, per the Purchase Volume CLI Argument requirement.

## Risks / Trade-offs

- [A negative purchase volume is not explicitly rejected by the decision table itself — it simply falls into the `[0..499]` "Standard" row's interval only if `>= 0`, and otherwise matches nothing, triggering a `CompletenessError`] → Acceptable for an example: the resulting error is still a clear, non-crashing failure mode, and negative purchase volumes are outside the contract clause's domain.
- [Hand-maintaining tier boundaries in two places (the DMN file's rules and this design doc's restatement of them) risks drifting from the contract clause if either is edited later] → Mitigated by keeping the DMN file as the single source of truth for the boundaries; the spec's scenarios and design notes are descriptive, not duplicated logic.
