package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Tool describes an MCP tool.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ListTools returns all available MCP tools.
func (s *Server) ListTools() []Tool {
	return []Tool{
		{Name: "query_contracts", Description: "Query active contracts by party and template",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"party":{"type":"string"},"template_id":{"type":"string"}},"required":["party"]}`)},
		{Name: "list_parties", Description: "List allocated parties",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`)},
		{Name: "ledger_status", Description: "Get node health and version",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`)},
		{Name: "allocate_party", Description: "Allocate a new party",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"},"display_name":{"type":"string"}},"required":["name","display_name"]}`)},
		{Name: "upload_dar", Description: "Upload a DAR package",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"dar_path":{"type":"string"}},"required":["dar_path"]}`)},
		{Name: "list_packages", Description: "List uploaded packages",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`)},
	}
}

// CallTool invokes a tool by name with the given arguments.
func (s *Server) CallTool(ctx context.Context, name string, args json.RawMessage) (any, error) {
	switch name {
	case "query_contracts":
		return s.toolQueryContracts(ctx, args)
	case "list_parties":
		return s.Ledger.ListParties(ctx)
	case "ledger_status":
		return s.Ledger.Health(ctx)
	case "allocate_party":
		return s.toolAllocateParty(ctx, args)
	case "upload_dar":
		return s.toolUploadDAR(ctx, args)
	case "list_packages":
		return s.Ledger.GetPackages(ctx)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) toolQueryContracts(ctx context.Context, args json.RawMessage) (any, error) {
	var params struct {
		Party      string `json:"party"`
		TemplateID string `json:"template_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	return s.Ledger.ListContracts(ctx, params.Party, params.TemplateID)
}

func (s *Server) toolAllocateParty(ctx context.Context, args json.RawMessage) (any, error) {
	var params struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	return s.Ledger.AllocateParty(ctx, params.Name, params.DisplayName)
}

func (s *Server) toolUploadDAR(ctx context.Context, args json.RawMessage) (any, error) {
	var params struct {
		DARPath string `json:"dar_path"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if err := s.Ledger.UploadDAR(ctx, params.DARPath); err != nil {
		return nil, err
	}
	return map[string]string{"status": "uploaded"}, nil
}
