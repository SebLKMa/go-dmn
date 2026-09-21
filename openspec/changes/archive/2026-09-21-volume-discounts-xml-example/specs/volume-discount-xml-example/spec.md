## Purpose

A runnable example demonstrating the `dmn-xml-interchange` capability's `DecodeXML` by encoding the volume-discount pricing clause from `examples/volume-discounts/contract-clause.md` as an OMG DMN XML decision table and evaluating it from a small Go CLI.

## ADDED Requirements

### Requirement: XML Decision Table Source
The example SHALL define its decision table as an OMG DMN XML file (not JSON) and SHALL load it using the `dmn-xml-interchange` capability's `DecodeXML`, producing a `DecisionTableDefinition` that is then compiled and evaluated the same way as any other decision table.

#### Scenario: Table loads from XML
- **WHEN** the example program starts
- **THEN** it decodes its decision table from an OMG DMN XML file via `DecodeXML` rather than parsing JSON

### Requirement: Discount Tier Decision Table
The example SHALL provide a DMN decision table, valid per the `dmn_decision_table` data model, with a single numeric input clause representing cumulative purchase volume, hit policy `U` (Unique), and rules encoding the contract clause's tiers: below 500 units yields a 0% discount, 500 to 1,499 units yields a 5% discount ("Silver"), 1,500 to 4,999 units yields a 10% discount ("Gold"), and 5,000 units or more yields a 15% discount ("Platinum").

#### Scenario: Below Silver threshold
- **WHEN** the decision table is evaluated with a purchase volume of `499`
- **THEN** it resolves to a `0` discount

#### Scenario: Silver tier lower boundary
- **WHEN** the decision table is evaluated with a purchase volume of `500`
- **THEN** it resolves to a `0.05` discount

#### Scenario: Silver tier upper boundary
- **WHEN** the decision table is evaluated with a purchase volume of `1499`
- **THEN** it resolves to a `0.05` discount

#### Scenario: Gold tier lower boundary
- **WHEN** the decision table is evaluated with a purchase volume of `1500`
- **THEN** it resolves to a `0.10` discount

#### Scenario: Gold tier upper boundary
- **WHEN** the decision table is evaluated with a purchase volume of `4999`
- **THEN** it resolves to a `0.10` discount

#### Scenario: Platinum tier
- **WHEN** the decision table is evaluated with a purchase volume of `5000` or more
- **THEN** it resolves to a `0.15` discount

### Requirement: Purchase Volume CLI Argument
The example program SHALL accept a purchase volume as its sole command-line argument.

#### Scenario: Valid numeric argument
- **WHEN** the program is run with a single non-negative integer argument
- **THEN** it evaluates the decision table using that value as the purchase volume

#### Scenario: Missing or invalid argument
- **WHEN** the program is run with zero arguments, or with an argument that is not a valid non-negative number
- **THEN** it prints a usage message and exits with a non-zero status, without evaluating the decision table

### Requirement: Pricing Breakdown Output
The example program SHALL evaluate the decision table for the given purchase volume and print `Unit-Price`, `Discount`, `Net-Unit-Price`, and `Volume-Discounted-Price`, where `Unit-Price` is the fixed base wholesale price of `$50.00`, `Net-Unit-Price` is `Unit-Price` reduced by the resolved `Discount`, and `Volume-Discounted-Price` is `Net-Unit-Price` multiplied by the purchase volume.

#### Scenario: Silver-tier pricing breakdown
- **WHEN** the program is run with a purchase volume of `1000`
- **THEN** it prints `Unit-Price` of `50.00`, `Discount` of `0.05`, `Net-Unit-Price` of `47.50`, and `Volume-Discounted-Price` of `47500.00`

#### Scenario: No-discount pricing breakdown
- **WHEN** the program is run with a purchase volume of `100`
- **THEN** it prints `Unit-Price` of `50.00`, `Discount` of `0`, `Net-Unit-Price` of `50.00`, and `Volume-Discounted-Price` of `5000.00`
