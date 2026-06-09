package ui

import (
	"fmt"
	"os"
)

// Colors for terminal output.
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[90m"
	ColorWhite  = "\033[97m"
)

// Header renders the unibrow header bar.
func Header(version, sessionName string) string {
	return fmt.Sprintf("%s┌─────────────────────────────────────────────────────────%s│%s [%s🤖%s][%s🧠%s][%s🌟%s]  unibrow %s%s          session: %s%s  │%s%s┌─────────────────────────────────────────────────────────%s│",
		ColorBlue, ColorReset, ColorCyan, ColorReset,
		ColorCyan, ColorReset, ColorCyan, ColorReset,
		ColorCyan, version, ColorReset, ColorYellow, sessionName, ColorReset,
		ColorBlue, ColorReset)
}

// StatusBar renders a status bar for agents.
func StatusBar(agentStatus map[string]string) string {
	var s string
	for name, status := range agentStatus {
		color := ColorGreen
		if status == "unavailable" {
			color = ColorGray
		} else if status == "busy" {
			color = ColorYellow
		}
		s += fmt.Sprintf("%s[%s]%s %s ", color, status, ColorReset, name)
	}
	return s
}

// PrintAgentOutput prints agent output with formatting.
func PrintAgentOutput(agentName, output string) {
	fmt.Printf("%s[%s]%s %s\n", ColorCyan, agentName, ColorReset, output)
}

// PrintError prints an error message.
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s✗ Error: %s%s\n", ColorRed, fmt.Sprintf(format, args...), ColorReset)
}

// PrintSuccess prints a success message.
func PrintSuccess(format string, args ...interface{}) {
	fmt.Printf("%s✓ %s%s\n", ColorGreen, fmt.Sprintf(format, args...), ColorReset)
}

// PrintInfo prints an info message.
func PrintInfo(format string, args ...interface{}) {
	fmt.Printf("%sℹ %s%s\n", ColorBlue, fmt.Sprintf(format, args...), ColorReset)
}

// PrintWarning prints a warning message.
func PrintWarning(format string, args ...interface{}) {
	fmt.Printf("%s⚠ %s%s\n", ColorYellow, fmt.Sprintf(format, args...), ColorReset)
}

// ASCIILogo returns the ASCII art logo.
func ASCIILogo() string {
	return fmt.Sprintf(`%s
   _   _      _ _         
  | \ | |    | | |        
  |  \| | ___| | | ___    
  | .   |/ _ \ | |/ _ \   
  | |\  |  __/ | | (_) |  
  |_| \_|\___|_|_|\___/   
      _   _ _____ ____    
     | \ | | ____|  _ \   
     |  \| |  _| | |_) |  
     | |\  | |___|  _ <   
     |_| \_|____|_| \_\  
%s%s
`, ColorCyan, ColorReset, "     Unified AI Agent CLI Hub")
}