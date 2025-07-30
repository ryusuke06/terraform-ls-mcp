package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// LSP Request/Response structures
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Client represents an LSP client
type Client struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	
	reqID     int64
	responses map[interface{}]chan Response
	mu        sync.RWMutex
	
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient creates a new LSP client for terraform-ls
func NewClient() (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	cmd := exec.CommandContext(ctx, "terraform-ls", "serve")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start terraform-ls: %w", err)
	}
	
	client := &Client{
		cmd:       cmd,
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
		responses: make(map[interface{}]chan Response),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	go client.readResponses()
	
	return client, nil
}

// Close closes the LSP client
func (c *Client) Close() error {
	c.cancel()
	
	if c.stdin != nil {
		c.stdin.Close()
	}
	
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	
	return nil
}

// SendRequest sends a request to the LSP server and returns the response
func (c *Client) SendRequest(ctx context.Context, method string, params interface{}) (*Response, error) {
	id := atomic.AddInt64(&c.reqID, 1)
	
	request := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	
	respChan := make(chan Response, 1)
	c.mu.Lock()
	c.responses[id] = respChan
	c.mu.Unlock()
	
	defer func() {
		c.mu.Lock()
		delete(c.responses, id)
		c.mu.Unlock()
	}()
	
	if err := c.writeMessage(request); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}
	
	select {
	case response := <-respChan:
		return &response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}
}

// SendNotification sends a notification to the LSP server
func (c *Client) SendNotification(method string, params interface{}) error {
	notification := Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	
	return c.writeMessage(notification)
}

func (c *Client) writeMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	
	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	
	if _, err := c.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	
	return nil
}

func (c *Client) readResponses() {
	reader := bufio.NewReader(c.stdout)
	
	for {
		// Read the Content-Length header
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Content-Length:") {
			continue
		}
		
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		
		length, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			continue
		}
		
		// Skip the empty line after headers
		_, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		
		// Read the JSON content
		jsonData := make([]byte, length)
		_, err = io.ReadFull(reader, jsonData)
		if err != nil {
			continue
		}
		
		var response Response
		if err := json.Unmarshal(jsonData, &response); err != nil {
			continue
		}
		
		c.mu.RLock()
		respChan, exists := c.responses[response.ID]
		c.mu.RUnlock()
		
		if exists {
			select {
			case respChan <- response:
			default:
			}
		}
	}
}