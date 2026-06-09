package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/fuleinist/unibrow/internal/session"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage unibrow sessions",
}

func init() {
	sessionCmd.AddCommand(sessionNewCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionResumeCmd)
}

var sessionNewCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new session",
	Args:  cobra.MaximumNArgs(1),
	RunE:  sessionNew,
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	RunE:  sessionList,
}

var sessionResumeCmd = &cobra.Command{
	Use:   "resume [name|id]",
	Short: "Resume a session",
	Args:  cobra.ExactArgs(1),
	RunE:  sessionResume,
}

func sessionNew(cmd *cobra.Command, args []string) error {
	name := "default"
	if len(args) > 0 {
		name = args[0]
	}

	mgr, err := session.NewManager(getSessionDBPath())
	if err != nil {
		return fmt.Errorf("open session manager: %w", err)
	}
	defer mgr.Close()

	// Generate ID from name
	id := fmt.Sprintf("session-%d", time.Now().UnixNano())

	s, err := mgr.Create(id, name)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	fmt.Printf("Created session: %s (ID: %s)\n", s.Name, s.ID)
	return nil
}

func sessionList(cmd *cobra.Command, args []string) error {
	mgr, err := session.NewManager(getSessionDBPath())
	if err != nil {
		return fmt.Errorf("open session manager: %w", err)
	}
	defer mgr.Close()

	sessions, err := mgr.List()
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	fmt.Println("Sessions:")
	for _, s := range sessions {
		age := time.Since(s.LastActive)
		fmt.Printf("  %s | %s | last active: %s ago\n", s.Name, s.ID, age.Round(time.Second))
	}
	return nil
}

func sessionResume(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	mgr, err := session.NewManager(getSessionDBPath())
	if err != nil {
		return fmt.Errorf("open session manager: %w", err)
	}
	defer mgr.Close()

	// Try to find by name first, then by ID
	s, err := mgr.GetByName(identifier)
	if err != nil {
		return fmt.Errorf("find session: %w", err)
	}
	if s == nil {
		s, err = mgr.Get(identifier)
		if err != nil {
			return fmt.Errorf("find session: %w", err)
		}
		if s == nil {
			return fmt.Errorf("session not found: %s", identifier)
		}
	}

	fmt.Printf("Resuming session: %s (ID: %s)\n", s.Name, s.ID)
	return nil
}

func getSessionDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		usr, err := user.Current()
		if err != nil {
			home = os.TempDir()
		} else {
			home = usr.HomeDir
		}
	}
	// Ensure directory exists
	dir := filepath.Join(home, ".unibrow")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "sessions.db")
}