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
	// Load configuration
	cfg, err := best.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create agent
	agent := best.CreateAgent("ScenarioBot")

	// Connect to server
	fmt.Println("Connecting to server...")
	if err := agent.Connect(); err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer agent.Disconnect()

	// Wait for spawn
	time.Sleep(2 * time.Second)
	fmt.Println("Connected!")

	// Create scenario runner with verbose output
	runner, err := scenario.NewRunner(agent, &cfg.AI,
		scenario.WithVerbose(true),
		scenario.WithStepTimeout(30*time.Second),
	)
	if err != nil {
		fmt.Printf("Failed to create scenario runner: %v\n", err)
		os.Exit(1)
	}
	defer runner.Close()

	// Run scenario from string
	scenarioText := `
シナリオ: 基本接続テスト

1. 接続されていることを確認する
2. /help コマンドを実行する
3. 3秒待機する
4. 体力が0より大きいことを確認する
5. ゲームモードがサバイバルであることを確認する
`

	fmt.Println("\n=== Running Scenario ===")
	fmt.Println(scenarioText)
	fmt.Println("========================\n")

	ctx := context.Background()
	result, err := runner.RunFromString(ctx, scenarioText)
	if err != nil {
		fmt.Printf("Scenario failed: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Println("\n=== Results ===")
	fmt.Printf("Total Steps: %d\n", result.TotalSteps)
	fmt.Printf("Passed: %d\n", result.Passed)
	fmt.Printf("Failed: %d\n", result.Failed)
	fmt.Printf("Duration: %v\n", result.Duration)

	if result.Failed > 0 {
		fmt.Println("\nFailed steps:")
		for i, step := range result.Steps {
			if step.Status == scenario.StepStatusFailed {
				fmt.Printf("  Step %d: %s - %s\n", i+1, step.Step.Description, step.Error)
			}
		}
		os.Exit(1)
	}

	fmt.Println("\nAll steps passed!")
}
