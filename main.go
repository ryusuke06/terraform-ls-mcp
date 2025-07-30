package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ryu-ch/terraform-ls-mcp/pkg/mcp"
	"github.com/ryu-ch/terraform-ls-mcp/pkg/terraform"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  serve  Start the MCP server\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		serve()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func serve() {
	ctx := context.Background()
	
	// Initialize terraform-ls client
	tfClient, err := terraform.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize terraform-ls client: %v", err)
	}
	defer tfClient.Close()

	// Initialize MCP server
	server := mcp.NewServer(tfClient)

	// Handle stdin/stdout communication
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var request mcp.Request
		if err := decoder.Decode(&request); err != nil {
			log.Printf("Failed to decode request: %v", err)
			break
		}

		response := server.HandleRequest(ctx, request)
		
		if err := encoder.Encode(response); err != nil {
			log.Printf("Failed to encode response: %v", err)
			break
		}
	}
}