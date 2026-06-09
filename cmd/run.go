package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/fuleinist/unibrow/internal/agents"
	"github.com/fuleinist/unibrow/internal/memory"
	"github.com/fuleinist/unibrow/internal/router"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [prompt]",
	Short: "Run a prompt with intelligent agent routing",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runPrompt,
}

var (
	sessionID string
	agentName string
)

func init() {
	runCmd.Flags().StringVarP(&sessionID, "session", "s", "default", "Session ID to use")
	runCmd.Flags().StringVarP(&agentName, "agent", "a", "", "Explicit agent to use (claude, codex, gemini)")
}

func runPrompt(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	// Initialize registry and router
	registry := agents.InitDefaultRegistry()
	r := router.NewRouter(registry)

	// Determine which agent to use
	var routeResult router.RouteResult
	if agentName != "" {
		agent, ok := registry.Get(agentName)
		if !ok {
			return fmt.Errorf("unknown agent: %s", agentName)
		}
		routeResult = router.RouteResult{AgentName: agentName, Agent: agent, Reason: "explicit flag"}
	} else {
		routeResult = r.Route(prompt)
	}

	if routeResult.Agent == nil {
		fmt.Println("No available agents. Install one of:")
		for _, a := range registry.List() {
			fmt.Printf("  - %s: %s", a.Name(), a.InstallInstructions())
		}
		return fmt.Errorf("no available agents")
	}

	// Check agent availability
	if !routeResult.Agent.IsAvailable() {
		fmt.Printf("Agent %s is not available.\n%s\n", routeResult.Agent.Name(), routeResult.Agent.InstallInstructions())
		return fmt.Errorf("agent %s not available", routeResult.Agent.Name())
	}

	fmt.Printf("Routing to %s (reason: %s)\n", routeResult.Agent.Name(), routeResult.Reason)

	// Load memory for context
	store, err := memory.NewStore(getMemoryDBPath())
	if err != nil {
		fmt.Printf("Warning: could not open memory store: %v\n", err)
	} else {
		defer store.Close()
	}

	var memEntries []agents.MemoryEntry
	if store != nil {
		entries, _ := store.List(sessionID)
		for _, e := range entries {
			memEntries = append(memEntries, agents.MemoryEntry{
				ID:        e.ID,
				SessionID: e.SessionID,
				Agent:     e.Agent,
				Content:   e.Content,
			})
		}
	}

	// Run the agent
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	output, err := routeResult.Agent.Run(ctx, prompt, memEntries)
	if err != nil {
		return fmt.Errorf("agent run failed: %w", err)
	}

	fmt.Println("\n--- Output ---")
	fmt.Println(output)

	// Store the interaction in memory
	if store != nil {
		store.Add(sessionID, routeResult.Agent.Name(), prompt)
	}

	return nil
}