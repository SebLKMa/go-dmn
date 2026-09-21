# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`go-dmn` is a selective Go implementation of the OMG DMN (Decision Model and
Notation) standard. It implements **only Decision Tables** evaluated with a
subset of S-FEEL (Simplified FEEL). It intentionally omits Decision
Requirement Diagrams, a full FEEL compiler, and Boxed Expressions (see the
package doc comment in `dmn.go`).

## Commands

```sh
go test ./...              # run all tests
go test ./... -run TestXxx # run a single test by name
go test -v ./...           # verbose output
go vet ./...               # static checks
make build                 # builds every examples/*/main.go into ./bin,
                            # copying any *.json fixtures alongside the binary
make clean                 # removes ./bin
```

There is no separate lint config; `go vet` is the only static check in use.

## Architecture

The engine is a pipeline over a single JSON-serializable definition type,
`DecisionTableDefinition` (`model.go`). Each stage lives in its own file and
the flow for a single evaluation is:

1. **`validate.go`** — `Validate` performs design-time checks (required
   fields, structural congruence between each rule's entry counts and the
   table's declared inputs/outputs, and that hit policies `P`/`O` have
   non-empty `allowedValues`).
2. **`table.go`** — `Compile` runs `Validate`, then parses every rule's
   `InputEntries` strings into `entryMatcher` values once, producing a
   `*CompiledTable`. This compile-once/evaluate-many split exists so callers
   evaluating the same table repeatedly don't re-parse S-FEEL entries per
   call — use `Compile` + `CompiledTable.Evaluate` in that case, and the
   package-level `Evaluate` (in `evaluate.go`) only for one-shot use.
3. **`feel.go`** — implements the S-FEEL input-entry grammar and holds all
   the `entryMatcher` implementations: wildcard (`-`), exact literal, comma
   list (logical OR), relational (`<`, `<=`, `>`, `>=`, `!=`), and interval
   (`[a..b]`, `(a..b)`, and mixed-bracket variants). `toComparable` is the
   shared numeric/date coercion used by relational and interval matchers.
4. **`typecheck.go`** — `checkClauseType` verifies a context value's dynamic
   Go type against a clause's declared `ClauseType` (string/number/
   boolean/date) at evaluation time, independent of S-FEEL parsing.
5. **`table.go`** (`MatchingRules`) — evaluates every compiled rule's
   matchers against an `EvaluationContext` (a `map[string]any`, keyed by
   clause name) with logical AND across a rule's inputs, and raises
   `*CompletenessError` if zero rules match.
6. **`resolve.go`** — `Resolve` applies the table's `HitPolicy` to the
   matched rules to produce the final result. Each hit policy
   (`U`/`F`/`A`/`P`/`C`/`C+`/`C<`/`C>`/`C#`) has its own `resolveXxx`
   method; `R`/`O` are accepted by the schema but return
   `*UnsupportedHitPolicyError`. `U` and `A` can raise `*ConflictError` when
   matched rules can't be resolved to one output.
7. **`evaluate.go`** ties the pipeline together: `CompiledTable.Evaluate` =
   `MatchingRules` + `Resolve`.

All engine errors are typed (`errors.go`): `ValidationError`,
`CompletenessError`, `TypeMismatchError`, `ConflictError`,
`UnsupportedHitPolicyError` — check with `errors.As`, not string matching.

### Examples

`examples/<name>/` directories are standalone `main` packages that load a
`*.dmn.json` file (via `dmn.Compile`) and evaluate it against CLI input. Each
is picked up automatically by `make build` (anything matching
`examples/*/main.go`). See `examples/volume-discounts/` for the reference
shape: a `main.go`, its `*.dmn.json` decision table, and a
`contract-clause.md` describing the business rule the table encodes.

## Spec-driven workflow (OpenSpec)

This repo uses OpenSpec (`openspec/`) as the source of truth for *why* the
engine behaves as it does — `openspec/specs/*/spec.md` are the current specs,
and `openspec/changes/archive/` holds past proposals with their design
rationale. Code comments throughout the engine cite these
(e.g. "spec.md section 3.3", "design.md - ..."); when changing hit-policy or
S-FEEL behavior, check the corresponding spec file first rather than
inferring intent from the code alone.

The change cycle (see `.claude/skills/openspec-*` and `README.md`):
`openspec-propose` → `openspec-apply` → `openspec-archive` (or the `/opsx:*`
slash-command equivalents).
