## 1. Fixtures

- [x] 1.1 Create `examples/volume-discounts-xml/contract-clause.md` as a copy of `examples/volume-discounts/contract-clause.md`, and verify the two files are identical
- [x] 1.2 Create `examples/volume-discounts-xml/volume-discount.dmn.xml`, an OMG DMN XML document encoding the same table as `examples/volume-discounts/volume-discount.dmn.json` (decision name `VolumeDiscount`, hit policy `UNIQUE`, one number input `purchaseVolume`, outputs `discount` (number) and `tier` (string), and the four tier rules), and verify `dmn.DecodeXML` decodes it without error in a throwaway test/script

## 2. CLI program

- [x] 2.1 Implement `examples/volume-discounts-xml/main.go`'s `loadTable()` to read `volume-discount.dmn.xml` (located the same way as the JSON example, via `runtime.Caller` relative to the source file) and decode it with `dmn.DecodeXML`, and verify `go build ./examples/volume-discounts-xml/...` succeeds
- [x] 2.2 Implement `pricingBreakdown` and `main()` mirroring `examples/volume-discounts/main.go`'s CLI argument handling (single purchase-volume argument, usage message and non-zero exit on a missing/invalid argument) and output format (`Unit-Price`, `Discount`, `Net-Unit-Price`, `Volume-Discounted-Price`), and verify `go run ./examples/volume-discounts-xml 1000` prints `Unit-Price: 50.00`, `Discount: 0.05`, `Net-Unit-Price: 47.50`, `Volume-Discounted-Price: 47500.00`

## 3. Tests

- [x] 3.1 Add `examples/volume-discounts-xml/main_test.go` with a `TestLoadTable_DecodesAndCompiles` test and a `TestVolumeDiscountTable_TierBoundaries` table-driven test covering all six tier-boundary cases (499, 500, 1499, 1500, 4999, 5000) from the spec, mirroring `examples/volume-discounts/main_test.go`, and verify `go test ./examples/volume-discounts-xml/...` passes

## 4. Verification

- [x] 4.1 Run `make build` and confirm it produces `bin/volume-discounts-xml`; note that the Makefile's fixture-copy loop only copies `*.json` today, so `volume-discount.dmn.xml` will not be copied into `bin/` — the binary still works because it loads the fixture via its own source-tree path (`runtime.Caller`), the same as the existing JSON example, so this is not a functional gap and the Makefile is left unchanged
- [x] 4.2 Run `go test ./...` and `go vet ./...` from the repo root and confirm all tests pass with no new vet warnings
