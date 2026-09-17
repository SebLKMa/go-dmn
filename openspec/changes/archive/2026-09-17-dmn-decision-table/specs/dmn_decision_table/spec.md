## ADDED Requirements

### Requirement: Decision Table Definition
The system SHALL accept a decision table definition consisting of a `tableName`, a `hitPolicy`, an `inputs` array, an `outputs` array, and a `rules` array, each clause having an `id`, `name`, and `type` (`string`, `number`, `boolean`, or `date`), and each rule having a `ruleId`, `inputEntries`, and `outputEntries`. `inputs`, `outputs`, and `rules` may be empty.

#### Scenario: Valid definition is accepted
- **WHEN** a decision table definition is provided with `tableName`, `hitPolicy`, `inputs`, `outputs`, and `rules` all present
- **THEN** the system loads the definition without error

#### Scenario: Definition missing a required field is rejected
- **WHEN** a decision table definition omits `tableName` or `hitPolicy`
- **THEN** the system rejects the definition before any evaluation is attempted

### Requirement: S-FEEL Input Entry Grammar
The system SHALL parse and evaluate each string in a rule's `inputEntries` using the following token rules: a wildcard `-` SHALL always evaluate to `true`; an exact literal SHALL be compared directly against the context value of the matching type; a relational operator (`<`, `>`, `<=`, `>=`, `!=`) followed by a number SHALL be evaluated as that comparison; an inclusive interval `[a..b]` SHALL evaluate to `true` when the context value is `>= a AND <= b`; an exclusive interval `(a..b)` SHALL evaluate to `true` when the context value is `> a AND < b`; a comma-separated list of values SHALL evaluate as a logical OR across the listed values.

#### Scenario: Wildcard always matches
- **WHEN** an input entry is `-`
- **THEN** the entry evaluates to `true` regardless of the context value

#### Scenario: Exact literal match
- **WHEN** an input entry is `"VIP"` and the context value for that input is `"VIP"`
- **THEN** the entry evaluates to `true`

#### Scenario: Relational operator
- **WHEN** an input entry is `< 50` and the context value for that input is `40`
- **THEN** the entry evaluates to `true`

#### Scenario: Inclusive interval
- **WHEN** an input entry is `[18..65]` and the context value for that input is `18` or `65`
- **THEN** the entry evaluates to `true`

#### Scenario: Exclusive interval
- **WHEN** an input entry is `(18..65)` and the context value for that input is `18` or `65`
- **THEN** the entry evaluates to `false`

#### Scenario: Value list as logical OR
- **WHEN** an input entry is `"Gold", "Silver"` and the context value for that input is `"Silver"`
- **THEN** the entry evaluates to `true`

### Requirement: Rule Matching
A rule SHALL match a given context if and only if every one of its `inputEntries` evaluates to `true` against the corresponding context value (logical AND across all input entries in the rule).

#### Scenario: All input entries true
- **WHEN** every input entry of a rule evaluates to `true` against the context
- **THEN** the rule is considered a match

#### Scenario: One input entry false
- **WHEN** at least one input entry of a rule evaluates to `false` against the context
- **THEN** the rule is not considered a match

### Requirement: Hit Policy - Unique (U)
When `hitPolicy` is `U`, exactly one rule SHALL match. If more than one rule matches, the system SHALL raise a `DMNRuntimeConflictException`.

#### Scenario: Single match returns its output
- **WHEN** `hitPolicy` is `U` and exactly one rule matches the context
- **THEN** the system returns that rule's output entries

#### Scenario: Multiple matches raise a conflict
- **WHEN** `hitPolicy` is `U` and more than one rule matches the context
- **THEN** the system raises a `DMNRuntimeConflictException`

### Requirement: Hit Policy - First (F)
When `hitPolicy` is `F`, the system SHALL evaluate rules sequentially top-to-bottom and return the output of the first matching rule, without evaluating subsequent rules.

#### Scenario: First match wins
- **WHEN** `hitPolicy` is `F` and two rules would match the context
- **THEN** the system returns the output entries of whichever matching rule appears first in the `rules` array

### Requirement: Hit Policy - Any (A)
When `hitPolicy` is `A`, multiple rules may match, but all matched rules SHALL produce identical output entries. If matched rules produce different outputs, the system SHALL raise a `DMNRuntimeConflictException`.

#### Scenario: Multiple matches with identical outputs
- **WHEN** `hitPolicy` is `A` and every matching rule has the same output entries
- **THEN** the system returns those output entries

#### Scenario: Multiple matches with differing outputs
- **WHEN** `hitPolicy` is `A` and matching rules have different output entries
- **THEN** the system raises a `DMNRuntimeConflictException`

### Requirement: Hit Policy - Priority (P)
When `hitPolicy` is `P`, multiple rules may match. The system SHALL return the output entries of the matching rule whose output value has the highest priority as defined by the order of values in the output clause's `allowedValues` array.

#### Scenario: Highest-priority match wins
- **WHEN** `hitPolicy` is `P` and multiple rules match with outputs present at different positions in `allowedValues`
- **THEN** the system returns the output entries of the match whose output value ranks highest in `allowedValues`

### Requirement: Hit Policy - Collect Family (C, C+, C<, C>, C#)
When `hitPolicy` is `C`, the system SHALL return an array of the output entries of all matching rules, in arbitrary order. When `hitPolicy` is `C+`, the system SHALL return the arithmetic sum of all matching rules' numeric output entries. When `hitPolicy` is `C<`, the system SHALL return the lowest numeric output value among matching rules. When `hitPolicy` is `C>`, the system SHALL return the highest numeric output value among matching rules. When `hitPolicy` is `C#`, the system SHALL return the count of matching rules as an integer.

#### Scenario: Collect returns all outputs
- **WHEN** `hitPolicy` is `C` and two rules match
- **THEN** the system returns an array containing both rules' output entries

#### Scenario: Collect sum
- **WHEN** `hitPolicy` is `C+` and matching rules have numeric outputs `10` and `20`
- **THEN** the system returns `30`

#### Scenario: Collect min
- **WHEN** `hitPolicy` is `C<` and matching rules have numeric outputs `10` and `20`
- **THEN** the system returns `10`

#### Scenario: Collect max
- **WHEN** `hitPolicy` is `C>` and matching rules have numeric outputs `10` and `20`
- **THEN** the system returns `20`

#### Scenario: Collect count
- **WHEN** `hitPolicy` is `C#` and three rules match
- **THEN** the system returns `3`

### Requirement: Unsupported Hit Policies (R, O)
The `hitPolicy` values `R` (Rule Order) and `O` (Output Order) SHALL be accepted by the schema but are not implemented by this change. The system SHALL raise a clear "unsupported hit policy" error at evaluation time when a decision table declares one of these values.

#### Scenario: Rule Order hit policy is rejected at evaluation
- **WHEN** a decision table's `hitPolicy` is `R` and evaluation is attempted
- **THEN** the system raises an "unsupported hit policy" error rather than silently applying different semantics

#### Scenario: Output Order hit policy is rejected at evaluation
- **WHEN** a decision table's `hitPolicy` is `O` and evaluation is attempted
- **THEN** the system raises an "unsupported hit policy" error rather than silently applying different semantics

### Requirement: Design-Time Validation
The system SHALL validate, before evaluation, that every rule's `inputEntries` and `outputEntries` lengths exactly match the lengths of the root `inputs` and `outputs` arrays respectively, and that when `hitPolicy` is `P` (or `O`, once supported), the corresponding output clause has a non-empty `allowedValues` array.

#### Scenario: Rule entry count mismatch is rejected
- **WHEN** a rule's `inputEntries` or `outputEntries` length does not match the number of declared `inputs` or `outputs`
- **THEN** the system rejects the definition before evaluation

#### Scenario: Priority hit policy without allowed values is rejected
- **WHEN** `hitPolicy` is `P` and the output clause has no `allowedValues` (or an empty array)
- **THEN** the system rejects the definition before evaluation

### Requirement: Runtime Completeness Checking
When an evaluation payload is provided and zero rules match, and no rule uses a fallback (`-`) entry that would have matched, the system SHALL raise a `DMNCompletenessException`.

#### Scenario: No rule matches and no fallback exists
- **WHEN** a context is evaluated against a decision table and no rule matches
- **THEN** the system raises a `DMNCompletenessException`

### Requirement: Runtime Type Checking
When the type of a value in the evaluation context does not match the type declared for the corresponding input clause, the system SHALL raise a `DMNTypeMismatchException` rather than attempting evaluation with mismatched types.

#### Scenario: Context value type does not match input type
- **WHEN** an input clause declares type `number` and the context supplies a `string` value for that input
- **THEN** the system raises a `DMNTypeMismatchException`
