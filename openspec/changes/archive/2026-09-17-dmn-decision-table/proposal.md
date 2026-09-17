## Why

`go-dmn` currently has no implementation — only a hand-authored technical specification (`openspec/specs/dmn_decision_table/spec.md`) describing a selective DMN Decision Table engine. To make the project usable as a Go library, that specification needs a first Go implementation: parsing decision table definitions, evaluating S-FEEL input entries against a context payload, and applying hit policies to produce a result.

## What Changes

- Add a Go package that models the DMN decision table definition (inputs, outputs, rules, hit policy) as described in the existing spec.
- Add a parser/evaluator for the selective S-FEEL input entry grammar: wildcard `-`, exact literals, relational operators (`<`, `>`, `<=`, `>=`, `!=`), inclusive/exclusive intervals, and comma-separated value lists.
- Add rule matching logic (logical AND across a rule's input entries; a rule matches when all its input entries evaluate to `true` against the context).
- Add hit policy resolution for all policies in the spec: `U`, `F`, `A`, `P`, `C`, `C+`, `C<`, `C>`, `C#` (`R` and `O` appear in the schema enum but are out of scope for this change — see Impact).
- Add design-time validation (structural congruence between clause counts and rule entry counts; `allowedValues` required when `hitPolicy` is `P`).
- Add runtime error conditions: no rule matches and no fallback (`-`) rule exists; context value type does not match the declared input type.
- Add a public Go API to load a decision table definition and evaluate it against a context payload, returning a result or a typed error.

## Capabilities

### New Capabilities
(none — the capability spec already exists in the repo)

### Modified Capabilities
- `dmn_decision_table`: formalizes the existing free-form specification as OpenSpec-tracked requirements (data model, S-FEEL grammar, evaluation/hit-policy semantics, and validation/error behavior) ahead of the first implementation. No behavior described in the existing spec is changed.

## Impact

- Affected code: none yet (greenfield) — this change adds a new Go module/package tree (no `go.mod` currently exists in the repo).
- Out of scope for this change: hit policies `R` (Rule Order) and `O` (Output Order), which are listed in the `hitPolicy` schema enum but have no resolution rule defined in section 3.3 of the existing spec. They are left unimplemented (return a clear "unsupported hit policy" error) until a follow-up change defines their semantics.
- No external dependencies are introduced; the S-FEEL subset is small enough to hand-write a parser rather than pull in a full FEEL library.
- No breaking changes (no prior public API exists).
