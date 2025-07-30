package terraform

import (
	"encoding/json"
	"testing"
)

func TestInitializeParams_Marshal(t *testing.T) {
	params := InitializeParams{
		ProcessID: nil,
		RootURI:   "file:///test/workspace",
		WorkspaceFolders: []WorkspaceFolder{
			{
				URI:  "file:///test/workspace",
				Name: "workspace",
			},
		},
		Capabilities: ClientCapabilities{
			TextDocument: &TextDocumentClientCapabilities{
				Completion: &CompletionClientCapabilities{
					CompletionItem: &CompletionItemClientCapabilities{
						SnippetSupport: true,
					},
				},
				Hover: &HoverClientCapabilities{
					ContentFormat: []string{"markdown", "plaintext"},
				},
			},
		},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Errorf("Failed to marshal InitializeParams: %v", err)
	}

	// Verify JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
	}

	// Check required fields
	if result["rootUri"] != "file:///test/workspace" {
		t.Errorf("Expected rootUri 'file:///test/workspace', got: %v", result["rootUri"])
	}

	workspaceFolders, ok := result["workspaceFolders"].([]interface{})
	if !ok {
		t.Error("Expected workspaceFolders to be array")
	} else if len(workspaceFolders) != 1 {
		t.Errorf("Expected 1 workspace folder, got: %d", len(workspaceFolders))
	}
}

func TestDiagnostic_Structure(t *testing.T) {
	diagnostic := Diagnostic{
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: 0, Character: 10},
		},
		Severity: 1, // Error
		Source:   "terraform-ls",
		Message:  "Test error message",
	}

	data, err := json.Marshal(diagnostic)
	if err != nil {
		t.Errorf("Failed to marshal Diagnostic: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
	}

	if result["message"] != "Test error message" {
		t.Errorf("Expected message 'Test error message', got: %v", result["message"])
	}

	if result["source"] != "terraform-ls" {
		t.Errorf("Expected source 'terraform-ls', got: %v", result["source"])
	}
}

func TestCompletionItem_Structure(t *testing.T) {
	item := CompletionItem{
		Label:         "resource",
		Kind:          14, // Keyword
		Detail:        "Terraform resource block",
		Documentation: "Define a resource in Terraform",
		InsertText:    "resource \"${1:type}\" \"${2:name}\" {\n\t$0\n}",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Errorf("Failed to marshal CompletionItem: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
	}

	if result["label"] != "resource" {
		t.Errorf("Expected label 'resource', got: %v", result["label"])
	}

	if result["detail"] != "Terraform resource block" {
		t.Errorf("Expected detail 'Terraform resource block', got: %v", result["detail"])
	}
}

func TestValidationResult_Structure(t *testing.T) {
	result := ValidationResult{
		URI: "file:///test/main.tf",
		Diagnostics: []Diagnostic{
			{
				Range: Range{
					Start: Position{Line: 1, Character: 0},
					End:   Position{Line: 1, Character: 10},
				},
				Severity: 1,
				Source:   "terraform-ls",
				Message:  "Invalid syntax",
			},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal ValidationResult: %v", err)
	}

	var unmarshaled ValidationResult
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal ValidationResult: %v", err)
	}

	if unmarshaled.URI != "file:///test/main.tf" {
		t.Errorf("Expected URI 'file:///test/main.tf', got: %s", unmarshaled.URI)
	}

	if len(unmarshaled.Diagnostics) != 1 {
		t.Errorf("Expected 1 diagnostic, got: %d", len(unmarshaled.Diagnostics))
	}
}