## 1. Decision Table

- [x] 1.1 Create `examples/volume-discounts/volume-discount.dmn.json` with input clause `purchaseVolume` (number), output clauses `discount` (number) and `tier` (string), hit policy `U`, and four rules for the `[0..499]`, `[500..1499]`, `[1500..4999]`, and `>= 5000` ranges per the Discount Tier Decision Table requirement, and verify it decodes via `dmn.Compile` in a Go test
- [x] 1.2 Verify the table's tier boundaries with a table-driven Go test evaluating `dmn.Evaluate` directly against the file's decoded definition for `499`, `500`, `1499`, `1500`, `4999`, and `5000`, asserting the `discount` output matches the Discount Tier Decision Table requirement's scenarios

## 2. CLI Program

- [x] 2.1 Create `examples/volume-discounts/main.go` that reads the purchase volume from `os.Args[1]` via `strconv.ParseFloat`, and verify `go run . 1000` runs without error
- [x] 2.2 Add usage/error handling for a missing or non-numeric argument (print usage to stderr, `os.Exit(1)`) per the Purchase Volume CLI Argument requirement, and verify with `go run .` (no args) and `go run . abc` both exiting non-zero without evaluating the table
- [x] 2.3 Load and compile the decision table file at program startup (path relative to the executable's source directory) and evaluate it with the parsed purchase volume, and verify a `CompletenessError` (e.g. from a negative volume) is reported as a clean error message rather than a panic

## 3. Pricing Output

- [x] 3.1 Define the `Unit-Price` constant ($50.00) in `main.go` and compute `Net-Unit-Price` (`Unit-Price * (1 - discount)`) and `Volume-Discounted-Price` (`Net-Unit-Price * purchaseVolume`) from the decision table's resolved `discount` output, per the Pricing Breakdown Output requirement
- [x] 3.2 Print `Unit-Price`, `Discount`, `Net-Unit-Price`, and `Volume-Discounted-Price` in a clear labeled format, and verify with `go run . 1000` that the output matches the Silver-tier pricing breakdown scenario (`50.00`, `0.05`, `47.50`, `47500.00`)
- [x] 3.3 Verify the no-discount pricing breakdown scenario with `go run . 100`, asserting output `50.00`, `0`, `50.00`, `5000.00`

## 4. Verification

- [x] 4.1 Run `go build ./...` and `go vet ./...` from the module root and verify both succeed with the new example package included
- [x] 4.2 Run `go test ./...` and verify all tests (existing `dmn` package tests plus the new example tests) pass
