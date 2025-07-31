package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ryu-ch/terraform-ls-mcp/pkg/mcp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <test_name>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Tests: initialize, list_tools, validate\n")
		os.Exit(1)
	}

	testName := os.Args[1]
	
	// Start MCP server
	cmd := exec.Command("./terraform-ls-mcp", "serve")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Failed to create stdin pipe: %v\n", err)
		os.Exit(1)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Failed to create stdout pipe: %v\n", err)
		os.Exit(1)
	}
	
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start MCP server: %v\n", err)
		os.Exit(1)
	}
	defer cmd.Process.Kill()

	switch testName {
	case "initialize":
		testInitialize(stdin, stdout)
	case "list_tools":
		testInitialize(stdin, stdout)
		testListTools(stdin, stdout)
	case "validate":
		testInitialize(stdin, stdout)
		testValidate(stdin, stdout)
	default:
		fmt.Printf("Unknown test: %s\n", testName)
	}
}

func testInitialize(stdin io.WriteCloser, stdout io.ReadCloser) {
	fmt.Println("Testing initialize...")
	
	request := mcp.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion": "2024-11-05", "capabilities": {"tools": {}}}`),
	}
	
	sendRequest(stdin, request)
	response := readResponse(stdout)
	
	fmt.Printf("Initialize response: %+v\n", response)
}

func testListTools(stdin io.WriteCloser, stdout io.ReadCloser) {
	fmt.Println("Testing list tools...")
	
	request := mcp.Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	
	sendRequest(stdin, request)
	response := readResponse(stdout)
	
	fmt.Printf("List tools response: %+v\n", response)
}

func testValidate(stdin io.WriteCloser, stdout io.ReadCloser) {
	fmt.Println("Testing validate...")
	
	// Get current working directory and construct relative paths
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		return
	}
	
	workspacePath := filepath.Join(cwd, "test-workspace")
	filePath := filepath.Join(workspacePath, "main.tf")
	content := `resource "aws_instance" "test" {
  ami = "ami-12345"
  instance_type = "t3.micro"
  
  tags = {
    Name = "test-instance"
  }
}`

	params := map[string]interface{}{
		"name": "terraform_validate",
		"arguments": map[string]interface{}{
			"workspace_path": workspacePath,
			"file_path":      filePath,
			"content":        content,
		},
	}
	
	paramsBytes, _ := json.Marshal(params)
	
	request := mcp.Request{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  json.RawMessage(paramsBytes),
	}
	
	sendRequest(stdin, request)
	response := readResponse(stdout)
	
	fmt.Printf("Validate response: %+v\n", response)
}

func sendRequest(stdin io.WriteCloser, request mcp.Request) {
	data, _ := json.Marshal(request)
	stdin.Write(data)
	stdin.Write([]byte("\n"))
}

func readResponse(stdout io.ReadCloser) map[string]interface{} {
	scanner := bufio.NewScanner(stdout)
	if scanner.Scan() {
		var response map[string]interface{}
		json.Unmarshal(scanner.Bytes(), &response)
		return response
	}
	return nil
}