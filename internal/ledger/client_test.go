package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/displacetech/cantool/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *HTTPLedgerClient) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := NewHTTPLedgerClient(srv.URL)
	return srv, client
}

// --- Health ---

func TestHealth_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		json.NewEncoder(w).Encode(HealthResponse{Status: "SERVING", Version: "3.4.11", Uptime: "1h"})
	})
	h, err := client.Health(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "SERVING", h.Status)
	assert.Equal(t, "3.4.11", h.Version)
}

func TestHealth_Unauthorized(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
	})
	_, err := client.Health(context.Background())
	var ce *output.CantoolError
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "CT6002", ce.Code)
}

func TestHealth_ServerError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	})
	_, err := client.Health(context.Background())
	var ce *output.CantoolError
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "CT6004", ce.Code)
}

func TestHealth_InvalidJSON(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("not json"))
	})
	_, err := client.Health(context.Background())
	var ce *output.CantoolError
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "CT6004", ce.Code)
}

func TestHealth_ConnectionRefused(t *testing.T) {
	client := NewHTTPLedgerClient("http://localhost:1") // nothing on port 1
	_, err := client.Health(context.Background())
	require.Error(t, err)
}

// --- Parties ---

func TestListParties_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/parties", r.URL.Path)
		json.NewEncoder(w).Encode([]Party{
			{ID: "alice", DisplayName: "Alice Corp", IsLocal: true},
		})
	})
	parties, err := client.ListParties(context.Background())
	require.NoError(t, err)
	require.Len(t, parties, 1)
	assert.Equal(t, "Alice Corp", parties[0].DisplayName)
}

func TestAllocateParty_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/parties/allocate", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		json.Unmarshal(body, &req)
		assert.Equal(t, "Alice", req["party"])
		assert.Equal(t, "Alice Corp", req["display_name"])

		json.NewEncoder(w).Encode(Party{ID: "alice", DisplayName: "Alice Corp"})
	})
	p, err := client.AllocateParty(context.Background(), "Alice", "Alice Corp")
	require.NoError(t, err)
	assert.Equal(t, "alice", p.ID)
}

func TestAllocateParty_WithToken(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer my-jwt", r.Header.Get("Authorization"))
		json.NewEncoder(w).Encode(Party{ID: "alice"})
	})
	client.token = "my-jwt"
	_, err := client.AllocateParty(context.Background(), "Alice", "Alice")
	require.NoError(t, err)
}

// --- Contracts ---

func TestListContracts_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/contracts/search", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		json.Unmarshal(body, &req)
		assert.Equal(t, "Alice", req["party"])
		assert.Equal(t, "Token", req["template_id"])

		json.NewEncoder(w).Encode([]Contract{
			{ContractID: "c1", TemplateID: "Token"},
		})
	})
	contracts, err := client.ListContracts(context.Background(), "Alice", "Token")
	require.NoError(t, err)
	require.Len(t, contracts, 1)
}

func TestListContracts_NoTemplateFilter(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		json.Unmarshal(body, &req)
		_, hasTemplate := req["template_id"]
		assert.False(t, hasTemplate)
		json.NewEncoder(w).Encode([]Contract{})
	})
	contracts, err := client.ListContracts(context.Background(), "Alice", "")
	require.NoError(t, err)
	assert.Empty(t, contracts)
}

// --- Packages ---

func TestGetPackages_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/packages", r.URL.Path)
		json.NewEncoder(w).Encode([]Package{{PackageID: "pkg1", Size: 1024}})
	})
	pkgs, err := client.GetPackages(context.Background())
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	assert.Equal(t, "pkg1", pkgs[0].PackageID)
}

func TestUploadDAR_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/packages", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
		w.WriteHeader(http.StatusOK)
	})

	dir := t.TempDir()
	darPath := filepath.Join(dir, "test.dar")
	require.NoError(t, os.WriteFile(darPath, []byte("fake dar"), 0644))

	err := client.UploadDAR(context.Background(), darPath)
	require.NoError(t, err)
}

func TestUploadDAR_FileNotFound(t *testing.T) {
	client := NewHTTPLedgerClient("http://localhost")
	err := client.UploadDAR(context.Background(), "/nonexistent/test.dar")
	var ce *output.CantoolError
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "CT6005", ce.Code)
}

func TestUploadDAR_ServerError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad dar"))
	})

	dir := t.TempDir()
	darPath := filepath.Join(dir, "test.dar")
	require.NoError(t, os.WriteFile(darPath, []byte("fake"), 0644))

	err := client.UploadDAR(context.Background(), darPath)
	var ce *output.CantoolError
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "CT6005", ce.Code)
}

// --- Client options ---

func TestWithToken(t *testing.T) {
	c := NewHTTPLedgerClient("http://localhost", WithToken("jwt123"))
	assert.Equal(t, "jwt123", c.token)
}

func TestWithTimeout(t *testing.T) {
	c := NewHTTPLedgerClient("http://localhost", WithTimeout(30*time.Second))
	assert.Equal(t, 30*time.Second, c.httpClient.Timeout)
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := NewHTTPLedgerClient("http://localhost", WithHTTPClient(custom))
	assert.Equal(t, custom, c.httpClient)
}

func TestUploadDAR_WithToken(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	})
	client.token = "tok"

	dir := t.TempDir()
	darPath := filepath.Join(dir, "test.dar")
	require.NoError(t, os.WriteFile(darPath, []byte("data"), 0644))

	err := client.UploadDAR(context.Background(), darPath)
	require.NoError(t, err)
}

func TestListParties_Error(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	})
	_, err := client.ListParties(context.Background())
	require.Error(t, err)
}

func TestAllocateParty_Error(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	})
	_, err := client.AllocateParty(context.Background(), "Alice", "Alice")
	require.Error(t, err)
}

func TestListContracts_Error(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	})
	_, err := client.ListContracts(context.Background(), "Alice", "")
	require.Error(t, err)
}

func TestGetPackages_Error(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	})
	_, err := client.GetPackages(context.Background())
	require.Error(t, err)
}

func TestListParties_InvalidJSON(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("not json"))
	})
	_, err := client.ListParties(context.Background())
	require.Error(t, err)
}

func TestContextTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
	}))
	t.Cleanup(srv.Close)

	client := NewHTTPLedgerClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Health(ctx)
	require.Error(t, err)
}
