package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/displacetech/cantool/internal/config"
	"github.com/displacetech/cantool/internal/ledger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLedger implements ledger.LedgerClient for testing.
type mockLedger struct {
	parties   []ledger.Party
	contracts []ledger.Contract
	packages  []ledger.Package
	health    *ledger.HealthResponse
}

func (m *mockLedger) Health(_ context.Context) (*ledger.HealthResponse, error) {
	return m.health, nil
}
func (m *mockLedger) ListParties(_ context.Context) ([]ledger.Party, error) {
	return m.parties, nil
}
func (m *mockLedger) AllocateParty(_ context.Context, name, display string) (*ledger.Party, error) {
	return &ledger.Party{ID: name, DisplayName: display}, nil
}
func (m *mockLedger) ListContracts(_ context.Context, _, _ string) ([]ledger.Contract, error) {
	return m.contracts, nil
}
func (m *mockLedger) UploadDAR(_ context.Context, _ string) error {
	return nil
}
func (m *mockLedger) GetPackages(_ context.Context) ([]ledger.Package, error) {
	return m.packages, nil
}

func newTestServer() *Server {
	return NewServer(&mockLedger{
		health:  &ledger.HealthResponse{Status: "SERVING", Version: "3.4.11"},
		parties: []ledger.Party{{ID: "alice", DisplayName: "Alice"}},
		packages: []ledger.Package{{PackageID: "pkg1"}},
	}, &config.Config{})
}

func sendRequest(t *testing.T, s *Server, method string, params any) jsonRPCResponse {
	t.Helper()
	var paramsJSON json.RawMessage
	if params != nil {
		b, _ := json.Marshal(params)
		paramsJSON = b
	}
	req := jsonRPCRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: paramsJSON}
	reqBytes, _ := json.Marshal(req)

	in := strings.NewReader(string(reqBytes) + "\n")
	var out bytes.Buffer

	s.serve(context.Background(), in, &out)

	var resp jsonRPCResponse
	require.NoError(t, json.Unmarshal(out.Bytes(), &resp))
	return resp
}

func TestServer_Initialize(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "initialize", nil)
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	assert.Equal(t, "2024-11-05", result["protocolVersion"])
}

func TestServer_ToolsList(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "tools/list", nil)
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	tools := result["tools"].([]any)
	assert.Len(t, tools, 6)
}

func TestServer_ToolCall_ListParties(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "tools/call", map[string]any{"name": "list_parties", "arguments": json.RawMessage(`{}`)})
	assert.Nil(t, resp.Error)
}

func TestServer_ToolCall_LedgerStatus(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "tools/call", map[string]any{"name": "ledger_status", "arguments": json.RawMessage(`{}`)})
	assert.Nil(t, resp.Error)
}

func TestServer_ToolCall_AllocateParty(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "tools/call", map[string]any{
		"name":      "allocate_party",
		"arguments": json.RawMessage(`{"name":"Bob","display_name":"Bob Corp"}`),
	})
	assert.Nil(t, resp.Error)
}

func TestServer_ToolCall_QueryContracts(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "tools/call", map[string]any{
		"name":      "query_contracts",
		"arguments": json.RawMessage(`{"party":"Alice"}`),
	})
	assert.Nil(t, resp.Error)
}

func TestServer_ToolCall_ListPackages(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "tools/call", map[string]any{
		"name":      "list_packages",
		"arguments": json.RawMessage(`{}`),
	})
	assert.Nil(t, resp.Error)
}

func TestServer_ToolCall_Unknown(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "tools/call", map[string]any{
		"name":      "nonexistent",
		"arguments": json.RawMessage(`{}`),
	})
	assert.NotNil(t, resp.Error)
	assert.Contains(t, resp.Error.Message, "unknown tool")
}

func TestServer_ResourcesList(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "resources/list", nil)
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	resources := result["resources"].([]any)
	assert.Len(t, resources, 3)
}

func TestServer_ResourceRead_Parties(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "resources/read", map[string]string{"uri": "canton://parties"})
	assert.Nil(t, resp.Error)
}

func TestServer_ResourceRead_Unknown(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "resources/read", map[string]string{"uri": "canton://unknown"})
	assert.NotNil(t, resp.Error)
}

func TestServer_UnknownMethod(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "nonexistent", nil)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, -32601, resp.Error.Code)
}

func TestServer_ParseError(t *testing.T) {
	s := newTestServer()
	in := strings.NewReader("not json\n")
	var out bytes.Buffer
	s.serve(context.Background(), in, &out)

	var resp jsonRPCResponse
	require.NoError(t, json.Unmarshal(out.Bytes(), &resp))
	assert.Equal(t, -32700, resp.Error.Code)
}
