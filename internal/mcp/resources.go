package mcp

import (
	"context"
	"fmt"
)

// Resource describes an MCP resource.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// ListResources returns all available MCP resources.
func (s *Server) ListResources() []Resource {
	return []Resource{
		{URI: "canton://contracts", Name: "Active Contracts", Description: "Active contracts for the current environment", MimeType: "application/json"},
		{URI: "canton://parties", Name: "Party Topology", Description: "Allocated parties", MimeType: "application/json"},
		{URI: "canton://packages", Name: "Packages", Description: "Uploaded DAR packages", MimeType: "application/json"},
	}
}

// ReadResource reads a resource by URI.
func (s *Server) ReadResource(ctx context.Context, uri string) (any, error) {
	switch uri {
	case "canton://contracts":
		return s.Ledger.ListContracts(ctx, "", "")
	case "canton://parties":
		return s.Ledger.ListParties(ctx)
	case "canton://packages":
		return s.Ledger.GetPackages(ctx)
	default:
		return nil, fmt.Errorf("unknown resource: %s", uri)
	}
}
