package main

import (
	"testing"

	dmn "github.com/seblkma/go-dmn"
)

func TestLoadTable_DecodesAndCompiles(t *testing.T) {
	def, err := loadTable()
	if err != nil {
		t.Fatalf("loadTable() error = %v", err)
	}
	if _, err := dmn.Compile(def); err != nil {
		t.Fatalf("dmn.Compile() error = %v", err)
	}
}

func TestVolumeDiscountTable_TierBoundaries(t *testing.T) {
	def, err := loadTable()
	if err != nil {
		t.Fatalf("loadTable() error = %v", err)
	}
	table, err := dmn.Compile(def)
	if err != nil {
		t.Fatalf("dmn.Compile() error = %v", err)
	}

	tests := []struct {
		volume       float64
		wantDiscount float64
		wantTier     string
	}{
		{499, 0, "Standard"},
		{500, 0.05, "Silver"},
		{1499, 0.05, "Silver"},
		{1500, 0.10, "Gold"},
		{4999, 0.10, "Gold"},
		{5000, 0.15, "Platinum"},
	}
	for _, tt := range tests {
		discount, _, _, err := pricingBreakdown(table, tt.volume)
		if err != nil {
			t.Fatalf("pricingBreakdown(%v) error = %v", tt.volume, err)
		}
		if discount != tt.wantDiscount {
			t.Errorf("pricingBreakdown(%v) discount = %v, want %v", tt.volume, discount, tt.wantDiscount)
		}
	}
}
