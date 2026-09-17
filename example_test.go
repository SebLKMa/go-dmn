package dmn_test

import (
	"fmt"

	dmn "github.com/seblkma/go-dmn"
)

func ExampleEvaluate() {
	def := dmn.DecisionTableDefinition{
		TableName: "Discount",
		HitPolicy: dmn.HitPolicyFirst,
		Inputs: []dmn.ClauseDefinition{
			{ID: "i1", Name: "customerType", Type: dmn.ClauseTypeString},
		},
		Outputs: []dmn.ClauseDefinition{
			{ID: "o1", Name: "discount", Type: dmn.ClauseTypeNumber},
		},
		Rules: []dmn.RuleDefinition{
			{RuleID: "r1", InputEntries: []string{`"VIP"`}, OutputEntries: []string{"20"}},
			{RuleID: "r2", InputEntries: []string{"-"}, OutputEntries: []string{"0"}},
		},
	}

	result, err := dmn.Evaluate(def, dmn.EvaluationContext{"customerType": "VIP"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output: map[discount:20]
}

func ExampleCompile() {
	def := dmn.DecisionTableDefinition{
		TableName: "AgeGroup",
		HitPolicy: dmn.HitPolicyFirst,
		Inputs: []dmn.ClauseDefinition{
			{ID: "i1", Name: "age", Type: dmn.ClauseTypeNumber},
		},
		Outputs: []dmn.ClauseDefinition{
			{ID: "o1", Name: "group", Type: dmn.ClauseTypeString},
		},
		Rules: []dmn.RuleDefinition{
			{RuleID: "r1", InputEntries: []string{"[18..65]"}, OutputEntries: []string{`"adult"`}},
			{RuleID: "r2", InputEntries: []string{"-"}, OutputEntries: []string{`"other"`}},
		},
	}

	table, err := dmn.Compile(def)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	// Compile once, evaluate many times against different contexts.
	for _, age := range []float64{10, 30} {
		result, err := table.Evaluate(dmn.EvaluationContext{"age": age})
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println(result)
	}
	// Output:
	// map[group:other]
	// map[group:adult]
}
