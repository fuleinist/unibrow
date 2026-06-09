package agents

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CodexAgent integrates with OpenAI Codex (CLI).
type CodexAgent struct{}

// NewCodexAgent creates a new Codex agent.
func NewCodexAgent() *CodexAgent {
	return &CodexAgent{}
}

// Name returns the agent name.
func (a *CodexAgent) Name() string {
	return "codex"
}

// IsAvailable checks if Codex CLI is installed.
func (a *CodexAgent) IsAvailable() bool {
	// Try "codex" or "openai codex" command
	cmd := exec.Command("codex", "--version")
	if cmd.Run() == nil {
		return true
	}
	cmd = exec.Command("openai", "codex", "--version")
	return cmd.Run() == nil
}

// Run executes Codex with the given prompt.
func (a *CodexAgent) Run(ctx context.Context, prompt string, memory []MemoryEntry) (string, error) {
	fullPrompt := buildContextPromptCodex(memory) + "\n\n" + prompt

	cmd := exec.CommandContext(ctx, "codex", "-p", fullPrompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("codex: %w", err)
	}
	return string(output), nil
}

// InstallInstructions returns installation instructions.
func (a *CodexAgent) InstallInstructions() string {
	return `To install Codex:
  openai install
  OR visit https://openai.com/index/introducing-codex
`
}

func buildContextPromptCodex(memory []MemoryEntry) string {
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