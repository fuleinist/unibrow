package cmd

import (
	"fmt"
	"strings"

	"github.com/fuleinist/unibrow/internal/memory"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage the shared context buffer",
}

var (
	contextSessionID string
)

func init() {
	contextCmd.AddCommand(contextShowCmd)
	contextCmd.AddCommand(contextAddCmd)
	contextCmd.AddCommand(contextRemoveCmd)
	contextCmd.AddCommand(contextClearCmd)

	contextCmd.PersistentFlags().StringVarP(&contextSessionID, "session", "s", "default", "Session ID")
}

var contextShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current context buffer",
	RunE:  contextShow,
}

var contextAddCmd = &cobra.Command{
	Use:   "add [content]",
	Short: "Add content to the context buffer",
	Args:  cobra.MinimumNArgs(1),
	RunE:  contextAdd,
}

var contextRemoveCmd = &cobra.Command{
	Use:   "remove [id]",
	Short: "Remove an entry from the context buffer",
	Args:  cobra.ExactArgs(1),
	RunE:  contextRemove,
}

var contextClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the entire context buffer",
	RunE:  contextClear,
}

func contextShow(cmd *cobra.Command, args []string) error {
	store, err := memory.NewStore(getMemoryDBPath())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer store.Close()

	entries, err := store.GetRecent(contextSessionID, 50)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}

	if len(entries) == 0 {
		fmt.Printf("Context buffer is empty for session %q\n", contextSessionID)
		return nil
	}

	total, err := store.Count(contextSessionID)
	if err != nil {
		return fmt.Errorf("count context: %w", err)
	}

	fmt.Printf("Context buffer for session %q:\n", contextSessionID)
	fmt.Println("---")
	for _, e := range entries {
		fmt.Printf("[%s] %s\n", e.Agent, e.Content)
	}
	fmt.Println("---")
	fmt.Printf("(Showing %d of %d entries)\n", len(entries), total)
	return nil
}

func contextAdd(cmd *cobra.Command, args []string) error {
	content := strings.Join(args, " ")

	store, err := memory.NewStore(getMemoryDBPath())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer store.Close()

	entry, err := store.Add(contextSessionID, "context", content)
	if err != nil {
		return fmt.Errorf("add context: %w", err)
	}

	fmt.Printf("Added to context buffer (#%d)\n", entry.ID)
	return nil
}

func contextRemove(cmd *cobra.Command, args []string) error {
	// Note: This would need a RemoveByID method in the store
	// For now, use memory clear
	fmt.Println("Use 'unibrow memory clear' to remove memory entries")
	return nil
}

func contextClear(cmd *cobra.Command, args []string) error {
	store, err := memory.NewStore(getMemoryDBPath())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer store.Close()

	if err := store.Clear(contextSessionID); err != nil {
		return fmt.Errorf("clear context: %w", err)
	}

	fmt.Printf("Cleared context buffer for session %q\n", contextSessionID)
	return nil
}