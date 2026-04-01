package ledger

import "context"

// ListContracts queries active contracts by party and optional template filter.
func (c *HTTPLedgerClient) ListContracts(ctx context.Context, party, templateID string) ([]Contract, error) {
	reqBody := map[string]string{"party": party}
	if templateID != "" {
		reqBody["template_id"] = templateID
	}
	body, err := encodeJSON(reqBody)
	if err != nil {
		return nil, err
	}
	data, err := c.doRequest(ctx, "POST", "/v2/contracts/search", body)
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]Contract](data)
}
