package main

import (
	"fmt"
	"syscall/js"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/workflows"
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
	repo := args[0].String()

	return js.ValueOf(workflows.FindWorkflows(repo, jwt))
}
