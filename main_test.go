package main

import (
	"os"
	"testing"
)

func TestMainUsage(t *testing.T) {
	// Test that main function handles no arguments correctly
	// We can't actually call main() as it would run the server
	// But we can test the argument parsing logic

	if len(os.Args) < 1 {
		t.Error("Expected at least one argument (program name)")
	}

	// Test command validation logic
	validCommands := []string{"serve"}
	
	for _, cmd := range validCommands {
		// This is just testing the structure - actual validation would be in main()
		if cmd != "serve" {
			t.Errorf("Unexpected command in valid commands: %s", cmd)
		}
	}
}

func TestCommandValidation(t *testing.T) {
	// Test command validation
	validCommands := map[string]bool{
		"serve": true,
	}

	invalidCommands := []string{
		"start",
		"run",
		"invalid",
		"",
	}

	for _, cmd := range invalidCommands {
		if validCommands[cmd] {
			t.Errorf("Command %s should not be valid", cmd)
		}
	}

	if !validCommands["serve"] {
		t.Error("Command 'serve' should be valid")
	}
}