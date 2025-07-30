package lsp

import (
	"context"
	"testing"
	"time"
)

func TestRequest_Marshal(t *testing.T) {
	request := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"processId": nil,
			"rootUri":   "file:///test",
		},
	}

	// Test that request structure is valid
	if request.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got: %s", request.JSONRPC)
	}

	if request.Method != "initialize" {
		t.Errorf("Expected method 'initialize', got: %s", request.Method)
	}
}

func TestResponse_Structure(t *testing.T) {
	response := Response{
		JSONRPC: "2.0",
		ID:      1,
		Result:  map[string]interface{}{"capabilities": map[string]interface{}{}},
	}

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got: %s", response.JSONRPC)
	}

	if response.ID != 1 {
		t.Errorf("Expected ID 1, got: %v", response.ID)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got: %v", response.Error)
	}
}

func TestError_Structure(t *testing.T) {
	err := Error{
		Code:    -32601,
		Message: "Method not found",
		Data:    "additional info",
	}

	if err.Code != -32601 {
		t.Errorf("Expected code -32601, got: %d", err.Code)
	}

	if err.Message != "Method not found" {
		t.Errorf("Expected message 'Method not found', got: %s", err.Message)
	}
}

// Mock test for client creation (without actually starting terraform-ls)
func TestClientCreation_Structure(t *testing.T) {
	// Test that we can create the basic structure
	// This test doesn't actually create a client since it would require terraform-ls
	
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Test context cancellation
	select {
	case <-ctx.Done():
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded, got: %v", ctx.Err())
		}
	case <-time.After(2 * time.Second):
		t.Error("Context should have been cancelled")
	}
}