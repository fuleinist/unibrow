package router

import (
	"testing"

	"github.com/fuleinist/unibrow/internal/agents"
)

func TestRouter_Route_ExplicitPrefix(t *testing.T) {
	registry := agents.InitDefaultRegistry()
	r := NewRouter(registry)

	tests := []struct {
		name         string
		prompt       string
		expectedName string
	}{
		{"claude prefix /c", "/c fix this bug", "claude"},
		{"claude prefix /claude", "/claude write tests", "claude"},
		{"codex prefix /x", "/x complete this", "codex"},
		{"codex prefix /codex", "/codex fix typo", "codex"},
		{"gemini prefix /g", "/g explain this", "gemini"},
		{"gemini prefix /gemini", "/gemini document", "gemini"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.Route(tt.prompt)
			if result.AgentName != tt.expectedName {
				t.Errorf("Route(%q) = %q, want %q", tt.prompt, result.AgentName, tt.expectedName)
			}
		})
	}
}

func TestRouter_Route_Heuristics(t *testing.T) {
	registry := agents.InitDefaultRegistry()
	r := NewRouter(registry)

	tests := []struct {
		name         string
		prompt       string
		expectedName string
	}{
		{"refactor", "refactor this function", "claude"},
		{"implement", "implement a new feature", "claude"},
		{"create", "create a new API endpoint", "claude"},
		{"build", "build the authentication system", "claude"},
		{"write code", "write code for me", "claude"},
		{"complete", "complete this line", "codex"},
		{"fix bug", "fix the login bug", "codex"},
		{"typo", "fix the typo in README", "codex"},
		{"quick fix", "quick fix for crash", "codex"},
		{"simple change", "make a simple change", "codex"},
		{"explain", "explain how this works", "gemini"},
		{"document", "document this module", "gemini"},
		{"what does", "what does this function do", "gemini"},
		{"how does", "how does this work", "gemini"},
		{"tutorial", "tutorial on Go", "gemini"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.Route(tt.prompt)
			if result.AgentName != tt.expectedName {
				t.Errorf("Route(%q) = %q, want %q (reason: %s)", tt.prompt, result.AgentName, tt.expectedName, result.Reason)
			}
		})
	}
}

func TestRouter_Route_DefaultFallback(t *testing.T) {
	registry := agents.InitDefaultRegistry()
	r := NewRouter(registry)

	// Random prompt should default to claude
	result := r.Route("do something random")
	if result.AgentName != "claude" {
		t.Errorf("Route(random) = %q, want claude", result.AgentName)
	}
}

func TestParseAgentPrefix(t *testing.T) {
	tests := []struct {
		name          string
		prompt        string
		wantAgent     string
		wantRest      string
		wantHasPrefix bool
	}{
		{"/c prompt", "/c hello", "claude", "hello", true},
		{"/claude prompt", "/claude hello", "claude", "hello", true},
		{"/x prompt", "/x hello", "codex", "hello", true},
		{"/codex prompt", "/codex hello", "codex", "hello", true},
		{"/g prompt", "/g hello", "gemini", "hello", true},
		{"/gemini prompt", "/gemini hello", "gemini", "hello", true},
		{"no prefix", "hello", "", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, rest, hasPrefix := ParseAgentPrefix(tt.prompt)
			if agent != tt.wantAgent {
				t.Errorf("ParseAgentPrefix(%q) agent = %q, want %q", tt.prompt, agent, tt.wantAgent)
			}
			if rest != tt.wantRest {
				t.Errorf("ParseAgentPrefix(%q) rest = %q, want %q", tt.prompt, rest, tt.wantRest)
			}
			if hasPrefix != tt.wantHasPrefix {
				t.Errorf("ParseAgentPrefix(%q) hasPrefix = %v, want %v", tt.prompt, hasPrefix, tt.wantHasPrefix)
			}
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		s         string
		substrs   []string
		expected  bool
	}{
		{"hello world", []string{"hello", "hi"}, true},
		{"hello world", []string{"hi", "hey"}, false},
		{"refactor this", []string{"refactor", "rewrite"}, true},
		{"simple change", []string{"simple", "change"}, true},
	}

	for _, tt := range tests {
		result := containsAny(tt.s, tt.substrs)
		if result != tt.expected {
			t.Errorf("containsAny(%q, %v) = %v, want %v", tt.s, tt.substrs, result, tt.expected)
		}
	}
}