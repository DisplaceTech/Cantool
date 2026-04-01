package ledger

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/displacetech/cantool/internal/output"
)

// Health returns the node health status.
func (c *HTTPLedgerClient) Health(ctx context.Context) (*HealthResponse, error) {
	data, err := c.doRequest(ctx, "GET", "/health", nil)
	if err != nil {
		return nil, err
	}
	return decodeJSONPtr[HealthResponse](data)
}

// GetPackages returns all uploaded packages.
func (c *HTTPLedgerClient) GetPackages(ctx context.Context) ([]Package, error) {
	data, err := c.doRequest(ctx, "GET", "/v2/packages", nil)
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]Package](data)
}

// UploadDAR uploads a DAR file to the ledger.
func (c *HTTPLedgerClient) UploadDAR(ctx context.Context, darPath string) error {
	f, err := os.Open(darPath)
	if err != nil {
		return output.Wrapf(err, "CT6005",
			fmt.Sprintf("Check that %s exists", darPath),
			"cannot open DAR file")
	}
	defer f.Close()

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		part, err := mw.CreateFormFile("dar", darPath)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, f); err != nil {
			pw.CloseWithError(err)
			return
		}
		mw.Close()
		pw.Close()
	}()

	url := c.baseURL + "/v2/packages"
	req, err := http.NewRequestWithContext(ctx, "POST", url, pr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return output.Wrapf(err, "CT6005",
			"Check that the node is running",
			"DAR upload failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return output.Errorf("CT6005",
			"Check the DAR file is valid",
			"DAR upload failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
