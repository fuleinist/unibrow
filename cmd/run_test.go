package cmd

import (
	"testing"
)

func TestParseAgentPrefix(t *testing.T) {
	tests := []struct {
		name         string
		prompt       string
		wantAgent    string
		wantRest     string
	}{
		{"empty prompt", "", "", ""},
		{"no prefix", "hello world", "", "hello world"},
		{"plain /", "/", "", "/"},
		{"claude short", "/c hello", "claude", "hello"},
		{"claude full", "/claude hello", "claude", "hello"},
		{"claude no space", "/c", "claude", ""},
		{"codex short", "/x hello", "codex", "hello"},
		{"codex full", "/codex hello", "codex", "hello"},
		{"gemini short", "/g hello", "gemini", "hello"},
		{"gemini full", "/gemini hello", "gemini", "hello"},
		{"unknown prefix", "/unknown hello", "", "/unknown hello"},
		{"single char no prefix", "c hello", "", "c hello"},
		{"prefix with only space", "/c", "claude", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAgent, gotRest := parseAgentPrefix(tt.prompt)
			if gotAgent != tt.wantAgent {
				t.Errorf("parseAgentPrefix(%q) agent = %q, want %q", tt.prompt, gotAgent, tt.wantAgent)
			}
			if gotRest != tt.wantRest {
				t.Errorf("parseAgentPrefix(%q) rest = %q, want %q", tt.prompt, gotRest, tt.wantRest)
			}
		})
	}
}

func TestIsAllCommand(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   bool
	}{
		{"plain /all", "/all", true},
		{"/all with space", "/all hello world", true},
		{"plain /a", "/a", true},
		{"/a with space", "/a hello world", true},
		{"random /all in text", "something /all else", false},
		{"plain hello", "hello", false},
		{"empty", "", false},
		{"just spaces", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllCommand(tt.prompt)
			if got != tt.want {
				t.Errorf("isAllCommand(%q) = %v, want %v", tt.prompt, got, tt.want)
			}
		})
	}
}
func TestResolveRunPrompt(t *testing.T) {
	tests := []struct {
		name      string
		rawPrompt string
		args      []string
		wantAgent string
		wantRest  string
	}{
		{"single arg no prefix", "hello", []string{"hello"}, "", "hello"},
		{"multi arg joins with spaces", "hello", []string{"hello", "world"}, "", "hello world"},
		{"three arg join", "fix", []string{"fix", "the", "bug"}, "", "fix the bug"},
		{"prefix consumes single arg", "/c hello", []string{"/c", "hello"}, "claude", "hello"},
		{"prefix with multi args joins rest", "/c", []string{"/c", "hello", "world"}, "claude", "hello world"},
		{"unknown prefix falls through to single arg", "/unknown foo", []string{"/unknown", "foo"}, "", "/unknown foo"},
		{"empty args slice", "", []string{}, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAgent, gotRest := resolveRunPrompt(tt.rawPrompt, tt.args)
			if gotAgent != tt.wantAgent {
				t.Errorf("resolveRunPrompt(%q, %v) agent = %q, want %q", tt.rawPrompt, tt.args, gotAgent, tt.wantAgent)
			}
			if gotRest != tt.wantRest {
				t.Errorf("resolveRunPrompt(%q, %v) rest = %q, want %q", tt.rawPrompt, tt.args, gotRest, tt.wantRest)
			}
		})
	}
}
