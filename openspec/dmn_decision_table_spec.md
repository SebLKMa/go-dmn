# OpenSpec: Selective DMN Decision Table Engine Specification

## 1. Overview & Scope
This specification defines a **lightweight, selective implementation** of the OMG Decision Model and Notation (DMN) standard, focused exclusively on **Decision Tables** using a subset of **S-FEEL (Simplified Friendly Enough Expression Language)**. It intentionally omits Decision Requirement Diagrams (DRDs), Full FEEL compilers, and Boxed Expressions to optimize for low-latency execution and ease of integration.

---

## 2. Data Models (JSON Schema)

### 2.1 Decision Table Definition
The structural blueprint of a decision table.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "DMNDecisionTableDefinition",
  "type": "object",
  "properties": {
    "tableName": { "type": "string" },
    "hitPolicy": { 
      "type": "string", 
      "enum": ["U", "F", "A", "P", "C", "C+", "C<", "C>", "C#", "R", "O"] 
    },
    "inputs": {
      "type": "array",
      "items": {
        "$ref": "#/$defs/ClauseDefinition"
      }
    },
    "outputs": {
      "type": "array",
      "items": {
        "$ref": "#/$defs/ClauseDefinition"
      }
    },
    "rules": {
      "type": "array",
      "items": {
        "$ref": "#/$defs/RuleDefinition"
      }
    }
  },
  "required": ["tableName", "hitPolicy", "inputs", "outputs", "rules"],
  "$defs": {
    "ClauseDefinition": {
      "type": "object",
      "properties": {
        "id": { "type": "string" },
        "name": { "type": "string" },
        "type": { "type": "string", "enum": ["string", "number", "boolean", "date"] },
        "allowedValues": { "type": "array", "items": { "type": "string" } }
      },
      "required": ["id", "name", "type"]
    },
    "RuleDefinition": {
      "type": "object",
      "properties": {
        "ruleId": { "type": "string" },
        "inputEntries": {
          "type": "array",
          "items": { "type": "string" }
        },
        "outputEntries": {
          "type": "array",
          "items": { "type": "string" }
        }
      },
      "required": ["ruleId", "inputEntries", "outputEntries"]
    }
  }
}
```

### 2.2 Execution Payload
The context data passed into the engine to run the evaluation.

```json
{
  "title": "DMNEvaluationRequest",
  "type": "object",
  "properties": {
    "context": {
      "type": "object",
      "additionalProperties": {
        "type": ["string", "number", "boolean"]
      }
    }
  },
  "required": ["context"]
}
```

---

## 3. Component Specifications

### 3.1 Input Entry Grammar (Selective S-FEEL Subset)
The engine must parse and evaluate the strings within `inputEntries` using the following token rules:

* **Wildcard (`-`):** Automatically evaluates to `true`. Skips evaluation for that input.
* **Exact Literals:** Direct comparison matching the type (e.g., `"VIP"`, `100`, `true`).
* **Relational Operators:** Mathematical operators `<, >, <=, >=, !=` followed by a number (e.g., `< 50`).
* **Intervals / Ranges:**
  * Inclusive: `[18..65]` (translates to `>= 18 AND <= 65`).
  * Exclusive: `(18..65)` (translates to `> 18 AND < 65`).
* **Value Lists:** Comma-separated values evaluated as a logical OR (e.g., `"Gold", "Silver"`).

### 3.2 Evaluation Logic Matrix
A single rule row matches **if and only if** all of its individual `inputEntries` evaluate to `true` against the provided `context` (logical AND).

### 3.3 Hit Policy Implementation Requirements
When multiple rules evaluate to `true`, the engine must apply the selected `hitPolicy` strictly as defined below:

| Code | Hit Policy | Single/Multiple | Engine Resolution Rule |
| :--- | :--- | :--- | :--- |
| **`U`** | Unique | Single | Exactly one rule must match. If >1 match, throw `RuntimeConflictException`. |
| **`F`** | First | Single | Evaluate sequentially top-to-bottom. Return immediately upon the first match. |
| **`A`** | Any | Single | Multiple rules can match, but all matched outputs must be identical. If different, throw `RuntimeConflictException`. |
| **`P`** | Priority | Single | Multiple rules can match. Return the output with the highest matching value defined in `allowedValues`. |
| **`C`** | Collect | Multiple | Return an array containing all matching output entries in an arbitrary order. |
| **`C+`** | Collect (Sum) | Multiple | Compute the arithmetic sum of all matching numeric output entries. |
| **`C<`** | Collect (Min) | Multiple | Return the lowest numeric output value among the matches. |
| **`C>`** | Collect (Max) | Multiple | Return the highest numeric output value among the matches. |
| **`C#`** | Collect (Count)| Multiple | Return an integer count of the total matching rules. |

---

## 4. Validation Rules (Guards & Exceptions)

### 4.1 Design-Time Validation (Schema Guards)
* **Structural Congruence:** The length of `inputEntries` and `outputEntries` within every object in the `rules` array *must* exactly match the length of the root `inputs` and `outputs` arrays respectively.
* **Priority Bound:** If `hitPolicy` is set to `P` or `O`, the corresponding output clause *must* include an populated `allowedValues` array to act as the sorting hierarchy.

### 4.2 Runtime Exception Triggers
* **`DMNCompletenessException`**: Occurs when a payload is passed and zero rules are matched, provided no fallback rule (`-`) exists.
* **`DMNTypeMismatchException`**: Occurs when the data type of the variable inside the `context` payload does not match the designated type defined in the table's `inputs` schema.