package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gollilla/best"
	"github.com/gollilla/best/pkg/scenario"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <scenario-file> [scenario-file...]")
		os.Exit(1)
	}

	cfg, _ := best.LoadConfig()
	agent := best.CreateAgent("ScenarioBot")
	defer agent.Disconnect()

	runner, err := scenario.NewRunner(agent, &cfg.AI,
		scenario.WithVerbose(true),
		scenario.WithStepTimeout(30*time.Second),
		scenario.WithWebhook(&cfg.Webhook),
	)
	if err != nil {
		fmt.Printf("Failed: %v\n", err)
		os.Exit(1)
	}
	defer runner.Close()

	ctx := context.Background()
	summary, _ := runner.RunMultipleFromFiles(ctx, os.Args[1:])

	// LLM によるサマリー生成
	fmt.Println("\n=== Test Summary ===")
	text, err := runner.GenerateSummary(ctx, summary)
	if err != nil {
		fmt.Printf("Summary generation failed: %v\n", err)
	} else {
		fmt.Println(text)
	}

	if !summary.Success() {
		os.Exit(1)
	}
}
