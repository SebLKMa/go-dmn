// Command volume-discounts evaluates the tiered volume-discount pricing
// clause in contract-clause.md against a purchase volume, using the
// volume-discount.dmn.json decision table and the dmn package.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	dmn "github.com/seblkma/go-dmn"
)

// unitPrice is the fixed base wholesale price per unit, per the contract
// clause's Base Pricing section. It does not vary by volume tier, so it is
// not part of the decision table.
const unitPrice = 50.00

const dmnFileName = "volume-discount.dmn.json"

// sourceDir returns the directory containing this source file, so the DMN
// file can be found regardless of the caller's working directory.
func sourceDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func loadTable() (dmn.DecisionTableDefinition, error) {
	path := filepath.Join(sourceDir(), dmnFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return dmn.DecisionTableDefinition{}, fmt.Errorf("reading %s: %w", path, err)
	}
	var def dmn.DecisionTableDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return dmn.DecisionTableDefinition{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	return def, nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: volume-discounts <purchase-volume>")
}

// pricingBreakdown computes the Unit-Price, Discount, Net-Unit-Price, and
// Volume-Discounted-Price for a given purchase volume by evaluating the
// volume-discount decision table.
func pricingBreakdown(table *dmn.CompiledTable, purchaseVolume float64) (discount, netUnitPrice, volumeDiscountedPrice float64, err error) {
	result, err := table.Evaluate(dmn.EvaluationContext{"purchaseVolume": purchaseVolume})
	if err != nil {
		return 0, 0, 0, err
	}
	values, ok := result.(map[string]any)
	if !ok {
		return 0, 0, 0, fmt.Errorf("unexpected result type %T", result)
	}
	discount, ok = values["discount"].(float64)
	if !ok {
		return 0, 0, 0, fmt.Errorf("unexpected discount type %T", values["discount"])
	}
	netUnitPrice = unitPrice * (1 - discount)
	volumeDiscountedPrice = netUnitPrice * purchaseVolume
	return discount, netUnitPrice, volumeDiscountedPrice, nil
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	purchaseVolume, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		usage()
		os.Exit(1)
	}

	def, err := loadTable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	table, err := dmn.Compile(def)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	discount, netUnitPrice, volumeDiscountedPrice, err := pricingBreakdown(table, purchaseVolume)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Printf("Unit-Price: %.2f\n", unitPrice)
	fmt.Printf("Discount: %v\n", discount)
	fmt.Printf("Net-Unit-Price: %.2f\n", netUnitPrice)
	fmt.Printf("Volume-Discounted-Price: %.2f\n", volumeDiscountedPrice)
}
