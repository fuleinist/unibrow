package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
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
	fmt.Fprintf(os.Stderr, "DEBUG rawPrompt=%q args=%v\n", rawPrompt, args)

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

	// Handle /all to run all available agents (supports /all foo bar and /a foo bar)
	if isAllCommand(rawPrompt) {
		return runAllAgents(cmd, args)
	}

	// Handle agent prefix routing: /c, /x, /g
	prefixedAgent, prompt := resolveRunPrompt(rawPrompt, args)

	// Initialize registry and router
	registry := agents.InitDefaultRegistry()
	r := router.NewRouter(registry)

	// Determine which agent to use.
	// Order: --agent flag (wins) → prefix → heuristic.
	var routeResult router.RouteResult
	switch {
	case agentName != "":
		ag, ok := registry.Get(agentName)
		if !ok {
			return fmt.Errorf("unknown agent: %s", agentName)
		}
		routeResult = router.RouteResult{AgentName: agentName, Agent: ag, Reason: "explicit flag"}
	case prefixedAgent != "":
		ag, ok := registry.Get(prefixedAgent)
		if !ok {
			return fmt.Errorf("unknown agent: %s", prefixedAgent)
		}
		routeResult = router.RouteResult{AgentName: prefixedAgent, Agent: ag, Reason: "explicit prefix"}
	default:
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

// resolveRunPrompt resolves the agent name and full prompt for runPrompt.
// If rawPrompt starts with a known agent prefix (/c, /x, /g), the prefix is
// stripped and that agent name is returned. The remaining args (after the
// prefix token, if any) are joined with spaces so multi-word prompts
// (e.g. `unibrow run hello world` or `unibrow run /c hello world`) are not
// truncated to the first word.
func resolveRunPrompt(rawPrompt string, args []string) (string, string) {
	prefixedAgent, prompt := parseAgentPrefix(rawPrompt)
	if prefixedAgent != "" {
		// The prefix was a standalone token in args[0]; the rest of the
		// prompt is args[1:].
		if len(args) > 1 {
			prompt = strings.Join(args[1:], " ")
		}
		return prefixedAgent, prompt
	}
	if len(args) > 1 {
		prompt = strings.Join(args, " ")
	}
	return prefixedAgent, prompt
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
	target := strings.TrimSpace(parts[1])

	registry := agents.InitDefaultRegistry()
	agent, ok := registry.Get(target)
	if !ok {
		return fmt.Errorf("unknown agent: %s", target)
	}

	if !agent.IsAvailable() {
		return fmt.Errorf("agent %s is not available", target)
	}

	fmt.Printf("Delegating to %s: %s\n", target, task)
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

// isAllCommand reports whether the raw prompt is a /all or /a command
// (handles both /all alone and /all followed by content).
func isAllCommand(prompt string) bool {
	trimmed := strings.TrimSpace(prompt)
	return trimmed == "/all" || trimmed == "/a" ||
		strings.HasPrefix(trimmed, "/all ") || strings.HasPrefix(trimmed, "/a ")
}

// runAllAgents runs the prompt on all available agents in parallel.
func runAllAgents(cmd *cobra.Command, args []string) error {
	// Reconstruct prompt: strip leading /all or /a token from args
	combined := strings.TrimSpace(strings.Join(args, " "))
	for _, p := range []string{"/all", "/a"} {
		if combined == p {
			combined = ""
			break
		}
		if strings.HasPrefix(combined, p+" ") {
			combined = strings.TrimSpace(combined[len(p):])
			break
		}
	}
	if combined == "" {
		return fmt.Errorf("/all requires a prompt. Use: unibrow run /all [prompt]")
	}
	prompt := combined

	registry := agents.InitDefaultRegistry()
	available := registry.Available()

	if len(available) == 0 {
		return fmt.Errorf("no available agents")
	}

	fmt.Printf("Running on %d available agents in parallel...\n", len(available))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	type result struct {
		name   string
		output string
		err    error
	}
	results := make(chan result, len(available))
	var wg sync.WaitGroup

	for _, ag := range available {
		wg.Add(1)
		go func(a agents.Agent) {
			defer wg.Done()
			out, runErr := a.Run(ctx, prompt, nil)
			results <- result{name: a.Name(), output: out, err: runErr}
		}(ag)
	}

	wg.Wait()
	close(results)

	for res := range results {
		fmt.Printf("\n--- %s ---\n", strings.ToUpper(res.name))
		if res.err != nil {
			fmt.Printf("Error: %v\n", res.err)
			continue
		}
		fmt.Println(res.output)
	}
	return nil
}