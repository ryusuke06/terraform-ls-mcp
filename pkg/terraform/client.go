package terraform

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ryu-ch/terraform-ls-mcp/pkg/lsp"
)

// Client represents a terraform-ls client
type Client struct {
	lspClient *lsp.Client
	workspaceRoot string
}

// NewClient creates a new terraform-ls client
func NewClient() (*Client, error) {
	lspClient, err := lsp.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create LSP client: %w", err)
	}
	
	client := &Client{
		lspClient: lspClient,
	}
	
	return client, nil
}

// Close closes the terraform-ls client
func (c *Client) Close() error {
	if c.lspClient != nil {
		return c.lspClient.Close()
	}
	return nil
}

// Initialize initializes the terraform-ls server with workspace
func (c *Client) Initialize(ctx context.Context, workspaceRoot string) error {
	c.workspaceRoot = workspaceRoot
	
	initParams := InitializeParams{
		ProcessID:    nil,
		RootURI:      fmt.Sprintf("file://%s", workspaceRoot),
		WorkspaceFolders: []WorkspaceFolder{
			{
				URI:  fmt.Sprintf("file://%s", workspaceRoot),
				Name: filepath.Base(workspaceRoot),
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
	
	resp, err := c.lspClient.SendRequest(ctx, "initialize", initParams)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	
	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}
	
	// Send initialized notification
	if err := c.lspClient.SendNotification("initialized", struct{}{}); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}
	
	return nil
}

// ValidateDocument validates a Terraform document
func (c *Client) ValidateDocument(ctx context.Context, uri, content string) (*ValidationResult, error) {
	// Open document
	if err := c.openDocument(ctx, uri, content); err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}
	
	// Get diagnostics (validation results)
	resp, err := c.lspClient.SendRequest(ctx, "textDocument/diagnostic", DiagnosticParams{
		TextDocument: TextDocumentIdentifier{
			URI: uri,
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get diagnostics: %w", err)
	}
	
	if resp.Error != nil {
		return nil, fmt.Errorf("diagnostic error: %s", resp.Error.Message)
	}
	
	var diagnostics []Diagnostic
	if resp.Result != nil {
		// Parse diagnostics from response
		// This is a simplified version - actual implementation would need proper parsing
	}
	
	return &ValidationResult{
		URI:         uri,
		Diagnostics: diagnostics,
	}, nil
}

// FormatDocument formats a Terraform document
func (c *Client) FormatDocument(ctx context.Context, uri, content string) (*FormatResult, error) {
	// Open document
	if err := c.openDocument(ctx, uri, content); err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}
	
	resp, err := c.lspClient.SendRequest(ctx, "textDocument/formatting", DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{
			URI: uri,
		},
		Options: FormattingOptions{
			TabSize:      2,
			InsertSpaces: true,
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to format document: %w", err)
	}
	
	if resp.Error != nil {
		return nil, fmt.Errorf("format error: %s", resp.Error.Message)
	}
	
	// Parse text edits from response
	var textEdits []TextEdit
	// This is a simplified version - actual implementation would need proper parsing
	
	return &FormatResult{
		URI:   uri,
		Edits: textEdits,
	}, nil
}

// GetCompletion gets completion suggestions for a position in document
func (c *Client) GetCompletion(ctx context.Context, uri, content string, line, character int) (*CompletionResult, error) {
	// Open document
	if err := c.openDocument(ctx, uri, content); err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}
	
	resp, err := c.lspClient.SendRequest(ctx, "textDocument/completion", CompletionParams{
		TextDocument: TextDocumentIdentifier{
			URI: uri,
		},
		Position: Position{
			Line:      line,
			Character: character,
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get completion: %w", err)
	}
	
	if resp.Error != nil {
		return nil, fmt.Errorf("completion error: %s", resp.Error.Message)
	}
	
	// Parse completion items from response
	var completionItems []CompletionItem
	// This is a simplified version - actual implementation would need proper parsing
	
	return &CompletionResult{
		URI:   uri,
		Items: completionItems,
	}, nil
}

func (c *Client) openDocument(ctx context.Context, uri, content string) error {
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        uri,
			LanguageID: "terraform",
			Version:    1,
			Text:       content,
		},
	}
	
	return c.lspClient.SendNotification("textDocument/didOpen", params)
}