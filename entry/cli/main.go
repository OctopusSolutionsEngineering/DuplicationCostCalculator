package main

import (
	"os"
	"strings"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/workflows"
)

func main() {
	args := os.Args

	if len(args) < 3 {
		println("Usage: app <repo1> <repo2> ... <repoN>")
		return
	}

	client := workflows.GetClientLocal()

	report := workflows.GenerateReport(client, args[1:])

	for sourceRepo, comparison := range report.Comparisons {
		println(sourceRepo, "Contributors:", len(report.Contributors[sourceRepo]))
		for repoName, measurements := range comparison {
			println("  ", repoName)
			println("    Steps that indicate duplication risk:", measurements.StepsThatIndicateDuplicationRisk)
			println("    Steps with different versions:", measurements.StepsWithDifferentVersionsCount, strings.Join(measurements.StepsWithDifferentVersions, ", "))
			println("    Steps with similar config:", measurements.StepsWithSimilarConfigCount, strings.Join(measurements.StepsWithSimilarConfig, ", "))
		}
	}
}
