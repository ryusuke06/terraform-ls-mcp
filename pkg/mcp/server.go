package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/ryu-ch/terraform-ls-mcp/pkg/terraform"
)

// Server represents an MCP server
type Server struct {
	tfClient *terraform.Client
}

// NewServer creates a new MCP server
func NewServer(tfClient *terraform.Client) *Server {
	return &Server{
		tfClient: tfClient,
	}
}

// HandleRequest handles incoming MCP requests
func (s *Server) HandleRequest(ctx context.Context, request Request) Response {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(ctx, request)
	case "tools/list":
		return s.handleListTools(ctx, request)
	case "tools/call":
		return s.handleCallTool(ctx, request)
	default:
		return Response{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &Error{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", request.Method),
			},
		}
	}
}

func (s *Server) handleInitialize(ctx context.Context, request Request) Response {
	var params InitializeParams
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return Response{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &Error{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: &ServerInfo{
			Name:    "terraform-ls-mcp",
			Version: "0.1.0",
		},
	}

	return Response{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

func (s *Server) handleListTools(ctx context.Context, request Request) Response {
	tools := []Tool{
		{
			Name:        "terraform_validate",
			Description: "Validate Terraform configuration files",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspace_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the Terraform workspace directory",
					},
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the specific Terraform file to validate",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content of the Terraform file to validate",
					},
				},
				"required": []string{"workspace_path", "file_path", "content"},
			},
		},
		{
			Name:        "terraform_format",
			Description: "Format Terraform configuration files",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspace_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the Terraform workspace directory",
					},
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the specific Terraform file to format",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content of the Terraform file to format",
					},
				},
				"required": []string{"workspace_path", "file_path", "content"},
			},
		},
		{
			Name:        "terraform_completion",
			Description: "Get completion suggestions for Terraform configuration",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspace_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the Terraform workspace directory",
					},
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the specific Terraform file",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content of the Terraform file",
					},
					"line": map[string]interface{}{
						"type":        "integer",
						"description": "Line number (0-based)",
					},
					"character": map[string]interface{}{
						"type":        "integer",
						"description": "Character position (0-based)",
					},
				},
				"required": []string{"workspace_path", "file_path", "content", "line", "character"},
			},
		},
	}

	return Response{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: ListToolsResult{
			Tools: tools,
		},
	}
}

func (s *Server) handleCallTool(ctx context.Context, request Request) Response {
	var params CallToolParams
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return Response{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &Error{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	switch params.Name {
	case "terraform_validate":
		return s.handleValidateTool(ctx, request.ID, params.Arguments)
	case "terraform_format":
		return s.handleFormatTool(ctx, request.ID, params.Arguments)
	case "terraform_completion":
		return s.handleCompletionTool(ctx, request.ID, params.Arguments)
	default:
		return Response{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &Error{
				Code:    -32602,
				Message: fmt.Sprintf("Unknown tool: %s", params.Name),
			},
		}
	}
}

func (s *Server) handleValidateTool(ctx context.Context, requestID interface{}, args map[string]interface{}) Response {
	workspacePath, ok := args["workspace_path"].(string)
	if !ok {
		return s.errorResponse(requestID, -32602, "workspace_path is required and must be a string")
	}

	filePath, ok := args["file_path"].(string)
	if !ok {
		return s.errorResponse(requestID, -32602, "file_path is required and must be a string")
	}

	content, ok := args["content"].(string)
	if !ok {
		return s.errorResponse(requestID, -32602, "content is required and must be a string")
	}

	// Initialize terraform-ls with workspace
	if err := s.tfClient.Initialize(ctx, workspacePath); err != nil {
		return s.errorResponse(requestID, -32603, fmt.Sprintf("Failed to initialize terraform-ls: %v", err))
	}

	// Create file URI
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return s.errorResponse(requestID, -32603, fmt.Sprintf("Failed to get absolute path: %v", err))
	}
	uri := fmt.Sprintf("file://%s", absPath)

	// Validate document
	result, err := s.tfClient.ValidateDocument(ctx, uri, content)
	if err != nil {
		return s.errorResponse(requestID, -32603, fmt.Sprintf("Failed to validate document: %v", err))
	}

	return Response{
		JSONRPC: "2.0",
		ID:      requestID,
		Result: CallToolResult{
			Content: []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Validation completed for %s. Found %d diagnostic(s).", filePath, len(result.Diagnostics)),
				},
			},
		},
	}
}

func (s *Server) handleFormatTool(ctx context.Context, requestID interface{}, args map[string]interface{}) Response {
	workspacePath, ok := args["workspace_path"].(string)
	if !ok {
		return s.errorResponse(requestID, -32602, "workspace_path is required and must be a string")
	}

	filePath, ok := args["file_path"].(string)
	if !ok {
		return s.errorResponse(requestID, -32602, "file_path is required and must be a string")
	}

	content, ok := args["content"].(string)
	if !ok {
		return s.errorResponse(requestID, -32602, "content is required and must be a string")
	}

	// Initialize terraform-ls with workspace
	if err := s.tfClient.Initialize(ctx, workspacePath); err != nil {
		return s.errorResponse(requestID, -32603, fmt.Sprintf("Failed to initialize terraform-ls: %v", err))
	}

	// Create file URI
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return s.errorResponse(requestID, -32603, fmt.Sprintf("Failed to get absolute path: %v", err))
	}
	uri := fmt.Sprintf("file://%s", absPath)

	// Format document
	result, err := s.tfClient.FormatDocument(ctx, uri, content)
	if err != nil {
		return s.errorResponse(requestID, -32603, fmt.Sprintf("Failed to format document: %v", err))
	}

	return Response{
		JSONRPC: "2.0",
		ID:      requestID,
		Result: CallToolResult{
			Content: []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Formatting completed for %s. Applied %d edit(s).", filePath, len(result.Edits)),
				},
			},
		},
	}
}

func (s *Server) handleCompletionTool(ctx context.Context, requestID interface{}, args map[string]interface{}) Response {
	workspacePath, ok := args["workspace_path"].(string)
	if !ok {
		return s.errorResponse(requestID, -32602, "workspace_path is required and must be a string")
	}

	filePath, ok := args["file_path"].(string)
	if !ok {
		return s.errorResponse(requestID, -32602, "file_path is required and must be a string")
	}

	content, ok := args["content"].(string)
	if !ok {
		return s.errorResponse(requestID, -32602, "content is required and must be a string")
	}

	lineFloat, ok := args["line"].(float64)
	if !ok {
		return s.errorResponse(requestID, -32602, "line is required and must be a number")
	}
	line := int(lineFloat)

	characterFloat, ok := args["character"].(float64)
	if !ok {
		return s.errorResponse(requestID, -32602, "character is required and must be a number")
	}
	character := int(characterFloat)

	// Initialize terraform-ls with workspace
	if err := s.tfClient.Initialize(ctx, workspacePath); err != nil {
		return s.errorResponse(requestID, -32603, fmt.Sprintf("Failed to initialize terraform-ls: %v", err))
	}

	// Create file URI
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return s.errorResponse(requestID, -32603, fmt.Sprintf("Failed to get absolute path: %v", err))
	}
	uri := fmt.Sprintf("file://%s", absPath)

	// Get completion
	result, err := s.tfClient.GetCompletion(ctx, uri, content, line, character)
	if err != nil {
		return s.errorResponse(requestID, -32603, fmt.Sprintf("Failed to get completion: %v", err))
	}

	return Response{
		JSONRPC: "2.0",
		ID:      requestID,
		Result: CallToolResult{
			Content: []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Completion completed for %s at line %d, character %d. Found %d suggestion(s).", filePath, line, character, len(result.Items)),
				},
			},
		},
	}
}

func (s *Server) errorResponse(id interface{}, code int, message string) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
}