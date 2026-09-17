## 1. Module Setup

- [x] 1.1 Initialize the Go module (`go mod init github.com/seblkma/go-dmn`) and verify `go build ./...` succeeds on an empty package
- [x] 1.2 Create the `dmn` package skeleton (single flat package per design.md) with a top-level doc comment summarizing scope, and verify `go vet ./...` passes

## 2. Definition Model

- [x] 2.1 Define `DecisionTableDefinition`, `ClauseDefinition`, and `RuleDefinition` Go structs with JSON tags mirroring spec.md section 2.1, and verify a sample JSON fixture decodes via `encoding/json` in a unit test
- [x] 2.2 Define the `DMNEvaluationRequest`/context payload type per spec.md section 2.2, and verify a sample context JSON decodes correctly in a unit test

## 3. Design-Time Validation

- [x] 3.1 Implement structural congruence checks (rule `inputEntries`/`outputEntries` length must match `inputs`/`outputs` length) per the Design-Time Validation requirement, and verify with unit tests for both a matching and a mismatched rule
- [x] 3.2 Implement the priority-bound check (`hitPolicy` `P` requires a non-empty `allowedValues` on the output clause) per the Design-Time Validation requirement, and verify with unit tests for present/missing/empty `allowedValues`
- [x] 3.3 Wire both checks into a `Validate(DecisionTableDefinition) error` function called before evaluation, and verify a definition failing either check returns an error without attempting evaluation

## 4. S-FEEL Input Entry Parser

- [x] 4.1 Implement wildcard (`-`) and exact literal parsing/evaluation, and verify with table-driven tests against string, number, and boolean context values
- [x] 4.2 Implement relational operator parsing (`<`, `>`, `<=`, `>=`, `!=`) against numeric values, and verify with table-driven tests including boundary values
- [x] 4.3 Implement inclusive (`[a..b]`) and exclusive (`(a..b)`) interval parsing, and verify with table-driven tests covering values at, inside, and outside each boundary
- [x] 4.4 Implement comma-separated value list parsing evaluated as logical OR, and verify with table-driven tests including a value present and absent from the list
- [x] 4.5 Compile each rule's input entries once at load time (per design.md) into an evaluable form, and verify via a benchmark or unit test that repeated evaluation against different contexts does not re-parse the entry string

## 5. Runtime Type Checking

- [x] 5.1 Implement context-value-to-clause-type checking (`string`, `number`, `boolean`, `date`) that runs before input-entry evaluation for each input, and verify with unit tests for a matching type and a mismatched type
- [x] 5.2 Implement `TypeMismatchError` (implementing `error`, matching `DMNTypeMismatchException` from spec.md section 4.2) and verify it is returned (and distinguishable via `errors.As`) on a type mismatch

## 6. Rule Matching

- [x] 6.1 Implement per-rule matching (logical AND across all of a rule's input entries) per the Rule Matching requirement, and verify with unit tests for all-true, one-false, and empty-entries cases
- [x] 6.2 Implement the "no rule matches and no fallback" completeness check and `CompletenessError` (matching `DMNCompletenessException`), and verify with a unit test where zero rules match and none use `-` as a catch-all

## 7. Hit Policy Resolution

- [x] 7.1 Implement `U` (Unique) resolution including `ConflictError` on multiple matches, and verify with unit tests for single-match and multi-match cases
- [x] 7.2 Implement `F` (First) resolution, and verify with a unit test that only the first matching rule's output is returned when multiple rules would match
- [x] 7.3 Implement `A` (Any) resolution including `ConflictError` on differing outputs, and verify with unit tests for identical-output and differing-output cases
- [x] 7.4 Implement `P` (Priority) resolution using output clause `allowedValues` order as the ranking, and verify with a unit test asserting the documented ranking direction (last element = highest priority, per design.md)
- [x] 7.5 Implement `C` (Collect) resolution returning all matching outputs, and verify with a unit test on a multi-match context
- [x] 7.6 Implement `C+`, `C<`, `C>`, `C#` numeric collect resolutions, and verify each with a unit test using the sum/min/max/count example values from spec.md section 3.3
- [x] 7.7 Implement `UnsupportedHitPolicyError` returned for `R` and `O` at evaluation time, and verify with unit tests for both values

## 8. Public API

- [x] 8.1 Implement a top-level `Evaluate(DecisionTableDefinition, context) (result any, err error)` (or equivalent) function that runs validation, rule matching, type checking, and hit-policy resolution in sequence, and verify with an end-to-end unit test per hit policy using a small fixture table
- [x] 8.2 Add package-level godoc examples (`Example...` test functions) demonstrating loading a definition and evaluating a context, and verify they compile and run via `go test ./...`

## 9. Test Coverage Verification

- [x] 9.1 Run `go test ./... -cover` and verify all requirements/scenarios in `specs/dmn_decision_table/spec.md` have at least one corresponding test case
