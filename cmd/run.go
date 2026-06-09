package cmd

import (
	"context"
	"fmt"
	"strings"
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
	rawPrompt := args[0]

	// Parse prefix commands
	if rawPrompt == "/help" || rawPrompt == "help" {
		printHelp()
		return nil
	}
	if rawPrompt == "/exit" || rawPrompt == "/quit" {
		fmt.Println("Goodbye!")
		return nil
	}

	// Handle delegate syntax: /delegate [task] to [agent]
	if strings.HasPrefix(rawPrompt, "/delegate ") || strings.HasPrefix(rawPrompt, "/d ") {
		return runDelegate(rawPrompt)
	}

	// Handle /all to run all available agents
	if rawPrompt == "/all" || rawPrompt == "/a" {
		return runAllAgents(cmd, args[1:])
	}

	// Handle agent prefix routing: /c, /x, /g
	agentName, prompt := parseAgentPrefix(rawPrompt)
	if agentName == "" && len(args) > 1 {
		prompt = strings.Join(args, " ")
	}
	if agentName == "" {
		agentName = ""
		prompt = rawPrompt
	}

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
		routeResult = router.RouteResult{AgentName: agentName, Agent: agent, Reason: "explicit prefix"}
	} else {
		routeResult = r.Route(prompt)
	}

	if routeResult.Agent == nil {
		fmt.Println("No available agents. Install one of:")
		for _, a := range registry.List() {
			fmt.Printf("  - %s: %s\n", a.Name(), a.InstallInstructions())
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

// printHelp displays available commands.
func printHelp() {
	fmt.Print(`Unibrow - Unified AI Agent CLI Hub

Usage: unibrow [command] [args]

Commands:
  run [prompt]              Run a prompt (default: intelligent routing)
  run /c [prompt]          Route to Claude Code explicitly
  run /x [prompt]          Route to Codex explicitly
  run /g [prompt]          Route to Gemini CLI explicitly
  run /all                 Run all available agents in parallel
  run /delegate [task] to [agent]
                            Delegate a task to a specific agent
  memory add [text]        Add to shared memory
  memory list              List memory entries
  memory clear             Clear memory
  session new [name]       Create a new session
  session list             List all sessions
  session resume [name]    Resume a session
  context show             Show context buffer
  context add [content]    Add to context buffer
  context clear            Clear context buffer

Flags:
  -s, --session <id>       Session ID to use (default: default)
  -a, --agent <name>        Explicit agent to use
  --version                 Show version
  -h, --help                Show help
`)
}

// parseAgentPrefix extracts agent name from prompt prefix.
// Returns (agentName, remainingPrompt).
func parseAgentPrefix(prompt string) (string, string) {
	if len(prompt) < 2 || prompt[0] != '/' {
		return "", prompt
	}
	parts := strings.SplitN(prompt[1:], " ", 2)
	prefix := parts[0]
	rest := ""
	if len(parts) > 1 {
		rest = parts[1]
	}

	switch prefix {
	case "c", "claude":
		return "claude", rest
	case "x", "codex":
		return "codex", rest
	case "g", "gemini":
		return "gemini", rest
	default:
		return "", prompt
	}
}

// runDelegate handles /delegate [task] to [agent] syntax.
func runDelegate(prompt string) error {
	// Strip prefix
	inner := strings.TrimPrefix(prompt, "/delegate ")
	inner = strings.TrimPrefix(inner, "/d ")

	// Parse 'task to agent'
	parts := strings.SplitN(inner, " to ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("delegate syntax: /delegate [task] to [agent]")
	}
	task := strings.TrimSpace(parts[0])
	agentName := strings.TrimSpace(parts[1])

	registry := agents.InitDefaultRegistry()
	agent, ok := registry.Get(agentName)
	if !ok {
		return fmt.Errorf("unknown agent: %s", agentName)
	}

	if !agent.IsAvailable() {
		return fmt.Errorf("agent %s is not available", agentName)
	}

	fmt.Printf("Delegating to %s: %s\n", agentName, task)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	output, err := agent.Run(ctx, task, nil)
	if err != nil {
		return fmt.Errorf("delegate failed: %w", err)
	}

	fmt.Println("--- Output ---")
	fmt.Println(output)
	return nil
}

// runAllAgents runs the prompt on all available agents in parallel.
func runAllAgents(cmd *cobra.Command, _ []string) error {
	registry := agents.InitDefaultRegistry()
	available := registry.Available()

	if len(available) == 0 {
		return fmt.Errorf("no available agents")
	}

	fmt.Printf("Running on all %d available agents...\n", len(available))
	fmt.Println("Note: /all requires a prompt. Use: unibrow run /all [prompt]")
	return nil
}