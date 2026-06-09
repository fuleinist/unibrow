package agents

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GeminiAgent integrates with Gemini CLI.
type GeminiAgent struct{}

// NewGeminiAgent creates a new Gemini CLI agent.
func NewGeminiAgent() *GeminiAgent {
	return &GeminiAgent{}
}

// Name returns the agent name.
func (a *GeminiAgent) Name() string {
	return "gemini"
}

// IsAvailable checks if Gemini CLI is installed.
func (a *GeminiAgent) IsAvailable() bool {
	cmd := exec.Command("gemini", "--version")
	return cmd.Run() == nil
}

// Run executes Gemini CLI with the given prompt.
func (a *GeminiAgent) Run(ctx context.Context, prompt string, memory []MemoryEntry) (string, error) {
	fullPrompt := buildContextPromptGemini(memory) + "\n\n" + prompt

	cmd := exec.CommandContext(ctx, "gemini", "-p", fullPrompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("gemini: %w", err)
	}
	return string(output), nil
}

// InstallInstructions returns installation instructions.
func (a *GeminiAgent) InstallInstructions() string {
	return `To install Gemini CLI:
  npm install -g @google/gemini-cli
  OR visit https://ai.google.dev/gemini-api/docs
`
}

func buildContextPromptGemini(memory []MemoryEntry) string {
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