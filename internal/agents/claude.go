package agents

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ClaudeAgent integrates with Claude Code.
type ClaudeAgent struct{}

// NewClaudeAgent creates a new Claude Code agent.
func NewClaudeAgent() *ClaudeAgent {
	return &ClaudeAgent{}
}

// Name returns the agent name.
func (a *ClaudeAgent) Name() string {
	return "claude"
}

// IsAvailable checks if Claude Code is installed.
func (a *ClaudeAgent) IsAvailable() bool {
	cmd := exec.Command("claude", "--version")
	return cmd.Run() == nil
}

// Run executes Claude Code with the given prompt.
func (a *ClaudeAgent) Run(ctx context.Context, prompt string, memory []MemoryEntry) (string, error) {
	// Prepend memory context to prompt
	fullPrompt := buildContextPrompt(memory) + "\n\nTask: " + prompt

	cmd := exec.CommandContext(ctx, "claude", "-p", fullPrompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude code: %w", err)
	}
	return string(output), nil
}

// InstallInstructions returns installation instructions.
func (a *ClaudeAgent) InstallInstructions() string {
	return `To install Claude Code:
  npm install -g @anthropic-ai/claude-code
  OR visit https://docs.anthropic.com/en/docs/claude-code
`
}

func buildContextPrompt(memory []MemoryEntry) string {
	if len(memory) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Shared Context\n")
	for _, entry := range memory {
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", entry.Agent, entry.Content))
	}
	sb.WriteString("\n")
	return sb.String()
}