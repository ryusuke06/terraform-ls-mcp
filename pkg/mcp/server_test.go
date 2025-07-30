package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ryu-ch/terraform-ls-mcp/pkg/terraform"
)

func TestServer_HandleInitialize(t *testing.T) {
	// Create a mock terraform client
	tfClient := &terraform.Client{}
	server := NewServer(tfClient)

	// Create initialize request
	initParams := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities: ClientCapabilities{
			Tools: &ToolsCapability{},
		},
		ClientInfo: &ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	paramsBytes, err := json.Marshal(initParams)
	if err != nil {
		t.Fatalf("Failed to marshal init params: %v", err)
	}

	request := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  paramsBytes,
	}

	ctx := context.Background()
	response := server.HandleRequest(ctx, request)

	// Verify response
	if response.Error != nil {
		t.Errorf("Expected no error, got: %v", response.Error)
	}

	if response.ID != 1 {
		t.Errorf("Expected ID 1, got: %v", response.ID)
	}

	// Check if result is InitializeResult
	result, ok := response.Result.(InitializeResult)
	if !ok {
		t.Errorf("Expected InitializeResult, got: %T", response.Result)
	}

	if result.ProtocolVersion != "2024-11-05" {
		t.Errorf("Expected protocol version '2024-11-05', got: %s", result.ProtocolVersion)
	}

	if result.ServerInfo == nil {
		t.Error("Expected ServerInfo to be set")
	} else {
		if result.ServerInfo.Name != "terraform-ls-mcp" {
			t.Errorf("Expected server name 'terraform-ls-mcp', got: %s", result.ServerInfo.Name)
		}
	}
}

func TestServer_HandleListTools(t *testing.T) {
	tfClient := &terraform.Client{}
	server := NewServer(tfClient)

	request := Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	ctx := context.Background()
	response := server.HandleRequest(ctx, request)

	// Verify response
	if response.Error != nil {
		t.Errorf("Expected no error, got: %v", response.Error)
	}

	if response.ID != 2 {
		t.Errorf("Expected ID 2, got: %v", response.ID)
	}

	// Check if result is ListToolsResult
	result, ok := response.Result.(ListToolsResult)
	if !ok {
		t.Errorf("Expected ListToolsResult, got: %T", response.Result)
	}

	expectedTools := []string{"terraform_validate", "terraform_format", "terraform_completion"}
	if len(result.Tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(result.Tools))
	}

	for i, expectedTool := range expectedTools {
		if i >= len(result.Tools) {
			t.Errorf("Missing tool: %s", expectedTool)
			continue
		}
		if result.Tools[i].Name != expectedTool {
			t.Errorf("Expected tool name %s, got %s", expectedTool, result.Tools[i].Name)
		}
	}
}

func TestServer_HandleUnknownMethod(t *testing.T) {
	tfClient := &terraform.Client{}
	server := NewServer(tfClient)

	request := Request{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "unknown_method",
	}

	ctx := context.Background()
	response := server.HandleRequest(ctx, request)

	// Verify error response
	if response.Error == nil {
		t.Error("Expected error for unknown method")
	}

	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got: %d", response.Error.Code)
	}

	if response.ID != 3 {
		t.Errorf("Expected ID 3, got: %v", response.ID)
	}
}

func TestServer_HandleCallToolInvalidParams(t *testing.T) {
	tfClient := &terraform.Client{}
	server := NewServer(tfClient)

	// Invalid JSON params
	invalidParams := json.RawMessage(`{"invalid": "json"`)

	request := Request{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  invalidParams,
	}

	ctx := context.Background()
	response := server.HandleRequest(ctx, request)

	// Verify error response
	if response.Error == nil {
		t.Error("Expected error for invalid params")
	}

	if response.Error.Code != -32602 {
		t.Errorf("Expected error code -32602, got: %d", response.Error.Code)
	}
}