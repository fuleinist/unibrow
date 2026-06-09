# Unibrow — Unified AI Agent CLI Hub

##1. Concept& Vision

Unibrow is a CLI hub that orchestrates multiple AI coding agents (Claude Code, Codex, Gemini CLI) from a single terminal. It presents one unified context buffer, shared memory across agents, and intelligent task-based routing — so developers stop asking "which agent do I use?" and start getting things done.

The experience feels like a command center: one terminal, all agents, shared brain.

## 2. Design Language

- **Aesthetic**: Terminal-native with monospace elegance. Think `htop` meets a modern IDE's activity bar.
- **Colors**: Dark background (`#0d1117`), cyan primary (`#58a6ff`), magenta accent (`#f778ba`), white text (`#e6edf3`), muted gray (`#8b949e`)
- **Typography**: Monospace throughout (JetBrains Mono or system monospace)
- **Spatial system**: Dense but scannable — status bar at top, agent output in middle, command input at bottom
- **Motion**: Minimal — output streams in real-time, status indicators pulse when agents are active
- **Visual assets**: ASCII-art logo, agent icons as unicode emoji (🤖 Claude, 🧠 Codex, 🌟 Gemini)

## 3. Layout & Structure

```
┌─────────────────────────────────────────────────────────┐
│ [🤖][🧠][🌟]  unibrow v0.1.0          session: proj-x  │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  [agent output / shared context / history]              │
│                                                         │
├─────────────────────────────────────────────────────────┤
│ unibrow > _                                            │
└─────────────────────────────────────────────────────────┘
```

- **Header bar**: Active agents (toggleable), version, session name
- **Main area**: Scrollable output from all agents, unified context view
- **Input area**: Unified prompt input, `/agent claude` to switch, `/context` to view shared memory
- **Sidebar** (future): Agent status, routing suggestions

## 4. Features & Interactions

### Core Features

1. **Unified Prompt Input**
   - Single CLI entry point for all agents
   - Agent selection via prefix: `/c` → Claude Code, `/x` → Codex, `/g` → Gemini CLI
   - Without prefix: intelligent routing based on task type

2. **Shared Memory Layer**
   - SQLite-backed context store
   - All agents contribute to and read from shared context
   - `/memory add <text>`, `/memory list`, `/memory clear`
   - Context automatically prepended to each agent invocation

3. **Agent Orchestration**
   - Run multiple agents sequentially or in parallel
   - Delegate sub-tasks: `/delegate write-tests to claude`
   - Parallel mode: `/run all` — fires to all agents, collects results

4. **Session Management**
   - Named sessions: `/session new <name>` or auto-generated
   - Session history: `/history`, `/export`
   - Resume session: `unibrow --session <name>`

5. **Context Buffer**
   - Unified view of current project context
   - Auto-detects project type, relevant files
   - `/context show` — view current context
   - `/context add <file>` — manually add files to context

### Interaction Details

- **Agent switching**: `/c`, `/x`, `/g` prefix or tab-complete agent names
- **Memory commands**: `/memory` as top-level command
- **Session commands**: `/session` as top-level command
- **Exit**: `Ctrl+C` or `/exit`
- **Help**: `/help` or `unibrow --help`

### Edge Cases

- Agent not installed → show install instructions + skip gracefully
- Agent API key missing → warn but allow other agents to run
- Shared context too large → auto-summarize oldest entries
- Agent times out → show warning, allow retry or skip

## 5. Component Inventory

### `unibrow` CLI (root command)
- **States**: idle, agent-running, error
- Displays version, help, session info

### `unibrow run <prompt>` — run a prompt
- Intelligent routing or explicit agent selection
- Streams output, shows which agent handled it

### `unibrow memory` — memory subcommands
- `add`, `list`, `clear`, `show`
- SQLite persistence

### `unibrow session` — session subcommands
- `new`, `list`, `resume`, `export`

### `unibrow context` — context subcommands
- `show`, `add`, `remove`, `clear`

### Status Indicators
- `●` green = agent ready
- `◐` yellow = agent busy
- `○` gray = agent not available
- `✗` red = agent error

## 6. Technical Approach

### Stack
- **Language**: Go 1.21+
- **CLI Framework**: Cobra + Viper (standard Go CLI)
- **Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **Agent Integration**: CLI subprocess spawning + IPC

### Architecture

```
unibrow/
├── cmd/
│   ├── root.go          # root command, global flags
│   ├── run.go           # run prompt command
│   ├── memory.go        # memory subcommands
│   ├── session.go      # session subcommands
│   └── context.go       # context subcommands
├── internal/
│   ├── agents/
│   │   ├── registry.go  # agent discovery + availability
│   │   ├── claude.go    # Claude Code integration
│   │   ├── codex.go     # Codex integration
│   │   └── gemini.go    # Gemini CLI integration
│   ├── memory/
│   │   └── store.go     # SQLite memory store
│   ├── session/
│   │   └── manager.go   # session management
│   ├── router/
│   │   └── router.go    # intelligent task routing
│   └── ui/
│       └── tui.go       # terminal UI helpers
├── unibrow.go           # main.go entry point
└── go.mod
```

### Agent Integration Pattern
Each agent is a Go interface:
```go
type Agent interface {
    Name() string
    IsAvailable() bool
    Run(ctx context.Context, prompt string, memory []MemoryEntry) (string, error)
    InstallInstructions() string
}
```

### Routing Heuristics (v1)
- Code generation, refactoring → Claude Code
- Fast completions, simple edits → Codex
- Documentation, explanations → Gemini CLI
- Unknown → default to Claude Code

### Data Model

**MemoryEntry** (SQLite):
```
id, session_id, agent, content, created_at, usage_count
```

**Session** (SQLite):
```
id, name, created_at, last_active, context_summary
```

## 7. Acceptance Criteria

- [ ] `unibrow --help` shows all commands and flags
- [ ] `unibrow run "hello world"` executes on default agent (Claude Code)
- [ ] `unibrow run "/c hello"` routes to Claude Code explicitly
- [ ] `unibrow memory add "test"` persists and `unibrow memory list` shows it
- [ ] `unibrow session new test-session` creates a named session
- [ ] Shared memory is prepended to agent prompts
- [ ] Graceful handling when agent binary is not found
- [ ] Version flag `unibrow --version` works
- [ ] Code compiles with `go build ./...`
- [ ] Unit tests pass for core logic (router, memory store)
