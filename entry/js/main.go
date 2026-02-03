package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/workflows"
	"github.com/samber/lo"
)

func main() {
	// Expose a Go function to JavaScript
	js.Global().Set("processPerson", js.FuncOf(findWorkflows))

	// Keep the Go program alive
	<-make(chan struct{})
}

func findWorkflows(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		fmt.Println("Error: Missing argument")
		return nil
	}

	jwt := args[0].String()

	repos := lo.Map(args[1:], func(item js.Value, index int) string {
		return item.String()
	})

	client := workflows.GetClient(jwt)

	report := workflows.GenerateReport(client, repos)

	reportJson, err := json.Marshal(report)

	if err != nil {
		fmt.Println("Error marshalling report to JSON:", err)
		return nil
	}

	return reportJson
}
