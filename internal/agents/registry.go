package agents

import (
	"context"
	"fmt"
)

// Agent represents an AI coding agent integration.
type Agent interface {
	// Name returns the agent's display name.
	Name() string
	// IsAvailable checks if the agent binary/SDK is installed and accessible.
	IsAvailable() bool
	// Run executes the agent with the given prompt and context.
	Run(ctx context.Context, prompt string, memory []MemoryEntry) (string, error)
	// InstallInstructions returns how to install this agent.
	InstallInstructions() string
}

// MemoryEntry represents a shared memory item.
type MemoryEntry struct {
	ID         int64
	SessionID  string
	Agent      string
	Content    string
	UsageCount int
}

// Registry manages available agents.
type Registry struct {
	agents map[string]Agent
}

// NewRegistry creates a new agent registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]Agent),
	}
}

// Register adds an agent to the registry.
func (r *Registry) Register(agent Agent) {
	r.agents[agent.Name()] = agent
}

// Get retrieves an agent by name.
func (r *Registry) Get(name string) (Agent, bool) {
	agent, ok := r.agents[name]
	return agent, ok
}

// List returns all registered agents.
func (r *Registry) List() []Agent {
	agents := make([]Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}

// Available returns all available (installed) agents.
func (r *Registry) Available() []Agent {
	var available []Agent
	for _, agent := range r.agents {
		if agent.IsAvailable() {
			available = append(available, agent)
		}
	}
	return available
}

// InitDefaultRegistry creates a registry with all known agents.
func InitDefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(NewClaudeAgent())
	r.Register(NewCodexAgent())
	r.Register(NewGeminiAgent())
	return r
}

// StatusIndicator returns a status indicator for an agent.
func StatusIndicator(agent Agent) string {
	if !agent.IsAvailable() {
		return "○"
	}
	return "●"
}

// StatusColor returns a color code for agent status.
func StatusColor(available bool) string {
	if !available {
		return "\033[90m" // muted gray
	}
	return "\033[32m" // green
}

// FormatAgentStatus formats an agent's status for display.
func FormatAgentStatus(agent Agent) string {
	status := "○"
	color := "\033[90m"
	if agent.IsAvailable() {
		status = "●"
		color = "\033[32m"
	}
	return fmt.Sprintf("%s[%s] %s\033[0m", color, status, agent.Name())
}