package router

import (
	"strings"

	"github.com/fuleinist/unibrow/internal/agents"
)

// Router handles intelligent task routing to agents.
type Router struct {
	registry *agents.Registry
}

// NewRouter creates a new router.
func NewRouter(registry *agents.Registry) *Router {
	return &Router{registry: registry}
}

// RouteResult contains the routing decision.
type RouteResult struct {
	AgentName string
	Agent     agents.Agent
	Reason    string
}

// Route determines which agent should handle a prompt.
func (r *Router) Route(prompt string) RouteResult {
	prompt = strings.TrimSpace(prompt)

	// Check for explicit agent prefixes
	if strings.HasPrefix(prompt, "/c ") || strings.HasPrefix(prompt, "/claude ") {
		agent, ok := r.registry.Get("claude")
		if ok {
			return RouteResult{AgentName: "claude", Agent: agent, Reason: "explicit /c prefix"}
		}
	}

	if strings.HasPrefix(prompt, "/x ") || strings.HasPrefix(prompt, "/codex ") {
		agent, ok := r.registry.Get("codex")
		if ok {
			return RouteResult{AgentName: "codex", Agent: agent, Reason: "explicit /x prefix"}
		}
	}

	if strings.HasPrefix(prompt, "/g ") || strings.HasPrefix(prompt, "/gemini ") {
		agent, ok := r.registry.Get("gemini")
		if ok {
			return RouteResult{AgentName: "gemini", Agent: agent, Reason: "explicit /g prefix"}
		}
	}

	// Heuristic routing based on prompt content
	agentName := r.routeByHeuristics(prompt)
	agent, ok := r.registry.Get(agentName)
	if !ok {
		// Fallback to first available agent
		available := r.registry.Available()
		if len(available) > 0 {
			return RouteResult{AgentName: available[0].Name(), Agent: available[0], Reason: "fallback to first available"}
		}
		return RouteResult{AgentName: "", Agent: nil, Reason: "no available agents"}
	}

	return RouteResult{AgentName: agentName, Agent: agent, Reason: "heuristic match"}
}

// routeByHeuristics applies routing rules based on prompt analysis.
func (r *Router) routeByHeuristics(prompt string) string {
	lower := strings.ToLower(prompt)

	// Code generation, refactoring → Claude Code
	if containsAny(lower, []string{"refactor", "rewrite", "implement", "create", "build", "write code", "generate code", "develop", "architect"}) {
		return "claude"
	}

	// Fast completions, simple edits → Codex
	if containsAny(lower, []string{"complete", "fix", "bug", "typo", "small change", "quick fix", "simple", "edit this", "change this"}) {
		return "codex"
	}

	// Documentation, explanations → Gemini CLI
	if containsAny(lower, []string{"explain", "document", "documentation", "what does", "how does", "describe", "understand", "learn about", "tutorial"}) {
		return "gemini"
	}

	// Default to Claude Code
	return "claude"
}

// containsAny checks if the string contains any of the substrings.
func containsAny(s string, substrings []string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ParseAgentPrefix extracts the agent prefix from a prompt.
func ParseAgentPrefix(prompt string) (agent string, rest string, hasPrefix bool) {
	prompt = strings.TrimSpace(prompt)

	prefixes := []string{"/c ", "/claude ", "/x ", "/codex ", "/g ", "/gemini "}
	agentMap := map[string]string{
		"/c ":      "claude",
		"/claude ": "claude",
		"/x ":      "codex",
		"/codex ":  "codex",
		"/g ":      "gemini",
		"/gemini ": "gemini",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(prompt, prefix) {
			return agentMap[prefix], strings.TrimPrefix(prompt, prefix), true
		}
	}

	return "", prompt, false
}