## Context

This is a greenfield implementation — no `go.mod` or Go source exists yet in the repo. The target behavior is fully described in `openspec/specs/dmn_decision_table/spec.md` and formalized as requirements in this change's `specs/dmn_decision_table/spec.md`. See `proposal.md` - Why / What Changes for motivation and scope; this document covers how the Go package is structured.

## Goals / Non-Goals

**Goals:**
- A dependency-free Go package (`go-dmn`) that loads a decision table definition and evaluates it against a context payload.
- A hand-written S-FEEL input-entry parser scoped strictly to the tokens in the spec (wildcard, literal, relational, interval, list) — no general FEEL expression support.
- Clear, typed Go errors for the two runtime exceptions (`DMNCompletenessException`, `DMNTypeMismatchException`) and the two conflict cases (`DMNRuntimeConflictException` for `U` and `A`), plus an explicit unsupported-hit-policy error for `R`/`O`.

**Non-Goals:**
- No DRD (Decision Requirement Diagram) support, no Full FEEL compiler, no Boxed Expressions — per the existing spec's stated scope.
- No implementation of `R` (Rule Order) or `O` (Output Order) hit policies — the existing spec's hit-policy table (section 3.3) defines no resolution rule for them; this change requires a decision from the user before behavior can be authored (see proposal.md - Impact).
- No JSON (de)serialization concerns beyond mapping the existing JSON Schema shapes to Go structs with struct tags — no schema-validation library.

## Decisions

**Package layout: single flat package, e.g. `dmn` at the module root.**
The spec is small and cohesive (one capability: decision tables). A flat package avoids premature internal API boundaries. Alternative considered: splitting `parser` / `evaluator` / `model` into sub-packages — rejected for now since nothing external needs to import them independently, and Go idiom favors fewer, cohesive packages until a real reuse boundary appears.

**Definition model: exported Go structs mirroring the JSON Schema in spec.md section 2.1, decoded via `encoding/json`.**
`DecisionTableDefinition`, `ClauseDefinition`, `RuleDefinition` map 1:1 to the schema. Using standard `encoding/json` with struct tags keeps the dependency graph at zero. Alternative considered: a schema-validation library (e.g. `santhosh-tekuri/jsonschema`) to enforce the JSON Schema directly — rejected as unnecessary; the required structural checks (section 4.1) are a handful of length/presence checks cheaper to hand-write than to pull in a validator for.

**S-FEEL input entry evaluation: a small hand-rolled tokenizer/parser per entry string, compiled once per rule at load time into a closure or small AST node, then evaluated per context value.**
Given the fixed, small grammar (5 token shapes), a hand-written recursive-descent-style parser is simpler and more auditable than a parser-generator or regex-only approach. Regex alone was considered and rejected: intervals and multi-value OR lists need structural parsing (extracting bounds, splitting on top-level commas) that regex captures awkwardly and error-reports poorly.

**Type checking: input clause `type` (`string`, `number`, `boolean`, `date`) is checked against the Go dynamic type of the context value at evaluation time**, using `int`/`float64`/`string`/`bool`/`time.Time` (or a `date`-parseable `string`) as the accepted Go representations. Mismatches raise `DMNTypeMismatchException` before any input-entry parsing runs for that value, per Runtime Type Checking requirement.

**Hit policy resolution: a single `resolve(hitPolicy, matches []RuleDefinition) (any, error)` step run after rule matching, dispatching on the `hitPolicy` string via a switch.**
Keeps hit-policy semantics (section 3.3 of the spec) in one place, separate from rule matching, so each policy's behavior maps directly to one switch case and is independently testable. `P` priority resolution reads clause `allowedValues` order as the ranking; `C+`/`C<`/`C>` require the matched output values to parse as numbers, reusing the same numeric coercion used for relational/interval input parsing.

**Errors: typed error values/structs (not just `errors.New` strings) implementing the standard `error` interface**, e.g. `CompletenessError`, `TypeMismatchError`, `ConflictError`, `UnsupportedHitPolicyError`, so callers can `errors.As` to distinguish exception kinds, matching the distinct exception names in the spec (section 4.2) without introducing a custom exception hierarchy (Go has no exceptions).

## Risks / Trade-offs

- [Hand-written S-FEEL parser could have subtle grammar bugs, especially around interval boundary inclusivity and negative numbers] → Mitigate with table-driven unit tests covering every token shape from the spec, including boundary values (e.g. exactly `18` and `65` for both inclusive and exclusive intervals).
- [`P` priority hit policy depends on `allowedValues` order as an implicit ranking, which is easy to get backwards (highest-first vs. lowest-first)] → The spec says "highest matching value defined in `allowedValues`"; treat array order as ascending priority (last element = highest) and document this explicitly in the requirement/godoc so it's unambiguous, and cover it with a test.
- [Type checking against Go's dynamic `any` context values is inherently loose (e.g. is `10` an `int` or a `float64`? is a date a `string` or `time.Time`?)] → Normalize accepted context value shapes explicitly in the public API's input type (documented in godoc) rather than trying to accept every possible Go numeric/date representation.

## Open Questions

- Should `R` and `O` hit policies be defined and implemented in a follow-up change, or is "unsupported" the intended permanent behavior? Left for a future proposal since it doesn't affect this change's specs, approach, or tasks.
