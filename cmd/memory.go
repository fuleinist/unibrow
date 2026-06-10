package cmd

import (
	"fmt"
	"os"
	"os/user"

	"github.com/fuleinist/unibrow/internal/memory"
	"github.com/spf13/cobra"
)

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Manage shared memory across agents",
}

var (
	memorySessionID string
)

func init() {
	memoryCmd.AddCommand(memoryAddCmd)
	memoryCmd.AddCommand(memoryListCmd)
	memoryCmd.AddCommand(memoryClearCmd)

	memoryCmd.PersistentFlags().StringVarP(&memorySessionID, "session", "s", "default", "Session ID")
}

var memoryAddCmd = &cobra.Command{
	Use:   "add [text]",
	Short: "Add a memory entry",
	Args:  cobra.MinimumNArgs(1),
	RunE:  memoryAdd,
}

var memoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List memory entries",
	RunE:  memoryList,
}

var memoryClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear memory entries for a session",
	RunE:  memoryClear,
}

func memoryAdd(cmd *cobra.Command, args []string) error {
	text := args[0]
	store, err := memory.NewStore(getMemoryDBPath())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer store.Close()

	entry, err := store.Add(memorySessionID, "user", text)
	if err != nil {
		return fmt.Errorf("add memory: %w", err)
	}

	fmt.Printf("Added memory entry #%d to session %q\n", entry.ID, memorySessionID)
	return nil
}

func memoryList(cmd *cobra.Command, args []string) error {
	store, err := memory.NewStore(getMemoryDBPath())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer store.Close()

	entries, err := store.List(memorySessionID)
	if err != nil {
		return fmt.Errorf("list memory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Printf("No memory entries for session %q\n", memorySessionID)
		return nil
	}

	fmt.Printf("Memory entries for session %q:\n", memorySessionID)
	for _, e := range entries {
		fmt.Printf("  #%d [%s] %s (used %d times)\n", e.ID, e.Agent, e.Content, e.UsageCount)
	}
	return nil
}

func memoryClear(cmd *cobra.Command, args []string) error {
	store, err := memory.NewStore(getMemoryDBPath())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer store.Close()

	if err := store.Clear(memorySessionID); err != nil {
		return fmt.Errorf("clear memory: %w", err)
	}

	fmt.Printf("Cleared memory for session %q\n", memorySessionID)
	return nil
}

func getMemoryDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback for systems without home dir
		usr, err := user.Current()
		if err != nil {
			home = os.TempDir()
		} else {
			home = usr.HomeDir
		}
	}
	dbDir := home + string(os.PathSeparator) + ".unibrow"
	// Ensure the directory exists before SQLite opens the database
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		// Fallback to temp dir if home is not writable
		dbDir = os.TempDir() + string(os.PathSeparator) + ".unibrow"
		os.MkdirAll(dbDir, 0755)
	}
	return dbDir + string(os.PathSeparator) + "memory.db"
}